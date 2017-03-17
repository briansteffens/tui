package tui

import (
	"errors"
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)

const (
	CommandMode = 0
	InsertMode  = 1
)

type Char struct {
	Char rune
	Fg   termbox.Attribute
	Bg   termbox.Attribute
}

type TextChangedEvent func(*EditBox)
type CursorMovedEvent func(*EditBox)

type EditBox struct {
	Bounds        Rect
	Lines         [][]Char
	OnTextChanged TextChangedEvent
	OnCursorMoved CursorMovedEvent
	cursorLine    int
	cursorChar    int
	scroll        int
	focus         bool
	mode          int
	chord         []escapebox.Event
}

func (e *EditBox) GetCursor() int {
	ret := 0

	for l := 0; l < e.cursorLine; l++ {
		ret += len(e.Lines[l]) + 1
	}

	return ret + e.cursorChar
}

// Find the line and char index where the given character is
func (e *EditBox) indexToChar(index int) (int, int, bool) {
	for l := 0; l < len(e.Lines); l++ {
		// "+ 1" for implicit newline
		lineWidth := len(e.Lines[l]) + 1

		if index >= lineWidth {
			index -= lineWidth
			continue
		}

		return l, index, true
	}

	return -1, -1, false
}

func (e *EditBox) GetChar(index int) (*Char, error) {
	lineIndex, charIndex, ok := e.indexToChar(index)

	if !ok {
		return nil, errors.New("Index out of range")
	}

	line := e.Lines[lineIndex]

	// Implicit newline
	if charIndex == len(line) {
		return &Char { Char: '\n' }, nil
	}

	return &line[charIndex], nil
}

func (e *EditBox) GetText() string {
	ret := ""

	for i, line := range e.Lines {
		if i > 0 {
			ret += "\n"
		}

		for _, char := range line {
			ret += string(char.Char)
		}
	}

	return ret
}

func (e *EditBox) SetText(raw string) {
	e.Lines = [][]Char{}

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
			e.Lines = append(e.Lines, line)
			line = []Char{}
		}
	}

	if e.OnTextChanged != nil {
		e.OnTextChanged(e)
	}
}

func (e *EditBox) Render() {
	textWidth := e.Bounds.Width
	textHeight := e.Bounds.Height - 1 // Bottom line free for modes/notices

	// Generate virtual lines and map the cursor to them.
	virtualLines := make([][]Char, 0)

	cursorRow := 0
	cursorCol := 0

	for lineIndex, line := range e.Lines {
		virtualLineCount := len(line) / textWidth + 1

		if e.cursorLine == lineIndex {
			cursorRow = len(virtualLines) +
				    e.cursorChar / textWidth
			cursorCol = e.cursorChar % textWidth
		}

		for i := 0; i < virtualLineCount; i++ {
			start := i * textWidth
			stop := min(len(line), (i + 1) * textWidth)
			virtualLines = append(virtualLines, line[start:stop])
		}
	}

	if cursorRow < e.scroll {
		e.scroll = cursorRow
	}

	if cursorRow >= e.scroll + textHeight {
		e.scroll = cursorRow - textHeight + 1
	}

	scrollEnd := min(len(virtualLines), e.scroll + textHeight)

	for l := e.scroll; l < scrollEnd; l++ {
		for c, ch := range virtualLines[l] {
			termbox.SetCell(e.Bounds.Left + c,
					e.Bounds.Top + l - e.scroll, ch.Char,
					ch.Fg, ch.Bg)
		}
	}

	if e.focus {
		termbox.SetCursor(e.Bounds.Left + cursorCol,
				  e.Bounds.Top + cursorRow - e.scroll)
	}

	if e.mode == InsertMode {
		termPrint(e.Bounds.Left, e.Bounds.Bottom(),
			  "-- INSERT --")
	}
}

func (e *EditBox) SetFocus() {
	e.focus = true
}

func (e *EditBox) UnsetFocus() {
	e.focus = false
}

func (e *EditBox) HandleEvent(ev escapebox.Event) bool {
	if ev.Type != termbox.EventKey {
		return false
	}

	oldCursorLine := e.cursorLine
	oldCursorChar := e.cursorChar

	handled := false

	if !handled && ev.Key == termbox.KeyHome {
		e.cursorChar = 0
		handled = true
	}

	if !handled && ev.Key == termbox.KeyEnd {
		e.cursorChar = len(e.Lines[e.cursorLine]) - 1
		handled = true
	}

	if !handled && e.mode == CommandMode {
		handled = e.handleCommandModeEvent(ev)
	} else if !handled && e.mode == InsertMode {
		handled = e.handleInsertModeEvent(ev)
	}

	// Clamp the cursor to its constraints
	e.cursorLine = max(0, e.cursorLine)
	e.cursorLine = min(len(e.Lines) - 1, e.cursorLine)

	e.cursorChar = max(0, e.cursorChar)

	minChar := len(e.Lines[e.cursorLine])
	if e.mode == CommandMode && minChar > 0 {
	    minChar--
	}

	e.cursorChar = min(minChar, e.cursorChar)

	// Detect and fire OnCursorMoved
	if e.OnCursorMoved != nil &&
	   (oldCursorLine != e.cursorLine || oldCursorChar != e.cursorChar) {
		e.OnCursorMoved(e)
	}

	return handled
}

func (e *EditBox) handleChord_d() bool {
	if e.chord[1].Ch == 'd' {
		// Delete current line
		if len(e.Lines) == 0 {
			return true
		}

		if len(e.Lines) == 1 {
			e.Lines[0] = []Char {}
			return true
		}

		newLines := make([][]Char, len(e.Lines) - 1)

		for i := 0; i < e.cursorLine; i++ {
			newLines[i] = e.Lines[i]
		}

		for i := e.cursorLine + 1; i < len(e.Lines); i++ {
			newLines[i - 1] = e.Lines[i]
		}

		e.Lines = newLines
	}

	return true
}

func (e *EditBox) handleCommandModeEvent(ev escapebox.Event) bool {
	// Start a chord
	if len(e.chord) == 0 && ev.Ch == 'd' {
		e.chord = []escapebox.Event { ev }
		return true
	}

	// Continue a chord
	if len(e.chord) > 0 {
		// End chord
		if ev.Key == termbox.KeyEsc {
			e.chord = []escapebox.Event {}
			return true
		}

		e.chord = append(e.chord, ev)
		consumed := false

		switch e.chord[0].Ch {
		case 'd':
			consumed = e.handleChord_d()
		}

		// Chord consumed
		if consumed {
			e.chord = []escapebox.Event {}
		}

		return true
	}

	switch ev.Ch {
	case 'h':
		e.cursorChar--
		return true
	case 'l':
		e.cursorChar++
		return true
	case 'k':
		e.cursorLine--
		return true
	case 'j':
		e.cursorLine++
		return true
	case '0':
		e.cursorChar = 0
		return true
	case 'i':
		e.mode = InsertMode
		return true
	case 'A':
		e.cursorChar = len(e.Lines[e.cursorLine])
		e.mode = InsertMode
		return true
	case 'o':
		// Make room for another line
		e.Lines = append(e.Lines, []Char {})

		// Shift lines after cursorLine down
		for i := len(e.Lines) - 2; i > e.cursorLine; i-- {
			Log("%d", i)
			e.Lines[i + 1] = e.Lines[i]
		}

		// Add the new line
		e.Lines[e.cursorLine + 1] = []Char {}

		e.cursorLine++
		e.cursorChar = 0
		e.mode = InsertMode
		return true
	}

	return false
}

func (e *EditBox) fireTextChanged() {
	if e.OnTextChanged != nil {
		e.OnTextChanged(e)
	}
}

func (e *EditBox) insertAt(lineIndex, charIndex int, newText string) {
	line := e.Lines[lineIndex]

	pre  := line[0:charIndex]
	post := line[charIndex:len(line)]

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
		e.Lines[lineIndex] = newLines[0]
		e.fireTextChanged()
		return
	}

	// If lines were added, the whole lines array needs to be shifted down
	oldLinesLen := len(e.Lines)

	// Make room at the end of the array
	for i := 0; i < len(newLines) - 1; i++ {
		Log("newline")
		e.Lines = append(e.Lines, []Char {})
	}

	// Shift post lines to the end of the array to make room for new lines
	shiftDistance := len(newLines) - 1

	for i := oldLinesLen - 1; i >= lineIndex + 1; i-- {
		e.Lines[i + shiftDistance] = e.Lines[i]
	}

	// Copy new lines into place
	for i := 0; i < len(newLines); i++ {
		e.Lines[lineIndex + i] = newLines[i]
	}

	e.fireTextChanged()
}

func (e *EditBox) Insert(newText string) {
	e.insertAt(e.cursorLine, e.cursorChar, newText)
}

func (e *EditBox) handleInsertModeEvent(ev escapebox.Event) bool {
	line := e.Lines[e.cursorLine]

	pre  := line[0:e.cursorChar]
	post := line[e.cursorChar:len(line)]

	preLines := e.Lines[0:e.cursorLine]
	postLines := e.Lines[e.cursorLine + 1:len(e.Lines)]

	if ev.Key == termbox.KeyTab {
		e.Insert("    ")
		e.cursorChar += 4
		return true
	}

	if ev.Key == termbox.KeyEsc {
		e.mode = CommandMode
		e.cursorChar--
		return true
	}

	if renderableChar(ev) {
		e.Insert(string(ev.Ch))
		e.cursorChar++
		return true
	}

	switch (ev.Key) {
	case termbox.KeyArrowLeft:
		e.cursorChar--
		return true

	case termbox.KeyArrowRight:
		e.cursorChar++
		return true

	case termbox.KeyArrowUp:
		e.cursorLine--
		return true

	case termbox.KeyArrowDown:
		e.cursorLine++
		return true

	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if len(pre) > 0 {
			e.Lines[e.cursorLine] = append(pre[0:len(pre) - 1],
						       post...)
			e.cursorChar--
		} else if e.cursorLine > 0 {
			newLines := make([][]Char, len(e.Lines) - 1)
			j := 0

			for i := 0; i < len(preLines); i++ {
				newLines[j] = preLines[i]
				j++
			}

			for i := 0; i < len(postLines); i++ {
				newLines[j] = postLines[i]
				j++
			}

			e.Lines = newLines

			e.cursorLine--
			e.cursorChar = len(e.Lines[e.cursorLine])
			e.Lines[e.cursorLine] = append(e.Lines[e.cursorLine],
						       post...)
		}

		e.fireTextChanged()

		return true

	case termbox.KeyEnter:
		e.Insert("\n")
		e.cursorLine++
		e.cursorChar = 0

		return true

	case termbox.KeySpace:
		e.Insert(" ")
		e.cursorChar++

		return true

	case termbox.KeyTab:
		return true
	}

	return false
}
