package scene

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/pkg/engine/queues"
)

type GameScene struct {
	BgColor color.Color
}

func NewGameScene() *GameScene {
	return &GameScene{
		BgColor: color.RGBA{0x00, 0x00, 0x00, 0xff},
	}
}

func (s *GameScene) Update(dt time.Duration) {
	queues.DefaultUpdateQueue.Execute(dt)
}

func (s *GameScene) Draw(screen *ebiten.Image) {
	screen.Fill(s.BgColor)
	queues.DefaultQueue.Execute(screen)
}
