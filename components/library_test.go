package components

import (
	"testing"
	"testing/fstest"

	"github.com/mcbalaam/ebitter/events"
	"github.com/mcbalaam/ebitter/queues"
)

func TestSpawnWithUsesSceneServices(t *testing.T) {
	fsys := fstest.MapFS{
		"entities/actor.yml": &fstest.MapFile{Data: []byte("id: actor\ncomponents:\n  transform: { x: 1, y: 2 }\n")},
	}
	library := NewBlueprintLibrary(fsys, "entities")
	bus := &events.SignalBus{}
	queueSet := queues.NewQueueSet()

	entity, err := library.SpawnWith("actor", 17, 29, bus, queueSet)
	if err != nil {
		t.Fatal(err)
	}
	if entity.Signals() != bus {
		t.Fatal("SpawnWith did not assign the scene signal bus")
	}
	if entity.Queues != queueSet {
		t.Fatal("SpawnWith did not assign the scene queue set")
	}
	transform := entity.Transform()
	if transform == nil || transform.X != 17 || transform.Y != 29 {
		t.Fatalf("spawn transform = %#v, want position (17, 29)", transform)
	}
}
