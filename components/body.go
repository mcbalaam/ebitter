package components

import (
	"math"
	"time"
)

// Body applies forces to the entity's Velocity: constant acceleration, gravity,
// drag and a speed cap. Grounded is a flag the game (or a collision pass) may
// set; the body itself does not resolve collisions.
type Body struct {
	AccelX   float64 `yaml:"accelX"`
	AccelY   float64 `yaml:"accelY"`
	Gravity  float64 `yaml:"gravity"`
	Drag     float64 `yaml:"drag"`
	MaxSpeed float64 `yaml:"maxSpeed"`
	Grounded bool    `yaml:"grounded"`
}

func (b *Body) Type() string { return "body" }

func (b *Body) Update(e *Entity, dt time.Duration) {
	v := e.Velocity()
	if v == nil {
		return
	}
	s := dt.Seconds()
	v.X += b.AccelX * s
	v.Y += (b.AccelY + b.Gravity) * s
	if b.Drag > 0 {
		f := 1 - b.Drag*s
		if f < 0 {
			f = 0
		}
		v.X *= f
		v.Y *= f
	}
	if b.MaxSpeed > 0 {
		if sp := math.Hypot(v.X, v.Y); sp > b.MaxSpeed {
			k := b.MaxSpeed / sp
			v.X *= k
			v.Y *= k
		}
	}
}
