package tui

import (
	"os"
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

// Non-standard escape sequences
const (
	SeqShiftTab = 1
)

func renderableChar(ev escapebox.Event) bool {
	return ev.Type == termbox.EventKey && ev.Key == 0 && ev.Ch != 0
}

func setCell(x, y int, r rune) {
	termbox.SetCell(x, y, r, termbox.ColorWhite, termbox.ColorBlack)
}

func termPrintf(x, y int, format string, args ...interface{}) {
	termPrint(x, y, fmt.Sprintf(format, args...))
}

func termPrintColorf(x, y int, fg, bg termbox.Attribute, format string,
		     args ...interface{}) {
	termPrintColor(x, y, fg, bg, fmt.Sprintf(format, args...))
}

func termPrint(x, y int, content string) {
	termPrintColor(x, y, termbox.ColorWhite, termbox.ColorBlack, content)
}

func termPrintColor(x, y int, fg, bg termbox.Attribute, content string) {
	for i, c := range content {
		termbox.SetCell(x + i, y, c, fg, bg)
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

type BorderRenderer func(c Control)

type Control interface {
	Render()
}

type Focusable interface {
	Control
	SetFocus()
	UnsetFocus()
	HandleEvent(escapebox.Event)
}

var outFile *os.File

func Log(format string, args ...interface{}) {
	outFile.WriteString(fmt.Sprintf(format + "\n", args...))
}

func Init() {
	var err error

	outFile, err = os.Create("outfile")
	if err != nil {
		panic(err)
	}

	err = termbox.Init()
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
	outFile.Close()
}

func MainLoop(c *Container) {
	c.FocusNext()

	c.Width, c.Height = termbox.Size()
	if c.ResizeHandler != nil {
		c.ResizeHandler()
	}

	c.Refresh()

	loop: for {
		ev := escapebox.PollEvent()

		handled := false

		switch ev.Seq {
		case escapebox.SeqNone:
			switch ev.Type {
			case termbox.EventResize:
				c.Width = ev.Width
				c.Height = ev.Height

				if c.ResizeHandler != nil {
					c.ResizeHandler()
				}
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyCtrlC:
					break loop
				case termbox.KeyTab:
					c.FocusNext()
					handled = true
				}
			}
		case SeqShiftTab:
			c.FocusPrevious()
			handled = true
		}

		if !handled && c.Focused != nil {
			c.Focused.HandleEvent(ev)
		}

		c.Refresh()
	}
}

