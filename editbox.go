package main

import (
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)

type Editbox struct {
	Bounds     Rect
	lines      []string
	cursorLine int
	cursorChar int
	scroll     int
	focus      bool
}

func splitRows(line string, textWidth int) []string {
	rows := len(line) / textWidth + 1
	ret := make([]string, rows)

	for i := 0; i < rows; i++ {
		start := i * textWidth
		stop := min((i + 1) * textWidth, len(line))
		ret[i] = line[start:stop]
	}

	return ret
}

func (e *Editbox) Render() {
	textWidth := e.Bounds.Width - 2
	textHeight := e.Bounds.Height - 2

	// Generate virtual lines and map the cursor to them.
	virtualLines := make([]string, 0)

	cursorRow := 0
	cursorCol := 0

	for lineIndex, line := range e.lines {
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

	scrollEnd := e.scroll + textHeight

	for i := e.scroll; i < scrollEnd; i++ {
		termPrintf(e.Bounds.Left + 1, e.Bounds.Top + 1 + i - e.scroll,
			   virtualLines[i])
	}

	RenderBorder(e.Bounds)

	if e.focus {
		termbox.SetCursor(e.Bounds.Left + 1 + cursorCol,
				  e.Bounds.Top  + 1 + cursorRow - e.scroll)
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

	switch ev.Ch {
	case 'h', 'H':
		e.cursorChar = max(0, e.cursorChar - 1)
	case 'l', 'L':
		e.cursorChar++
	case 'k', 'K':
		e.cursorLine = max(0, e.cursorLine - 1)

	case 'j', 'J':
		e.cursorLine = min(len(e.lines) - 1, e.cursorLine + 1)
	case '0':
		e.cursorChar = 0
	}

	e.cursorChar = min(len(e.lines[e.cursorLine]) - 1, e.cursorChar)
}
