package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)

type Textbox struct {
	Bounds Rect
	Value  string
	cursor int
	scroll int
	focus  bool
}

func renderableChar(k termbox.Key) bool {
	return k != termbox.KeyEnter  &&
	       k != termbox.KeyPgup   &&
	       k != termbox.KeyPgdn   &&
	       k != termbox.KeyInsert
}

func (t* Textbox) maxVisibleChars() int {
	return t.Bounds.Width - 2
}

func (t* Textbox) visibleChars() int {
	return min(t.maxVisibleChars(), len(t.Value) - t.scroll)
}

func (t* Textbox) lastVisible() int {
	return t.scroll + t.visibleChars() - 1
}

func (t* Textbox) Render() {
	RenderBorder(t.Bounds)
	termPrintf(t.Bounds.Left + 1, t.Bounds.Top + 1,
		   t.Value[t.scroll:t.lastVisible() + 1])

	if t.focus {
		termbox.SetCursor(t.Bounds.Left + 1 + t.cursor - t.scroll,
				  t.Bounds.Top + 1)
	}
}

func (t* Textbox) SetFocus() {
	t.focus = true
}

func (t* Textbox) UnsetFocus() {
	t.focus = false
}

func (t* Textbox) HandleEvent(ev escapebox.Event) {
	pre := t.Value[0:t.cursor]
	post := t.Value[t.cursor:len(t.Value)]

	switch ev.Type {
	case termbox.EventKey:
		char := fmt.Sprintf("%c", ev.Ch)

		switch ev.Key {
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			if len(pre) > 0 {
				t.Value = pre[0:len(pre)-1] + post
				t.cursor--
			}
		case termbox.KeyDelete:
			if len(post) > 0 {
				t.Value = pre + post[1:len(post)]
			}
		case termbox.KeyArrowLeft:
			t.cursor--
		case termbox.KeyArrowRight:
			t.cursor++
		case termbox.KeyHome:
			t.cursor = 0
		case termbox.KeyEnd:
			t.cursor = len(t.Value)
		default:
			if renderableChar(ev.Key) {
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
}

