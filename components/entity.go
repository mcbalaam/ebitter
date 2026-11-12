package components

import (
	"sort"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/events"
	"github.com/mcbalaam/ebitter/queues"
	"github.com/mcbalaam/ebitter/render"
)

type componentEntry struct {
	c     Component
	order int
}

// Entity is a data-driven container of components. Components are updated and
// drawn in ascending order (ties keep insertion order). Every component
// receives a pointer to its owning entity so it can read and write siblings.
type Entity struct {
	ID    string
	Layer int
	Tags  []string
	State map[string]any

	// Bus, when set, is used for signals emitted by the entity (collisions,
	// health, triggers, ...). Nil falls back to events.MasterSignalBus.
	Bus *events.SignalBus
	// Queues selects the queues used by Schedule, Destroy and deferred deletion.
	// Nil falls back to the package-wide defaults.
	Queues *queues.QueueSet

	entries []componentEntry
	index   map[string]Component
	sorted  bool
}

// NewEntity returns an entity with an empty component set.
func NewEntity(id string) *Entity {
	return &Entity{
		ID:    id,
		State: map[string]any{},
		index: map[string]Component{},
	}
}

// Add attaches c using its registered default order.
func (e *Entity) Add(c Component) {
	if c == nil {
		return
	}
	order := 0
	if o, ok := DefaultOrder(c.Type()); ok {
		order = o
	}
	e.AddWithOrder(c, order)
}

// AddWithOrder attaches c at an explicit order, overriding the registry default.
func (e *Entity) AddWithOrder(c Component, order int) {
	if c == nil {
		return
	}
	if e.index == nil {
		e.index = map[string]Component{}
	}
	name := c.Type()
	if _, exists := e.index[name]; exists {
		for i := range e.entries {
			if e.entries[i].c.Type() == name {
				e.entries[i] = componentEntry{c: c, order: order}
				break
			}
		}
	} else {
		e.entries = append(e.entries, componentEntry{c: c, order: order})
	}
	e.index[name] = c
	e.sorted = false
	if init, ok := c.(Initializer); ok {
		init.Init(e)
	}
}

// Get returns the component registered under name, or nil.
func (e *Entity) Get(name string) Component {
	if e == nil || e.index == nil {
		return nil
	}
	return e.index[name]
}

// Has reports whether a component is attached under name.
func (e *Entity) Has(name string) bool { return e.Get(name) != nil }

// Remove detaches the component registered under name.
func (e *Entity) Remove(name string) {
	if e == nil || e.index == nil {
		return
	}
	if _, ok := e.index[name]; !ok {
		return
	}
	delete(e.index, name)
	for i := range e.entries {
		if e.entries[i].c.Type() == name {
			e.entries = append(e.entries[:i], e.entries[i+1:]...)
			break
		}
	}
	e.sorted = false
}

// Components returns the attached components in update/draw order.
func (e *Entity) Components() []Component {
	e.ensureSorted()
	out := make([]Component, len(e.entries))
	for i, en := range e.entries {
		out[i] = en.c
	}
	return out
}

func (e *Entity) ensureSorted() {
	if e == nil || e.sorted {
		return
	}
	sort.SliceStable(e.entries, func(i, j int) bool {
		return e.entries[i].order < e.entries[j].order
	})
	e.sorted = true
}

// Update runs every component's Update in order.
func (e *Entity) Update(dt time.Duration) {
	if e == nil {
		return
	}
	e.ensureSorted()
	for _, en := range e.entries {
		if u, ok := en.c.(Updater); ok {
			u.Update(e, dt)
		}
	}
}

// Draw runs every component's Draw in order.
func (e *Entity) Draw(screen *ebiten.Image) {
	if e == nil {
		return
	}
	e.ensureSorted()
	for _, en := range e.entries {
		if d, ok := en.c.(Drawer); ok {
			d.Draw(e, screen)
		}
	}
}

// OnSignal forwards a signal to every component implementing SignalHandler.
func (e *Entity) OnSignal(sig events.Signal) {
	if e == nil {
		return
	}
	e.ensureSorted()
	for _, en := range e.entries {
		if h, ok := en.c.(SignalHandler); ok {
			h.OnSignal(e, sig)
		}
	}
}

// IsParallelSafe reports that entities must be updated sequentially.
func (e *Entity) IsParallelSafe() bool { return false }

// Schedule adds the entity to the update queue and the layered draw queue.
func (e *Entity) Schedule() {
	q := e.queueSet()
	q.Update.Schedule(e)
	q.Draw.ScheduleAt(e, e.Layer)
}

// signalBus returns the bus the entity emits on.
func (e *Entity) signalBus() *events.SignalBus {
	if e.Bus != nil {
		return e.Bus
	}
	return &events.MasterSignalBus
}

// Signals returns the entity's configured signal bus, or MasterSignalBus when
// no scene-specific bus was assigned.
func (e *Entity) Signals() *events.SignalBus { return e.signalBus() }

func (e *Entity) queueSet() *queues.QueueSet {
	if e.Queues != nil {
		return e.Queues
	}
	return queues.DefaultQueueSet
}

// Destroy removes the entity from the queues and drops its components.
func (e *Entity) Destroy() {
	q := e.queueSet()
	q.Draw.Unschedule(e)
	q.Update.Unschedule(e)
	e.entries = nil
	e.index = nil
}

// CheckCollisionsWithList emits signalName for each other entity whose hitbox
// intersects this entity's hitbox. Source is e, Data is the other entity.
func (e *Entity) CheckCollisionsWithList(others []*Entity, signalName string) {
	self := e.Collider()
	if self == nil {
		return
	}
	bus := e.signalBus()
	for _, other := range others {
		if other == nil || other == e {
			continue
		}
		oc := other.Collider()
		if oc == nil {
			continue
		}
		if self.CollidesWith(oc) {
			bus.Emit(signalName, e, other)
		}
	}
}

// --- typed accessors -------------------------------------------------------

func (e *Entity) Transform() *Transform { c, _ := ComponentOf[*Transform](e, "transform"); return c }
func (e *Entity) Velocity() *Velocity   { c, _ := ComponentOf[*Velocity](e, "velocity"); return c }
func (e *Entity) Sprite() *Sprite       { c, _ := ComponentOf[*Sprite](e, "sprite"); return c }
func (e *Entity) Hitbox() *Hitbox       { c, _ := ComponentOf[*Hitbox](e, "hitbox"); return c }
func (e *Entity) Body() *Body           { c, _ := ComponentOf[*Body](e, "body"); return c }
func (e *Entity) Animator() *Animator   { c, _ := ComponentOf[*Animator](e, "animator"); return c }
func (e *Entity) Trigger() *Trigger     { c, _ := ComponentOf[*Trigger](e, "trigger"); return c }

func (e *Entity) Interactable() *Interactable {
	c, _ := ComponentOf[*Interactable](e, "interactable")
	return c
}

func (e *Entity) Lifetime() *Lifetime { c, _ := ComponentOf[*Lifetime](e, "lifetime"); return c }

// Collider returns the entity's hitbox collider, or nil.
func (e *Entity) Collider() *Collider {
	h := e.Hitbox()
	if h == nil {
		return nil
	}
	return h.Collider
}

// DrawSpriteOnCollider aligns an animated icon to a transform and hitbox,
// centering the sprite on the collider and honouring its offset.
func DrawSpriteOnCollider(screen *ebiten.Image, icon *render.AnimatedIcon, transform *Transform, collider *Collider) {
	if icon == nil || transform == nil || collider == nil {
		return
	}
	offsetVec := Vec{X: collider.OffsetX, Y: collider.OffsetY}
	transformedOffset := TransformPoint(offsetVec, transform.ScaleX, transform.ScaleY, transform.Rotation, 0, 0)
	centerX := transform.X - (collider.Width*transform.ScaleX)/2
	centerY := transform.Y - (collider.Height*transform.ScaleY)/2
	icon.Draw(screen, centerX+transformedOffset.X, centerY+transformedOffset.Y,
		transform.ScaleX, transform.ScaleY, transform.Rotation)
}
