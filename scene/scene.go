// Package scene provides minimal scene and scene manager implementation
package scene

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Scene interface {
	Update(dt time.Duration)
	Draw(screen *ebiten.Image)
}

const defaultFadeDuration = 400 * time.Millisecond

type fadePhase int

const (
	fadeNone fadePhase = iota
	fadeOut
	fadeIn
)

// SceneManager is a stack of scenes. The top scene is the active one.
// Scene switches can be wrapped in a fade-to-black transition: the screen
// fades out, the stack is mutated at full black, then fades back in. The
// manager also fades in from black automatically when the first scene
// appears.
type SceneManager struct {
	// FadeDuration is the length of a single fade-out or fade-in.
	// The zero value uses a default.
	FadeDuration time.Duration

	// FadeColor is the color transitions fade through. Nil means black.
	FadeColor color.Color

	stack   []Scene
	started bool

	phase      fadePhase
	fadeAlpha  float64 // 0 = fully clear, 1 = fully black
	pending    Scene
	pendingPop bool
}

func (sm *SceneManager) Push(s Scene) {
	if current := sm.Top(); current != nil {
		pause(current)
	}
	sm.stack = append(sm.stack, s)
	resume(s)
}

func (sm *SceneManager) Pop() {
	if len(sm.stack) > 0 {
		pause(sm.stack[len(sm.stack)-1])
		sm.stack = sm.stack[:len(sm.stack)-1]
		resume(sm.Top())
	}
}

func (sm *SceneManager) Top() Scene {
	if len(sm.stack) == 0 {
		return nil
	}
	return sm.stack[len(sm.stack)-1]
}

// IsTransitioning reports whether a fade transition is currently playing.
func (sm *SceneManager) IsTransitioning() bool {
	return sm.phase != fadeNone
}

// SwitchTo fades the screen to black, replaces the current scene with s and
// fades back in. With an empty stack s is simply pushed (the startup fade-in
// still applies).
func (sm *SceneManager) SwitchTo(s Scene) {
	if s == nil {
		return
	}
	if sm.phase == fadeNone && sm.Top() == nil {
		sm.Push(s)
		return
	}
	sm.pending, sm.pendingPop = s, false
	if sm.phase != fadeOut {
		sm.phase = fadeOut
	}
}

// PopFaded fades the screen to black, pops the current scene and fades back
// in to the scene underneath.
func (sm *SceneManager) PopFaded() {
	if len(sm.stack) == 0 {
		return
	}
	sm.pending, sm.pendingPop = nil, true
	if sm.phase != fadeOut {
		sm.phase = fadeOut
	}
}

func (sm *SceneManager) Update(dt time.Duration) {
	// Fade in from black when the first scene appears.
	if !sm.started && sm.Top() != nil {
		sm.started = true
		sm.fadeAlpha = 1
		sm.phase = fadeIn
	}

	step := dt.Seconds() / sm.fadeDur().Seconds()
	switch sm.phase {
	case fadeOut:
		sm.fadeAlpha += step
		if sm.fadeAlpha >= 1 {
			sm.fadeAlpha = 1
			sm.applyPending()
			sm.phase = fadeIn
		}
	case fadeIn:
		sm.fadeAlpha -= step
		if sm.fadeAlpha <= 0 {
			sm.fadeAlpha = 0
			sm.phase = fadeNone
		}
	}

	// The outgoing scene is frozen while fading out. During fade-in the new
	// scene already updates so it can warm up (e.g. build its layers) behind
	// the fading overlay.
	if sm.phase != fadeOut {
		if s := sm.Top(); s != nil {
			s.Update(dt)
		}
	}
}

func (sm *SceneManager) Draw(screen *ebiten.Image) {
	if s := sm.Top(); s != nil {
		s.Draw(screen)
	} else {
		screen.Fill(color.Black)
	}

	if sm.fadeAlpha > 0 {
		if a := uint8(sm.fadeAlpha * 255); a > 0 {
			fade := sm.FadeColor
			if fade == nil {
				fade = color.Black
			}
			r, g, b, _ := fade.RGBA()
			bounds := screen.Bounds()
			vector.DrawFilledRect(screen, 0, 0, float32(bounds.Dx()), float32(bounds.Dy()),
				color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: a}, false)
		}
	}
}

func (sm *SceneManager) fadeDur() time.Duration {
	if sm.FadeDuration > 0 {
		return sm.FadeDuration
	}
	return defaultFadeDuration
}

// applyPending performs the queued stack mutation at full black.
func (sm *SceneManager) applyPending() {
	if sm.pendingPop {
		sm.Pop()
		sm.pendingPop = false
	}
	if sm.pending != nil {
		pause(sm.Top())
		if len(sm.stack) > 0 {
			sm.stack[len(sm.stack)-1] = sm.pending
		} else {
			sm.stack = append(sm.stack, sm.pending)
		}
		resume(sm.pending)
		sm.pending = nil
	}
}

func pause(s Scene) {
	if lifecycle, ok := s.(LifecycleScene); ok {
		lifecycle.Pause()
	}
}

func resume(s Scene) {
	if lifecycle, ok := s.(LifecycleScene); ok {
		lifecycle.Resume()
	}
}
