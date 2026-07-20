// Package scene provides minimal scene and scene manager implementation
package scene

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Scene interface {
	Update(dt time.Duration)
	Draw(screen *ebiten.Image)
}

type SceneManager struct {
	stack []Scene
}

func (sm *SceneManager) Push(s Scene) {
	sm.stack = append(sm.stack, s)
}

func (sm *SceneManager) Pop() {
	if len(sm.stack) > 0 {
		sm.stack = sm.stack[:len(sm.stack)-1]
	}
}

func (sm *SceneManager) Top() Scene {
	if len(sm.stack) == 0 {
		return nil
	}
	return sm.stack[len(sm.stack)-1]
}

func (sm *SceneManager) Update(dt time.Duration) {
	if s := sm.Top(); s != nil {
		s.Update(dt)
	}
}

func (sm *SceneManager) Draw(screen *ebiten.Image) {
	if s := sm.Top(); s != nil {
		s.Draw(screen)
	}
}
