package tiled

import (
	"testing"

	gotiled "github.com/lafriks/go-tiled"
)

func TestCollectSpawns(t *testing.T) {
	groups := []*gotiled.ObjectGroup{{
		Objects: []*gotiled.Object{
			{
				Name: "frog_spot",
				X:    10,
				Y:    20,
				Properties: gotiled.Properties{
					{Name: "blueprint", Type: "string", Value: "froggit"},
					{Name: "facing", Type: "string", Value: "left"},
				},
			},
			{Name: "plain", X: 1, Y: 1},
		},
	}}

	spawns := collectSpawns(groups, 2, DefaultSchema())
	if len(spawns) != 1 {
		t.Fatalf("got %d spawns, want 1", len(spawns))
	}
	s := spawns[0]
	if s.Blueprint != "froggit" {
		t.Errorf("blueprint = %q", s.Blueprint)
	}
	if s.Name != "frog_spot" {
		t.Errorf("name = %q", s.Name)
	}
	if s.X != 20 || s.Y != 40 {
		t.Errorf("scaled position = (%v, %v), want (20, 40)", s.X, s.Y)
	}
	if s.Property("facing") != "left" {
		t.Errorf("facing = %q", s.Property("facing"))
	}
}

func TestSpawnObjectsAreNotPoints(t *testing.T) {
	groups := []*gotiled.ObjectGroup{{
		Objects: []*gotiled.Object{
			{
				Name: "froggit",
				X:    5,
				Y:    5,
				Properties: gotiled.Properties{
					{Name: "blueprint", Type: "string", Value: "froggit"},
				},
			},
		},
	}}

	if pts := collectPoints(groups, 1, DefaultSchema()); len(pts) != 0 {
		t.Errorf("spawn object collected as point: %v", pts)
	}
}
