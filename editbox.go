package tui

import (
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)

const (
	CommandMode = 0
	InsertMode  = 1
)

type Editbox struct {
	Bounds     Rect
	Lines      []string
	cursorLine int
	cursorChar int
	scroll     int
	focus      bool
	mode       int
}

func (e *Editbox) Render() {
	textWidth := e.Bounds.Width - 2
	textHeight := e.Bounds.Height - 3

	// Generate virtual lines and map the cursor to them.
	virtualLines := make([]string, 0)

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

	for i := e.scroll; i < scrollEnd; i++ {
		termPrintf(e.Bounds.Left + 1, e.Bounds.Top + 1 + i - e.scroll,
			   virtualLines[i])
	}

	RenderBorder(e.Bounds)

	if e.focus {
		termbox.SetCursor(e.Bounds.Left + 1 + cursorCol,
				  e.Bounds.Top  + 1 + cursorRow - e.scroll)
	}

	if e.mode == InsertMode {
		termPrintf(e.Bounds.Left + 1, e.Bounds.Bottom() - 1,
			   "-- INSERT --")
	}
}

func (e *Editbox) SetFocus() {
	e.focus = true
}

func (e *Editbox) UnsetFocus() {
	e.focus = false
}

func (e *Editbox) HandleEvent(ev escapebox.Event) {
	if ev.Type != termbox.EventKey {
		return
	}

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
}

func (e *Editbox) handleCommandModeEvent(ev escapebox.Event) {
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

func (e *Editbox) handleInsertModeEvent(ev escapebox.Event) {
	line := e.Lines[e.cursorLine]

	pre  := line[0:e.cursorChar]
	post := line[e.cursorChar:len(line)]

	preLines := e.Lines[0:e.cursorLine]
	postLines := e.Lines[e.cursorLine + 1:len(e.Lines)]

	if ev.Key == termbox.KeyEsc {
		e.mode = CommandMode
		e.cursorChar--
		return
	} else if renderableChar(ev.Key) {
		e.Lines[e.cursorLine] = pre + string(ev.Ch) + post
		e.cursorChar++
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
			e.Lines[e.cursorLine] = pre[0:len(pre) - 1] + post
			e.cursorChar--
		} else if e.cursorLine > 0 {
			newLines := make([]string, len(e.Lines) - 1)
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
			e.Lines[e.cursorLine] += post
		}
	case termbox.KeyEnter:
		newLines := make([]string, len(e.Lines) + 1)
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
	}
}
