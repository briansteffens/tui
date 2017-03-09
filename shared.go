package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func renderableChar(k termbox.Key) bool {
	return k != termbox.KeyEnter      &&
	       k != termbox.KeyPgup       &&
	       k != termbox.KeyPgdn       &&
	       k != termbox.KeyInsert     &&
	       k != termbox.KeyArrowUp    &&
	       k != termbox.KeyArrowDown  &&
	       k != termbox.KeyArrowLeft  &&
	       k != termbox.KeyArrowRight &&
	       k != termbox.KeyBackspace  &&
	       k != termbox.KeyBackspace2
}

func setCell(x, y int, r rune) {
	termbox.SetCell(x, y, r, termbox.ColorWhite, termbox.ColorBlack)
}

func termPrintf(x, y int, format string, args ...interface{}) {
	termPrintColorf(x, y, termbox.ColorWhite, termbox.ColorBlack, format,
			args...)
}

func termPrintColorf(x, y int, fg, bg termbox.Attribute, format string,
		     args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	for i, c := range s {
		termbox.SetCell(x + i, y, c, fg, bg)
	}
}

func RenderBorder(r Rect) {
	// Corners
	setCell(r.Left, r.Top, '┌')
	setCell(r.Right(), r.Top, '┓')
	setCell(r.Left, r.Bottom(), '└')
	setCell(r.Right(), r.Bottom(), '┘')

	// Horizontal borders
	for x := r.Left + 1; x < r.Right(); x++ {
		setCell(x, r.Top, '-')
		setCell(x, r.Bottom(), '-')
	}

	// Vertical borders
	for y := r.Top + 1; y < r.Bottom(); y++ {
		setCell(r.Left, y, '┃')
		setCell(r.Right(), y, '┃')
	}
}

type Rect struct {
	Left, Top, Width, Height int
}

func (r *Rect) Right() int {
	return r.Left + r.Width - 1
}

func (r *Rect) Bottom() int {
	return r.Top + r.Height - 1
}

type Control interface {
	Render()
}

type Focusable interface {
	Control
	SetFocus()
	UnsetFocus()
	HandleEvent(escapebox.Event)
}
