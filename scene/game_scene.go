package scene

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type GameScene struct {
	BgColor color.Color
	runtime *Runtime
}

func NewGameScene() *GameScene {
	return &GameScene{
		BgColor: color.Black,
		runtime: NewRuntime(),
	}
}

func (s *GameScene) Runtime() *Runtime {
	if s.runtime == nil {
		s.runtime = NewRuntime()
	}
	return s.runtime
}

func (s *GameScene) Pause() {
	s.Runtime().Queues.Pause()
}

func (s *GameScene) Resume() {
	s.Runtime().Queues.Resume()
}

func (s *GameScene) Update(dt time.Duration) {
	s.Runtime().Queues.ExecuteUpdate(dt)
	s.Runtime().Queues.ExecuteDelete()
	if s.Runtime().Dialog.Active() {
		s.Runtime().Dialog.Update(dt)
	}
}

func (s *GameScene) Draw(screen *ebiten.Image) {
	screen.Fill(s.BgColor)
	s.Runtime().Queues.ExecuteDraw(screen)
}
