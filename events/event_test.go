package events

import "testing"

func TestSignalBusesAreIndependent(t *testing.T) {
	var firstCount, secondCount int
	first := &SignalBus{}
	second := &SignalBus{}
	first.Subscribe("change", nil, func(Signal) { firstCount++ })
	second.Subscribe("change", nil, func(Signal) { secondCount++ })

	first.Emit("change", nil)
	if firstCount != 1 || secondCount != 0 {
		t.Fatalf("first bus counts = (%d, %d), want (1, 0)", firstCount, secondCount)
	}
	second.Emit("change", nil)
	if firstCount != 1 || secondCount != 1 {
		t.Fatalf("after second emit counts = (%d, %d), want (1, 1)", firstCount, secondCount)
	}
}
