package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func main() {
	exe, _ := os.Executable()
	cfg, err := loadConfig(filepath.Join(filepath.Dir(exe), "koha.json"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "koha:", err)
		fmt.Fprintln(os.Stderr, "Tekan Enter untuk keluar...")
		fmt.Scanln()
		os.Exit(1)
	}

	restore := enableRaw()
	chosen := runMenu(cfg)
	restore()
	if chosen == nil {
		fmt.Print("\x1b[0m\x1b[?25h")
		return
	}
	if err := runItem(*chosen); err != nil {
		fmt.Fprintln(os.Stderr, "koha:", err)
		fmt.Scanln()
	}
}

type level struct {
	items []Item
	sel   int
}

func runMenu(cfg *Config) *Item {
	th := themes[cfg.Theme]
	stack := []level{{items: cfg.Items}}
	for {
		cur := &stack[len(stack)-1]
		draw(cur, th)
		switch readKey() {
		case keyUpArrow:
			cur.sel = (cur.sel - 1 + len(cur.items)) % len(cur.items)
		case keyDownArrow:
			cur.sel = (cur.sel + 1) % len(cur.items)
		case keyEnter:
			it := cur.items[cur.sel]
			if len(it.Children) > 0 {
				stack = append(stack, level{items: it.Children})
			} else {
				return &it
			}
		case keyEsc:
			if len(stack) == 1 {
				return nil
			}
			stack = stack[:len(stack)-1]
		case keyQuit:
			return nil
		}
	}
}

func draw(l *level, th Theme) {
	width := 28
	for _, it := range l.items {
		if n := utf8.RuneCountInString(it.Label) + 6; n > width {
			width = n
		}
	}
	resizeConsole(width, len(l.items))

	var b strings.Builder
	b.WriteString("\x1b[?25l" + bgSeq(th.bg) + "\x1b[2J\x1b[H")
	for i, it := range l.items {
		label := it.Label
		if len(it.Children) > 0 {
			label += " >"
		}
		var line string
		if i == l.sel {
			b.WriteString(bgSeq(th.selBg) + fgSeq(th.selFg))
			line = "> " + label
		} else {
			b.WriteString(bgSeq(th.bg) + fgSeq(th.fg))
			line = "  " + label
		}
		b.WriteString(line + strings.Repeat(" ", width-utf8.RuneCountInString(line)))
		if i < len(l.items)-1 {
			b.WriteString("\r\n")
		}
	}
	os.Stdout.WriteString(b.String())
}

const (
	keyOther = iota
	keyUpArrow
	keyDownArrow
	keyEnter
	keyEsc
	keyQuit
)

func readKey() int {
	var buf [16]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil || n == 0 {
		return keyQuit
	}
	s := string(buf[:n])
	switch {
	case s == "\x1b[A" || s == "\x1bOA" || s == "\x1b[Z" || s == "k":
		return keyUpArrow
	case s == "\x1b[B" || s == "\x1bOB" || s == "\t" || s == "j":
		return keyDownArrow
	case s == "\r":
		return keyEnter
	case s == "\x1b":
		return keyEsc
	case s == "\x03":
		return keyQuit
	}
	return keyOther
}
