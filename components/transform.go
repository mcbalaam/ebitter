package components

// Transform places a component in the world. ScaleX/ScaleY default to 1.
type Transform struct {
	X        float64
	Y        float64
	ScaleX   float64
	ScaleY   float64
	Rotation float64
}

func (t *Transform) Type() string { return "transform" }

func (t *Transform) Move(dx, dy float64) {
	t.X += dx
	t.Y += dy
}

func (t *Transform) SetPosition(x, y float64) {
	t.X = x
	t.Y = y
}

func (t *Transform) Rotate(angle float64) {
	t.Rotation += angle
}

func (t *Transform) SetRotation(angle float64) {
	t.Rotation = angle
}

func (t *Transform) SetScale(s float64) {
	t.ScaleX = s
	t.ScaleY = s
}
