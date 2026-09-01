package tween

import (
	"testing"
	"time"

	"github.com/mcbalaam/ebitter/queues"
)

func TestScheduleIsIdempotent(t *testing.T) {
	queues.DefaultUpdateQueue = &queues.UpdateQueue{}
	var v float64
	tw := NewFloat(0, 10, time.Second, func(f float64) { v = f }).Schedule()
	tw.Schedule()
	// Only one entry should exist: driving it once must move the value.
	queues.DefaultUpdateQueue.Execute(100 * time.Millisecond)
	if got := v; got != 1 {
		t.Fatalf("value after one 100ms tick = %v, want 1 (tween scheduled twice?)", got)
	}
}

func TestRetargetContinuesFromCurrentValue(t *testing.T) {
	queues.DefaultUpdateQueue = &queues.UpdateQueue{}
	var v float64
	tw := NewFloat(0, 10, time.Second, func(f float64) { v = f }).Schedule()
	queues.DefaultUpdateQueue.Execute(500 * time.Millisecond) // halfway: v = 5
	if v != 5 {
		t.Fatalf("midpoint = %v, want 5", v)
	}
	tw.Retarget(20)
	// From current (5) toward 20 over 1s; half a second → 12.5.
	queues.DefaultUpdateQueue.Execute(500 * time.Millisecond)
	if v != 12.5 {
		t.Fatalf("after retarget midpoint = %v, want 12.5", v)
	}
}

func TestStopRemovesFromQueue(t *testing.T) {
	queues.DefaultUpdateQueue = &queues.UpdateQueue{}
	var v float64
	tw := NewFloat(0, 100, time.Second, func(f float64) { v = f }).Schedule()
	Stop(tw)
	queues.DefaultUpdateQueue.Execute(time.Second)
	if v != 0 {
		t.Fatalf("value after Stop = %v, want 0 (still updating)", v)
	}
	if queues.DefaultUpdateQueue.IsScheduled(tw) {
		t.Fatal("stopped tween still scheduled")
	}
}
