package main

import (
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)

type Column struct {
	Name  string
	Width int
}

type Detailview struct {
	Bounds     Rect
	focus      bool
	scroll     int
	cursorCol  int
	cursorRow  int
	Columns    []Column
	Rows	   [][]string
}

func renderValue(src string, maxWidth int) string {
	maxLen := min(maxWidth, len(src))
	return src[0:maxLen]
}

func (d *Detailview) heightForRows() int {
	return d.Bounds.Height - 3 // 2 borders and column line
}

func (d *Detailview) scrollEnd() int {
	return min(len(d.Rows), d.scroll + d.heightForRows())
}

func (d *Detailview) Render() {
	RenderBorder(d.Bounds)

	top := d.Bounds.Top + 1
	left := d.Bounds.Left + 1

	for _, col := range d.Columns {
		termPrintf(left, top, renderValue(col.Name, col.Width))
		left += col.Width
	}

	cursorX := 0
	cursorY := 0

	for r := d.scroll; r < d.scrollEnd(); r++ {
		left = d.Bounds.Left + 1
		top++

		for ci, col := range d.Columns {
			termPrintf(left, top,
				   renderValue(d.Rows[r][ci], col.Width))

			if d.cursorCol == ci && d.cursorRow == r {
				cursorX = left
				cursorY = top
			}

			left += col.Width
		}
	}

	if d.focus {
		termbox.SetCursor(cursorX, cursorY)
	}
}

func (d *Detailview) SetFocus() {
	d.focus = true
}

func (d *Detailview) UnsetFocus() {
	d.focus = false
}

func (d *Detailview) HandleEvent(ev escapebox.Event) {
	if ev.Type != termbox.EventKey {
		return
	}

	switch ev.Ch {
	case 'k':
		d.cursorRow--
	case 'j':
		d.cursorRow++
	case 'h':
		d.cursorCol--
	case 'l':
		d.cursorCol++
	}

	// Clamp cursor
	d.cursorRow = max(0, d.cursorRow)
	d.cursorRow = min(len(d.Rows) - 1, d.cursorRow)

	d.cursorCol = max(0, d.cursorCol)
	d.cursorCol = min(len(d.Columns) - 1, d.cursorCol)

	// Clamp scroll
	if d.cursorRow < d.scroll {
		d.scroll = d.cursorRow
	}

	if d.cursorRow >= d.scrollEnd() {
		d.scroll = d.cursorRow - d.heightForRows() + 1
	}
}
