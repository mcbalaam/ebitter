package main

import (
	"embed"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/pkg/embedfs"
	"github.com/mcbalaam/ebitter/pkg/engine/queues"
	"github.com/mcbalaam/ebitter/pkg/engine/scene"
)

//go:embed media
var mediaFS embed.FS

func init() {
	embedfs.SetFS(mediaFS)
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
	ebiten.SetWindowTitle("Ebitter")

	game := &Game{sm: &scene.SceneManager{}}
	mapScene, err := scene.NewTiledMapScene("media/maps/ruins.tmx", 2)
	if err != nil {
		log.Fatalf("map: %v", err)
	}
	game.sm.Push(mapScene)

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
