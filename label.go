package tui

type Label struct {
	Bounds Rect
	Text   string
}

func (l *Label) Render() {
	termPrintf(l.Bounds.Left, l.Bounds.Top, l.Text)
}
