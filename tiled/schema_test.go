package tiled

import (
	"testing"

	gotiled "github.com/lafriks/go-tiled"
)

func TestCustomSchemaInteraction(t *testing.T) {
	schema := DefaultSchema()
	schema.InteractionKeywords = []string{"talk"}
	schema.SignalProperty = "emit"
	schema.DefaultSignal = "chat"

	groups := []*gotiled.ObjectGroup{{
		Name: "talk_zone",
		Objects: []*gotiled.Object{{
			Name: "npc", X: 0, Y: 0, Width: 10, Height: 10,
		}},
	}}

	zones := collectInteractions(groups, schema)
	if len(zones) != 1 {
		t.Fatalf("got %d zones, want 1", len(zones))
	}
	if zones[0].Signal != "chat" {
		t.Errorf("signal = %q, want chat (default)", zones[0].Signal)
	}
	// The default schema must not pick up a "talk" group as an interaction
	// when no signal/interaction property is present.
	if got := collectInteractions(groups, DefaultSchema()); len(got) != 0 {
		t.Errorf("default schema collected %d zones, want 0", len(got))
	}
}

func TestCustomSchemaCollisionKeyword(t *testing.T) {
	schema := DefaultSchema()
	schema.CollisionKeywords = []string{"solid"}

	groups := []*gotiled.ObjectGroup{{
		Name:    "solid_walls",
		Objects: []*gotiled.Object{{X: 1, Y: 2, Width: 3, Height: 4}},
	}}

	if boxes := collectColliders(groups, schema); len(boxes) != 1 {
		t.Fatalf("custom keyword: got %d, want 1", len(boxes))
	}
	// With the default schema the same group is not a collision layer.
	if boxes := collectColliders(groups, DefaultSchema()); len(boxes) != 0 {
		t.Errorf("default schema: got %d boxes, want 0", len(boxes))
	}
}

func TestCustomSchemaZoomProperties(t *testing.T) {
	schema := DefaultSchema()
	schema.MinZoomProperty = "cam_min"
	schema.MaxZoomProperty = "cam_max"

	spec := &gotiled.Map{Properties: &gotiled.Properties{
		{Name: "cam_min", Type: "float", Value: "0.25"},
		{Name: "cam_max", Type: "float", Value: "3"},
	}}
	m := &GameMap{spec: spec, schema: schema}
	if m.MinZoom() != 0.25 || m.MaxZoom() != 3 {
		t.Errorf("zoom = (%v, %v), want (0.25, 3)", m.MinZoom(), m.MaxZoom())
	}
}
