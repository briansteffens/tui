package tui

type Rect struct {
	Left, Top, Width, Height int
}

func (r Rect) Right() int {
	return r.Left + r.Width - 1
}

func (r Rect) Bottom() int {
	return r.Top + r.Height - 1
}

func (r Rect) ContainsRect(other Rect) bool {
	return other.Left >= r.Left && other.Right() <= r.Right() &&
		other.Top >= r.Top && other.Bottom() <= r.Bottom()
}

func (r Rect) ContainsPoint(x, y int) bool {
	return x >= 0 && x < r.Width &&
		y >= 0 && y < r.Height
}
