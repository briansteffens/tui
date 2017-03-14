package tui

import (
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
}

func (e *EditBox) GetCursor() int {
	ret := 0

	for l := 0; l < e.cursorLine; l++ {
		ret += len(e.Lines[l])
	}

	return ret + e.cursorChar
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

func (e *EditBox) HandleEvent(ev escapebox.Event) {
	if ev.Type != termbox.EventKey {
		return
	}

	oldCursorLine := e.cursorLine
	oldCursorChar := e.cursorChar

	if e.mode == CommandMode {
		e.handleCommandModeEvent(ev)
	} else if e.mode == InsertMode {
		e.handleInsertModeEvent(ev)
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
}

func (e *EditBox) handleCommandModeEvent(ev escapebox.Event) {
	switch ev.Ch {
	case 'h':
		e.cursorChar--
	case 'l':
		e.cursorChar++
	case 'k':
		e.cursorLine--
	case 'j':
		e.cursorLine++
	case '0':
		e.cursorChar = 0
	case 'i':
		e.mode = InsertMode
	case 'A':
		e.cursorChar = len(e.Lines[e.cursorLine])
		e.mode = InsertMode
	}
}

func (e *EditBox) handleInsertModeEvent(ev escapebox.Event) {
	line := e.Lines[e.cursorLine]

	pre  := line[0:e.cursorChar]
	post := line[e.cursorChar:len(line)]

	preLines := e.Lines[0:e.cursorLine]
	postLines := e.Lines[e.cursorLine + 1:len(e.Lines)]

	if ev.Key == termbox.KeyEsc {
		e.mode = CommandMode
		e.cursorChar--
		return
	} else if renderableChar(ev) {
		newLine := make([]Char, len(pre) + len(post) + 1)

		j := 0
		for i := 0; i < len(pre); i++ {
			newLine[j] = pre[i]
			j++
		}

		newLine[j] = Char {
			Char: ev.Ch,
			Fg: termbox.ColorWhite,
			Bg: termbox.ColorBlack,
		}
		j++

		for i := 0; i < len(post); i++ {
			newLine[j] = post[i]
			j++
		}

		e.Lines[e.cursorLine] = newLine

		e.cursorChar++

		if e.OnTextChanged != nil {
			e.OnTextChanged(e)
		}

		return
	}

	switch (ev.Key) {
	case termbox.KeyArrowLeft:
		e.cursorChar--
	case termbox.KeyArrowRight:
		e.cursorChar++
	case termbox.KeyArrowUp:
		e.cursorLine--
	case termbox.KeyArrowDown:
		e.cursorLine++
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

		if e.OnTextChanged != nil {
			e.OnTextChanged(e)
		}
	case termbox.KeyEnter:
		newLines := make([][]Char, len(e.Lines) + 1)
		j := 0

		for i := 0; i < len(preLines); i++ {
			newLines[j] = preLines[i]
			j++
		}

		newLines[j] = pre
		j++

		newLines[j] = post
		j++

		for i := 0; i < len(postLines); i++ {
			newLines[j] = postLines[i]
			j++
		}

		e.Lines = newLines

		e.cursorLine++
		e.cursorChar = 0

		if e.OnTextChanged != nil {
			e.OnTextChanged(e)
		}
	case termbox.KeySpace:
		char := Char {
			Char: ' ',
			Fg: termbox.ColorWhite,
			Bg: termbox.ColorBlack,
		}
		e.Lines[e.cursorLine] = append(pre, char)
		e.Lines[e.cursorLine] = append(e.Lines[e.cursorLine], post...)
		e.cursorChar++

		if e.OnTextChanged != nil {
			e.OnTextChanged(e)
		}
	}
}
