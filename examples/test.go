package main

import (
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/tui"
)

func buttonClickHandler(b *tui.Button) {
	panic("clicked!")
}

func main() {
	tui.Init()
	defer tui.Close()

	edit1 := tui.EditBox {
		Bounds: tui.Rect { Left: 2, Top: 6, Width: 30, Height: 10 },
	}

	edit1.SetText("abcdefgh")

	l := tui.Label {
		Bounds: tui.Rect { Left: 2, Top: 1, Width: 20, Height: 1 },
		Text: "Greetings:",
	}

	t := tui.TextBox {
		Bounds: tui.Rect { Left: 2, Top: 2, Width: 5, Height: 3 },
		Value: "12",
	}

	t2 := tui.TextBox {
		Bounds: tui.Rect { Left: 10, Top: 2, Width: 15, Height: 3},
		Value: "Greetings!",
	}

	checkbox1 := tui.CheckBox {
		Bounds: tui.Rect { Left: 27, Top: 1, Width: 30, Height: 1},
		Text: "Enable the whateverthing",
	}

	button1 := tui.Button {
		Bounds: tui.Rect { Left: 27, Top: 2, Width: 10, Height: 3},
		Text: "Continue!",
		ClickHandler: buttonClickHandler,
	}

	dv := tui.DetailView {
		Bounds: tui.Rect { Left: 2, Top: 16, Width: 25, Height: 3 },
		Columns: []tui.Column {
			tui.Column { Name: "ID", Width: 3 },
			tui.Column { Name: "Name", Width: 5 },
			tui.Column { Name: "More Data", Width: 20 },
		},
		Rows: [][]string {
			[]string { "3", "A", "Other details" },
			[]string { "7", "B", "Yes very many details" },
			[]string { "13", "C", "Such an informative table" },
			[]string { "17", "D", "Abcdefghijklmnopqrst" },
		},
		RowBg: termbox.Attribute(0),
		RowBgAlt: termbox.Attribute(236),
		SelectedBg: termbox.Attribute(22),
	}

	c := tui.Container {
		Controls: []tui.Control {&t, &dv, &edit1, &l, &t2, &checkbox1,
					 &button1},
		KeyBindingExit: tui.KeyBinding { Key: termbox.KeyCtrlC },
		KeyBindingFocusNext: tui.KeyBinding { Key: termbox.KeyTab },
		KeyBindingFocusPrevious: tui.KeyBinding {
			Seq: tui.SeqShiftTab,
		},
	}

	tui.MainLoop(&c)
}
