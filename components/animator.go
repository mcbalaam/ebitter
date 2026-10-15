package components

import (
	"time"

	"github.com/mcbalaam/ebitter/events"
)

// Animator drives an entity's Sprite animation state. It switches to the state
// named by On[signal] whenever that signal is delivered to the entity, and
// falls back to Default when no state has been chosen yet.
type Animator struct {
	Default string            `yaml:"default"`
	On      map[string]string `yaml:"on"`

	current string
}

func (a *Animator) Type() string { return "animator" }

func (a *Animator) Init(e *Entity) {
	if a.current == "" && a.Default != "" {
		a.Play(e, a.Default)
	}
}

func (a *Animator) Update(e *Entity, dt time.Duration) {
	if a.current == "" && a.Default != "" {
		a.Play(e, a.Default)
	}
}

// Play switches the entity's sprite to the named animation state.
func (a *Animator) Play(e *Entity, state string) {
	a.current = state
	if s := e.Sprite(); s != nil && s.Icon != nil {
		_ = s.Icon.SetIconState(state)
	}
}

// State returns the currently playing state.
func (a *Animator) State() string { return a.current }

func (a *Animator) OnSignal(e *Entity, sig events.Signal) {
	if a.On == nil {
		return
	}
	if state, ok := a.On[sig.Name]; ok {
		a.Play(e, state)
	}
}
