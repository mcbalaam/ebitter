package scene

import (
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type dummyScene struct {
	name    string
	updates int
}

func (d *dummyScene) Update(time.Duration)  { d.updates++ }
func (d *dummyScene) Draw(*ebiten.Image)    {}

const testFade = 100 * time.Millisecond

func newTestManager() *SceneManager {
	return &SceneManager{FadeDuration: testFade}
}

// step advances the manager n ticks of dt.
func step(sm *SceneManager, n int, dt time.Duration) {
	for i := 0; i < n; i++ {
		sm.Update(dt)
	}
}

func TestStartupFadeIn(t *testing.T) {
	sm := newTestManager()
	a := &dummyScene{name: "a"}
	sm.Push(a)

	sm.Update(testFade / 2)
	if !sm.IsTransitioning() {
		t.Fatal("expected a fade-in after the first scene appears")
	}
	if sm.fadeAlpha != 0.5 {
		t.Fatalf("fadeAlpha = %v, want 0.5", sm.fadeAlpha)
	}
	// The incoming scene updates during fade-in so it can warm up.
	if a.updates != 1 {
		t.Fatalf("updates = %d, want 1", a.updates)
	}

	step(sm, 2, testFade/2)
	if sm.IsTransitioning() {
		t.Fatal("fade-in should have finished")
	}
	if sm.fadeAlpha != 0 {
		t.Fatalf("fadeAlpha = %v, want 0", sm.fadeAlpha)
	}
}

func TestSwitchToReplacesSceneAtBlack(t *testing.T) {
	sm := newTestManager()
	a := &dummyScene{name: "a"}
	b := &dummyScene{name: "b"}
	sm.Push(a)
	step(sm, 2, testFade) // finish the startup fade-in

	sm.SwitchTo(b)
	if !sm.IsTransitioning() {
		t.Fatal("expected a transition after SwitchTo")
	}
	if got := sm.Top(); got != a {
		t.Fatalf("Top() = %v, want a until the fade-out completes", got)
	}

	before := a.updates
	sm.Update(testFade / 2)
	if a.updates != before {
		t.Fatal("outgoing scene must stay frozen during fade-out")
	}

	sm.Update(testFade / 2) // reach full black: swap happens here
	if got := sm.Top(); got != b {
		t.Fatalf("Top() = %v, want b after the fade-out", got)
	}
	if len(sm.stack) != 1 {
		t.Fatalf("stack len = %d, want 1 (replace, not push)", len(sm.stack))
	}

	step(sm, 2, testFade/2) // finish the fade-in
	if sm.IsTransitioning() {
		t.Fatal("transition should have finished")
	}
	if b.updates == 0 {
		t.Fatal("new scene should have updated during the fade-in")
	}
}

func TestPopFaded(t *testing.T) {
	sm := newTestManager()
	a := &dummyScene{name: "a"}
	b := &dummyScene{name: "b"}
	sm.Push(a)
	sm.Push(b)
	step(sm, 2, testFade)

	sm.PopFaded()
	sm.Update(testFade) // fade out to black, pop, start fade-in
	if got := sm.Top(); got != a {
		t.Fatalf("Top() = %v, want a after the pop", got)
	}

	step(sm, 2, testFade)
	if sm.IsTransitioning() {
		t.Fatal("transition should have finished")
	}
}

func TestSwitchToRetargetsInFlightTransition(t *testing.T) {
	sm := newTestManager()
	a := &dummyScene{name: "a"}
	b := &dummyScene{name: "b"}
	c := &dummyScene{name: "c"}
	sm.Push(a)
	step(sm, 2, testFade)

	sm.SwitchTo(b)
	sm.Update(testFade / 2) // mid fade-out
	sm.SwitchTo(c)          // change our mind mid-transition
	sm.Update(testFade)     // cross the black point and fade back in
	if got := sm.Top(); got != c {
		t.Fatalf("Top() = %v, want c after retargeting", got)
	}
}
