package engine

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/tween"
)

// Camera is a world-coordinate camera. X, Y is the world point shown at the
// center of the view — the source of truth for where the camera is. It is
// clamped to the map bounds (SetBounds with the screen size from SetViewport),
// so X, Y always equals the on-screen center; at map edges the camera stops
// instead of revealing the void. Zoom scales the view (larger = zoomed in).
//
// When Smooth is enabled, every follow (CenterOn) and zoom (SetZoom) transition
// is animated through the tween package on the default update queue. When
// Smooth is disabled, transitions snap instantly.
type Camera struct {
	X, Y float64
	Zoom float64

	worldW, worldH float64
	sw, sh         int
	minZoom        float64
	maxZoom        float64
	smooth         bool
	followDur      time.Duration
	zoomDur        time.Duration

	easing     tween.EasingFunc
	zoomFilter ebiten.Filter

	targetX, targetY float64
	targetZoom       float64

	posTween  *tween.Tween[tween.Vec2]
	zoomTween *tween.Tween[float64]

	offscreen *ebiten.Image
	offW      int
	offH      int
}

// NewCamera returns a 1:1, non-smooth camera centered on the origin.
func NewCamera() *Camera {
	return &Camera{
		Zoom:       1,
		targetZoom: 1,
		minZoom:    1,
		maxZoom:    1,
		followDur:  200 * time.Millisecond,
		zoomDur:    150 * time.Millisecond,
		easing:     tween.EaseOutQuad,
		zoomFilter: ebiten.FilterNearest,
	}
}

// SetEasing sets the easing function used for smooth follow and zoom tweens.
// A nil easing restores the default (EaseOutQuad).
func (c *Camera) SetEasing(easing tween.EasingFunc) { c.easing = easing }

// SetZoomFilter sets the filter used when upscaling the zoomed view. The
// default is ebiten.FilterNearest (crisp pixel art).
func (c *Camera) SetZoomFilter(filter ebiten.Filter) { c.zoomFilter = filter }

func (c *Camera) ease() tween.EasingFunc {
	if c.easing == nil {
		return tween.EaseOutQuad
	}
	return c.easing
}

// SetBounds clamps the camera center to a map of worldW x worldH pixels. Pass a
// non-positive size to disable clamping (free scroll).
func (c *Camera) SetBounds(worldW, worldH float64) {
	c.worldW, c.worldH = worldW, worldH
	c.clampTargets()
}

// SetViewport informs the camera of the screen size it renders to, which
// determines how far from the map edges the center can sit. Apply refreshes it
// automatically from the target image.
func (c *Camera) SetViewport(w, h int) {
	if w > 0 && h > 0 {
		c.sw, c.sh = w, h
		c.clampTargets()
	}
}

// SetZoomLimits restricts zoom to [min, max]. A max below min disables limits
// (any positive zoom is allowed).
func (c *Camera) SetZoomLimits(min, max float64) {
	if min <= 0 {
		min = 1
	}
	if max < min {
		min, max = 1, math.MaxFloat64
	}
	c.minZoom, c.maxZoom = min, max
	c.clampTargets()
}

// SetSmooth toggles tweened follow and zoom. Turning it off snaps the camera to
// its current targets immediately.
func (c *Camera) SetSmooth(on bool) {
	if c.smooth == on {
		return
	}
	c.smooth = on
	if !on {
		c.killTweens()
		c.X, c.Y = c.targetX, c.targetY
		if c.targetZoom > 0 {
			c.Zoom = c.targetZoom
		}
	}
}

// Smooth reports whether tweened transitions are active.
func (c *Camera) Smooth() bool { return c.smooth }

// FollowDuration sets the tween duration used for position transitions.
func (c *Camera) FollowDuration(d time.Duration) { c.followDur = d }

// ZoomDuration sets the tween duration used for zoom transitions.
func (c *Camera) ZoomDuration(d time.Duration) { c.zoomDur = d }

// CenterOn moves the camera to center on the world point (x, y). The result is
// clamped so that X, Y remain the real on-screen center of the view.
func (c *Camera) CenterOn(x, y float64) {
	c.targetX, c.targetY = x, y
	c.clampTargets()
	if c.smooth {
		if c.posTween == nil {
			c.posTween = tween.NewVec2(
				tween.Vec2{X: c.X, Y: c.Y},
				tween.Vec2{X: c.targetX, Y: c.targetY},
				c.followDur,
				func(v tween.Vec2) { c.X, c.Y = v.X, v.Y },
			).SetEasing(c.ease()).Schedule()
		} else {
			c.posTween.SetDuration(c.followDur).Retarget(tween.Vec2{X: c.targetX, Y: c.targetY})
		}
		return
	}
	c.killPosTween()
	c.X, c.Y = c.targetX, c.targetY
}

// SetZoom sets the zoom level, clamped to the configured limits. The camera
// center is re-clamped against the new viewport size.
func (c *Camera) SetZoom(z float64) {
	if z < c.minZoom {
		z = c.minZoom
	}
	if z > c.maxZoom {
		z = c.maxZoom
	}
	c.targetZoom = z
	c.clampTargets()
	if c.smooth {
		if c.zoomTween == nil {
			c.zoomTween = tween.NewFloat(c.Zoom, z, c.zoomDur,
				func(v float64) { c.Zoom = v },
			).SetEasing(c.ease()).Schedule()
		} else {
			c.zoomTween.SetDuration(c.zoomDur).Retarget(z)
		}
		return
	}
	c.killZoomTween()
	c.Zoom = z
}

// ZoomBy multiplies the current target zoom by factor and applies the result.
func (c *Camera) ZoomBy(factor float64) {
	c.SetZoom(c.targetZoom * factor)
}

// clampTargets restricts the pending center/zoom so the view stays on the map.
// Until the first frame (no viewport known yet) only the zoom is restricted.
func (c *Camera) clampTargets() {
	if c.targetZoom < c.minZoom {
		c.targetZoom = c.minZoom
	}
	if c.targetZoom > c.maxZoom {
		c.targetZoom = c.maxZoom
	}
	if c.sw > 0 && c.sh > 0 && c.targetZoom > 0 {
		c.targetX = clampCenter(c.targetX, float64(c.sw)/c.targetZoom, c.worldW)
		c.targetY = clampCenter(c.targetY, float64(c.sh)/c.targetZoom, c.worldH)
	}
	c.clamp()
}

// clamp restricts the active center to the current viewport.
func (c *Camera) clamp() {
	if c.sw <= 0 || c.sh <= 0 || c.Zoom <= 0 {
		return
	}
	c.X = clampCenter(c.X, float64(c.sw)/c.Zoom, c.worldW)
	c.Y = clampCenter(c.Y, float64(c.sh)/c.Zoom, c.worldH)
}

// clampCenter restricts a view center so the viewport of viewSize pixels stays
// inside a world of world pixels. When the world is smaller than the view the
// world itself is centered, pinning the center to its midpoint.
func clampCenter(center, viewSize, world float64) float64 {
	if world <= 0 {
		return center
	}
	if world <= viewSize {
		return world / 2
	}
	lo, hi := viewSize/2, world-viewSize/2
	if center < lo {
		return lo
	}
	if center > hi {
		return hi
	}
	return center
}

func (c *Camera) killTweens() {
	c.killPosTween()
	c.killZoomTween()
}

func (c *Camera) killPosTween() {
	if c.posTween != nil {
		tween.Stop(c.posTween)
		c.posTween = nil
	}
}

func (c *Camera) killZoomTween() {
	if c.zoomTween != nil {
		tween.Stop(c.zoomTween)
		c.zoomTween = nil
	}
}

// ViewportOrigin returns the top-left world coordinate rendered at the
// screen's top-left corner for a screen of the given pixel size. The result is
// the camera center offset by half the (zoom-adjusted) viewport, clamped so the
// view never leaves the map bounds. When the view is wider or taller than the
// map, the map is centered on that axis instead.
func (c *Camera) ViewportOrigin(screenW, screenH int) (float64, float64) {
	return viewportOrigin(c.X, float64(screenW)/c.Zoom, c.worldW),
		viewportOrigin(c.Y, float64(screenH)/c.Zoom, c.worldH)
}

func viewportOrigin(center, viewSize, world float64) float64 {
	desired := center - viewSize/2
	if world <= 0 {
		return desired
	}
	maxOrigin := world - viewSize
	if maxOrigin < 0 {
		return maxOrigin / 2
	}
	if desired < 0 {
		return 0
	}
	if desired > maxOrigin {
		return maxOrigin
	}
	return desired
}

// Apply runs draw to render the world, mapping world coordinates through the
// camera's zoom and pan. When zoom is 1 the world is drawn directly onto
// screen; otherwise it is rendered to an offscreen buffer scaled up with
// nearest-neighbour filtering to keep pixels crisp. It also keeps the camera
// viewport (and thus the center clamping) in sync with the screen it draws to.
func (c *Camera) Apply(screen *ebiten.Image, draw func(target *ebiten.Image, ox, oy float64)) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	if sw > 0 && sh > 0 && (sw != c.sw || sh != c.sh) {
		c.sw, c.sh = sw, sh
		c.clampTargets()
	}
	c.clamp() // keep X, Y a true center for the (possibly mid-tween) zoom
	tx, ty := c.ViewportOrigin(sw, sh)

	if c.Zoom == 1 {
		draw(screen, -tx, -ty)
		return
	}

	vw := float64(sw) / c.Zoom
	vh := float64(sh) / c.Zoom
	bw, bh := int(math.Ceil(vw)), int(math.Ceil(vh))
	if bw < 1 {
		bw = 1
	}
	if bh < 1 {
		bh = 1
	}
	if c.offscreen == nil || c.offW != bw || c.offH != bh {
		c.offscreen = ebiten.NewImage(bw, bh)
		c.offW, c.offH = bw, bh
	}
	c.offscreen.Clear()
	draw(c.offscreen, -tx, -ty)

	op := &ebiten.DrawImageOptions{}
	op.Filter = c.zoomFilter
	op.GeoM.Scale(float64(sw)/float64(bw), float64(sh)/float64(bh))
	screen.DrawImage(c.offscreen, op)
}
