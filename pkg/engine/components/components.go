package components

import (
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/mcbalaam/ebitter/pkg/render"
)

type Vec struct {
	X, Y float64
}

type Transform struct {
	X        float64
	Y        float64
	ScaleX   float64
	ScaleY   float64
	Rotation float64
}

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

type Velocity struct {
	X float64
	Y float64
}

type Sprite struct {
	Icon *render.AnimatedIcon
}

func (s *Sprite) Update(dt time.Duration) {
	if s.Icon != nil {
		s.Icon.Update(dt)
	}
}

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

func TransformPoint(p Vec, sx, sy, rot, tx, ty float64) Vec {
	x, y := p.X*sx, p.Y*sy
	c, s := math.Cos(rot), math.Sin(rot)
	rx := x*c - y*s
	ry := x*s + y*c
	return Vec{rx + tx, ry + ty}
}

func projectPoly(axis Vec, verts []Vec) (min, max float64) {
	min, max = math.Inf(1), math.Inf(-1)
	for _, v := range verts {
		proj := v.X*axis.X + v.Y*axis.Y
		if proj < min {
			min = proj
		}
		if proj > max {
			max = proj
		}
	}
	return
}

func overlapOnAxis(aMin, aMax, bMin, bMax float64) bool {
	return !(aMax < bMin || bMax < aMin)
}

func polygonsIntersect(a, b []Vec) bool {
	checkAxes := func(verts []Vec) bool {
		n := len(verts)
		for i := 0; i < n; i++ {
			j := (i + 1) % n
			ex := verts[j].X - verts[i].X
			ey := verts[j].Y - verts[i].Y
			axis := Vec{-ey, ex}
			aMin, aMax := projectPoly(axis, a)
			bMin, bMax := projectPoly(axis, b)
			if !overlapOnAxis(aMin, aMax, bMin, bMax) {
				return false
			}
		}
		return true
	}
	return checkAxes(a) && checkAxes(b)
}

func RectToPoly(width, height float64) []Vec {
	return []Vec{
		{0, 0},
		{width, 0},
		{width, height},
		{0, height},
	}
}

func TransformPoly(verts []Vec, sx, sy, rot, tx, ty float64) []Vec {
	out := make([]Vec, len(verts))
	for i, v := range verts {
		out[i] = TransformPoint(v, sx, sy, rot, tx, ty)
	}
	return out
}
