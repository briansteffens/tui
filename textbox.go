package tui

import (
	//"fmt"
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)

type TextBox struct {
	Bounds     Rect
	Value      string
	cursor     int
	scroll     int
	focus      bool
}

func (t *TextBox) maxVisibleChars() int {
	return t.Bounds.Width - 2
}

func (t *TextBox) visibleChars() int {
	return min(t.maxVisibleChars(), len(t.Value) - t.scroll)
}

func (t *TextBox) lastVisible() int {
	return t.scroll + t.visibleChars() - 1
}

func (t *TextBox) Render() {
	termPrint(t.Bounds.Left + 1, t.Bounds.Top + 1,
		  t.Value[t.scroll:t.lastVisible() + 1])

	if t.focus {
		termbox.SetCursor(t.Bounds.Left + 1 + t.cursor - t.scroll,
				  t.Bounds.Top + 1)
	}
}

func (t *TextBox) SetFocus() {
	t.focus = true
}

func (t *TextBox) UnsetFocus() {
	t.focus = false
}

func (t *TextBox) HandleEvent(ev escapebox.Event) bool {
	pre := t.Value[0:t.cursor]
	post := t.Value[t.cursor:len(t.Value)]

	handled := false

	switch ev.Type {
	case termbox.EventKey:
		char := string(ev.Ch)

		switch ev.Key {
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			if len(pre) > 0 {
				t.Value = pre[0:len(pre)-1] + post
				t.cursor--
			}
			handled = true
		case termbox.KeyDelete:
			if len(post) > 0 {
				t.Value = pre + post[1:len(post)]
			}
			handled = true
		case termbox.KeyArrowLeft:
			t.cursor--
			handled = true
		case termbox.KeyArrowRight:
			t.cursor++
			handled = true
		case termbox.KeyHome:
			t.cursor = 0
			handled = true
		case termbox.KeyEnd:
			t.cursor = len(t.Value)
			handled = true
		default:
			if renderableChar(ev) {
				t.Value = pre + char + post
				t.cursor++
			}
		}
	}

	if t.cursor < 0 {
		t.cursor = 0
	}

	if t.cursor > len(t.Value) {
		t.cursor = len(t.Value)
	}

	if t.cursor < t.scroll {
		t.scroll = t.cursor
	}

	if t.cursor >= t.scroll + t.maxVisibleChars() {
		t.scroll = t.cursor - t.maxVisibleChars() + 1
	}

	return handled
}

