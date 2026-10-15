package components

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/events"
)

// Component is the base interface every entity component implements. Type
// returns the component's registry name, which is also the key used in YAML
// blueprints.
type Component interface {
	Type() string
}

// Initializer is optionally implemented by components that need a one-time
// setup step after being attached to an entity.
type Initializer interface {
	Init(e *Entity)
}

// Updater is implemented by components that need a per-frame update. e is the
// owning entity, giving access to sibling components.
type Updater interface {
	Update(e *Entity, dt time.Duration)
}

// Drawer is implemented by components that render. e is the owning entity.
type Drawer interface {
	Draw(e *Entity, screen *ebiten.Image)
}

// SignalHandler is implemented by components that react to signals on the bus.
type SignalHandler interface {
	OnSignal(e *Entity, sig events.Signal)
}

// ComponentOf returns the entity's component registered under name, asserted to
// type T. It returns the zero value and false when the entity has no such
// component or it is of a different type.
func ComponentOf[T Component](e *Entity, name string) (T, bool) {
	var zero T
	if e == nil {
		return zero, false
	}
	c := e.Get(name)
	if c == nil {
		return zero, false
	}
	v, ok := c.(T)
	if !ok {
		return zero, false
	}
	return v, true
}
