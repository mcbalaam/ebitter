package render

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// A frame is an image and its duration.
type Frame struct {
	Image *ebiten.Image
	Time  time.Duration
}

// An icon state is a set of frames and additional info on how to iterate through them.
type IconState struct {
	Name            string
	CurrentFrame    int
	CurrentFrameRef *Frame
	Frames          []Frame
	Mode            AnimationMode
	dir             int
	elapsed         time.Duration
	Continuous      bool
}

// An animated icon is a set of icon states.
type AnimatedIcon struct {
	CurrentState *IconState
	IconStates   map[string]*IconState
}

// Looks up the provided icon and state in the cache. Tries to cut and cache the icon if not already cached.
func NewAnimatedIconFromPath(path string, stateKey string) (*AnimatedIcon, error) {
	if err := MasterAtlasManager.CacheIconStates(path); err != nil {
		return nil, fmt.Errorf("cache icon states: %w", err)
	}

	iconStates := make(map[string]*IconState)
	MasterAtlasManager.mu.RLock()
	for k, v := range MasterAtlasManager.IconStateCache {
		st := v
		iconStates[k] = &st
	}
	MasterAtlasManager.mu.RUnlock()

	initial, ok := iconStates[stateKey]
	if !ok {
		return nil, fmt.Errorf("get icon state %q: not found", stateKey)
	}

	initial.CurrentFrame = 0
	initial.CurrentFrameRef = &initial.Frames[initial.CurrentFrame]
	initial.elapsed = 0
	initial.dir = 1

	return &AnimatedIcon{
		CurrentState: initial,
		IconStates:   iconStates,
	}, nil
}

// Iterates through the frames.
func (a *AnimatedIcon) Update(dt time.Duration) {
	s := a.CurrentState
	if s == nil || len(s.Frames) == 0 {
		return
	}

	s.elapsed += dt
	if s.CurrentFrame < 0 {
		s.CurrentFrame = 0
	}
	for s.Frames[s.CurrentFrame].Time > 0 && s.elapsed >= s.Frames[s.CurrentFrame].Time {
		s.elapsed -= s.Frames[s.CurrentFrame].Time
		switch s.Mode {
		case AnimationModeLoop:
			s.CurrentFrame++
			if s.CurrentFrame >= len(s.Frames) {
				s.CurrentFrame = 0
			}
		case AnimationModeOnce:
			if s.CurrentFrame < len(s.Frames)-1 {
				s.CurrentFrame++
			} else {
				s.elapsed = 0
				return
			}
		case AnimationModePingPong:
			if s.dir == 0 {
				s.dir = 1
			}
			s.CurrentFrame += s.dir
			if s.CurrentFrame >= len(s.Frames) {
				s.CurrentFrame = len(s.Frames) - 2
				s.dir = -1
			}
			if s.CurrentFrame < 0 {
				s.CurrentFrame = 1
				s.dir = 1
			}
		default:
			s.CurrentFrame++
			if s.CurrentFrame >= len(s.Frames) {
				s.CurrentFrame = 0
			}
		}
		s.CurrentFrameRef = &s.Frames[s.CurrentFrame]
	}
}

func (a *AnimatedIcon) Draw(screen *ebiten.Image, x, y, scaleX, scaleY, tilt float64) {
	a.DrawWithColorScale(screen, x, y, scaleX, scaleY, tilt, 1, 1, 1, 1)
}

// Colors the icon when rendering.
func (a *AnimatedIcon) DrawWithColorScale(screen *ebiten.Image, x, y, scaleX, scaleY, tilt, r, g, b, alpha float64) {
	s := a.CurrentState
	if s == nil || len(s.Frames) == 0 {
		return
	}
	frame := s.Frames[s.CurrentFrame]
	if frame.Image == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Rotate(tilt)
	op.GeoM.Translate(x, y)
	if r != 1 || g != 1 || b != 1 || alpha != 1 {
		op.ColorScale.Scale(float32(r), float32(g), float32(b), float32(alpha))
	}
	screen.DrawImage(frame.Image, op)
}

// Changes the icon's active state.
func (a *AnimatedIcon) SetIconState(state string) error {
	ns, ok := a.IconStates[state]
	if !ok {
		return fmt.Errorf("state %q not found", state)
	}

	if a.CurrentState == ns {
		return nil
	}

	ns.CurrentFrame = 0
	ns.CurrentFrameRef = &ns.Frames[ns.CurrentFrame]
	if !ns.Continuous {
		ns.elapsed = 0
	}
	ns.dir = 1
	a.CurrentState = ns

	return nil
}
