package tui

import (
	"github.com/nsf/termbox-go"
)

type DrawTarget interface {
	Width() int
	Height() int
	SetCell(x, y int, foreground, background termbox.Attribute,
		char rune) error
	Print(x, y int, foreground, background termbox.Attribute, text string,
		args ...interface{})
}

func Bounds(target DrawTarget) Rect {
	return Rect{0, 0, target.Width(), target.Height()}
}
