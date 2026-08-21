package queues

import (
	"sync"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type recorder struct {
	name  string
	order *[]string
	mu    *sync.Mutex
}

func (r *recorder) Draw(*ebiten.Image) {
	r.mu.Lock()
	*r.order = append(*r.order, r.name)
	r.mu.Unlock()
}

func TestDrawQueueLayerOrderDeterministic(t *testing.T) {
	q := &DrawQueue{}
	var order []string
	var mu sync.Mutex

	highA := &recorder{name: "high-a", order: &order, mu: &mu}
	highB := &recorder{name: "high-b", order: &order, mu: &mu}
	low := &recorder{name: "low", order: &order, mu: &mu}

	q.ScheduleAt(highA, LayerOverlay)
	q.ScheduleAt(low, LayerBackground)
	q.ScheduleAt(highB, LayerOverlay)

	for i := 0; i < 20; i++ {
		q.Execute(nil)
	}

	// Layers drawn in ascending order; within a layer, insertion order preserved.
	want := []string{"low", "high-a", "high-b"}
	if len(order) != len(want)*20 {
		t.Fatalf("expected %d draws, got %d", len(want)*20, len(order))
	}
	for i, got := range order[:3] {
		if got != want[i] {
			t.Fatalf("draw order mismatch at %d: got %q want %q (full: %v)", i, got, want[i], order)
		}
	}
}

func TestUpdateQueueUnschedule(t *testing.T) {
	q := &UpdateQueue{}
	obj := &scheduledUpdater{}
	q.Schedule(obj)
	q.Unschedule(obj)
	q.Execute(0)
	if obj.calls != 0 {
		t.Errorf("unscheduled object was updated %d times", obj.calls)
	}
}

type scheduledUpdater struct{ calls int }

func (u *scheduledUpdater) Update(dt time.Duration) { u.calls++ }

func TestDeleteQueueRuns(t *testing.T) {
	q := &DeleteQueue{}
	obj := &destroyRecorder{}
	q.objects = append(q.objects, obj)
	q.Execute()
	if obj.calls != 1 {
		t.Errorf("expected 1 destroy, got %d", obj.calls)
	}
}

type destroyRecorder struct{ calls int }

func (d *destroyRecorder) Destroy() { d.calls++ }
