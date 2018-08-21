package tui

import (
	"github.com/briansteffens/escapebox"
	"github.com/nsf/termbox-go"
)

type ButtonClickEvent func(*Button)

type Button struct {
	Bounds       Rect
	Text         string
	focus        bool
	ClickHandler ButtonClickEvent
}

func (b *Button) GetBounds() *Rect {
	return &b.Bounds
}

func (b *Button) Draw(target *DrawTarget) {
	target.Print(2, 1, termbox.ColorWhite, termbox.ColorBlack, b.Text)

	if b.focus {
		termbox.SetCursor(b.Bounds.Left+1, b.Bounds.Top+1)
	}
}

func (b *Button) SetFocus() {
	b.focus = true
}

func (b *Button) UnsetFocus() {
	b.focus = false
}

func (b *Button) HandleEvent(ev escapebox.Event) bool {
	switch ev.Type {
	case termbox.EventKey:
		switch ev.Key {
		case termbox.KeyEnter, termbox.KeySpace:
			if b.ClickHandler != nil {
				b.ClickHandler(b)
			}
			return true
		}
	}

	return false
}
