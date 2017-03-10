package tui

import (
	"github.com/nsf/termbox-go"
	"github.com/briansteffens/escapebox"
)

type Column struct {
	Name  string
	Width int
}

type DetailView struct {
	Bounds     Rect
	focus      bool
	scrollCol  int
	scrollRow  int
	cursorCol  int
	cursorRow  int
	Columns    []Column
	Rows	   [][]string
	RowBg	   termbox.Attribute
	RowBgAlt   termbox.Attribute
	SelectedBg termbox.Attribute
}

func (d *DetailView) SetFocus() {
	d.focus = true
}

func (d *DetailView) UnsetFocus() {
	d.focus = false
}

func renderValue(src string, maxWidth int) string {
	maxLen := min(maxWidth, len(src))
	return src[0:maxLen]
}

func (d *DetailView) viewHeight() int {
	return d.Bounds.Height - 1 // Header row
}

func (d *DetailView) viewWidth() int {
	return d.Bounds.Width
}

func (d *DetailView) lastVisibleRow() int {
	return min(len(d.Rows), d.scrollRow + d.viewHeight())
}

func (d *DetailView) firstVisibleCol() (int, int) {
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

func (d *DetailView) lastVisibleCol() (int, int) {
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

func (d *DetailView) scrollColEnd() int {
	return d.scrollCol + d.viewWidth() - 1
}

func (d *DetailView) columnLeft(colIndex int) int {
	ret := 0

	for i := 0; i < colIndex; i++ {
		ret += d.Columns[i].Width
	}

	return ret
}

func (d *DetailView) columnRight(colIndex int) int {
	return d.columnLeft(colIndex) + d.Columns[colIndex].Width - 1
}

func (d *DetailView) totalWidth() int {
	ret := 0

	for _, col := range d.Columns {
		ret += col.Width
	}

	return ret
}

func (d *DetailView) Render() {
	top := d.Bounds.Top
	left := d.Bounds.Left

	firstCol, firstOffset := d.firstVisibleCol()
	lastCol, lastOffset := d.lastVisibleCol()

	for ci := firstCol; ci <= lastCol; ci++ {
		col := d.Columns[ci]

		name := col.Name

		if ci == firstCol {
			if len(name) - firstOffset >= 0 {
				name = name[firstOffset:len(name)]
			} else {
				name = ""
			}
		}

		maxLen := col.Width

		if ci == lastCol {
			maxLen = min(maxLen, col.Width - lastOffset)
		}

		maxLen = min(maxLen, d.viewWidth())

		termPrintColorf(left, top,
				termbox.ColorWhite | termbox.AttrBold,
				termbox.ColorBlack, renderValue(name, maxLen))

		left += col.Width

		if ci == firstCol {
			left -= firstOffset
		}
	}

	for r := d.scrollRow; r < d.lastVisibleRow(); r++ {
		left = d.Bounds.Left
		top++

		rowColor := d.RowBg
		if r % 2 == 0 {
			rowColor = d.RowBgAlt
		}

		for ci := firstCol; ci <= lastCol; ci++ {
			col := d.Columns[ci]

			colColor := rowColor

			if d.cursorCol == ci && d.cursorRow == r && d.focus {
				colColor = d.SelectedBg
			}

			val := d.Rows[r][ci]

			if ci == firstCol {
				if len(val) - firstOffset <= 0 {
					val = ""
				} else {
					val = val[firstOffset:len(val)]
				}
			}

			maxLen := col.Width

			if ci == lastCol {
				maxLen = min(maxLen, col.Width - lastOffset)
			}

			maxLen = min(maxLen, d.viewWidth())

			if len(val) > maxLen {
				val = val[0:maxLen]
			}

			for {
				if len(val) >= maxLen {
					break
				}

				val = val + " "
			}

			termPrintColorf(left, top, termbox.ColorWhite,
					colColor, val)

			left += col.Width

			if ci == firstCol {
				left -= firstOffset
			}
		}
	}

	if d.focus {
		termbox.HideCursor()
	}
}

func (d *DetailView) HandleEvent(ev escapebox.Event) {
	if ev.Type != termbox.EventKey {
		return
	}

	oldCursorRow := d.cursorRow
	oldCursorCol := d.cursorCol

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

	switch ev.Key {
	case termbox.KeyArrowRight:
		d.scrollCol++
	case termbox.KeyArrowLeft:
		d.scrollCol--
	}

	// Clamp cursor
	d.cursorRow = max(0, d.cursorRow)
	d.cursorRow = min(len(d.Rows) - 1, d.cursorRow)

	d.cursorCol = max(0, d.cursorCol)
	d.cursorCol = min(len(d.Columns) - 1, d.cursorCol)

	cursorChanged := oldCursorRow != d.cursorRow ||
			 oldCursorCol != d.cursorCol

	// Clamp scroll
	d.scrollCol = max(d.scrollCol, 0)
	d.scrollRow = max(d.scrollRow, 0)

	maxScrollCol := max(0, d.totalWidth() - d.viewWidth())
	d.scrollCol = min(d.scrollCol, maxScrollCol)
	d.scrollRow = min(d.scrollRow, len(d.Rows) - 1)

	if cursorChanged {
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
			d.scrollCol = d.columnLeft(d.cursorCol)
		}

		if d.columnRight(d.cursorCol) > d.scrollColEnd() &&
		   d.columnLeft(d.cursorCol) > d.scrollCol {
			d.scrollCol = min(
				d.columnLeft(d.cursorCol),
				d.columnRight(d.cursorCol) - d.scrollColEnd())
		}
	}
}
