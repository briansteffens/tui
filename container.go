package main

import (
	"github.com/nsf/termbox-go"
)

type Container struct {
	Controls []Control
	Focused  Focusable
}

func (c *Container) focus(f Focusable) {
	if c.Focused != nil {
		c.Focused.UnsetFocus()
	}

	c.Focused = f
	c.Focused.SetFocus()
}

func (c *Container) FocusNext() {
	currentIndex := 0

	// Find index of currently focused control
	if c.Focused != nil {
		for index, ctrl := range c.Controls {
			if ctrl == c.Focused {
				currentIndex = index
				break
			}
		}
	}

	// Scan list after focused control for another Focusable control
	for i := currentIndex + 1; i < len(c.Controls); i++ {
		f, ok := c.Controls[i].(Focusable)
		if ok {
			c.focus(f)
			return
		}
	}

	// Scan list before focused control (loop around)
	for i := 0; i <= currentIndex; i++ {
		f, ok := c.Controls[i].(Focusable)
		if ok {
			c.focus(f)
			return
		}
	}
}

func (c *Container) FocusPrevious() {
	currentIndex := 0

	// Find index of currently focused control
	if c.Focused != nil {
		for index, ctrl := range c.Controls {
			if ctrl == c.Focused {
				currentIndex = index
				break
			}
		}
	}

	// Scan list before focused control for another Focusable control
	for i := currentIndex - 1; i >= 0; i-- {
		f, ok := c.Controls[i].(Focusable)
		if ok {
			c.focus(f)
			return
		}
	}

	// Scan list after focused control (loop around)
	for i := len(c.Controls) - 1; i >= currentIndex; i-- {
		f, ok := c.Controls[i].(Focusable)
		if ok {
			c.focus(f)
			return
		}
	}
}

func refresh(c Container) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for _, v := range c.Controls {
		v.Render()
	}

	termbox.Flush()
	termbox.Sync()
}
