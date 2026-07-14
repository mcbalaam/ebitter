package main

import (
	"embed"
	"image/color"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/pkg/embedfs"
	"github.com/mcbalaam/ebitter/pkg/engine"
	"github.com/mcbalaam/ebitter/pkg/engine/queues"
	"github.com/mcbalaam/ebitter/pkg/sound"
)

//go:embed media
var mediaFS embed.FS

func init() {
	embedfs.SetFS(mediaFS)
}

const version = "v0.98"

type Game struct {
	sm     *engine.SceneManager
	splash *engine.TextString
}

func (g *Game) Update() error {
	if g.splash != nil && !g.splash.Visible {
		g.splash = nil
	}
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

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Ebitter")
	game := &Game{sm: &engine.SceneManager{}}
	game.sm.Push(&engine.GameScene{})

	splashStyle := engine.TextStyle{
		FontName:    "determination",
		ScaleX:      0.1,
		ScaleY:      0.1,
		FontHeight:  24,
		CharSpacing: 8,
		Color:       color.RGBA{255, 255, 255, 255},
	}
	var splashElapsed float64
	splash := engine.NewTextString("ebitter  "+version, 320, 240, splashStyle)
	splash.Centered = true
	splash.UpdateFunc = func(ts *engine.TextString, dt time.Duration) {
		splashElapsed += dt.Seconds()
		pulse := 1.0 + math.Sin(splashElapsed*3)*0.1
		ts.Style.ScaleX = 0.4 * pulse
		ts.Style.ScaleY = 0.4 * pulse
		ts.Rotation = math.Sin(splashElapsed*2) * 0.15
	}
	splash.Show()
	game.splash = splash

	cornerStyle := engine.TextStyle{
		FontName:    "determination",
		ScaleX:      0.3,
		ScaleY:      0.3,
		FontHeight:  24,
		CharSpacing: 8,
		Color:       color.RGBA{255, 255, 255, 255},
	}
	corner := engine.NewTextString("github.com/mcbalaam/ebitter", 10, 430, cornerStyle)
	corner.Show()
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
