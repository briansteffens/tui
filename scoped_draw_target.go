package tui

import (
	"errors"
	"github.com/nsf/termbox-go"
)

// A DrawTarget represents a drawable portion of the terminal window. Drawing
// with methods like SetCell and Print will automatically translate from local
// coordinates to screen coordinates and clip drawing to the drawable region.
// It can be further subdivided by calling ChildDrawTarget().
type ScopedDrawTarget struct {
	width  int
	height int

	offsetLeft int
	offsetTop  int

	parent DrawTarget
}

func (target ScopedDrawTarget) Width() int {
	return target.width
}

func (target ScopedDrawTarget) Height() int {
	return target.height
}

// Set one terminal cell. If (x, y) is out of bounds, an error will be returned
// and the terminal will be unchanged.
func (target ScopedDrawTarget) SetCell(x, y int,
	foreground, background termbox.Attribute, char rune) error {
	if !Bounds(target).ContainsPoint(x, y) {
		return errors.New(
			"Coordinates are out of bounds for the ScopedDrawTarget")
	}

	parentX, parentY := target.localToParentCoords(x, y)
	target.parent.SetCell(parentX, parentY, foreground, background, char)
	return nil
}

// Write formatted text to the terminal using the "fmt" package formatting
// style. The text will be automatically clipped to the ScopedDrawTarget's drawable
// region.
func (target ScopedDrawTarget) Print(x, y int,
	foreground, background termbox.Attribute, text string,
	args ...interface{}) {
	parentX, parentY := target.localToParentCoords(x, y)
	target.parent.Print(parentX, parentY, foreground, background, text,
		args...)
}

// Create a ScopedDrawTarget which allows drawing to a portion of the parent's
// ScopedDrawTarget area. childBounds should be specified in the parent's local
// coordinates.
//
// Note: this is mostly needed if you're writing a control that contains other
// controls.
func Scope(parent DrawTarget, childBounds Rect) (*ScopedDrawTarget, error) {
	if !Bounds(parent).ContainsRect(childBounds) {
		return nil, errors.New("Provided child bounds would exceed " +
			"the parent's dimensions.")
	}

	return &ScopedDrawTarget{
		parent:	    parent,
		offsetLeft: childBounds.Left,
		offsetTop:  childBounds.Top,
		width:      childBounds.Width,
		height:     childBounds.Height,
	}, nil
}

func (target ScopedDrawTarget) localToParentCoords(x, y int) (int, int) {
	return target.offsetLeft + x, target.offsetTop + y
}
