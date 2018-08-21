package tui

import (
	"errors"
	"github.com/nsf/termbox-go"
)

type ScrolledDrawTarget struct {
	width  int
	height int

	scrollLeft int
	scrollTop  int

	parent DrawTarget
}

func (target ScrolledDrawTarget) Width() int {
	return target.width
}

func (target ScrolledDrawTarget) Height() int {
	return target.height
}

func (target ScrolledDrawTarget) ScrollLeft() int {
	return target.scrollLeft
}

func (target ScrolledDrawTarget) ScrollTop() int {
	return target.scrollTop
}

func (target ScrolledDrawTarget) SetCell(x, y int,
	foreground, background termbox.Attribute, char rune) error {
	if !Bounds(target).ContainsPoint(x, y) {
		return errors.New("Coordinates are out of bounds for the " +
			"ScrolledDrawTarget")
	}

	parentX, parentY := target.scrollCoords(x, y)
	target.parent.SetCell(parentX, parentY, foreground, background, char)
	return nil
}

// Write formatted text to the terminal using the "fmt" package formatting
// style. The text will be automatically clipped to the ScrolledDrawTarget's drawable
// region.
func (target ScrolledDrawTarget) Print(x, y int,
	foreground, background termbox.Attribute, text string,
	args ...interface{}) {
	parentX, parentY := target.scrollCoords(x, y)
	target.parent.Print(parentX, parentY, foreground, background, text,
		args...)
}

func Scroll(parent DrawTarget, width, height int) (ScrolledDrawTarget, error) {
	return ScrolledDrawTarget{
		width:      width,
		height:     height,

		scrollLeft: 0,
		scrollTop:  0,

		parent:	    parent,
	}, nil
}

func (target ScrolledDrawTarget) scrollCoords(x, y int) (int, int) {
	return x - target.scrollLeft, y - target.scrollTop
}
