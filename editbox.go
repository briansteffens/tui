package tui

import (
	"os"
	"fmt"
	"errors"
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)

const (
	CommandMode    = 0
	InsertMode     = 1
	VisualLineMode = 2
	tabWidth       = 4
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

func (c *Char) clone() Char {
	return Char {
		Char:    c.Char,
		Fg:      c.Fg,
		Bg:      c.Bg,
		Quote:   c.Quote,
		Escaped: c.Escaped,
	}
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

type TextChangedEvent func(*EditBox)
type CursorMovedEvent func(*EditBox)
type Highlighter      func(*EditBox)

type EditBox struct {
	Bounds        Rect
	Lines         [][]Char
	OnTextChanged TextChangedEvent
	OnCursorMoved CursorMovedEvent
	Highlighter   Highlighter
	Dialect       Dialect

	cursorLine      int
	cursorChar      int
	scroll          int
	focus           bool
	mode            int
	chord           []escapebox.Event
	visualLineStart int
	clipBoard       [][]Char
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
	ret := 0

	for l := 0; l < e.cursorLine; l++ {
		ret += len(e.Lines[l]) + 1
	}

	return ret + e.cursorChar
}

func (e *EditBox) CursorChar() *Char {
	line := e.Lines[e.cursorLine]

	if e.cursorChar == len(line) {
		return &Char { Char: '\n' }
	}

	return &line[e.cursorChar]
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

	e.fireTextChanged()
}

func (e *EditBox) Render() {
	textWidth := e.Bounds.Width
	textHeight := e.Bounds.Height - 1 // Bottom line free for modes/notices

	visualLineStart := e.visualLineStart
	visualLineStop := e.cursorLine
	if visualLineStart > visualLineStop {
		temp := visualLineStart
		visualLineStart = visualLineStop
		visualLineStop = temp
	}

	virtualVisualLineStart := -1
	virtualVisualLineStop := -1

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

		if e.mode == VisualLineMode && lineIndex == visualLineStart {
			virtualVisualLineStart = len(virtualLines)
		}

		for i := 0; i < virtualLineCount; i++ {
			start := i * textWidth
			stop := min(len(line), (i + 1) * textWidth)
			virtualLines = append(virtualLines, line[start:stop])
		}

		if e.mode == VisualLineMode && lineIndex == visualLineStop {
			virtualVisualLineStop = len(virtualLines) - 1
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
			fg := ch.Fg
			bg := ch.Bg

			if l >= virtualVisualLineStart &&
				l <= virtualVisualLineStop {
				bg = termbox.ColorYellow
				fg = termbox.ColorBlack
			}

			termbox.SetCell(e.Bounds.Left + c,
					e.Bounds.Top + l - e.scroll, ch.Char,
					fg, bg)
		}
	}

	if e.focus {
		f, err := os.OpenFile("out", os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}

		defer f.Close()

		s := fmt.Sprintf("%d, %d\n", e.Bounds.Left + cursorCol,
			e.Bounds.Top + cursorRow - e.scroll)

		if _, err = f.WriteString(s); err != nil {
			panic(err)
		}

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

func (e *EditBox) fireTextChanged() {
	if e.Highlighter != nil {
		e.Highlighter(e)
	}

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

func (e *EditBox) Delete() bool {
	line := e.Lines[e.cursorLine]

	// Nothing to delete
	if len(e.Lines) == 1 && len(line) == 0 {
		return false
	}

	// Delete from current line
	if e.cursorChar < len(line) {
		newLine := make([]Char, len(line) - 1)

		j := 0
		for i := 0; i < len(line); i++ {
			if i == e.cursorChar {
				continue
			}

			newLine[j] = line[i]
			j++
		}

		e.Lines[e.cursorLine] = newLine
		e.fireTextChanged()

		return true
	}

	if e.cursorLine == len(e.Lines) - 1 {
		return true
	}

	// Cursor is on the implicit newline at the end of a line and there
	// are more lines after it. Concat this and the next line.
	nextLine := e.Lines[e.cursorLine + 1]
	newLines := make([][]Char, len(e.Lines) - 1)

	j := 0
	for i := 0; i < e.cursorLine; i++ {
		newLines[j] = e.Lines[i]
		j++
	}

	newLine := []Char {}
	newLine = append(newLine, line...)
	newLine = append(newLine, nextLine...)

	newLines[j] = newLine
	j++

	for i := e.cursorLine + 2; i < len(e.Lines); i++ {
		newLines[j] = e.Lines[i]
		j++
	}

	e.Lines = newLines
	e.fireTextChanged()

	return true
}

func removeFromLeft(src []Char, toRemove int) []Char {
	ret := make([]Char, len(src) - toRemove)

	for i := toRemove; i < len(ret); i++ {
		ret[i - toRemove] = src[i]
	}

	return ret
}

func (e *EditBox) shiftTab() {
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
}

func (e *EditBox) CursorNext() bool {
	e.cursorChar++

	if e.cursorChar <= len(e.Lines[e.cursorLine]) {
		return true
	}

	// No more lines, undo the move
	if e.cursorLine == len(e.Lines) - 1 {
		e.cursorChar--
		return false
	}

	// Advance to the next line
	e.cursorLine++
	e.cursorChar = 0
	return true
}

func (e *EditBox) CursorPrevious() bool {
	e.cursorChar--

	if e.cursorChar >= 0 {
		return true
	}

	// No more lines, undo the move
	if e.cursorLine == 0 {
		e.cursorChar++
		return false
	}

	// Move to the previous line
	e.cursorLine--
	e.cursorChar = len(e.Lines[e.cursorLine])
	return true
}

func (e *EditBox) cursorAtBeginning() bool {
	return e.cursorLine == 0 && e.cursorChar == 0
}

func (e *EditBox) cursorAtEnd() bool {
	return e.cursorLine >= len(e.Lines) - 1 &&
	       e.cursorChar >= len(e.Lines[e.cursorLine]) - 1
}

func isDelimiter(c Char) bool {
	return c.Char == ' ' || c.Char == '\t' || c.Char == '\n'
}

func (e *EditBox) nextWord() {
	for !isDelimiter(*e.CursorChar()) {
		e.CursorNext()
	}

	for e.CursorNext() && isDelimiter(*e.CursorChar()) &&
	    e.CursorChar().Char != '\n' {
	}
}

func (e *EditBox) previousWord() {
	if !e.CursorPrevious() {
		return
	}

	// Rewind through delimiters until we reach a normal character or \n
	for isDelimiter(*e.CursorChar()) && e.CursorChar().Char != '\n' {
		if !e.CursorPrevious() {
			return
		}
	}

	if e.CursorChar().Char == '\n' {
		e.CursorPrevious()

		if e.CursorChar().Char == '\n' {
			e.CursorNext()
			return
		}
	}

	// Rewind through normal characters (previous string) until a delimiter
	// is reached
	for !isDelimiter(*e.CursorChar()) {
		if !e.CursorPrevious() {
			return
		}
	}

	if e.cursorAtBeginning() {
		return
	}

	e.CursorNext()
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

	if !handled && ev.Key == termbox.KeyArrowLeft {
		e.cursorChar--
		handled = true
	}

	if !handled && ev.Key == termbox.KeyArrowRight {
		e.cursorChar++
		handled = true
	}

	if !handled && ev.Key == termbox.KeyArrowUp {
		e.cursorLine--
		handled = true
	}

	if !handled && ev.Key == termbox.KeyArrowDown {
		e.cursorLine++
		handled = true
	}

	if !handled && ev.Key == termbox.KeyDelete {
		e.Delete()
		handled = true
	}

	if !handled && (ev.Key == termbox.KeyBackspace ||
			ev.Key == termbox.KeyBackspace2) {
		if e.CursorPrevious() {
			e.Delete()
		}
		handled = true
	}

	if !handled {
		switch e.mode {
		case CommandMode:
			handled = e.handleCommandModeEvent(ev)
			break
		case InsertMode:
			handled = e.handleInsertModeEvent(ev)
			break
		case VisualLineMode:
			handled = e.handleVisualLineModeEvent(ev)
			break
		}
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
	if oldCursorLine != e.cursorLine || oldCursorChar != e.cursorChar {
		e.fireCursorMoved()
	}

	return handled
}

func (e *EditBox) fireCursorMoved() {
	if e.OnCursorMoved != nil {
		e.OnCursorMoved(e)
	}
}

func (e *EditBox) validateLineRange(start, stop int) {
	if start > stop {
		panic("Can't delete this range")
	}

	if start < 0 || stop < 0 ||
	   start >= len(e.Lines) || stop >= len(e.Lines) {
		panic("Line deletion out of range")
	}

	if stop - start + 1 > len(e.Lines) {
		panic("Can't delete more lines than currently exist")
	}
}

func (e *EditBox) deleteLines(start, stop int) {
	e.validateLineRange(start, stop)

	toDelete := stop - start + 1
	newSize := len(e.Lines) - toDelete

	if newSize == 0 {
		e.Lines = make([][]Char, 1)
		e.Lines[0] = []Char{}
		return
	}

	newLines := make([][]Char, len(e.Lines) - toDelete)

	dst := 0
	for src := 0; src < len(e.Lines); src++ {
		if src >= start && src <= stop {
			continue
		}

		newLines[dst] = e.Lines[src]
		dst++
	}

	e.Lines = newLines
	e.fireCursorMoved()
}

func (e *EditBox) copyLinesToClipBoard(start, stop int) {
	e.validateLineRange(start, stop)

	e.clipBoard = [][]Char{}

	for i := start; i <= stop; i++ {
		line := []Char{}

		for _, c := range e.Lines[i] {
			line = append(line, c.clone())
		}

		e.clipBoard = append(e.clipBoard, line)
	}
}

func (e *EditBox) paste() {
	if len(e.clipBoard) == 0 {
		return
	}

	newLines := [][]Char{}

	for i := 0; i <= e.cursorLine; i++ {
		newLines = append(newLines, e.Lines[i])
	}

	for _, l := range e.clipBoard {
		newLines = append(newLines, l)
	}

	for i := e.cursorLine + 1; i < len(e.Lines); i++ {
		newLines = append(newLines, e.Lines[i])
	}

	e.Lines = newLines
	e.cursorLine++
}

func (e *EditBox) handleChord_d() bool {
	if e.chord[1].Ch == 'd' {
		// Delete current line
		e.copyLinesToClipBoard(e.cursorLine, e.cursorLine)
		e.deleteLines(e.cursorLine, e.cursorLine)
	}

	return true
}

func (e *EditBox) handleChord_c() bool {
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

	return true
}

func (e *EditBox) handleChord_g() bool {
	if e.chord[1].Ch == 'g' {
		// Jump to beginning of file
		e.cursorLine = 0
		e.cursorChar = 0
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
	case 'w':
		e.nextWord()
		return true
	case 'b':
		e.previousWord()
		return true
	case 'x':
		e.Delete()
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
	case 'G':
		e.cursorLine = len(e.Lines) - 1
		e.cursorChar = len(e.Lines[e.cursorLine]) - 1
		return true
	case 'V':
		e.mode = VisualLineMode
		e.visualLineStart = e.cursorLine
		return true
	case 'p':
		e.paste()
		return true
	}

	return false
}

func (e *EditBox) handleInsertModeEvent(ev escapebox.Event) bool {
	if ev.Key == termbox.KeyTab {
		e.Insert("    ")
		e.cursorChar += tabWidth
		return true
	}

	if ev.Seq == SeqShiftTab {
		e.shiftTab()
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

func (e *EditBox) handleVisualLineModeEvent(ev escapebox.Event) bool {
	if ev.Key == termbox.KeyEsc {
		e.mode = CommandMode
		return true
	}

	switch ev.Ch {
	case 'k':
		e.cursorLine--
		return true
	case 'j':
		e.cursorLine++
		return true
	case 'd':
		start := e.visualLineStart
		stop := e.cursorLine

		if start > stop {
			temp := start
			start = stop
			stop = temp
		}

		e.copyLinesToClipBoard(start, stop)
		e.deleteLines(start, stop)
		e.mode = CommandMode
		return true
	}

	return false
}

func (e *EditBox) AllChars() []*Char {
	ret := []*Char {}

	for l := 0; l < len(e.Lines); l++ {
		line := &e.Lines[l]

		for c := 0; c < len(*line); c++ {
			ret = append(ret, &(*line)[c])
		}

		ret = append(ret, &Char {
			Char: '\n',
		})
	}

	return ret
}

func BasicHighlighter(e *EditBox) {
	delimiters := []rune { ' ', '\n', '(', ')', ',', ';' }

	var cur, next *Char
	var quote rune
	var quoteStartIndex int

	word := ""

	chars := e.AllChars()

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
