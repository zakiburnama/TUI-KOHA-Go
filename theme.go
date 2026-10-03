package main

import "fmt"

type Theme struct{ bg, fg, selBg, selFg [3]uint8 }

func rgb(h uint32) [3]uint8 { return [3]uint8{uint8(h >> 16), uint8(h >> 8), uint8(h)} }

// Warna sama dengan THEMES di koha.ahk.
var themes = map[string]Theme{
	"game_boy":         {rgb(0x9BBC0F), rgb(0x0F380F), rgb(0x0F380F), rgb(0x9BBC0F)},
	"amber":            {rgb(0x1A0F00), rgb(0xFFB000), rgb(0xFFB000), rgb(0x1A0F00)},
	"green_term":       {rgb(0x0A0A0A), rgb(0x33FF33), rgb(0x33FF33), rgb(0x0A0A0A)},
	"catppuccin-mocha": {rgb(0x1E1E2E), rgb(0xCDD6F4), rgb(0xCBA6F7), rgb(0x1E1E2E)},
	"gruvbox":          {rgb(0x282828), rgb(0xEBDBB2), rgb(0xFE8019), rgb(0x282828)},
	"vague":            {rgb(0x141415), rgb(0xCDCDCD), rgb(0x6E94B2), rgb(0x141415)},
	"tokyonight":       {rgb(0x1A1B26), rgb(0xC0CAF5), rgb(0x7AA2F7), rgb(0x1A1B26)},
}

func fgSeq(c [3]uint8) string { return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", c[0], c[1], c[2]) }
func bgSeq(c [3]uint8) string { return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", c[0], c[1], c[2]) }
