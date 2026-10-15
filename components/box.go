package components

// AABB is an axis-aligned bounding box in world coordinates.
type AABB struct {
	X, Y, W, H float64
}

// Overlaps reports whether two boxes intersect.
func (a AABB) Overlaps(b AABB) bool {
	return a.X < b.X+b.W && a.X+a.W > b.X && a.Y < b.Y+b.H && a.Y+a.H > b.Y
}

// Center returns the box center.
func (a AABB) Center() (float64, float64) { return a.X + a.W/2, a.Y + a.H/2 }
