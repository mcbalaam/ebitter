package main

import (
	"embed"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/pkg/embedfs"
	"github.com/mcbalaam/ebitter/pkg/engine/queues"
	"github.com/mcbalaam/ebitter/pkg/engine/scene"
	"github.com/mcbalaam/ebitter/pkg/engine/text"
	"github.com/mcbalaam/ebitter/pkg/render"
	"github.com/mcbalaam/ebitter/pkg/sound"
)

//go:embed media
var mediaFS embed.FS

func init() {
	embedfs.SetFS(mediaFS)
}

const version = "v0.98"

type Game struct {
	sm        *scene.SceneManager
	splash    *text.TextString
	pressIcon *render.AnimatedIcon
}

func (g *Game) Update() error {
	if g.splash != nil && !g.splash.Visible {
		g.splash = nil
	}
	g.sm.Update(time.Second / 60)
	text.DefaultDialog.Update(time.Second / 60)
	if g.pressIcon != nil {
		g.pressIcon.Update(time.Second / 60)
	}
	queues.DefaultDeleteQueue.Execute()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.sm.Draw(screen)
	if g.pressIcon != nil && text.DefaultDialog.RevealedAll() {
		g.pressIcon.Draw(screen, 380, 220, 1, 1, 0)
	}
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

	player.RegisterNewSound("media/sound/snd_text.wav", "text")
	text.DefaultDialog.SoundPlayer = player

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Ebitter")
	pi, err := render.NewAnimatedIconFromPath("media/sprites/press_z", "idle")
	if err != nil {
		log.Printf("press_z: %v", err)
	}
	game := &Game{sm: &scene.SceneManager{}, pressIcon: pi}
	game.sm.Push(&scene.GameScene{})

	splashStyle := text.TextStyle{
		FontName:    "determination",
		ScaleX:      0.4,
		ScaleY:      0.4,
		FontHeight:  24,
		CharSpacing: 8,
	}
	var splashElapsed float64
	splash := text.NewTextString("ebitter  "+version, 320, 240, splashStyle)
	splash.Centered = true
	splash.UpdateFunc = func(ts *text.TextString, dt time.Duration) {
		splashElapsed += dt.Seconds()
		pulse := 1.0 + math.Sin(splashElapsed*3)*0.1
		ts.Style.ScaleX = 0.4 * pulse
		ts.Style.ScaleY = 0.4 * pulse
		ts.Rotation = math.Sin(splashElapsed*2) * 0.15
	}
	splash.Show()
	game.splash = splash

	cornerStyle := text.TextStyle{
		FontName:    "determination",
		ScaleX:      0.3,
		ScaleY:      0.3,
		FontHeight:  24,
		CharSpacing: 8,
	}
	corner := text.NewTextString("github.com/mcbalaam/ebitter", 10, 430, cornerStyle)
	corner.Show()

	dialogStyle := text.TextStyle{
		FontName:     "determination",
		ScaleX:       0.3,
		ScaleY:       0.3,
		StartX:       20,
		StartY:       0,
		FontHeight:   24,
		LineSpacing:  90,
		DefaultDelay: 0.08,
		CharSpacing:  2,
	}
	text.DefaultDialog.Show(
		"Heya.$p600 $ce5d20dmcbalaam$cffffff speaking.$n$p600The engine is in beta,$p200 and$nI don't have a cool demo yet.$e",
		dialogStyle, "text")
	text.DefaultDialog.Show(
		"Stay tuned!$p600 Check out the $n$ce5d20ddocumentation$cffffff on pkg.go.dev in$nthe meanwhile.$e",
		dialogStyle, "text")
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
