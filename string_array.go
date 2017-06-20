package tui

import (
	"errors"
	"github.com/nsf/termbox-go"
)

type StringArray struct {
	lines         [][]Char
	cursorLine    int
	cursorChar    int
	onDataChanged DataChangedEvent
}

func (d *StringArray) SetDataChangedEvent(e DataChangedEvent) {
	d.onDataChanged = e
}

func (d *StringArray) GetText() string {
	ret := ""

	for i, line := range d.lines {
		if i > 0 {
			ret += "\n"
		}

		for _, char := range line {
			ret += string(char.Char)
		}
	}

	return ret
}

func (d *StringArray) SetText(raw string) {
	d.lines = [][]Char{}

	line := []Char{}

	for i, c := range raw {
		isEnd := i == len(raw) - 1

		if c != '\n' {
			char := Char {
				Char: c,
				Fg: termbox.ColorWhite,
				Bg: termbox.ColorBlack,
			}

			line = append(line, char)
		}

		if c == '\n' || isEnd {
			d.lines = append(d.lines, line)
			line = []Char{}
		}
	}
}

// Find the line and char index where the given character is
func (d *StringArray) indexToChar(index int) (int, int, bool) {
	for l := 0; l < len(d.lines); l++ {
		// "+ 1" for implicit newline
		lineWidth := len(d.lines[l]) + 1

		if index >= lineWidth {
			index -= lineWidth
			continue
		}

		return l, index, true
	}

	return -1, -1, false
}

func (d *StringArray) GetChar(index int) (*Char, error) {
	lineIndex, charIndex, ok := d.indexToChar(index)

	if !ok {
		return nil, errors.New("Index out of range")
	}

	line := d.lines[lineIndex]

	// Implicit newline
	if charIndex == len(line) {
		return &Char { Char: '\n' }, nil
	}

	return &line[charIndex], nil
}

func (d *StringArray) GetCursor() int {
	ret := 0

	for l := 0; l < d.cursorLine; l++ {
		ret += len(d.lines[l]) + 1
	}

	return ret + d.cursorChar
}

func (d *StringArray) Insert(newText string) {
	line := d.lines[d.cursorLine]

	pre  := line[0:d.cursorChar]
	post := line[d.cursorChar:len(line)]

	newLines := [][]Char {}

	newLine := make([]Char, len(pre))

	// Copy the pre part of the line being inserted into
	for i := 0; i < len(pre); i++ {
		newLine[i] = pre[i]
	}

	// Copy the new text into place
	for _, r := range newText {
		if r != '\n' {
			newLine = append(newLine, Char { Char: r })
			continue
		}

		newLines = append(newLines, newLine)
		newLine = []Char {}
	}

	// Copy the post part of the line being inserted into
	for i := 0; i < len(post); i++ {
		newLine = append(newLine, post[i])
	}

	newLines = append(newLines, newLine)

	// If no new lines were added (just one line modified), update that one
	// line in place to save a linear copy
	if len(newLines) == 1 {
		d.lines[d.cursorLine] = newLines[0]
		return
	}

	// If lines were added, the whole lines array needs to be shifted down
	oldLinesLen := len(d.lines)

	// Make room at the end of the array
	for i := 0; i < len(newLines) - 1; i++ {
		d.lines = append(d.lines, []Char {})
	}

	// Shift post lines to the end of the array to make room for new lines
	shiftDistance := len(newLines) - 1

	for i := oldLinesLen - 1; i >= d.cursorLine + 1; i-- {
		d.lines[i + shiftDistance] = d.lines[i]
	}

	// Copy new lines into place
	for i := 0; i < len(newLines); i++ {
		d.lines[d.cursorLine + i] = newLines[i]
	}
}

func (d *StringArray) Delete() {
	line := d.lines[d.cursorLine]

	// Nothing to delete
	if len(d.lines) == 1 && len(line) == 0 {
		return
	}

	// Delete from current line
	if d.cursorChar < len(line) {
		newLine := make([]Char, len(line) - 1)

		j := 0
		for i := 0; i < len(line); i++ {
			if i == d.cursorChar {
				continue
			}

			newLine[j] = line[i]
			j++
		}

		d.lines[d.cursorLine] = newLine
		d.fireTextChanged()

		return
	}

	if d.cursorLine == len(d.lines) - 1 {
		return
	}

	// Cursor is on the implicit newline at the end of a line and there
	// are more lines after it. Concat this and the next line.
	nextLine := d.lines[d.cursorLine + 1]
	newLines := make([][]Char, len(d.lines) - 1)

	j := 0
	for i := 0; i < d.cursorLine; i++ {
		newLines[j] = d.lines[i]
		j++
	}

	newLine := []Char {}
	newLine = append(newLine, line...)
	newLine = append(newLine, nextLine...)

	newLines[j] = newLine
	j++

	for i := d.cursorLine + 2; i < len(d.lines); i++ {
		newLines[j] = d.lines[i]
		j++
	}

	d.lines = newLines
	d.fireTextChanged()

	return
}

func (d *StringArray) CursorNext() bool {
	d.cursorChar++

	if d.cursorChar <= len(d.lines[d.cursorLine]) {
		return true
	}

	// No more lines, undo the move
	if d.cursorLine == len(d.lines) - 1 {
		d.cursorChar--
		return false
	}

	// Advance to the next line
	d.cursorLine++
	d.cursorChar = 0
	return true
}

func (d *StringArray) CursorPrevious() bool {
	d.cursorChar--

	if d.cursorChar >= 0 {
		return true
	}

	// No more lines, undo the move
	if d.cursorLine == 0 {
		d.cursorChar++
		return false
	}

	// Move to the previous line
	d.cursorLine--
	d.cursorChar = len(d.lines[d.cursorLine])
	return true
}

func (d *StringArray) CursorBeginningOfLine() {
	d.cursorChar = 0
}

func (d *StringArray) CursorEndOfLine() {
	d.cursorChar = len(d.lines[d.cursorLine]) - 1
}

func (d *StringArray) LinePrevious() {
	d.cursorLine--
}

func (d *StringArray) LineNext() {
	d.cursorLine++
}

func (d *StringArray) ClampCursor() {
	d.cursorLine = max(0, d.cursorLine)
	d.cursorLine = min(len(d.lines) - 1, d.cursorLine)

	d.cursorChar = max(0, d.cursorChar)

	minChar := len(d.lines[d.cursorLine])
	//if e.mode == CommandMode && minChar > 0 {
	//    minChar--
	//}

	d.cursorChar = min(minChar, d.cursorChar)
}

func (d *StringArray) CursorAtEnd() bool {
	return d.cursorLine >= len(d.lines) - 1 &&
	       d.cursorChar >= len(d.lines[d.cursorLine]) - 1
}

func (d *StringArray) LineCount() int {
	return len(d.lines)
}

func (d *StringArray) CursorBeginning() {
	d.cursorLine = 0
	d.cursorChar = 0
}

func (d *StringArray) AllChars() []*Char {
	ret := []*Char {}

	for l := 0; l < len(d.lines); l++ {
		line := &d.lines[l]

		for c := 0; c < len(*line); c++ {
			ret = append(ret, &(*line)[c])
		}

		ret = append(ret, &Char {
			Char: '\n',
		})
	}

	return ret
}

func (d *StringArray) fireTextChanged() {
	if d.onDataChanged != nil {
		d.onDataChanged(d)
	}
}
