package main

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter"
	"github.com/mcbalaam/ebitter/assetfs"
	"github.com/mcbalaam/ebitter/queues"
	"github.com/mcbalaam/ebitter/scene"
)

func init() {
	assetfs.SetFS(ebitter.Media)
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
	mapScene, err := NewDemoScene("media/maps/ruins.tmx", 2)
	if err != nil {
		log.Fatalf("map: %v", err)
	}
	game.sm.Push(mapScene)

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
