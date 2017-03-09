package main

import (
	"os"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)

// Non-standard escape sequences
const (
	SeqShiftTab = 1
)

func buttonClickHandler(b *button) {
	panic("clicked!")
}

var outFile *os.File

func log(format string, args ...interface{}) {
	outFile.WriteString(fmt.Sprintf(format + "\n", args...))
}

func main() {
	var err error

	outFile, err = os.Create("outfile")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc) // | termbox.InputMouse)

	escapebox.Init()
	defer escapebox.Close()

	escapebox.Register(SeqShiftTab, 91, 90)

	edit1 := Editbox {
		Bounds: Rect { Left: 2, Top: 6, Width: 30, Height: 10 },
		lines: []string {
			"Hello! This is a file which has a line!",
			"And here is another line woah",
			"And yet another",
            "",
			"Guess what",
			"They keep going!",
			"Again! Another wrapping line gogogo",
			"Ok one more",
		},
		scroll: 0,
		cursorLine: 4,
		cursorChar: 3,
	}

	l := Label {
		Bounds: Rect { Left: 2, Top: 1, Width: 20, Height: 1 },
		Text: "Greetings:",
	}

	t := Textbox {
		Bounds: Rect { Left: 2, Top: 2, Width: 5, Height: 3 },
		Value: "12",
		cursor: 2,
		scroll: 0,
	}

	t2 := Textbox {
		Bounds: Rect { Left: 10, Top: 2, Width: 15, Height: 3},
		Value: "Greetings!",
		cursor: 0,
		scroll: 0,
	}

	checkbox1 := Checkbox {
		Bounds: Rect { Left: 27, Top: 1, Width: 30, Height: 1},
		Text: "Enable the whateverthing",
	}

	button1 := button {
		Bounds: Rect { Left: 27, Top: 2, Width: 10, Height: 3},
		Text: "Continue!",
		ClickHandler: buttonClickHandler,
	}

	dv := Detailview {
		Bounds: Rect { Left: 2, Top: 20, Width: 7, Height: 4 },
		Columns: []Column {
			Column { Name: "ID", Width: 3 },
			Column { Name: "Name", Width: 5 },
			Column { Name: "More Data", Width: 20 },
		},
		Rows: [][]string {
			[]string { "3", "A", "Other details" },
			[]string { "7", "B", "Yes very many details" },
			[]string { "13", "C", "Such an informative table" },
			[]string { "17", "D", "Abcdefghijklmnopqrst" },
		},
		scrollRow: 2,
		scrollCol: 1,
		cursorRow: 2,
	}

	c := Container {
		Controls: []Control {&t, &dv, &edit1, &l, &t2, &checkbox1, &button1},
	}

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
				case termbox.KeyCtrlA:
					l.Text = ""
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
