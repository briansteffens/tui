package tui

import (
	"github.com/briansteffens/escapebox"
	"github.com/nsf/termbox-go"
)

type CheckBox struct {
	Bounds  Rect
	Text    string
	Checked bool
	focus   bool
}

func (c *CheckBox) GetBounds() *Rect {
	return &c.Bounds
}

func (c *CheckBox) Draw(target DrawTarget) {
	checkContent := " "

	if c.Checked {
		checkContent = "X"
	}

	target.Print(0, 0, termbox.ColorWhite, termbox.ColorBlack,
		"[%s] %s", checkContent, c.Text)

	if c.focus {
		termbox.SetCursor(c.Bounds.Left+1, c.Bounds.Top)
	}
}

func (c *CheckBox) SetFocus() {
	c.focus = true
}

func (c *CheckBox) UnsetFocus() {
	c.focus = false
}

func (c *CheckBox) HandleEvent(ev escapebox.Event) bool {
	switch ev.Type {
	case termbox.EventKey:
		switch ev.Key {
		case termbox.KeySpace:
			c.Checked = !c.Checked
			return true
		}
	}

	return false
}
