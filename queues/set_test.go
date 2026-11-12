package queues

import (
	"testing"
	"time"
)

type countUpdate struct{ count int }

func (u *countUpdate) Update(time.Duration) { u.count++ }

type countDestroy struct{ count int }

func (d *countDestroy) Destroy() { d.count++ }

func TestQueueSetPauseAndResume(t *testing.T) {
	set := NewQueueSet()
	update := &countUpdate{}
	destroy := &countDestroy{}
	set.Update.Schedule(update)
	set.Delete.Schedule(destroy)

	set.Pause()
	set.ExecuteUpdate(time.Second)
	set.ExecuteDelete()
	if update.count != 0 || destroy.count != 0 {
		t.Fatalf("paused queue set executed work: updates=%d destroys=%d", update.count, destroy.count)
	}

	set.Resume()
	set.ExecuteUpdate(time.Second)
	set.ExecuteDelete()
	if update.count != 1 || destroy.count != 1 {
		t.Fatalf("resumed queue set lost work: updates=%d destroys=%d", update.count, destroy.count)
	}
}
