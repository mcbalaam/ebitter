package components

import (
	"strings"
	"testing"
	"time"

	"github.com/mcbalaam/ebitter/events"
	"github.com/mcbalaam/ebitter/queues"
)

func TestRegistryDefaultOrder(t *testing.T) {
	cases := map[string]int{
		"body":         orderBody,
		"velocity":     orderVelocity,
		"transform":    orderTransform,
		"animator":     orderAnimator,
		"sprite":       orderSprite,
		"hitbox":       orderHitbox,
		"trigger":      orderTrigger,
		"interactable": orderInteractable,
		"lifetime":     orderLifetime,
	}
	for name, want := range cases {
		if !Registered(name) {
			t.Errorf("%q not registered", name)
			continue
		}
		got, ok := DefaultOrder(name)
		if !ok || got != want {
			t.Errorf("%q order = %d (ok=%v), want %d", name, got, ok, want)
		}
	}
}

func TestComponentsSortedByOrder(t *testing.T) {
	e := NewEntity("x")
	e.Add(&Lifetime{TTL: 1})                          // 100
	e.Add(&Transform{ScaleX: 1})                      // 0
	e.Add(&Velocity{})                                // -10
	e.Add(&Hitbox{Collider: NewCollider(1, 1, 0, 0)}) // 30

	var types []string
	for _, c := range e.Components() {
		types = append(types, c.Type())
	}
	want := []string{"velocity", "transform", "hitbox", "lifetime"}
	if len(types) != len(want) {
		t.Fatalf("got %v, want %v", types, want)
	}
	for i := range want {
		if types[i] != want[i] {
			t.Fatalf("got %v, want %v", types, want)
		}
	}
}

func TestOrderOverride(t *testing.T) {
	e := NewEntity("x")
	e.Add(&Transform{ScaleX: 1})    // default 0
	e.AddWithOrder(&Velocity{}, 50) // override -> after transform
	got := e.Components()
	if got[0].Type() != "transform" || got[1].Type() != "velocity" {
		t.Fatalf("order override not applied: %v", got)
	}
}

func TestVelocityIntegratesTransform(t *testing.T) {
	e := NewEntity("x")
	e.Add(&Transform{ScaleX: 1, ScaleY: 1})
	e.Add(&Velocity{X: 100, Y: -50})

	e.Update(time.Second)

	tr := e.Transform()
	if tr == nil || tr.X != 100 || tr.Y != -50 {
		t.Fatalf("transform = %+v, want {100 -50}", tr)
	}
}

func TestLoadBlueprintAndInstantiate(t *testing.T) {
	const doc = `
id: froggit
layer: entity
tags: [enemy]
state: { met_before: false }
components:
  transform: { x: 10, y: 20, scale: 2 }
  velocity: { x: 1, y: 2, order: 5 }
  hitbox: { w: 42, h: 24, layer: enemy, mask: [player] }
`
	bp, err := LoadBlueprint(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("LoadBlueprint: %v", err)
	}
	if bp.ID != "froggit" {
		t.Fatalf("id = %q", bp.ID)
	}

	e, err := bp.Instantiate()
	if err != nil {
		t.Fatalf("Instantiate: %v", err)
	}

	if e.Layer != queues.LayerEntity {
		t.Errorf("layer = %d, want %d", e.Layer, queues.LayerEntity)
	}
	if len(e.Tags) != 1 || e.Tags[0] != "enemy" {
		t.Errorf("tags = %v", e.Tags)
	}
	if met, ok := e.State["met_before"]; !ok || met != false {
		t.Errorf("state met_before = %v", met)
	}

	tr := e.Transform()
	if tr == nil || tr.X != 10 || tr.Y != 20 || tr.ScaleX != 2 || tr.ScaleY != 2 {
		t.Errorf("transform = %+v", tr)
	}
	if h := e.Hitbox(); h == nil || h.Width != 42 || h.Height != 24 || h.Layer != "enemy" {
		t.Errorf("hitbox = %+v", h)
	}

	// order override: velocity (5) must run after transform (0).
	comps := e.Components()
	if comps[0].Type() != "transform" || comps[1].Type() != "velocity" {
		t.Errorf("component order = %v", comps)
	}
}

func TestBlueprintUnknownComponent(t *testing.T) {
	const doc = `
id: bad
components:
  nope: {}
`
	if _, err := LoadBlueprint(strings.NewReader(doc)); err == nil {
		t.Fatal("expected error for unknown component during blueprint validation")
	}
}

func TestBlueprintValidatesComponentConfigAtLoad(t *testing.T) {
	const doc = `
id: bad-config
components:
  velocity: { x: 1, surprise: true }
`
	if _, err := LoadBlueprint(strings.NewReader(doc)); err == nil || !strings.Contains(err.Error(), "components.velocity") {
		t.Fatalf("LoadBlueprint error = %v, want component config path", err)
	}
	const wrongType = `id: bad-type
components:
  velocity: { x: fast }
`
	if _, err := LoadBlueprint(strings.NewReader(wrongType)); err == nil {
		t.Fatal("expected invalid component value to fail during blueprint validation")
	}
	const wrongVector = `id: bad-vector
components:
  sprite: { anchor: [0.5] }
`
	if _, err := LoadBlueprint(strings.NewReader(wrongVector)); err == nil {
		t.Fatal("expected malformed vector to fail during blueprint validation")
	}
	const unknownTopLevel = `id: bad-root
componnets: {}
`
	if _, err := LoadBlueprint(strings.NewReader(unknownTopLevel)); err == nil {
		t.Fatal("expected unknown blueprint field to fail during load")
	}
}

func TestTriggerEmitsOnTouch(t *testing.T) {
	bus := &events.SignalBus{}
	e := NewEntity("t")
	e.Bus = bus
	e.Add(&Transform{ScaleX: 1, ScaleY: 1})
	tr := &Trigger{W: 10, H: 10, Signal: "ping"}
	e.Add(tr)

	var got events.Signal
	bus.Subscribe("ping", nil, func(s events.Signal) { got = s })

	box := AABB{X: -5, Y: -5, W: 10, H: 10}
	if !tr.Check(e, box, false, nil) {
		t.Fatal("trigger did not fire on touch")
	}
	if got.Name != "ping" {
		t.Fatalf("signal = %q", got.Name)
	}
	if tr.Check(e, box, false, nil) {
		t.Fatal("trigger fired twice without leaving")
	}
}

func TestTriggerButtonMode(t *testing.T) {
	bus := &events.SignalBus{}
	e := NewEntity("t")
	e.Bus = bus
	e.Add(&Transform{ScaleX: 1, ScaleY: 1})
	tr := &Trigger{W: 10, H: 10, Signal: "use", Mode: "button"}
	e.Add(tr)

	var count int
	bus.Subscribe("use", nil, func(events.Signal) { count++ })

	box := AABB{X: -5, Y: -5, W: 10, H: 10}
	if tr.Check(e, box, false, nil) {
		t.Fatal("button trigger fired without press")
	}
	if !tr.Check(e, box, true, nil) {
		t.Fatal("button trigger did not fire on press")
	}
	// Still inside: must not fire again until the box leaves.
	if tr.Check(e, box, true, nil) {
		t.Fatal("button trigger fired twice while inside")
	}
	if count != 1 {
		t.Fatalf("signal count = %d, want 1", count)
	}
}

func TestLifetimeDestroysEntity(t *testing.T) {
	e := NewEntity("l")
	e.Add(&Transform{ScaleX: 1})
	e.Add(&Lifetime{TTL: 0.5})

	e.Update(600 * time.Millisecond)
	queues.DefaultDeleteQueue.Execute()

	if e.Transform() != nil {
		t.Fatal("entity should have been destroyed")
	}
}

func TestAnimatorSwitchesSpriteState(t *testing.T) {
	// A nil icon still records the logical state.
	e := NewEntity("a")
	a := &Animator{Default: "idle", On: map[string]string{"move": "walk"}}
	e.Add(a)

	if a.State() != "idle" {
		t.Fatalf("state = %q, want idle", a.State())
	}
	a.OnSignal(e, events.Signal{Name: "move"})
	if a.State() != "walk" {
		t.Fatalf("state = %q, want walk", a.State())
	}
}
