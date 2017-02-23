package main

import "github.com/nsf/termbox-go"
import "fmt"

func setCell(x, y int, r rune) {
    termbox.SetCell(x, y, r, termbox.ColorWhite, termbox.ColorBlack)
}

func termPrint(x, y int, msg string) {
	for _, c := range msg {
        setCell(x, y, c)
		x++
	}
}

func termPrintf(x, y int, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	termPrint(x, y, s)
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
    handleEvent(termbox.Event)
}

type container struct {
    controls []control
    focused  focusable
}

type label struct {
    bounds rect
    text   string
}

func (l* label) render() {
    termPrint(l.bounds.left, l.bounds.top, l.text)
}

type textbox struct {
    bounds rect
    value  string
    cursor int
    scroll int
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
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
    termPrint(t.bounds.left + 1, t.bounds.top + 1,
              t.value[t.scroll:t.lastVisible() + 1])
    termbox.SetCursor(t.bounds.left + 1 + t.cursor - t.scroll,
                      t.bounds.top + 1)

    termPrintf(1, 10, "value: %s", t.value)
    termPrintf(1, 11, "cursor: %d", t.cursor)
    termPrintf(1, 12, "scroll: %d", t.scroll)
    termPrintf(1, 13, "maxVisibleChars(): %d", t.maxVisibleChars())
    termPrintf(1, 14, "lastVisible(): %d", t.lastVisible())
}

func (t* textbox) setFocus() {
}

func renderableChar(k termbox.Key) bool {
    return k != termbox.KeyEnter  &&
           k != termbox.KeyPgup   &&
           k != termbox.KeyPgdn   &&
           k != termbox.KeyInsert
}

func (t* textbox) handleEvent(ev termbox.Event) {
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
    //lastVisible := t.lastVisible()
    //if t.cursor > lastVisible {
    //    t.scroll = t.cursor - lastVisible
    //}
}

func refresh(c container) {
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

    for _, v := range c.controls {
        v.render()
    }

    termbox.Flush()
    termbox.Sync()
}

func main() {
    err := termbox.Init()
    if err != nil {
        panic(err)
    }
    defer termbox.Close()

    l := label {
        bounds: rect { left: 5, top: 1, width: 20, height: 1 },
        text: "A textbox:",
    }

    t := textbox {
        bounds: rect { left: 5, top: 2, width: 5, height: 3 },
        value: "12",
        cursor: 2,
        scroll: 0,
    }

    c := container {
        controls: []control {&l, &t},
        focused: &t,
    }

    termbox.SetInputMode(termbox.InputEsc) // | termbox.InputMouse)
    refresh(c)

    loop: for {
        ev := termbox.PollEvent()

        handled := false

        switch ev.Type {
        case termbox.EventKey:
            switch ev.Key {
            case termbox.KeyCtrlC:
                break loop
            case termbox.KeyTab:
                handled = true
            }
        }

        if !handled {
            c.focused.handleEvent(ev)
        }

        refresh(c)
    }
}
