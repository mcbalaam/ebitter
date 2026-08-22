package components

import (
	"math"
	"testing"
)

func TestColliderOffsetAppliedToWorldVerts(t *testing.T) {
	c := NewCollider(10, 10, 5, -3)
	tf := &Transform{X: 100, Y: 200, ScaleX: 1, ScaleY: 1, Rotation: 0}

	c.UpdateWorldVerts(tf)

	// Local top-left vert is (-5,-5), centered at origin; offset shifts the
	// hitbox. With no rotation/scale: world = local + offset + transform.
	got := c.WorldVerts[0]
	if math.Abs(got.X-100) > 1e-9 || math.Abs(got.Y-192) > 1e-9 {
		t.Errorf("offset not applied: got %v want {100 192}", got)
	}
}

func TestColliderOffsetRotatesWithTransform(t *testing.T) {
	c := NewCollider(10, 10, 10, 0)
	tf := &Transform{X: 0, Y: 0, ScaleX: 1, ScaleY: 1, Rotation: math.Pi / 2}

	c.UpdateWorldVerts(tf)

	// 90deg rotation maps offset (10, 0) to (0, 10) and local (-5,-5) to (5,-5).
	got := c.WorldVerts[0]
	if math.Abs(got.X-5) > 1e-9 || math.Abs(got.Y-5) > 1e-9 {
		t.Errorf("offset should rotate with transform: got %v want {5 5}", got)
	}
}

func TestColliderNoOffsetKeepsCentroid(t *testing.T) {
	c := NewCollider(10, 10, 0, 0)
	tf := &Transform{X: 100, Y: 100, ScaleX: 2, ScaleY: 2, Rotation: 0}

	c.UpdateWorldVerts(tf)

	// Top-left local vert (-5,-5) scaled by 2 -> (-10,-10), translated to (100,100).
	got := c.WorldVerts[0]
	if got.X != 90 || got.Y != 90 {
		t.Errorf("unexpected world vert: %v", got)
	}
}
