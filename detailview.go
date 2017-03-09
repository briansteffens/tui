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
	scrollCol  int
	scrollRow  int
	cursorCol  int
	cursorRow  int
	Columns    []Column
	Rows	   [][]string
}

func (d *Detailview) SetFocus() {
	d.focus = true
}

func (d *Detailview) UnsetFocus() {
	d.focus = false
}

func renderValue(src string, maxWidth int) string {
	maxLen := min(maxWidth, len(src))
	return src[0:maxLen]
}

func (d *Detailview) viewHeight() int {
	return d.Bounds.Height - 3 // 2 borders and column line
}

func (d *Detailview) viewWidth() int {
	return d.Bounds.Width - 2 // 2 borders
}

func (d *Detailview) lastVisibleRow() int {
	return min(len(d.Rows), d.scrollRow + d.viewHeight())
}

func (d *Detailview) firstVisibleCol() (int, int) {
	left := 0
	right := 0

	for ci, col := range d.Columns {
		right = left + col.Width

		if d.scrollCol < right {
			return ci, d.scrollCol - left
		}

		left += col.Width
	}

	return 0, 0
}

func (d *Detailview) lastVisibleCol() (int, int) {
	first, _ := d.firstVisibleCol()
	right := -1

	for ci, col := range d.Columns {
		right += col.Width

		if ci < first {
			continue
		}

		if d.scrollColEnd() <= right {
			if ci == 1 {
				//panic(d.scrollColEnd())
				//panic(right)
			}
			return ci, right - d.scrollColEnd()
		}
	}

	return len(d.Columns) - 1, 0
}

func (d *Detailview) scrollColEnd() int {
	return d.scrollCol + d.viewWidth() - 1
}

func (d *Detailview) columnLeft(colIndex int) int {
	ret := 0

	for i := 0; i < colIndex; i++ {
		ret += d.Columns[i].Width
	}

	return ret
}

func (d *Detailview) columnRight(colIndex int) int {
	return d.columnLeft(colIndex) + d.Columns[colIndex].Width - 1
}

func (d *Detailview) Render() {
	RenderBorder(d.Bounds)

	top := d.Bounds.Top + 1
	left := d.Bounds.Left + 1

	firstCol, firstOffset := d.firstVisibleCol()
	lastCol, lastOffset := d.lastVisibleCol()

	for ci := firstCol; ci <= lastCol; ci++ {
		col := d.Columns[ci]

		name := col.Name

		if ci == firstCol {
			name = name[firstOffset:len(name)]
		}

		maxLen := col.Width

		if ci == lastCol {
			maxLen = min(maxLen, col.Width - lastOffset)
		}

		termPrintf(left, top, renderValue(name, maxLen))

		left += col.Width

		if ci == firstCol {
			left -= firstOffset
		}
	}

	cursorX := 0
	cursorY := 0

	for r := d.scrollRow; r < d.lastVisibleRow(); r++ {
		left = d.Bounds.Left + 1
		top++

		for ci := firstCol; ci <= lastCol; ci++ {
			col := d.Columns[ci]

			val := d.Rows[r][ci]

			if ci == firstCol {
				val = val[firstOffset:len(val)]
			}

			maxLen := col.Width

			if ci == lastCol {
				maxLen = min(maxLen, col.Width - lastOffset)
			}

			termPrintf(left, top, renderValue(val, maxLen))

			if d.cursorCol == ci && d.cursorRow == r {
				cursorX = left
				cursorY = top
			}

			left += col.Width

			if ci == firstCol {
				left -= firstOffset
			}
		}
	}

	if d.focus {
		termbox.SetCursor(cursorX, cursorY)
	}
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
	if d.cursorRow < d.scrollRow {
		d.scrollRow = d.cursorRow
	}

	if d.cursorRow >= d.lastVisibleRow() {
		d.scrollRow = d.cursorRow - d.viewHeight() + 1
	}

	if d.columnLeft(d.cursorCol) < d.scrollCol {
		d.scrollCol = d.columnLeft(d.cursorCol)
	}

	if d.columnLeft(d.cursorCol) >= d.scrollColEnd() {
		d.scrollCol = d.columnLeft(d.cursorCol) // - d.viewWidth() + 1
	}

	if d.columnRight(d.cursorCol) > d.scrollColEnd() {
		d.scrollCol = min(
			d.columnLeft(d.cursorCol),
			d.columnRight(d.cursorCol) - d.scrollColEnd())
	}
}
