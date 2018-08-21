package tui

import (
	"fmt"
	"errors"
	"github.com/nsf/termbox-go"
	"golang.org/x/text/unicode/norm"
	"unicode/utf8"
)

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
func newTermboxDrawTarget() TermboxDrawTarget {
	terminalWidth, terminalHeight := termbox.Size()

	return TermboxDrawTarget{
		width:      terminalWidth,
		height:     terminalHeight,
	}
}

// Set one terminal cell. If (x, y) is out of bounds, an error will be returned
// and the terminal will be unchanged.
func (target TermboxDrawTarget) SetCell(x, y int,
	foreground, background termbox.Attribute, char rune) error {
	if !Bounds(target).ContainsPoint(x, y) {
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
