package components

import "time"

// Velocity moves the entity's transform by (X, Y) world units per second.
type Velocity struct {
	X float64 `yaml:"x"`
	Y float64 `yaml:"y"`
}

func (v *Velocity) Type() string { return "velocity" }

func (v *Velocity) Update(e *Entity, dt time.Duration) {
	t := e.Transform()
	if t == nil {
		return
	}
	s := dt.Seconds()
	t.X += v.X * s
	t.Y += v.Y * s
}
