package tiled

import (
	"os"
	"testing"

	gotiled "github.com/lafriks/go-tiled"
	"github.com/mcbalaam/ebitter/pkg/embedfs"
	"github.com/mcbalaam/ebitter/pkg/systems"
)

func setupFS(t *testing.T) {
	t.Helper()
	embedfs.SetFS(os.DirFS("../.."))
}

// testMapPath is a stable fixture; media/maps/demo.tmx is live game content
// and must not be used from tests.
const testMapPath = "pkg/tiled/testdata/testmap.tmx"

func TestLoadMapSpec(t *testing.T) {
	setupFS(t)

	m, err := LoadMapSpec(testMapPath, NewTilesetCache())
	if err != nil {
		t.Fatalf("load map: %v", err)
	}

	if m.Width() != 160 {
		t.Errorf("Width() = %d, want 160", m.Width())
	}
	if m.Height() != 128 {
		t.Errorf("Height() = %d, want 128", m.Height())
	}
	if m.TileWidth() != 16 || m.TileHeight() != 16 {
		t.Errorf("tile size = %dx%d, want 16x16", m.TileWidth(), m.TileHeight())
	}

	// Spec-only load must not pre-render layers.
	if n := len(m.Layers()); n != 0 {
		t.Errorf("Layers() = %d, want 0 before BuildLayers", n)
	}

	if n := len(m.Colliders()); n != 2 {
		t.Errorf("Colliders() = %d, want 2", n)
	}

	zones := m.Interactions()
	if len(zones) != 2 {
		t.Fatalf("Interactions() = %d, want 2", len(zones))
	}
	if zones[0].Trigger != TriggerTouch || zones[0].Signal != "interact" {
		t.Errorf("zone[0] = %+v, want touch/interact", zones[0])
	}
	if zones[1].Trigger != TriggerButton || zones[1].Signal != "interact" {
		t.Errorf("zone[1] = %+v, want button/interact", zones[1])
	}
	// Zone names come from the Tiled object names.
	if zones[0].Name != "touch_zone" || zones[1].Name != "button_zone" {
		t.Errorf("zone names = %q, %q, want touch_zone, button_zone",
			zones[0].Name, zones[1].Name)
	}

	// The "Start" point object marks the player spawn.
	p, ok := m.Point("start")
	if !ok {
		t.Fatal(`Point("start") not found`)
	}
	if p.X != 48 || p.Y != 96 {
		t.Errorf("start point = (%v, %v), want (48, 96)", p.X, p.Y)
	}

	// The external TSX tileset must have been resolved through the FS.
	spec := m.Spec()
	if len(spec.Tilesets) != 1 {
		t.Fatalf("Tilesets() = %d, want 1", len(spec.Tilesets))
	}
	if !spec.Tilesets[0].SourceLoaded {
		t.Error("external TSX tileset was not loaded")
	}
	if len(spec.Layers) != 2 {
		t.Errorf("spec layers = %d, want 2", len(spec.Layers))
	}
	ground := spec.Layers[0]
	if ground.IsEmpty() {
		t.Error("ground layer unexpectedly empty")
	}
	if len(ground.Tiles) != spec.Width*spec.Height {
		t.Errorf("ground tiles = %d, want %d", len(ground.Tiles), spec.Width*spec.Height)
	}
}

func TestColliderBox(t *testing.T) {
	a := &ColliderBox{X: 0, Y: 0, W: 10, H: 10}
	b := &ColliderBox{X: 5, Y: 5, W: 10, H: 10}
	if !a.Overlaps(b) {
		t.Error("boxes overlap, expected true")
	}
	c := &ColliderBox{X: 20, Y: 20, W: 10, H: 10}
	if a.Overlaps(c) {
		t.Error("boxes do not overlap, expected false")
	}
	if !a.Contains(5, 5) || a.Contains(-1, 0) || a.Contains(11, 0) {
		t.Error("Contains gave wrong result")
	}
}

func TestInteractionTouch(t *testing.T) {
	const sig = "_test_touch"
	zone := &InteractionZone{
		Name:    "t",
		Signal:  sig,
		Trigger: TriggerTouch,
		Rect:    ColliderBox{X: 0, Y: 0, W: 10, H: 10},
	}
	box := &ColliderBox{X: 1, Y: 1, W: 2, H: 2}

	fired := 0
	systems.MasterSignalBus.Subscribe(sig, t, func(systems.Signal) { fired++ })

	if !zone.Check(box, false, t) {
		t.Error("expected first trigger on entering the zone")
	}
	if zone.Check(box, false, t) {
		t.Error("must not re-fire while still inside")
	}
	box.X = 50
	if zone.Check(box, false, t) {
		t.Error("must not fire while outside")
	}
	box.X = 1
	if got := zone.Check(box, false, t); !got {
		t.Error("expected trigger after re-entering")
	}
	if fired != 2 {
		t.Errorf("fired = %d, want 2", fired)
	}
}

func TestInteractionButton(t *testing.T) {
	const sig = "_test_button"
	zone := &InteractionZone{
		Name:    "b",
		Signal:  sig,
		Trigger: TriggerButton,
		Rect:    ColliderBox{X: 0, Y: 0, W: 10, H: 10},
	}
	box := &ColliderBox{X: 1, Y: 1, W: 2, H: 2}

	fired := 0
	systems.MasterSignalBus.Subscribe(sig, t, func(systems.Signal) { fired++ })

	if zone.Check(box, false, t) {
		t.Error("must not fire without the button")
	}
	if !zone.Check(box, true, t) {
		t.Error("expected trigger on press inside the zone")
	}
	if zone.Check(box, true, t) {
		t.Error("must not re-fire while holding the button over the zone")
	}
	box.X = 50
	box.Y = 50
	if zone.Check(box, true, t) {
		t.Error("must not fire outside the zone")
	}
	box.X = 1
	box.Y = 1
	if !zone.Check(box, true, t) {
		t.Error("expected trigger after re-entering")
	}
	if fired != 2 {
		t.Errorf("fired = %d, want 2", fired)
	}
}

func TestCollectCollidersFromProperties(t *testing.T) {
	g := &gotiled.ObjectGroup{
		Name: "walls",
		Objects: []*gotiled.Object{{
			X: 4, Y: 8, Width: 16, Height: 24,
		}},
	}
	g.Properties = gotiled.Properties{
		{Name: "collision", Type: "boolean", Value: "true"},
	}
	boxes := collectColliders([]*gotiled.ObjectGroup{g})
	if len(boxes) != 1 {
		t.Fatalf("collectColliders = %d, want 1", len(boxes))
	}
	b := boxes[0]
	if b.X != 4 || b.Y != 8 || b.W != 16 || b.H != 24 {
		t.Errorf("collider = %+v", b)
	}

	// Non-rectangle / zero-size objects are skipped.
	ellipseG := &gotiled.ObjectGroup{Name: "deco"}
	boxes = collectColliders([]*gotiled.ObjectGroup{ellipseG})
	if len(boxes) != 0 {
		t.Errorf("collectColliders(deco) = %d, want 0", len(boxes))
	}
}
