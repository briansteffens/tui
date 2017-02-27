package main

import "github.com/nsf/termbox-go"
import "fmt"
import "time"

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func setCell(x, y int, r rune) {
    termbox.SetCell(x, y, r, termbox.ColorWhite, termbox.ColorBlack)
}

func termPrintf(x, y int, format string, args ...interface{}) {
    s := fmt.Sprintf(format, args...)
    for i, c := range s {
        setCell(x + i, y, c)
    }
}

func renderBorder(r rect) {
    // Corners
    setCell(r.left, r.top, '+')
    setCell(r.right(), r.top, '+')
    setCell(r.left, r.bottom(), '+')
    setCell(r.right(), r.bottom(), '+')

    // Horizontal borders
    for x := r.left + 1; x < r.right(); x++ {
        setCell(x, r.top, '-')
        setCell(x, r.bottom(), '-')
	}

    // Vertical borders
    for y := r.top + 1; y < r.bottom(); y++ {
        setCell(r.left, y, '|')
        setCell(r.right(), y, '|')
    }
}

type rect struct {
    left, top, width, height int
}

func (r* rect) right() int {
    return r.left + r.width - 1
}

func (r* rect) bottom() int {
    return r.top + r.height - 1
}

type control interface {
    render()
}

type focusable interface {
    control
    setFocus()
    unsetFocus()
    handleEvent(event)
}

// label ----------------------------------------------------------------------

type label struct {
    bounds rect
    text   string
}

func (l* label) render() {
    termPrintf(l.bounds.left, l.bounds.top, l.text)
}

// textbox --------------------------------------------------------------------

type textbox struct {
    bounds rect
    value  string
    cursor int
    scroll int
    focus  bool
}

func renderableChar(k termbox.Key) bool {
    return k != termbox.KeyEnter  &&
           k != termbox.KeyPgup   &&
           k != termbox.KeyPgdn   &&
           k != termbox.KeyInsert
}

func (t* textbox) maxVisibleChars() int {
    return t.bounds.width - 2
}

func (t* textbox) visibleChars() int {
    return min(t.maxVisibleChars(), len(t.value) - t.scroll)
}

func (t* textbox) lastVisible() int {
    return t.scroll + t.visibleChars() - 1
}

func (t* textbox) render() {
    renderBorder(t.bounds)
    termPrintf(t.bounds.left + 1, t.bounds.top + 1,
               t.value[t.scroll:t.lastVisible() + 1])

    if t.focus {
        termbox.SetCursor(t.bounds.left + 1 + t.cursor - t.scroll,
                          t.bounds.top + 1)
    }
}

func (t* textbox) setFocus() {
    t.focus = true
}

func (t* textbox) unsetFocus() {
    t.focus = false
}

func (t* textbox) handleEvent(ev event) {
    pre := t.value[0:t.cursor]
    post := t.value[t.cursor:len(t.value)]

    switch ev.Type {
    case termbox.EventKey:
        char := fmt.Sprintf("%c", ev.Ch)

        switch ev.Key {
        case termbox.KeyBackspace, termbox.KeyBackspace2:
            if len(pre) > 0 {
                t.value = pre[0:len(pre)-1] + post
                t.cursor--
            }
        case termbox.KeyDelete:
            if len(post) > 0 {
                t.value = pre + post[1:len(post)]
            }
        case termbox.KeyArrowLeft:
            t.cursor--
        case termbox.KeyArrowRight:
            t.cursor++
        case termbox.KeyHome:
            t.cursor = 0
        case termbox.KeyEnd:
            t.cursor = len(t.value)
        default:
            if renderableChar(ev.Key) {
                t.value = pre + char + post
                t.cursor++
            }
        }
    }

    if t.cursor < 0 {
        t.cursor = 0
    }

    if t.cursor > len(t.value) {
        t.cursor = len(t.value)
    }

    if t.cursor < t.scroll {
        t.scroll = t.cursor
    }

    if t.cursor >= t.scroll + t.maxVisibleChars() {
        t.scroll = t.cursor - t.maxVisibleChars() + 1
    }
}

// checkbox -------------------------------------------------------------------

type checkbox struct {
    bounds  rect
    text    string
    checked bool
    focus   bool
}

func (c* checkbox) render() {
    checkContent := " "

    if c.checked {
        checkContent = "X"
    }

	s := fmt.Sprintf("[%s] %s", checkContent, c.text)

    count := min(len(s), c.bounds.width)
    termPrintf(c.bounds.left, c.bounds.top, s[0:count])

    if c.focus {
        termbox.SetCursor(c.bounds.left + 1, c.bounds.top)
    }
}

func (c* checkbox) setFocus() {
    c.focus = true
}

func (c* checkbox) unsetFocus() {
    c.focus = false
}

func (c* checkbox) handleEvent(ev event) {
    switch ev.Type {
    case termbox.EventKey:
        switch ev.Key {
        case termbox.KeySpace:
            c.checked = !c.checked
        }
    }
}

// button ---------------------------------------------------------------------

type ButtonClickEvent func(*button)

type button struct {
    bounds       rect
    text         string
    focus        bool
    clickHandler ButtonClickEvent
}

func (b* button) render() {
    renderBorder(b.bounds)

    count := min(len(b.text), b.bounds.width - 4)
    termPrintf(b.bounds.left + 2, b.bounds.top + 1, b.text[0:count])

    if b.focus {
        termbox.SetCursor(b.bounds.left + 1, b.bounds.top + 1)
    }
}

func (b* button) setFocus() {
    b.focus = true
}

func (b* button) unsetFocus() {
    b.focus = false
}

func (b* button) handleEvent(ev event) {
    switch ev.Type {
    case termbox.EventKey:
        switch ev.Key {
        case termbox.KeyEnter, termbox.KeySpace:
            if b.clickHandler != nil {
                b.clickHandler(b)
            }
        }
    }
}

// container ------------------------------------------------------------------

type container struct {
    controls []control
    focused  focusable
}

func (c* container) focus(f focusable) {
    if c.focused != nil {
        c.focused.unsetFocus()
    }

    c.focused = f
    c.focused.setFocus()
}

func (c* container) focusNext() {
    currentIndex := 0

    // Find index of currently focused control
    if c.focused != nil {
        for index, ctrl := range c.controls {
            if ctrl == c.focused {
                currentIndex = index
                break
            }
        }
    }

    // Scan list after focused control for another focusable control
    for i := currentIndex + 1; i < len(c.controls); i++ {
        f, ok := c.controls[i].(focusable)
        if ok {
            c.focus(f)
            return
        }
    }

    // Scan list before focused control (loop around)
    for i := 0; i <= currentIndex; i++ {
        f, ok := c.controls[i].(focusable)
        if ok {
            c.focus(f)
            return
        }
    }
}

func (c* container) focusPrevious() {
    currentIndex := 0

    // Find index of currently focused control
    if c.focused != nil {
        for index, ctrl := range c.controls {
            if ctrl == c.focused {
                currentIndex = index
                break
            }
        }
    }

    // Scan list before focused control for another focusable control
    for i := currentIndex - 1; i >= 0; i-- {
        f, ok := c.controls[i].(focusable)
        if ok {
            c.focus(f)
            return
        }
    }

    // Scan list after focused control (loop around)
    for i := len(c.controls) - 1; i >= currentIndex; i-- {
        f, ok := c.controls[i].(focusable)
        if ok {
            c.focus(f)
            return
        }
    }
}

func refresh(c container) {
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

    for _, v := range c.controls {
        v.render()
    }

    termbox.Flush()
    termbox.Sync()
}

// Non-standard escape sequences
const (
    SeqNone     = 0
    SeqShiftTab = 1
)

type event struct {
    termbox.Event
    Seq int
}

// Convert termbox.Event to event
func makeEvent(e termbox.Event) event {
    var ret event

    ret.Type   = e.Type
    ret.Mod    = e.Mod
    ret.Key    = e.Key
    ret.Ch     = e.Ch
    ret.Width  = e.Width
    ret.Height = e.Height
    ret.Err    = e.Err
    ret.MouseX = e.MouseX
    ret.MouseY = e.MouseY
    ret.N      = e.N
    ret.Seq    = SeqNone

    return ret
}

// Check a list of termbox events to see if they match any known non-standard
// escape sequences
func detectSequence(events []event) event {
    var ret event

    if len(events) == 3 &&
       events[0].Type == termbox.EventKey &&
       events[0].Key == termbox.KeyEsc &&
       events[1].Type == termbox.EventKey &&
       events[1].Key == 0 &&
       events[1].Ch == 91 &&
       events[2].Type == termbox.EventKey &&
       events[2].Key == 0 &&
       events[2].Ch == 90 {
        ret.Seq = SeqShiftTab
    }

    return ret
}

// Channel-ize termbox.PollEvent
func pollTermboxEvents(c chan termbox.Event) {
    for {
        c <- termbox.PollEvent()
    }
}

// Read termbox events and try to detect non-standard escape sequences
func pollEvents(c chan event) {
    termboxEvents := make(chan termbox.Event)
    go pollTermboxEvents(termboxEvents)

    escapeSequenceMaxDuration := time.Millisecond

    // TODO: some kind of nil timer to start?
    escapeSequenceTimer := time.NewTimer(escapeSequenceMaxDuration)
    <-escapeSequenceTimer.C

    inEscapeSequence := false

    var sequenceBuffer [10]event
    sequenceBufferLen := 0

    for {
        select {
        case e := <-termboxEvents:
            ev := makeEvent(e)

            if ev.Type == termbox.EventKey && ev.Key == termbox.KeyEsc {
                // If already in escape sequence and we see another escape key,
                // flush the existing buffer and start a new escape sequence.
                if inEscapeSequence {
                    // Flush buffer
                    for i := 0; i < sequenceBufferLen; i++ {
                        c <- sequenceBuffer[i]
                    }
                }

                escapeSequenceTimer.Reset(escapeSequenceMaxDuration)
                inEscapeSequence = true
                sequenceBufferLen = 0
            }

            if inEscapeSequence {
                sequenceBuffer[sequenceBufferLen] = ev
                sequenceBufferLen++

                seq := detectSequence(sequenceBuffer[0:sequenceBufferLen])

                if seq.Seq != SeqNone {
                    // If an escape sequence was detected, return it and stop
                    // the timer.
                    c <- seq
                    sequenceBufferLen = 0
                    escapeSequenceTimer.Stop()
                    inEscapeSequence = false
                }

                break
            }

            // Not in possible escape sequence: handle event immediately.
            c <- ev

        case <-escapeSequenceTimer.C:
            // Escape sequence timeout reached. Assume no escape sequence is
            // coming. Flush buffer.
            inEscapeSequence = false

            // Flush buffer
            for i := 0; i < sequenceBufferLen; i++ {
                c <- sequenceBuffer[i]
            }
        }
    }
}

func buttonClickHandler(b *button) {
    panic("clicked!")
}

func main() {
    err := termbox.Init()
    if err != nil {
        panic(err)
    }
    defer termbox.Close()

    termbox.SetInputMode(termbox.InputEsc) // | termbox.InputMouse)

    l := label {
        bounds: rect { left: 5, top: 1, width: 20, height: 1 },
        text: "",
    }

    t := textbox {
        bounds: rect { left: 5, top: 2, width: 5, height: 3 },
        value: "12",
        cursor: 2,
        scroll: 0,
    }

    t2 := textbox {
        bounds: rect { left: 5, top: 7, width: 15, height: 3},
        value: "Greetings!",
        cursor: 0,
        scroll: 0,
    }

    checkbox1 := checkbox {
        bounds: rect { left: 5, top: 11, width: 30, height: 1},
        text: "Enable the whateverthing",
    }

    button1 := button {
        bounds: rect { left: 5, top: 15, width: 10, height: 3},
        text: "Continue!",
        clickHandler: buttonClickHandler,
    }

    c := container {
        controls: []control {&l, &t, &t2, &checkbox1, &button1},
    }

    c.focusNext()
    refresh(c)

    events := make(chan event)
    go pollEvents(events)

    loop: for {
        ev := <-events

        handled := false

        switch ev.Seq {
        case SeqNone:
            switch ev.Type {
            case termbox.EventKey:
                switch ev.Key {
                case termbox.KeyCtrlA:
                    l.text = ""
                case termbox.KeyCtrlC:
                    break loop
                case termbox.KeyTab:
                    c.focusNext()
                    handled = true
                }
            }
        case SeqShiftTab:
            c.focusPrevious()
            handled = true
        }

        if !handled && c.focused != nil {
            c.focused.handleEvent(ev)
        }

        refresh(c)
    }
}
