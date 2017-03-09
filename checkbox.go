package tui

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)

type Checkbox struct {
	Bounds  Rect
	Text    string
	Checked bool
	focus   bool
}

func (c *Checkbox) Render() {
	checkContent := " "

	if c.Checked {
		checkContent = "X"
	}

	s := fmt.Sprintf("[%s] %s", checkContent, c.Text)

	count := min(len(s), c.Bounds.Width)
	termPrintf(c.Bounds.Left, c.Bounds.Top, s[0:count])

	if c.focus {
		termbox.SetCursor(c.Bounds.Left + 1, c.Bounds.Top)
	}
}

func (c *Checkbox) SetFocus() {
	c.focus = true
}

func (c *Checkbox) UnsetFocus() {
	c.focus = false
}

func (c *Checkbox) HandleEvent(ev escapebox.Event) {
	switch ev.Type {
	case termbox.EventKey:
		switch ev.Key {
		case termbox.KeySpace:
			c.Checked = !c.Checked
		}
	}
}
