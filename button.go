package main

import (
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)

type ButtonClickEvent func(*button)

type button struct {
	Bounds       Rect
	Text         string
	focus        bool
	ClickHandler ButtonClickEvent
}

func (b* button) Render() {
	RenderBorder(b.Bounds)

	count := min(len(b.Text), b.Bounds.Width - 4)
	termPrintf(b.Bounds.Left + 2, b.Bounds.Top + 1, b.Text[0:count])

	if b.focus {
		termbox.SetCursor(b.Bounds.Left + 1, b.Bounds.Top + 1)
	}
}

func (b* button) SetFocus() {
	b.focus = true
}

func (b* button) UnsetFocus() {
	b.focus = false
}

func (b* button) HandleEvent(ev escapebox.Event) {
	switch ev.Type {
	case termbox.EventKey:
		switch ev.Key {
		case termbox.KeyEnter, termbox.KeySpace:
			if b.ClickHandler != nil {
				b.ClickHandler(b)
			}
		}
	}
}
