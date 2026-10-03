package main

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	k32 = syscall.NewLazyDLL("kernel32.dll")
	u32 = syscall.NewLazyDLL("user32.dll")

	pGetConsoleMode       = k32.NewProc("GetConsoleMode")
	pSetConsoleMode       = k32.NewProc("SetConsoleMode")
	pGetConsoleWindow     = k32.NewProc("GetConsoleWindow")
	pSetConsoleBufferSize = k32.NewProc("SetConsoleScreenBufferSize")
	pSetConsoleWindowInfo = k32.NewProc("SetConsoleWindowInfo")
	pFreeConsole          = k32.NewProc("FreeConsole")

	pGetWindowRect    = u32.NewProc("GetWindowRect")
	pMoveWindow       = u32.NewProc("MoveWindow")
	pGetSystemMetrics = u32.NewProc("GetSystemMetrics")
	pIsWindowVisible  = u32.NewProc("IsWindowVisible")
)

// enableRaw: input tanpa echo/line-buffer + escape sequence VT (panah = ESC [ A),
// output mendukung warna ANSI. Mengembalikan fungsi untuk memulihkan mode awal.
func enableRaw() func() {
	hin, hout := os.Stdin.Fd(), os.Stdout.Fd()
	var inMode, outMode uint32
	pGetConsoleMode.Call(hin, uintptr(unsafe.Pointer(&inMode)))
	pGetConsoleMode.Call(hout, uintptr(unsafe.Pointer(&outMode)))
	pSetConsoleMode.Call(hin, uintptr((inMode&^(0x1|0x2|0x4))|0x200))
	pSetConsoleMode.Call(hout, uintptr(outMode|0x1|0x4))
	return func() {
		pSetConsoleMode.Call(hin, uintptr(inMode))
		pSetConsoleMode.Call(hout, uintptr(outMode))
	}
}

type smallRect struct{ left, top, right, bottom int16 }

// resizeConsole mengatur ukuran jendela konsol ke cols x rows karakter lalu
// menaruhnya di tengah layar. Hanya berefek di conhost klasik; di Windows
// Terminal jendelanya dikendalikan WT, bukan program ini.
func resizeConsole(cols, rows int) {
	hout := os.Stdout.Fd()
	tiny := smallRect{0, 0, 1, 1}
	pSetConsoleWindowInfo.Call(hout, 1, uintptr(unsafe.Pointer(&tiny))) // kecilkan dulu, supaya buffer boleh menyusut
	size := uint32(uint16(cols)) | uint32(uint16(rows))<<16
	pSetConsoleBufferSize.Call(hout, uintptr(size))
	final := smallRect{0, 0, int16(cols - 1), int16(rows - 1)}
	pSetConsoleWindowInfo.Call(hout, 1, uintptr(unsafe.Pointer(&final)))

	hwnd, _, _ := pGetConsoleWindow.Call()
	if hwnd == 0 {
		return
	}
	if vis, _, _ := pIsWindowVisible.Call(hwnd); vis == 0 {
		return
	}
	var r struct{ l, t, r, b int32 }
	pGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	sw, _, _ := pGetSystemMetrics.Call(0)
	sh, _, _ := pGetSystemMetrics.Call(1)
	w, h := r.r-r.l, r.b-r.t
	pMoveWindow.Call(hwnd, uintptr((int32(sw)-w)/2), uintptr((int32(sh)-h)/2), uintptr(w), uintptr(h), 1)
}

// detachConsole menutup jendela menu (bila proses ini satu-satunya yang terhubung)
// supaya fokus kembali ke aplikasi sebelumnya sebelum keys/text dikirim.
func detachConsole() { pFreeConsole.Call() }
