package main

import (
	"embed"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/assetfs"
	"github.com/mcbalaam/ebitter/cmd/demo/scenes"
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

func (g *Game) Layout(_, _ int) (int, int) {
	return 640, 480
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Undebitter")

	game := &Game{sm: &scene.SceneManager{}}
	mapScene, err := scenes.NewDemoScene("media/maps/ruins.tmx", 2)
	if err != nil {
		log.Fatalf("map: %v", err)
	}
	game.sm.Push(mapScene)

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
