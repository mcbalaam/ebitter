package main

import (
	"image/color"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/mcbalaam/ebitter/engine"
	"github.com/mcbalaam/ebitter/queues"
	"github.com/mcbalaam/ebitter/render"
	"github.com/mcbalaam/ebitter/text"
	"github.com/mcbalaam/ebitter/tiled"
)

type demoScene struct {
	Map            *tiled.GameMap
	Camera         *engine.Camera
	BgColor        color.Color
	ShowColliders  bool
	ShowZones      bool
	ShowCamArrow   bool
	ShowStateLabel bool

	input         *engine.Input
	player        *tiled.ColliderBox
	color         color.RGBA
	speed         float64
	built         bool
	sprite        *render.AnimatedIcon
	facing        string
	hintSmooth    *text.TextString
	hintColliders *text.TextString
	hintZones     *text.TextString
	hintArrow     *text.TextString
	hintLabel     *text.TextString
	stateLabel    *text.TextString
	lastState     string
}

const (
	hudHintFollow    = "3 to toggle smooth camera"
	hudHintColliders = "1 to toggle hitboxes"
	hudHintZones     = "2 to toggle interaction zones"
	hudHintArrow     = "4 to toggle camera debug"
	hudHintLabel     = "5 to toggle state label"
)

var hudHintColor = color.RGBA{0xff, 0xff, 0xff, 0xff}
var hudHintActiveColor = color.RGBA{0xff, 0xd0, 0x3a, 0xff}

func hudStyle() text.TextStyle {
	return text.TextStyle{
		FontName:    "mars",
		ScaleX:      0.25,
		ScaleY:      0.25,
		FontHeight:  130,
		LineSpacing: 10,
		Color:       hudHintColor,
		CharWidths:  map[string]int{" ": 60},
	}
}

func NewDemoScene(mapPath string, mapScale float64) (*demoScene, error) {
	m, err := tiled.LoadMapSpecScaled(mapPath, tiled.DefaultTilesetCache, mapScale)
	if err != nil {
		return nil, err
	}

	s := &demoScene{
		Map:           m,
		BgColor:       color.Black,
		ShowColliders: false,
		ShowZones:     false,
		input:         engine.NewInput(),
		color:         color.RGBA{0xff, 0x3a, 0x3a, 0xff},
		speed:         160,
		facing:        "down",
	}

	const pw, ph = 42, 24
	s.player = &tiled.ColliderBox{X: 80, Y: 200, W: pw, H: ph}

	if p, ok := m.Point("start"); ok {
		s.player.X = p.X - pw/2
		s.player.Y = p.Y - ph/2
	}

	cam := engine.NewCamera()
	cam.SetBounds(float64(m.Width()), float64(m.Height()))
	cam.SetViewport(640, 480)
	cam.CenterOn(s.player.X+pw/2, s.player.Y+ph/2)
	s.Camera = cam

	hs := hudStyle()
	s.hintSmooth = text.NewTextString(hudHintFollow, 20, 40, hs)
	s.hintSmooth.Show()
	hs2 := hudStyle()
	s.hintColliders = text.NewTextString(hudHintColliders, 20, 0, hs2)
	s.hintColliders.Show()
	hs3 := hudStyle()
	s.hintZones = text.NewTextString(hudHintZones, 20, 20, hs3)
	s.hintZones.Show()
	hs4 := hudStyle()
	s.hintArrow = text.NewTextString(hudHintArrow, 20, 60, hs4)
	s.hintArrow.Show()
	hs5 := hudStyle()
	s.hintLabel = text.NewTextString(hudHintLabel, 20, 80, hs5)
	s.hintLabel.Show()

	s.stateLabel = text.NewTextString("down", 0, 0, text.TextStyle{
		FontName:    "mars",
		ScaleX:      0.2,
		ScaleY:      0.2,
		FontHeight:  179,
		LineSpacing: 4,
		Color:       color.White,
	})

	return s, nil
}

func (s *demoScene) Update(dt time.Duration) {
	s.input.Update()
	queues.DefaultUpdateQueue.Execute(dt)

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

	if text.DefaultDialog.Active() {
		text.DefaultDialog.Update(dt)
		s.followCamera()
		return
	}

	dx, dy := 0.0, 0.0
	if s.input.IsPressed(ebiten.KeyLeft) {
		dx -= 1
	}
	if s.input.IsPressed(ebiten.KeyRight) {
		dx += 1
	}
	if s.input.IsPressed(ebiten.KeyUp) {
		dy -= 1
	}
	if s.input.IsPressed(ebiten.KeyDown) {
		dy += 1
	}

	if s.input.JustPressed(ebiten.Key1) {
		s.ShowColliders = !s.ShowColliders
		s.setHintColor(s.hintColliders, s.ShowColliders)
	}
	if s.input.JustPressed(ebiten.Key2) {
		s.ShowZones = !s.ShowZones
		s.setHintColor(s.hintZones, s.ShowZones)
	}
	if s.input.JustPressed(ebiten.Key4) {
		s.ShowCamArrow = !s.ShowCamArrow
		s.setHintColor(s.hintArrow, s.ShowCamArrow)
	}
	if s.input.JustPressed(ebiten.Key5) {
		s.ShowStateLabel = !s.ShowStateLabel
		s.setHintColor(s.hintLabel, s.ShowStateLabel)
	}
	if s.input.JustPressed(ebiten.Key3) {
		s.setSmoothFollow(!s.Camera.Smooth())
	}

	seconds := dt.Seconds()
	solids := s.Map.Colliders()

	if dx != 0 || dy != 0 {
		s.moveAxis(0, dx*s.speed*seconds, solids)
		s.moveAxis(1, dy*s.speed*seconds, solids)
	}

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
		if state != s.lastState {
			s.lastState = state
			s.stateLabel.SetText(state)
		}
	}

	if s.input.JustPressed(ebiten.KeyZ) || s.input.JustPressed(ebiten.KeyEnter) {
		if zone := s.facedZone(); zone != nil {
			if d, ok := demoDialogs[zone.Name]; ok {
				for _, phrase := range d.phrases {
					text.DefaultDialog.Show(phrase, dialogStyle, d.sound)
				}
			}
		}
	}

	s.followCamera()
}

func (s *demoScene) facedZone() *tiled.InteractionZone {
	const reach = 12
	p := s.player
	probe := tiled.ColliderBox{X: p.X, Y: p.Y, W: p.W, H: p.H}
	switch s.facing {
	case "up":
		probe.Y -= reach
		probe.H = reach
	case "down":
		probe.Y += p.H
		probe.H = reach
	case "left":
		probe.X -= reach
		probe.W = reach
	case "right":
		probe.X += p.W
		probe.W = reach
	}
	for _, zone := range s.Map.Interactions() {
		if zone.Rect.Overlaps(&probe) {
			return zone
		}
	}
	return nil
}

func (s *demoScene) followCamera() {
	s.Camera.CenterOn(s.player.X+s.player.W/2, s.player.Y+s.player.H/2)
}

// setSmoothFollow toggles tweened camera transitions and recolors the follow
// hint yellow while smooth mode is active.
func (s *demoScene) setSmoothFollow(on bool) {
	s.Camera.SetSmooth(on)
	s.setHintColor(s.hintSmooth, on)
}

// setHintColor highlights a debug hint yellow while its mode is active. The
// hint image is baked at load time, so recoloring forces a re-bake via
// SetText; unchanged colors are skipped.
func (s *demoScene) setHintColor(ts *text.TextString, active bool) {
	clr := color.Color(hudHintColor)
	if active {
		clr = color.Color(hudHintActiveColor)
	}
	if ts.Style.Color == clr {
		return
	}
	ts.Style.Color = clr
	ts.SetText(ts.Text)
}

func (s *demoScene) moveAxis(axis int, delta float64, solids []*tiled.ColliderBox) {
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

func (s *demoScene) Draw(screen *ebiten.Image) {
	screen.Fill(s.BgColor)
	s.Camera.Apply(screen, func(target *ebiten.Image, sx, sy float64) {
		s.Map.DrawLayers(target, sx, sy)

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
				s.sprite.Draw(target, x, y, m, m, 0)
			}
		} else {
			vector.DrawFilledRect(target,
				float32(b.X+sx), float32(b.Y+sy), float32(b.W), float32(b.H),
				s.color, true)
		}

		if s.ShowStateLabel && s.stateLabel != nil {
			s.stateLabel.SetPosition(b.X+b.W+6+sx, b.Y-50+sy)
			s.stateLabel.Draw(target)
		}

		if s.ShowZones {
			for _, z := range s.Map.Interactions() {
				clr := color.RGBA{0x3a, 0xb8, 0xff, 0x60}
				if z.Trigger == tiled.TriggerButton {
					clr = color.RGBA{0xff, 0xd0, 0x3a, 0x60}
				}
				vector.DrawFilledRect(target,
					float32(z.Rect.X+sx), float32(z.Rect.Y+sy), float32(z.Rect.W), float32(z.Rect.H),
					clr, false)
			}
		}

		if s.ShowColliders {
			for _, c := range s.Map.Colliders() {
				vector.StrokeRect(target,
					float32(c.X+sx), float32(c.Y+sy), float32(c.W), float32(c.H),
					2, color.RGBA{0xff, 0x40, 0x40, 0xc0}, true)
			}
		}

		if s.ShowCamArrow {
			s.drawCamArrow(target, sx, sy)
		}
	})

	if text.DefaultDialog.Active() {
		drawTextbox(screen)
	}
	queues.DefaultQueue.Execute(screen)
}

// drawCamArrow renders a debug arrow from the camera center (the true
// on-screen center, kept clamped by the camera itself) to the player's hitbox
// center, plus a cross marking that point.
func (s *demoScene) drawCamArrow(target *ebiten.Image, sx, sy float64) {
	const (
		stroke = 2
		head   = 10.0
		cross  = 6.0
	)
	clr := color.RGBA{0x40, 0xff, 0x40, 0xff}

	fromX, fromY := s.Camera.X+sx, s.Camera.Y+sy
	toX, toY := s.player.X+s.player.W/2+sx, s.player.Y+s.player.H/2+sy

	vector.StrokeLine(target,
		float32(fromX-cross), float32(fromY), float32(fromX+cross), float32(fromY),
		stroke, clr, true)
	vector.StrokeLine(target,
		float32(fromX), float32(fromY-cross), float32(fromX), float32(fromY+cross),
		stroke, clr, true)

	vector.StrokeLine(target,
		float32(fromX), float32(fromY), float32(toX), float32(toY),
		stroke, clr, true)

	dx, dy := toX-fromX, toY-fromY
	if dx*dx+dy*dy > head*head {
		angle := math.Atan2(dy, dx)
		for _, spread := range []float64{math.Pi - 0.5, math.Pi + 0.5} {
			vector.StrokeLine(target,
				float32(toX), float32(toY),
				float32(toX+head*math.Cos(angle+spread)),
				float32(toY+head*math.Sin(angle+spread)),
				stroke, clr, true)
		}
	}
}
