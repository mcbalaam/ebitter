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

func (s *GameScene) Update(dt time.Duration) {
	queues.DefaultUpdateQueue.Execute(dt)
}

func (s *GameScene) Draw(screen *ebiten.Image) {
	if s.BgColor == nil {
		s.BgColor = color.RGBA{0x00, 0x00, 0x00, 0xff}
	}
	screen.Fill(s.BgColor)
	queues.DefaultQueue.Execute(screen)
}
