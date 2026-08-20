package scene

import (
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/mcbalaam/ebitter/pkg/engine"
	"github.com/mcbalaam/ebitter/pkg/render"
	"github.com/mcbalaam/ebitter/pkg/tiled"
)

// TiledMapScene renders a Tiled GameMap, moves a player (a red square) with the
// arrow keys/WASD and resolves collisions against the map's collision objects.
// The camera follows the player and is clamped to the map bounds. When the
// player enters an interaction zone, the zone's signal is emitted on the
// master signal bus.
type TiledMapScene struct {
	Map           *tiled.GameMap
	Camera        *engine.Camera
	BgColor       color.Color
	ShowColliders bool
	ShowZones     bool

	input   *engine.Input
	player  *tiled.ColliderBox
	color   color.RGBA
	speed   float64
	scrollX float64
	scrollY float64
	built   bool
	sprite  *render.AnimatedIcon
	facing  string
}

// NewTiledMapScene loads a map scaled by mapScale and sets up the demo
// player. Tile layers are pre-rendered lazily on the first update, after the
// Ebiten graphics context has started.
func NewTiledMapScene(mapPath string, mapScale float64) (*TiledMapScene, error) {
	m, err := tiled.LoadMapSpecScaled(mapPath, tiled.DefaultTilesetCache, mapScale)
	if err != nil {
		return nil, err
	}

	s := &TiledMapScene{
		Map:           m,
		BgColor:       color.Black,
		ShowColliders: true,
		ShowZones:     true,
		input:         engine.NewInput(),
		color:         color.RGBA{0xff, 0x3a, 0x3a, 0xff},
		speed:         160,
		facing:        "down",
	}

	const pw, ph = 24, 24
	s.player = &tiled.ColliderBox{X: 80, Y: 200, W: pw, H: ph}
	// Spawn at the map's "Start" point when present, centered on it.
	if p, ok := m.Point("start"); ok {
		s.player.X = p.X - pw/2
		s.player.Y = p.Y - ph/2
	}
	return s, nil
}

func (s *TiledMapScene) Update(dt time.Duration) {
	s.input.Update()

	if !s.built {
		s.built = true
		if err := s.Map.BuildLayers(tiled.DefaultTilesetCache); err != nil {
			log.Printf("tiled: prerender: %v", err)
		}
		icon, err := render.NewAnimatedIconFromPath("media/sprites/frisk", "down")
		if err != nil {
			log.Printf("frisk sprite: %v", err)
		} else {
			s.sprite = icon
		}
	}

	dx, dy := 0.0, 0.0
	if s.input.IsPressed(ebiten.KeyLeft) || s.input.IsPressed(ebiten.KeyA) {
		dx -= 1
	}
	if s.input.IsPressed(ebiten.KeyRight) || s.input.IsPressed(ebiten.KeyD) {
		dx += 1
	}
	if s.input.IsPressed(ebiten.KeyUp) || s.input.IsPressed(ebiten.KeyW) {
		dy -= 1
	}
	if s.input.IsPressed(ebiten.KeyDown) || s.input.IsPressed(ebiten.KeyS) {
		dy += 1
	}

	if s.input.JustPressed(ebiten.KeyF1) {
		s.ShowColliders = !s.ShowColliders
	}
	if s.input.JustPressed(ebiten.KeyF2) {
		s.ShowZones = !s.ShowZones
	}

	seconds := dt.Seconds()
	solids := s.Map.Colliders()

	// Move and resolve per axis to allow wall sliding.
	if dx != 0 || dy != 0 {
		normalized := 1.0
		dx *= normalized
		dy *= normalized

		s.moveAxis(0, dx*s.speed*seconds, solids)
		s.moveAxis(1, dy*s.speed*seconds, solids)
	}

	// Pick the animation state from the movement direction and advance it.
	if s.sprite != nil {
		switch {
		case dx < 0:
			s.facing = "left"
		case dx > 0:
			s.facing = "right"
		case dy < 0:
			s.facing = "up"
		case dy > 0:
			s.facing = "down"
		}
		state := s.facing
		if dx != 0 || dy != 0 {
			state += "_walk"
		}
		_ = s.sprite.SetIconState(state)
		s.sprite.Update(dt)
	}

	pressed := s.input.JustPressed(ebiten.KeyZ) || s.input.JustPressed(ebiten.KeyEnter)
	for _, zone := range s.Map.Interactions() {
		zone.Check(s.player, pressed, s)
	}

	s.updateCamera()
}

// updateCamera computes the world scroll offset that keeps the player centered
// and clamps it to the map bounds so the world never scrolls past its edges.
func (s *TiledMapScene) updateCamera() {
	const (
		screenW = 640
		screenH = 480
	)
	halfW := float64(screenW) / 2
	halfH := float64(screenH) / 2

	// Player's center.
	cx := s.player.X + s.player.W/2
	cy := s.player.Y + s.player.H/2

	// Top-left of the visible world, clamped to the map bounds.
	maxX := float64(s.Map.Width() - screenW)
	maxY := float64(s.Map.Height() - screenH)
	if maxX < 0 {
		maxX = 0
	}
	if maxY < 0 {
		maxY = 0
	}
	s.scrollX = clamp(cx-halfW, 0, maxX)
	s.scrollY = clamp(cy-halfH, 0, maxY)
}

// moveAxis moves the player along a single axis (0 = X, 1 = Y) by delta and
// pushes the player out of any solid it ends up overlapping.
func (s *TiledMapScene) moveAxis(axis int, delta float64, solids []*tiled.ColliderBox) {
	if axis == 0 {
		s.player.X += delta
		for _, c := range solids {
			if !s.player.Overlaps(c) {
				continue
			}
			if delta > 0 {
				s.player.X = c.X - s.player.W
			} else if delta < 0 {
				s.player.X = c.X + c.W
			}
		}
		return
	}

	s.player.Y += delta
	for _, c := range solids {
		if !s.player.Overlaps(c) {
			continue
		}
		if delta > 0 {
			s.player.Y = c.Y - s.player.H
		} else if delta < 0 {
			s.player.Y = c.Y + c.H
		}
	}
}

func (s *TiledMapScene) Draw(screen *ebiten.Image) {
	screen.Fill(s.BgColor)
	sx := -s.scrollX
	sy := -s.scrollY

	s.Map.DrawLayers(screen, sx, sy)

	// Player: sprite anchored to the hitbox's bottom center at the map scale,
	// or the debug square if the sprite failed to load.
	b := s.player
	if s.sprite != nil && s.sprite.CurrentState != nil && len(s.sprite.CurrentState.Frames) > 0 {
		m := s.Map.Scale()
		frame := s.sprite.CurrentState.Frames[s.sprite.CurrentState.CurrentFrame].Image
		if frame != nil {
			fb := frame.Bounds()
			fw := float64(fb.Dx()) * m
			fh := float64(fb.Dy()) * m
			x := b.X + sx + (b.W-fw)/2
			y := b.Y + sy + b.H - fh
			s.sprite.Draw(screen, x, y, m, m, 0)
		}
	} else {
		vector.DrawFilledRect(screen,
			float32(b.X+sx), float32(b.Y+sy), float32(b.W), float32(b.H),
			s.color, true)
	}

	if s.ShowZones {
		for _, z := range s.Map.Interactions() {
			clr := color.RGBA{0x3a, 0xb8, 0xff, 0x60}
			if z.Trigger == tiled.TriggerButton {
				clr = color.RGBA{0xff, 0xd0, 0x3a, 0x60}
			}
			vector.DrawFilledRect(screen,
				float32(z.Rect.X+sx), float32(z.Rect.Y+sy), float32(z.Rect.W), float32(z.Rect.H),
				clr, false)
		}
	}

	if s.ShowColliders {
		for _, c := range s.Map.Colliders() {
			vector.StrokeRect(screen,
				float32(c.X+sx), float32(c.Y+sy), float32(c.W), float32(c.H),
				2, color.RGBA{0xff, 0x40, 0x40, 0xc0}, true)
		}
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
