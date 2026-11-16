package scenes

import (
	"fmt"
	"image/color"
	"image/png"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/mcbalaam/ebitter/assetfs"
	democomponents "github.com/mcbalaam/ebitter/cmd/demo/components"
	"github.com/mcbalaam/ebitter/components"
	"github.com/mcbalaam/ebitter/queues"
	"github.com/mcbalaam/ebitter/render"
	"github.com/mcbalaam/ebitter/scene"
	"github.com/mcbalaam/ebitter/tween"
)

type demoScene struct {
	runtime        *scene.Runtime
	preview        *ebiten.Image
	buttons        map[string]*components.Entity
	buttonNames    []string
	selectedButton int
	menuActive     bool
	menuConfirmed  bool
	soul           *render.AnimatedIcon
	soulX, soulY   float64
	input          *demoInput
}

type demoInput struct {
	pressed   map[ebiten.Key]bool
	justPress map[ebiten.Key]bool
}

func newDemoInput() *demoInput {
	return &demoInput{pressed: make(map[ebiten.Key]bool), justPress: make(map[ebiten.Key]bool)}
}

func (i *demoInput) Update() {
	clear(i.justPress)
	for _, key := range inpututil.AppendJustPressedKeys(nil) {
		i.justPress[key] = true
	}
	for _, key := range []ebiten.Key{
		ebiten.KeyArrowLeft, ebiten.KeyArrowRight, ebiten.KeyArrowUp, ebiten.KeyArrowDown,
		ebiten.KeyZ, ebiten.KeyX,
	} {
		i.pressed[key] = ebiten.IsKeyPressed(key)
	}
}

func (i *demoInput) justPressed(key ebiten.Key) bool {
	return i.justPress[key]
}

type BattleBox struct {
	corners [4]tween.Vec2 // top-left, top-right, bottom-right, bottom-left
	stroke  float32
	color   color.Color
	updates *queues.UpdateQueue
	resize  *tween.Tween[[4]tween.Vec2]
}

func NewBattleBox(x, y, width, height float64, updates *queues.UpdateQueue) *BattleBox {
	return &BattleBox{
		corners: [4]tween.Vec2{{X: x, Y: y}, {X: x + width, Y: y}, {X: x + width, Y: y + height}, {X: x, Y: y + height}},
		stroke:  5,
		color:   color.White,
		updates: updates,
	}
}

func (b *BattleBox) Type() string { return "battleBox" }

// Resize tweens all four corners from their current positions to the supplied
// top-left, top-right, bottom-right, and bottom-left positions.
func (b *BattleBox) Resize(topLeft, topRight, bottomRight, bottomLeft tween.Vec2, duration time.Duration) {
	if b.resize != nil {
		tween.Stop(b.resize)
	}
	to := [4]tween.Vec2{topLeft, topRight, bottomRight, bottomLeft}
	if duration <= 0 {
		b.corners = to
		return
	}
	from := b.corners
	lerp := func(a, z [4]tween.Vec2, t float64) [4]tween.Vec2 {
		var out [4]tween.Vec2
		for i := range out {
			out[i] = tween.Vec2{X: a[i].X + (z[i].X-a[i].X)*t, Y: a[i].Y + (z[i].Y-a[i].Y)*t}
		}
		return out
	}
	b.resize = tween.New(from, to, duration, func(corners [4]tween.Vec2) {
		b.corners = corners
	}, lerp).SetEasing(tween.EaseInOutCubic).ScheduleOn(b.updates)
}

func (b *BattleBox) Draw(_ *components.Entity, screen *ebiten.Image) {
	// Draw axis-aligned boxes as four joined filled strips. Unlike a stroked
	// path, this produces square, pixel-aligned joins at every corner.
	if b.corners[0].Y == b.corners[1].Y &&
		b.corners[1].X == b.corners[2].X &&
		b.corners[2].Y == b.corners[3].Y &&
		b.corners[3].X == b.corners[0].X {
		left, top := b.corners[0].X, b.corners[0].Y
		right, bottom := b.corners[2].X, b.corners[2].Y
		width, height := right-left, bottom-top
		stroke := float64(b.stroke)
		vector.DrawFilledRect(screen, float32(left), float32(top), float32(width), float32(stroke), b.color, false)
		vector.DrawFilledRect(screen, float32(left), float32(bottom-stroke), float32(width), float32(stroke), b.color, false)
		vector.DrawFilledRect(screen, float32(left), float32(top+stroke), float32(stroke), float32(height-2*stroke), b.color, false)
		vector.DrawFilledRect(screen, float32(right-stroke), float32(top+stroke), float32(stroke), float32(height-2*stroke), b.color, false)
		return
	}

	var outline vector.Path
	outline.MoveTo(float32(b.corners[0].X), float32(b.corners[0].Y))
	for _, point := range b.corners[1:] {
		outline.LineTo(float32(point.X), float32(point.Y))
	}
	outline.Close()
	vector.StrokePath(screen, &outline,
		&vector.StrokeOptions{Width: b.stroke, LineJoin: vector.LineJoinMiter},
		&vector.DrawPathOptions{AntiAlias: false})
}

func NewDemoScene() (*demoScene, error) {
	s := &demoScene{
		runtime: scene.NewRuntime(), buttons: make(map[string]*components.Entity),
		buttonNames: []string{"fightButton", "actButton", "itemButton", "mercyButton"},
		input:       newDemoInput(),
	}
	previewFile, err := assetfs.FS.Open("media/sprites/preview.png")
	if err != nil {
		return nil, fmt.Errorf("open interface preview: %w", err)
	}
	previewSource, decodeErr := png.Decode(previewFile)
	closeErr := previewFile.Close()
	if decodeErr != nil {
		return nil, fmt.Errorf("decode interface preview: %w", decodeErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close interface preview: %w", closeErr)
	}
	s.preview = ebiten.NewImageFromImage(previewSource)
	s.preview.Clear()
	s.soul, err = render.NewAnimatedIconFromPath("media/sprites/soul", "idle")
	if err != nil {
		return nil, fmt.Errorf("load soul sprite: %w", err)
	}
	s.selectedButton = 0

	box := components.NewEntity("battleBox")
	box.Layer = queues.LayerArena
	box.Queues = s.runtime.Queues
	box.Add(NewBattleBox(32, 250, 576, 140, s.runtime.Queues.Update))
	box.Schedule()

	buttonLibrary := components.NewBlueprintLibrary(nil, "prototypes")
	buttons := []string{"actButton", "fightButton", "itemButton", "mercyButton"}
	for _, button := range buttons {
		blueprint, err := buttonLibrary.Load(button)
		if err != nil {
			return nil, fmt.Errorf("load %s prototype: %w", button, err)
		}
		entity, err := blueprint.InstantiateWith(s.runtime.Signals, s.runtime.Queues)
		if err != nil {
			return nil, fmt.Errorf("instantiate %s: %w", button, err)
		}
		entity.Schedule()
		s.buttons[button] = entity
	}
	s.syncButtonSelection()
	return s, nil
}

// SetButtonActive changes one of the prototype-created buttons via its signal.
// Names are actButton, fightButton, itemButton, and mercyButton.
func (s *demoScene) SetButtonActive(name string, active bool) error {
	button, ok := s.buttons[name]
	if !ok {
		return fmt.Errorf("battle button %q not found", name)
	}
	democomponents.SetBattleButtonActive(button, active)
	return nil
}

func (s *demoScene) Runtime() *scene.Runtime { return s.runtime }
func (s *demoScene) Pause()                  { s.runtime.Queues.Pause() }
func (s *demoScene) Resume()                 { s.runtime.Queues.Resume() }

func (s *demoScene) Update(dt time.Duration) {
	s.input.Update()
	if s.runtime.Dialog.Active() {
		s.runtime.Dialog.Update(dt)
		return
	}
	s.updateSoul(dt)
	s.runtime.Queues.ExecuteUpdate(dt)
	s.runtime.Queues.ExecuteDelete()
}

func (s *demoScene) updateSoul(dt time.Duration) {
	const (
		buttonY   = 475.0
		boxLeft   = 39.0
		boxTop    = 257.0
		boxRight  = 601.0
		boxBottom = 383.0
		speed     = 160.0
	)
	if s.soul == nil {
		return
	}
	if !s.menuActive {
		if s.input.justPressed(ebiten.KeyArrowLeft) {
			s.selectedButton--
		}
		if s.input.justPressed(ebiten.KeyArrowRight) {
			s.selectedButton++
		}
		if s.selectedButton < 0 {
			s.selectedButton = len(s.buttonNames) - 1
		}
		if s.selectedButton >= len(s.buttonNames) {
			s.selectedButton = 0
		}
		button := s.buttons[s.buttonNames[s.selectedButton]]
		if button != nil && button.Transform() != nil {
			s.soulX = button.Transform().X + 16
			s.soulY = buttonY - 22
		}
		s.syncButtonSelection()
		if s.input.justPressed(ebiten.KeyZ) {
			s.menuActive = true
			s.menuConfirmed = false
			s.soulX = boxLeft + 18
			s.soulY = (boxTop + boxBottom) / 2
		}
	} else {
		delta := speed * dt.Seconds()
		if s.input.pressed[ebiten.KeyArrowLeft] || s.input.justPressed(ebiten.KeyArrowLeft) {
			s.soulX -= delta
		}
		if s.input.pressed[ebiten.KeyArrowRight] || s.input.justPressed(ebiten.KeyArrowRight) {
			s.soulX += delta
		}
		if s.input.pressed[ebiten.KeyArrowUp] || s.input.justPressed(ebiten.KeyArrowUp) {
			s.soulY -= delta
		}
		if s.input.pressed[ebiten.KeyArrowDown] || s.input.justPressed(ebiten.KeyArrowDown) {
			s.soulY += delta
		}
		s.soulX = math.Max(boxLeft+8, math.Min(boxRight-8, s.soulX))
		s.soulY = math.Max(boxTop+8, math.Min(boxBottom-8, s.soulY))
		if s.input.justPressed(ebiten.KeyZ) {
			s.menuConfirmed = true
		}
		if s.input.justPressed(ebiten.KeyX) {
			s.menuActive = false
			s.menuConfirmed = false
			s.syncButtonSelection()
		}
	}
	s.soul.Update(dt)
}

func (s *demoScene) syncButtonSelection() {
	for i, name := range s.buttonNames {
		if button := s.buttons[name]; button != nil {
			democomponents.SetBattleButtonActive(button, i == s.selectedButton && !s.menuActive)
		}
	}
}

func (s *demoScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	s.runtime.Queues.ExecuteDraw(screen)
	if s.preview != nil {
		op := &ebiten.DrawImageOptions{}
		op.ColorScale.ScaleAlpha(0.3)
		screen.DrawImage(s.preview, op)
	}
	if s.soul != nil {
		s.soul.Draw(screen, s.soulX-8, s.soulY-8, 1, 1, 0)
	}
}
