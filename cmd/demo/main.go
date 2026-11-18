package main

import (
	"embed"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/assetfs"
	mainscene "github.com/mcbalaam/ebitter/cmd/demo/scenes/main"
	"github.com/mcbalaam/ebitter/components"
	"github.com/mcbalaam/ebitter/queues"
	"github.com/mcbalaam/ebitter/scene"
	"github.com/mcbalaam/ebitter/text"
)

//go:embed media prototypes
var demoMedia embed.FS

//go:generate go run ../fontgen -fonts media/fonts -output media/sprites

func init() {
	assetfs.SetFS(demoMedia)
	components.SpriteRoot = "media/sprites"
	text.FontRoot = "media/sprites"
}

type Game struct {
	sm *scene.SceneManager
}

func (g *Game) Update() error {
	g.sm.Update(time.Second / 60)
	queues.DefaultDeleteQueue.Execute()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.sm.Draw(screen)
}

// DrawFinalScreen uses nearest-neighbor scaling for the 640x480 logical image.
// This keeps the 2x window upscale crisp instead of applying Ebitengine's
// default box filter.
func (g *Game) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geoM ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{GeoM: geoM, Filter: ebiten.FilterNearest}
	screen.DrawImage(offscreen, op)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return 640, 480
}

func main() {
	ebiten.SetWindowSize(960, 720)
	ebiten.SetWindowTitle("Undebitter")

	game := &Game{sm: &scene.SceneManager{}}
	mapScene, err := mainscene.NewDemoScene()
	if err != nil {
		log.Fatalf("scene: %v", err)
	}
	game.sm.Push(mapScene)

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
