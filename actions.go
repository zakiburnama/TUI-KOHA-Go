package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

var (
	sh32          = syscall.NewLazyDLL("shell32.dll")
	pShellExecute = sh32.NewProc("ShellExecuteW")
	pSendInput    = u32.NewProc("SendInput")
	pLockWorkstn  = u32.NewProc("LockWorkStation")
	pSetSuspend   = syscall.NewLazyDLL("powrprof.dll").NewProc("SetSuspendState")
)

func runItem(it Item) error {
	switch it.Type {
	case "exe":
		verb := ""
		if it.Admin {
			verb = "runas"
		}
		return shellExec(verb, os.ExpandEnv(expandWinEnv(it.Target)), it.Args)
	case "url":
		return shellExec("", it.Target, "")
	case "keys", "text":
		detachConsole()
		d := it.DelayMs
		if d == 0 {
			d = 250
		}
		time.Sleep(time.Duration(d) * time.Millisecond)
		if it.Type == "keys" {
			return sendKeys(it.Target)
		}
		return sendText(it.Target)
	case "builtin":
		switch it.Target {
		case "lock":
			pLockWorkstn.Call()
		case "sleep":
			pSetSuspend.Call(0, 0, 0)
		default:
			return fmt.Errorf("builtin tidak dikenal: %s", it.Target)
		}
		return nil
	}
	return fmt.Errorf("type tidak dikenal: %q", it.Type)
}

// expandWinEnv mengubah %VAR% (gaya Windows) jadi ${VAR} agar os.ExpandEnv bisa memprosesnya.
func expandWinEnv(s string) string {
	parts := strings.Split(s, "%")
	if len(parts)%2 == 0 {
		return s
	}
	for i := 1; i < len(parts); i += 2 {
		parts[i] = "${" + parts[i] + "}"
	}
	return strings.Join(parts, "")
}

func shellExec(verb, file, args string) error {
	p := func(s string) uintptr {
		if s == "" {
			return 0
		}
		u, _ := syscall.UTF16PtrFromString(s)
		return uintptr(unsafe.Pointer(u))
	}
	r, _, _ := pShellExecute.Call(0, p(verb), p(file), p(args), 0, 1)
	if r <= 32 {
		return fmt.Errorf("ShellExecute gagal (kode %d) untuk %s", r, file)
	}
	return nil
}

// --- SendInput ---

type keybdInput struct {
	vk, scan uint16
	flags    uint32
	time     uint32
	extra    uintptr
}
type input struct {
	typ uint32
	_   uint32
	ki  keybdInput
	_   [8]byte // padding agar sizeof = 40 (ukuran INPUT di x64)
}

const (
	keyUp   = 0x2
	unicode = 0x4
)

func send(in []input) {
	if len(in) == 0 {
		return
	}
	pSendInput.Call(uintptr(len(in)), uintptr(unsafe.Pointer(&in[0])), unsafe.Sizeof(in[0]))
}

func vkEvent(vk uint16, up bool) input {
	var f uint32
	if up {
		f = keyUp
	}
	return input{typ: 1, ki: keybdInput{vk: vk, flags: f}}
}

var mods = map[string]uint16{"ctrl": 0x11, "alt": 0x12, "shift": 0x10, "win": 0x5B}
var named = map[string]uint16{
	"enter": 0x0D, "tab": 0x09, "esc": 0x1B, "space": 0x20, "backspace": 0x08,
	"delete": 0x2E, "up": 0x26, "down": 0x28, "left": 0x25, "right": 0x27,
	"home": 0x24, "end": 0x23, "pgup": 0x21, "pgdn": 0x22,
}

func keyVK(name string) (uint16, error) {
	if v, ok := named[name]; ok {
		return v, nil
	}
	if len(name) >= 2 && name[0] == 'f' {
		var n int
		if _, err := fmt.Sscanf(name[1:], "%d", &n); err == nil && n >= 1 && n <= 12 {
			return uint16(0x6F + n), nil
		}
	}
	if len(name) == 1 {
		c := name[0]
		switch {
		case c >= 'a' && c <= 'z':
			return uint16(c - 32), nil
		case c >= '0' && c <= '9':
			return uint16(c), nil
		}
	}
	return 0, fmt.Errorf("tombol tidak dikenal: %q", name)
}

// sendKeys: token dipisah spasi, tiap token "ctrl+shift+t" atau "enter";
// "wait:300" = jeda 300 ms.
func sendKeys(spec string) error {
	for _, tok := range strings.Fields(strings.ToLower(spec)) {
		if strings.HasPrefix(tok, "wait:") {
			var ms int
			fmt.Sscanf(tok[5:], "%d", &ms)
			time.Sleep(time.Duration(ms) * time.Millisecond)
			continue
		}
		parts := strings.Split(tok, "+")
		key, err := keyVK(parts[len(parts)-1])
		if err != nil {
			return err
		}
		var seq []input
		var held []uint16
		for _, m := range parts[:len(parts)-1] {
			v, ok := mods[m]
			if !ok {
				return fmt.Errorf("modifier tidak dikenal: %q", m)
			}
			held = append(held, v)
			seq = append(seq, vkEvent(v, false))
		}
		seq = append(seq, vkEvent(key, false), vkEvent(key, true))
		for i := len(held) - 1; i >= 0; i-- {
			seq = append(seq, vkEvent(held[i], true))
		}
		send(seq)
		time.Sleep(20 * time.Millisecond)
	}
	return nil
}

// sendText mengetik teks apa adanya (Unicode), tanpa menyentuh clipboard.
func sendText(s string) error {
	var seq []input
	for _, u := range utf16.Encode([]rune(s)) {
		switch u {
		case '\r':
			continue
		case '\n':
			seq = append(seq, vkEvent(0x0D, false), vkEvent(0x0D, true))
		default:
			seq = append(seq,
				input{typ: 1, ki: keybdInput{scan: u, flags: unicode}},
				input{typ: 1, ki: keybdInput{scan: u, flags: unicode | keyUp}})
		}
	}
	send(seq)
	return nil
}
