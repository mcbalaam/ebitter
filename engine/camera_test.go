package engine

import (
	"math"
	"testing"
)

func TestViewportOriginWithinBounds(t *testing.T) {
	c := NewCamera()
	c.SetBounds(840, 640)

	tests := []struct {
		name         string
		cx, cy       float64
		zoom         float64
		sw, sh       int
		wantX, wantY float64
	}{
		{"center", 420, 320, 1, 640, 480, 100, 80},
		{"top-left corner clamps to 0", 0, 0, 1, 640, 480, 0, 0},
		{"bottom-right clamps to max", 840, 640, 1, 640, 480, 200, 160},
		{"half zoom out widens view", 420, 320, 0.5, 640, 480, -220, -160},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.X, c.Y, c.Zoom = tt.cx, tt.cy, tt.zoom
			gotX, gotY := c.ViewportOrigin(tt.sw, tt.sh)
			if math.Abs(gotX-tt.wantX) > 1e-9 || math.Abs(gotY-tt.wantY) > 1e-9 {
				t.Fatalf("ViewportOrigin = (%v, %v), want (%v, %v)", gotX, gotY, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestViewportOriginCentersWhenMapSmallerThanView(t *testing.T) {
	c := NewCamera()
	c.SetBounds(100, 100)
	// Viewport is 640 wide vs a 100px map → origin is centered (negative).
	x, y := c.ViewportOrigin(640, 480)
	if want := (100.0 - 640) / 2; math.Abs(x-want) > 1e-9 {
		t.Fatalf("x = %v, want centered %v", x, want)
	}
	if want := (100.0 - 480) / 2; math.Abs(y-want) > 1e-9 {
		t.Fatalf("y = %v, want centered %v", y, want)
	}
}

func TestViewportOriginUnboundedWhenNoBounds(t *testing.T) {
	c := NewCamera()
	c.X, c.Y, c.Zoom = 420, 320, 1
	x, y := c.ViewportOrigin(640, 480)
	if x != 100 || y != 80 {
		t.Fatalf("free scroll origin = (%v, %v), want (100, 80)", x, y)
	}
}

func TestCenterOnClampsToViewport(t *testing.T) {
	c := NewCamera()
	c.SetBounds(840, 640)
	c.SetViewport(640, 480)

	c.CenterOn(420, 320)
	if c.X != 420 || c.Y != 320 {
		t.Fatalf("mid-map center = (%v, %v), want (420, 320)", c.X, c.Y)
	}

	// At the left/top edge the center itself is clamped to the margin, so
	// X, Y remain the true on-screen center.
	c.CenterOn(0, 0)
	if c.X != 320 || c.Y != 240 {
		t.Fatalf("edge center = (%v, %v), want (320, 240)", c.X, c.Y)
	}
	c.CenterOn(840, 640)
	if c.X != 520 || c.Y != 400 {
		t.Fatalf("edge center = (%v, %v), want (520, 400)", c.X, c.Y)
	}

	// Zooming out past the map size pins the center to the map midpoint.
	c.SetZoomLimits(0.5, 2)
	c.SetZoom(0.75)
	c.CenterOn(0, 0)
	if c.X != 420 || c.Y != 320 {
		t.Fatalf("overscroll center = (%v, %v), want map midpoint (420, 320)", c.X, c.Y)
	}

	// The rendered origin agrees with the center.
	tx, _ := c.ViewportOrigin(640, 480)
	if vw := float64(640) / c.Zoom; math.Abs(tx+vw/2-c.X) > 1e-9 {
		t.Fatalf("origin (%v) and center (%v) disagree on screen midpoint", tx, c.X)
	}
}

func TestCameraZoomLimits(t *testing.T) {
	c := NewCamera()
	c.SetZoomLimits(0.75, 3)

	c.SetZoom(5)
	if c.Zoom != 3 {
		t.Fatalf("zoom over max = %v, want 3", c.Zoom)
	}
	c.SetZoom(0.1)
	if c.Zoom != 0.75 {
		t.Fatalf("zoom under min = %v, want 0.75", c.Zoom)
	}
	c.ZoomBy(2)
	if c.Zoom != 1.5 {
		t.Fatalf("ZoomBy(2) from 0.75 = %v, want 1.5", c.Zoom)
	}
}

func TestCameraSmoothSnapsWhenDisabled(t *testing.T) {
	c := NewCamera()
	c.SetBounds(840, 640)
	c.CenterOn(100, 100)
	c.SetSmooth(true)
	c.CenterOn(500, 500)
	// Tween drives via the update queue; snap path is exercised by toggling off.
	c.SetSmooth(false)
	if c.X != 500 || c.Y != 500 {
		t.Fatalf("after disabling smooth, camera = (%v, %v), want target (500, 500)", c.X, c.Y)
	}
	if !(c.Zoom > 0) {
		t.Fatalf("after disabling smooth, zoom = %v, want > 0 (regression: targetZoom zeroed the view)", c.Zoom)
	}
	tx, ty := c.ViewportOrigin(640, 480)
	if math.IsInf(tx, 0) || math.IsInf(ty, 0) {
		t.Fatalf("viewport origin not finite: (%v, %v)", tx, ty)
	}
}
