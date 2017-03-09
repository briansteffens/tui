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

func renderableChar(k termbox.Key) bool {
	return k != termbox.KeyEnter      &&
	       k != termbox.KeyPgup       &&
	       k != termbox.KeyPgdn       &&
	       k != termbox.KeyInsert     &&
	       k != termbox.KeyArrowUp    &&
	       k != termbox.KeyArrowDown  &&
	       k != termbox.KeyArrowLeft  &&
	       k != termbox.KeyArrowRight &&
	       k != termbox.KeyBackspace  &&
	       k != termbox.KeyBackspace2
}

func setCell(x, y int, r rune) {
	termbox.SetCell(x, y, r, termbox.ColorWhite, termbox.ColorBlack)
}

func termPrintf(x, y int, format string, args ...interface{}) {
	termPrintColorf(x, y, termbox.ColorWhite, termbox.ColorBlack, format,
			args...)
}

func termPrintColorf(x, y int, fg, bg termbox.Attribute, format string,
		     args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	for i, c := range s {
		termbox.SetCell(x + i, y, c, fg, bg)
	}
}

func RenderBorder(r Rect) {
	// Corners
	setCell(r.Left, r.Top, '┌')
	setCell(r.Right(), r.Top, '┓')
	setCell(r.Left, r.Bottom(), '└')
	setCell(r.Right(), r.Bottom(), '┘')

	// Horizontal borders
	for x := r.Left + 1; x < r.Right(); x++ {
		setCell(x, r.Top, '-')
		setCell(x, r.Bottom(), '-')
	}

	// Vertical borders
	for y := r.Top + 1; y < r.Bottom(); y++ {
		setCell(r.Left, y, '┃')
		setCell(r.Right(), y, '┃')
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

func log(format string, args ...interface{}) {
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

func MainLoop(c Container) {
	c.FocusNext()
	refresh(c)

	loop: for {
		ev := escapebox.PollEvent()

		handled := false

		switch ev.Seq {
		case escapebox.SeqNone:
			switch ev.Type {
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

		refresh(c)
	}
}

