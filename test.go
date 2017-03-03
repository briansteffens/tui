package main

import "time"
import "github.com/nsf/termbox-go"
import "os"
import "fmt"

func pollTermboxEvents(c chan termbox.Event) {
    for {
        ev := termbox.PollEvent()
        c <- ev
    }
}

func fileWrite(f *os.File, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
    _, err := f.WriteString(s)
    if err != nil {
        panic(err)
    }
}

func formatEvent(ev event) string {
    return fmt.Sprintf("%d %d", ev.Key, ev.Ch)
}

const (
    SeqNone = 0
    SeqShiftTab = 1
)

type event struct {
    termbox.Event
    Seq int
}

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

func pollEvents(c chan event, outFile *os.File) {
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
            fileWrite(outFile, "Received: %s\n", formatEvent(ev))

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
            fileWrite(outFile, "Timer triggered\n")
            inEscapeSequence = false

            // Flush buffer
            for i := 0; i < sequenceBufferLen; i++ {
                c <- sequenceBuffer[i]
            }
        }
    }
}

func main() {
    err := termbox.Init()
    if err != nil {
        panic(err)
    }
    defer termbox.Close()

    outFile, err := os.Create("out")
    if err != nil {
        panic(err)
    }
    defer outFile.Close()

    events := make(chan event)
    go pollEvents(events, outFile)

    loop: for {
        ev, ok := <-events
        if !ok {
            break
        }

        fileWrite(outFile, "Handled: %s\n", formatEvent(ev))

        if ev.Type == termbox.EventKey && ev.Key == termbox.KeyCtrlC {
            break loop
        }
    }

/*
    termbox.SetInputMode(termbox.InputEsc) // | termbox.InputMouse)

    eventChannel := make(chan termbox.Event)
    go pollEvents(eventChannel)

    nothingCount := 0

    loop: for {
        select {
        case ev := <-eventChannel:
            if ev.Type == termbox.EventKey && ev.Key == termbox.KeyCtrlC {
                break loop
            }
            termPrintf(1, 1, "%d %d", ev.Key, ev.Ch)
        default:
            termPrintf(1, 2, "Nothing %d", nothingCount)
            nothingCount = nothingCount + 1
        }
        termbox.Flush()
        termbox.Sync()
    }
    */
}
/*
    loop: for {
        ev := termbox.PollEvent()
        l.text = fmt.Sprintf("%s%d %d %d,", l.text, ev.Mod, ev.Key, ev.Ch)

        handled := false

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

        if ev.Ch == 90 {
            c.focusPrevious()
            handled = true
        }

        if !handled && c.focused != nil {
            c.focused.handleEvent(ev)
        }

        refresh(c)
    }

    /*
    err := termbox.Init()
    if err != nil {
        panic(err)
    }
    defer termbox.Close()

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

    termbox.SetInputMode(termbox.InputEsc) // | termbox.InputMouse)
    refresh(c)

    loop: for {
        ev := termbox.PollEvent()
        l.text = fmt.Sprintf("%s%d %d %d,", l.text, ev.Mod, ev.Key, ev.Ch)

        handled := false

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

        if ev.Ch == 90 {
            c.focusPrevious()
            handled = true
        }

        if !handled && c.focused != nil {
            c.focused.handleEvent(ev)
        }

        refresh(c)
    }
    */
