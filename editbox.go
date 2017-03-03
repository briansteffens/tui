package main

import (
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)

type Editbox struct {
	Bounds Rect
	Value  string
	cursor int
	scroll int
	focus  bool
}

func (e *Editbox) Render() {
	RenderBorder(e.Bounds)
	termPrintf(e.Bounds.Left + 1, e.Bounds.Top + 1, e.Value)

	if e.focus {
		termbox.SetCursor(e.Bounds.Left + 1 + e.cursor - e.scroll,
				  e.Bounds.Top + 1)
	}
}

func (e *Editbox) SetFocus() {
	e.focus = true
}

func (e *Editbox) UnsetFocus() {
	e.focus = false
}

func (e *Editbox) HandleEvent(ev escapebox.Event) {
}
