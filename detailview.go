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
	Columns    []Column
	Rows	   [][]string
}

func renderValue(src string, maxWidth int) string {
	maxLen := min(maxWidth, len(src))
	return src[0:maxLen]
}

func (d *Detailview) Render() {
	RenderBorder(d.Bounds)

	top := d.Bounds.Top + 1
	left := d.Bounds.Left + 1

	for _, col := range d.Columns {
		termPrintf(left, top, renderValue(col.Name, col.Width))
		left += col.Width
	}

	heightForRows := d.Bounds.Height - 3 // 2 borders and column line

	scrollEnd := min(len(d.Rows), d.scroll + heightForRows)

	for r := d.scroll; r < scrollEnd; r++ {
		left = d.Bounds.Left + 1
		top++

		for ci, col := range d.Columns {
			termPrintf(left, top,
				   renderValue(d.Rows[r][ci], col.Width))

			left += col.Width
		}
	}

	if d.focus {
		termbox.SetCursor(d.Bounds.Left + 1, d.Bounds.Top + 1)
	}
}

func (d *Detailview) SetFocus() {
	d.focus = true
}

func (d *Detailview) UnsetFocus() {
	d.focus = false
}

func (d *Detailview) HandleEvent(ev escapebox.Event) {
}
