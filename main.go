package main

import (
	"embed"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/pkg/embedfs"
	"github.com/mcbalaam/ebitter/pkg/engine"
	"github.com/mcbalaam/ebitter/pkg/engine/queues"
	"github.com/mcbalaam/ebitter/pkg/render"
	"github.com/mcbalaam/ebitter/pkg/sound"
)

//go:embed media
var mediaFS embed.FS

func init() {
	embedfs.SetFS(mediaFS)
}

const version = "v0.98"

type SplashText struct {
	image     *ebiten.Image
	elapsed   float64
	centerX   float64
	centerY   float64
	baseScale float64
}

func NewSplashText(fontPath, text string, centerX, centerY, baseScale float64) (*SplashText, error) {
	icon, err := render.NewAnimatedIconFromPath(fontPath, " ")
	if err != nil {
		return nil, err
	}

	spacing := 10.0
	x := 0.0
	maxH := 0

	type glyphInfo struct {
		img *ebiten.Image
		x   float64
	}
	var glyphs []glyphInfo

	for _, r := range text {
		char := string(r)
		if err := icon.SetIconState(char); err != nil {
			continue
		}
		frame := icon.CurrentState.Frames[icon.CurrentState.CurrentFrame]
		h := frame.Image.Bounds().Dy()
		if h > maxH {
			maxH = h
		}
		glyphs = append(glyphs, glyphInfo{img: frame.Image, x: x})
		x += float64(frame.Image.Bounds().Dx()) + spacing
	}

	totalW := int(math.Ceil(x))
	off := ebiten.NewImage(totalW, maxH)
	for _, g := range glyphs {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(g.x, 0)
		off.DrawImage(g.img, op)
	}

	return &SplashText{
		image:     off,
		centerX:   centerX,
		centerY:   centerY,
		baseScale: baseScale,
	}, nil
}

func (s *SplashText) Update(dt time.Duration) {
	s.elapsed += dt.Seconds()
}

func (s *SplashText) Draw(screen *ebiten.Image) {
	pulse := 1.0 + math.Sin(s.elapsed*3)*0.1
	scale := s.baseScale * pulse
	tilt := math.Sin(s.elapsed*2) * 0.15

	w := float64(s.image.Bounds().Dx())
	h := float64(s.image.Bounds().Dy())

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(-w/2, -h/2)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Rotate(tilt)
	op.GeoM.Translate(s.centerX, s.centerY)
	screen.DrawImage(s.image, op)
}

type Game struct {
	sm *engine.SceneManager
}

func (g *Game) Update() error {
	g.sm.Update(time.Second / 60)
	queues.DefaultDeleteQueue.Execute()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.sm.Draw(screen)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return 640, 480
}

func main() {
	player, err := sound.NewSoundPlayer(44100)
	if err != nil {
		log.Fatalf("sound: %v", err)
	}
	if err := player.RegisterNewSound("media/sound/rare.wav", "bgm"); err != nil {
		log.Printf("bgm: %v", err)
	} else {
		player.PlayBackground("bgm", 1)
	}

	splash, err := NewSplashText("media/sprites/determination", "ebitter   "+version, 320, 240, 0.4)
	if err != nil {
		log.Fatalf("splash: %v", err)
	}
	queues.DefaultQueue.ScheduleAt(splash, queues.LayerOverlay)
	queues.DefaultUpdateQueue.Schedule(splash)

	textEngine := &engine.TextEngine{FontsLoaded: make(map[string]render.AnimatedIcon)}
	cornerStyle := engine.TextStyle{
		FontName:     "determination",
		StartX:       10,
		StartY:       430,
		ScaleX:       0.3,
		ScaleY:       0.3,
		FontHeight:   24,
		DefaultDelay: 0.03,
		Instant:      true,
	}
	textEngine.DisplayText(cornerStyle, "playable demo soon", nil, nil)

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Ebitter")
	game := &Game{sm: &engine.SceneManager{}}
	game.sm.Push(&engine.GameScene{})
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
