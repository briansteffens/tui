package tui

import (
	"fmt"
	"github.com/briansteffens/escapebox"
	"github.com/nsf/termbox-go"
	"os"
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

// Non-standard escape sequences
const (
	SeqShiftTab = 1
)

func renderableChar(ev escapebox.Event) bool {
	return ev.Type == termbox.EventKey && ev.Key == 0 && ev.Ch != 0
}

type KeyBinding struct {
	Seq escapebox.Sequence
	Key termbox.Key
	Ch  rune
}

// Check if a KeyBinding matches an event
func matchBinding(ev escapebox.Event, kb KeyBinding) bool {
	if ev.Type != termbox.EventKey {
		return false
	}

	if kb.Seq != 0 {
		return ev.Seq == kb.Seq
	}

	if kb.Ch != 0 {
		return ev.Ch == kb.Ch
	}

	if ev.Key != 0 {
		return ev.Key == kb.Key
	}

	return false
}

type Control interface {
	GetBounds() *Rect
	Draw(IDrawTarget)
}

type Focusable interface {
	Control
	SetFocus()
	UnsetFocus()
	HandleEvent(escapebox.Event) bool
}

func Init() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	termbox.SetInputMode(termbox.InputEsc) // | termbox.InputMouse)
	termbox.SetOutputMode(termbox.Output256)

	escapebox.Init()
	escapebox.Register(SeqShiftTab, 91, 90)
}

func Close() {
	escapebox.Close()
	termbox.Close()
}

func log(message string, args ...interface{}) {
	mode := os.O_APPEND | os.O_WRONLY | os.O_CREATE
	f, err := os.OpenFile(".tui-log", mode, 0644)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = fmt.Fprintf(f, message, args...); err != nil {
		panic(err)
	}

	if _, err = fmt.Fprintln(f); err != nil {
		panic(err)
	}
}

func Refresh(root *Container) {
	target := fullTerminalDrawTarget()
	root.Draw(target)
	termbox.Flush()
}

func MainLoop(c *Container) {
	c.FocusNext()

	c.Width, c.Height = termbox.Size()
	if c.ResizeHandler != nil {
		c.ResizeHandler()
	}

	Refresh(c)

loop:
	for {
		ev := escapebox.PollEvent()

		if matchBinding(ev, c.KeyBindingExit) {
			break loop
		}

		handled := false

		if ev.Type == termbox.EventResize {
			c.Width = ev.Width
			c.Height = ev.Height

			if c.ResizeHandler != nil {
				c.ResizeHandler()
			}

			handled = true
		}

		if !handled && c.Focused != nil {
			handled = c.Focused.HandleEvent(ev)
		}

		if !handled && matchBinding(ev, c.KeyBindingFocusNext) {
			c.FocusNext()
			handled = true
		}

		if !handled && matchBinding(ev, c.KeyBindingFocusPrevious) {
			c.FocusPrevious()
			handled = true
		}

		if !handled && c.HandleEvent != nil {
			handled = c.HandleEvent(c, ev)
		}

		Refresh(c)
	}
}
