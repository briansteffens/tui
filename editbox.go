package tui

import (
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)

const (
	CommandMode = 0
	InsertMode  = 1
	tabWidth    = 4
)

type CharClass int

const (
	ClassNormal     = 0
	ClassWhiteSpace = 1
	ClassSymbol     = 2
)

const QuoteNone   rune = 0
const QuoteSingle rune = '\''
const QuoteDouble rune = '"'

type Char struct {
	Char    rune
	Fg      termbox.Attribute
	Bg      termbox.Attribute

	// Highlighter data
	Quote   rune
	Escaped bool
}

const colorKeyword termbox.Attribute = termbox.ColorBlue
const colorType    termbox.Attribute = termbox.ColorRed

type Token int

const (
	TokenNone    Token = 0
	TokenKeyword Token = 1
	TokenType    Token = 2
)

type Dialect func(string) Token

type DataChangedEvent func(TextBackend)

type TextChangedEvent func(*EditBox)
type CursorMovedEvent func(*EditBox)
type Highlighter      func(*EditBox)

type TextBackend interface {
	GetText() string
	SetText(raw string)
	GetChar(index int) (*Char, error)
	GetCursor() int
	Insert(text string)
	Delete()
	CursorNext() bool
	CursorPrevious() bool
	CursorBeginning()
	CursorAtEnd() bool
	CursorBeginningOfLine()
	CursorEndOfLine()
	LinePrevious()
	LineNext()
	ClampCursor()
	LineCount() int
	AllChars() []*Char
	SetDataChangedEvent(DataChangedEvent)
}

type EditBox struct {
	Bounds        Rect
	OnTextChanged TextChangedEvent
	OnCursorMoved CursorMovedEvent
	Highlighter   Highlighter
	Dialect       Dialect

	Data          TextBackend
	scroll        int
	focus         bool
	mode          int
	chord         []escapebox.Event
}

var whitespace []rune = []rune { ' ', '\t' }
var symbols []rune = []rune { '!', '@', '#', '$', '%', '^', '*', '(', ')' }

func isRune(r rune, in []rune) bool {
	for _, i := range in {
		if r == i {
			return true
		}
	}

	return false
}

func getCharClass(r rune) CharClass {
	if isRune(r, symbols) {
		return ClassSymbol
	}

	if isRune(r, whitespace) {
		return ClassWhiteSpace
	}

	return ClassNormal
}

func (e *EditBox) GetCursor() int {
	return e.Data.GetCursor()
}

func (e *EditBox) CursorChar() *Char {
	ret, err := e.Data.GetChar(e.GetCursor())

	if err != nil {
		panic(err)
	}

	return ret
}

func (e *EditBox) GetText() string {
	return e.Data.GetText()
}

func (e *EditBox) SetText(raw string) {
	e.Data.SetText(raw)
	e.fireTextChanged()
}

func (e *EditBox) Render() {
/*
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
*/
}

func (e *EditBox) SetFocus() {
	e.focus = true
}

func (e *EditBox) UnsetFocus() {
	e.focus = false
}

func (e *EditBox) fireTextChanged() {
	if e.Highlighter != nil {
		e.Highlighter(e)
	}

	if e.OnTextChanged != nil {
		e.OnTextChanged(e)
	}
}

func (e *EditBox) Insert(newText string) {
	e.Data.Insert(newText)
	e.fireTextChanged()
}

func (e *EditBox) Delete() {
	e.Data.Delete()
	e.fireTextChanged()
}

func removeFromLeft(src []Char, toRemove int) []Char {
	ret := make([]Char, len(src) - toRemove)

	for i := toRemove; i < len(ret); i++ {
		ret[i - toRemove] = src[i]
	}

	return ret
}

func (e *EditBox) shiftTab() {
    /*
	line := e.Lines[e.cursorLine]

	if len(line) == 0 {
		return
	}

	if line[0].Char == '\t' {
		e.Lines[e.cursorLine] = removeFromLeft(line, 1)
		e.cursorChar--
		return
	}

	toRemove := 0

	for i := 0; i < len(line) && i < 4; i++ {
		toRemove++
	}

	if toRemove == 0 {
		return
	}

	e.Lines[e.cursorLine] = removeFromLeft(line, toRemove)
	e.cursorChar -= toRemove
	*/
}

func (e *EditBox) cursorAtBeginning() bool {
	return e.Data.GetCursor() == 0
}

func isDelimiter(c Char) bool {
	return c.Char == ' ' || c.Char == '\t' || c.Char == '\n'
}

func (e *EditBox) nextWord() {
	for !isDelimiter(*e.CursorChar()) {
		e.Data.CursorNext()
	}

	for e.Data.CursorNext() && isDelimiter(*e.CursorChar()) &&
	    e.CursorChar().Char != '\n' {
	}
}

func (e *EditBox) previousWord() {
	if !e.Data.CursorPrevious() {
		return
	}

	// Rewind through delimiters until we reach a normal character or \n
	for isDelimiter(*e.CursorChar()) && e.CursorChar().Char != '\n' {
		if !e.Data.CursorPrevious() {
			return
		}
	}

	if e.CursorChar().Char == '\n' {
		e.Data.CursorPrevious()

		if e.CursorChar().Char == '\n' {
			e.Data.CursorNext()
			return
		}
	}

	// Rewind through normal characters (previous string) until a delimiter
	// is reached
	for !isDelimiter(*e.CursorChar()) {
		if !e.Data.CursorPrevious() {
			return
		}
	}

	if e.cursorAtBeginning() {
		return
	}

	e.Data.CursorNext()
}

func (e *EditBox) HandleEvent(ev escapebox.Event) bool {
	if ev.Type != termbox.EventKey {
		return false
	}

	oldCursor := e.Data.GetCursor()

	handled := false

	if !handled && ev.Key == termbox.KeyHome {
		e.Data.CursorBeginningOfLine()
		handled = true
	}

	if !handled && ev.Key == termbox.KeyEnd {
		e.Data.CursorEndOfLine()
		handled = true
	}

	if !handled && ev.Key == termbox.KeyArrowLeft {
		e.Data.CursorPrevious()
		handled = true
	}

	if !handled && ev.Key == termbox.KeyArrowRight {
		e.Data.CursorNext()
		handled = true
	}

	if !handled && ev.Key == termbox.KeyArrowUp {
		e.Data.LinePrevious()
		handled = true
	}

	if !handled && ev.Key == termbox.KeyArrowDown {
		e.Data.LineNext()
		handled = true
	}

	if !handled && ev.Key == termbox.KeyDelete {
		e.Data.Delete()
		handled = true
	}

	if !handled && (ev.Key == termbox.KeyBackspace ||
			ev.Key == termbox.KeyBackspace2) {
		if e.Data.CursorPrevious() {
			e.Data.Delete()
		}
		handled = true
	}

	if !handled && e.mode == CommandMode {
		handled = e.handleCommandModeEvent(ev)
	} else if !handled && e.mode == InsertMode {
		handled = e.handleInsertModeEvent(ev)
	}

	e.Data.ClampCursor()

	// Detect and fire OnCursorMoved
	if oldCursor != e.Data.GetCursor() {
		e.fireCursorMoved()
	}

	return handled
}

func (e *EditBox) fireCursorMoved() {
	if e.OnCursorMoved != nil {
		e.OnCursorMoved(e)
	}
}

func (e *EditBox) handleChord_d() bool {
/*
	if e.chord[1].Ch == 'd' {
		// Delete current line
		if len(e.Lines) == 0 {
			return true
		}

		if len(e.Lines) == 1 {
			e.Lines[0] = []Char {}
			e.fireCursorMoved()
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

		e.fireCursorMoved()
	}
*/
	return true
}

func (e *EditBox) handleChord_c() bool {
    /*
	if e.chord[1].Ch == 'w' {
		// Delete current word
		deleteClass := getCharClass(e.CursorChar().Char)

		for {
			e.Delete()

			if getCharClass(e.CursorChar().Char) != deleteClass {
				break
			}

			if e.cursorAtEnd() {
				e.Delete()
				break
			}
		}

		e.mode = InsertMode
	}
	*/

	return true
}

func (e *EditBox) handleChord_g() bool {
	if e.chord[1].Ch == 'g' {
		e.Data.CursorBeginning()
	}

	return true
}

func (e *EditBox) handleCommandModeEvent(ev escapebox.Event) bool {
	// Start a chord
	if len(e.chord) == 0 &&
	   (ev.Ch == 'd' || ev.Ch == 'c' || ev.Ch == 'g') {
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
		case 'c':
			consumed = e.handleChord_c()
		case 'g':
			consumed = e.handleChord_g()
		}

		// Chord consumed
		if consumed {
			e.chord = []escapebox.Event {}
		}

		return true
	}

	switch ev.Ch {
	case 'h':
		e.Data.CursorPrevious()
		return true
	case 'l':
		e.Data.CursorNext()
		return true
	case 'k':
		e.Data.LinePrevious()
		return true
	case 'j':
		e.Data.LineNext()
		return true
	case '0':
		e.Data.CursorBeginningOfLine()
		return true
	case 'i':
		e.mode = InsertMode
		return true
	case 'A':
		e.Data.CursorEndOfLine()
		e.mode = InsertMode
		return true
	case 'w':
		e.nextWord()
		return true
	case 'b':
		e.previousWord()
		return true
	case 'x':
		e.Data.Delete()
		return true
	case 'o':
		// Make room for another line
		/*
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
		*/
		return true
	case 'G':
	    /*
		e.cursorLine = len(e.Lines) - 1
		e.cursorChar = len(e.Lines[e.cursorLine]) - 1
		*/
		return true
	}

	return false
}

func (e *EditBox) handleInsertModeEvent(ev escapebox.Event) bool {
	if ev.Key == termbox.KeyTab {
		e.Insert("    ")
		//e.cursorChar += tabWidth
		return true
	}

	if ev.Seq == SeqShiftTab {
		e.shiftTab()
		return true
	}

	if ev.Key == termbox.KeyEsc {
		e.mode = CommandMode
		e.Data.CursorPrevious()
		return true
	}

	if renderableChar(ev) {
		e.Insert(string(ev.Ch))
		e.Data.CursorNext()
		return true
	}

	switch (ev.Key) {
	case termbox.KeyEnter:
		e.Insert("\n")
		e.Data.LineNext()
		e.Data.CursorBeginningOfLine()

		return true

	case termbox.KeySpace:
		e.Insert(" ")
		e.Data.CursorNext()

		return true

	case termbox.KeyTab:
		return true
	}

	return false
}

func BasicHighlighter(e *EditBox) {
	delimiters := []rune { ' ', '\n', '(', ')', ',', ';' }

	var cur, next *Char
	var quote rune
	var quoteStartIndex int

	word := ""

	chars := e.Data.AllChars()

	// Loop over all chars plus one. i is always the index of 'next' so
	// the loop is basically running one char ahead. Run one extra time
	// to process the last character, which at that point will be in cur.
	for i := 0; i <= len(chars); i++ {
		cur = next

		if i < len(chars) {
			next = chars[i]
		} else {
			next = nil
		}

		// Skip first iteration because cur won't be set yet.
		if cur == nil {
			continue
		}

		// Is the next character:
		//   - Preceded by a slash
		nextSlashEscaped := next != nil && cur.Char == '\\'

		// Is the next character:
		//   - A quote char of the same type as the quote it's inside
		//   - Preceded by another of the same quote char type
		//   - Not the second character in a quote
		nextDoubleEscaped := next != nil && next.Char == quote &&
				     cur.Char == quote && quoteStartIndex < i

		// Is the next character:
		//   - Either slash- or double-escaped
		//   - Not preceded by another escaped character
		if next != nil {
			next.Escaped = !cur.Escaped &&
				       (nextSlashEscaped || nextDoubleEscaped)
		}

		// Is the current character:
		//   - A quote char
		//   - Not escaped
		//   - Not the first in a double-escaped sequence ('' or "")
		isCurQuote := !cur.Escaped && !nextDoubleEscaped &&
			      (cur.Char == QuoteSingle ||
			       cur.Char == QuoteDouble)

		quoteToggledThisLoop := false

		// Start of a quote
		if isCurQuote && quote == QuoteNone {
			quote = cur.Char
			quoteToggledThisLoop = true
			quoteStartIndex = i
		}

		cur.Quote = quote

		// Check for word delimiter
		isDelimiter := isRune(cur.Char, delimiters)

		// Reset word if we hit a delimiter or EOF
		if isDelimiter || next == nil {
			tokenType := TokenNone

			if e.Dialect != nil {
				tokenType = e.Dialect(word)
			}

			wordColor := termbox.ColorWhite

			switch (tokenType) {
			case TokenKeyword:
				wordColor = colorKeyword
			case TokenType:
				wordColor = colorType
			}

			if wordColor != termbox.ColorWhite {
				for j := i - 1; j >= i - len(word) - 1; j-- {
					chars[j].Fg = wordColor
				}
			}

			word = ""
		} else {
			word += string(cur.Char)
		}

		// Color quotes
		if quote != QuoteNone {
			cur.Fg = termbox.ColorGreen
		} else {
			cur.Fg = termbox.ColorWhite
		}

		// End quote
		if isCurQuote && quote != QuoteNone && !quoteToggledThisLoop &&
		   quote == cur.Char {
			quote = QuoteNone
		}
	}
}
