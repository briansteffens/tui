package tui

import (
	"errors"
	"fmt"
	"github.com/nsf/termbox-go"
	"golang.org/x/text/unicode/norm"
	"unicode/utf8"
)

type DrawTarget interface {
	Width() int
	Height() int
	SetCell(x, y int, foreground, background termbox.Attribute,
		char rune) error
	Print(x, y int, foreground, background termbox.Attribute, text string,
		args ...interface{})
}

func Bounds(target DrawTarget) *Rect {
	return &Rect{
		Left: 0,
		Top: 0,
		Width: target.Width(),
		Height: target.Height(),
	}
}

type TermboxDrawTarget struct {
	width, height int
}

func (target TermboxDrawTarget) Width() int {
	return target.width
}

func (target TermboxDrawTarget) Height() int {
	return target.height
}

// Create a draw target that allows drawing to the entire terminal window.
func newTermboxDrawTarget() *TermboxDrawTarget {
	terminalWidth, terminalHeight := termbox.Size()

	return &TermboxDrawTarget{
		width:      terminalWidth,
		height:     terminalHeight,
	}
}

func (target TermboxDrawTarget) Bounds() *Rect {
	return &Rect{
		Left: 0,
		Top: 0,
		Width: target.width,
		Height: target.height,
	}
}

// Set one terminal cell. If (x, y) is out of bounds, an error will be returned
// and the terminal will be unchanged.
func (target TermboxDrawTarget) SetCell(x, y int,
	foreground, background termbox.Attribute, char rune) error {
	if !target.Bounds().ContainsPoint(x, y) {
		return errors.New("Coordinates are out of bounds for the " +
			"TermboxDrawTarget")
	}

	termbox.SetCell(x, y, char, foreground, background)

	return nil
}

// Write formatted text to the terminal using the "fmt" package formatting
// style. The text will be automatically clipped to the ScopedDrawTarget's drawable
// region.
func (target TermboxDrawTarget) Print(x, y int,
	foreground, background termbox.Attribute, text string,
	args ...interface{}) {
	formatted := fmt.Sprintf(text, args...)
	normalized := normalizeString(formatted)
	for i, r := range normalized {
		target.SetCell(x+i, y, foreground, background, r)
	}
}

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
	formatted := fmt.Sprintf(text, args...)
	normalized := normalizeString(formatted)
	for i, r := range normalized {
		target.SetCell(x+i, y, foreground, background, r)
	}
}

// Create a ScopedDrawTarget which allows drawing to a portion of the parent's
// ScopedDrawTarget area. childBounds should be specified in the parent's local
// coordinates.
//
// Note: this is mostly needed if you're writing a control that contains other
// controls.
func Scope(parent DrawTarget, childBounds *Rect) (*ScopedDrawTarget, error) {
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

// The termbox API allows a single rune per terminal cell (x, y). Strings can
// contain grapheme clusters which are made up of multiple runes but occupy
// only the width of a single character when displayed. I don't know a way to
// map grapheme clusters to termbox's API without data loss or display issues.
//
// This function detects multi-rune grapheme clusters and replaces them with
// Unicode replacement characters in order to be explicit that the decode was
// not entirely successful.
//
// TODO: Is there a way to support grapheme clusters on the terminal? Can it
// be done with termbox or would it require switching libraries?
func normalizeString(input string) []rune {
	output := make([]rune, 0)

	var ia norm.Iter
	ia.InitString(norm.NFKD, input)

	for !ia.Done() {
		glyph := ia.Next()
		firstRune, decodedSize := utf8.DecodeRune(glyph)
		isGraphemeCluster := decodedSize < len(glyph)

		if isGraphemeCluster {
			firstRune = utf8.RuneError
		}

		output = append(output, firstRune)

	}

	return output
}
