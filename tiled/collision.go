package tiled

import (
	"strings"

	gotiled "github.com/lafriks/go-tiled"
)

// ColliderBox is an axis-aligned rectangle used for static map collision.
type ColliderBox struct {
	X, Y float64
	W, H float64
}

// Overlaps reports whether two boxes intersect.
func (b *ColliderBox) Overlaps(o *ColliderBox) bool {
	return b.X < o.X+o.W && b.X+b.W > o.X && b.Y < o.Y+o.H && b.Y+b.H > o.Y
}

// Contains reports whether the box contains the point (px, py).
func (b *ColliderBox) Contains(px, py float64) bool {
	return px >= b.X && px < b.X+b.W && py >= b.Y && py < b.Y+b.H
}

// resolveCollisions pushes the moving box out of the given solids.
// It is intended to be called separately for the X and Y axes to allow
// wall sliding. Returns the corrected position.
func resolveCollisions(box *ColliderBox, solids []*ColliderBox) (float64, float64) {
	x, y := box.X, box.Y

	for _, s := range solids {
		if !box.Overlaps(s) {
			continue
		}
		overlapX := s.X + s.W - box.X
		if box.X+box.W > s.X+s.W {
			overlapX = box.X + box.W - s.X
		}
		overlapY := s.Y + s.H - box.Y
		if box.Y+box.H > s.Y+s.H {
			overlapY = box.Y + box.H - s.Y
		}

		if overlapX <= overlapY {
			if overlapX > 0 {
				if box.X < s.X {
					x = s.X - box.W
				} else {
					x = s.X + s.W
				}
			}
		} else {
			if overlapY > 0 {
				if box.Y < s.Y {
					y = s.Y - box.H
				} else {
					y = s.Y + s.H
				}
			}
		}
		box.X, box.Y = x, y
	}

	return x, y
}

// collectColliders builds static collision boxes from Tiled object groups.
// A group contributes its objects when its name or class contains "collision"
// or it has a boolean "collision" property set to true.
func collectColliders(groups []*gotiled.ObjectGroup) []*ColliderBox {
	var out []*ColliderBox
	for _, g := range groups {
		if !groupCollides(g) {
			continue
		}
		for _, obj := range g.Objects {
			if obj.Width <= 0 || obj.Height <= 0 {
				continue
			}
			out = append(out, &ColliderBox{
				X: obj.X,
				Y: obj.Y,
				W: obj.Width,
				H: obj.Height,
			})
		}
	}
	return out
}

// groupCollides reports whether an object group is a collision layer.
// A group matches when its name or class contains "collision" or "hitbox",
// or it has a boolean "collision" property set to true.
func groupCollides(g *gotiled.ObjectGroup) bool {
	if g == nil {
		return false
	}
	name := strings.ToLower(g.Name)
	class := strings.ToLower(g.Class)
	if strings.Contains(name, "collision") || strings.Contains(name, "hitbox") ||
		strings.Contains(class, "collision") || strings.Contains(class, "hitbox") {
		return true
	}
	return g.Properties.GetBool("collision")
}
