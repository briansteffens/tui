package tui

import (
	"github.com/nsf/termbox-go"
)

type Label struct {
	Bounds		Rect
	Text		string
}

func (l *Label) GetBounds() *Rect {
	return &l.Bounds
}

func (l *Label) Draw(target *DrawTarget) {
	target.Print(0, 0, termbox.ColorWhite, termbox.ColorBlack, l.Text)
}
