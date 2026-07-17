package components

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Collider struct {
	LocalVerts []Vec
	WorldVerts []Vec
	Width      float64
	Height     float64
	OffsetX    float64
	OffsetY    float64
}

func NewCollider(width, height, offsetX, offsetY float64) *Collider {
	halfW := width / 2
	halfH := height / 2
	return &Collider{
		LocalVerts: []Vec{
			{-halfW, -halfH},
			{halfW, -halfH},
			{halfW, halfH},
			{-halfW, halfH},
		},
		WorldVerts: make([]Vec, 4),
		Width:      width,
		Height:     height,
		OffsetX:    offsetX,
		OffsetY:    offsetY,
	}
}

func (c *Collider) UpdateWorldVerts(t *Transform) {
	if c == nil || len(c.LocalVerts) == 0 {
		return
	}
	for i, local := range c.LocalVerts {
		c.WorldVerts[i] = TransformPoint(local, t.ScaleX, t.ScaleY, t.Rotation, t.X, t.Y)
	}
}

func (c *Collider) CollidesWith(other *Collider) bool {
	if c == nil || other == nil {
		return false
	}
	return polygonsIntersect(c.WorldVerts, other.WorldVerts)
}

func (c *Collider) GetBounds() (minX, minY, maxX, maxY float64) {
	if c == nil || len(c.WorldVerts) == 0 {
		return 0, 0, 0, 0
	}
	minX = math.Inf(1)
	minY = math.Inf(1)
	maxX = math.Inf(-1)
	maxY = math.Inf(-1)
	for _, v := range c.WorldVerts {
		if v.X < minX {
			minX = v.X
		}
		if v.Y < minY {
			minY = v.Y
		}
		if v.X > maxX {
			maxX = v.X
		}
		if v.Y > maxY {
			maxY = v.Y
		}
	}
	return
}

func (c *Collider) SetOffset(ox, oy float64) {
	c.OffsetX = ox
	c.OffsetY = oy
}

func (c *Collider) Center() {
	c.SetOffset(-c.Width/2, -c.Height/2)
}

func (c *Collider) DrawDebug(screen *ebiten.Image, clr color.Color) {
	if c == nil || len(c.WorldVerts) < 2 {
		return
	}
	for i := 0; i < len(c.WorldVerts); i++ {
		cur := c.WorldVerts[i]
		next := c.WorldVerts[(i+1)%len(c.WorldVerts)]
		vector.StrokeLine(screen, float32(cur.X), float32(cur.Y),
			float32(next.X), float32(next.Y), 2, clr, true)
	}
}
