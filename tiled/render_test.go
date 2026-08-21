package tiled

import (
	"os"
	"testing"

	"github.com/mcbalaam/ebitter/pkg/embedfs"
)

// TestBuildLayersPositions checks tile world placement against the CSV data.
// Pixel-level checks are not possible here: ebiten.Image.ReadPixels requires
// the running game loop.
func TestBuildLayersPositions(t *testing.T) {
	embedfs.SetFS(os.DirFS("../.."))

	m, err := LoadMap(testMapPath)
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}
	ground := m.Layers()[0]

	// Row 2, col 5 of the ground layer holds the only GID-2 tile.
	const wantX, wantY = 5 * 16, 2 * 16
	for _, tile := range ground.Tiles {
		if tile.X == wantX && tile.Y == wantY {
			return
		}
	}
	t.Errorf("no tile placed at (%d, %d); grid positioning is broken", wantX, wantY)
}

// TestLoadMapScaled verifies that loading with a scale factor multiplies all
// world-space values: map dimensions, tile placement, colliders and zones.
func TestLoadMapScaled(t *testing.T) {
	embedfs.SetFS(os.DirFS("../.."))

	m, err := LoadMapScaled(testMapPath, 2)
	if err != nil {
		t.Fatalf("LoadMapScaled: %v", err)
	}

	if m.Width() != 320 || m.Height() != 256 {
		t.Errorf("size = %dx%d, want 320x256", m.Width(), m.Height())
	}
	if m.TileWidth() != 32 || m.TileHeight() != 32 {
		t.Errorf("tile size = %dx%d, want 32x32", m.TileWidth(), m.TileHeight())
	}

	// Row 2, col 5 of the ground layer lands at scaled coordinates.
	const wantX, wantY = 5 * 32, 2 * 32
	found := false
	for _, tile := range m.Layers()[0].Tiles {
		if tile.Scale != 2 {
			t.Errorf("tile scale = %v, want 2", tile.Scale)
		}
		if tile.X == wantX && tile.Y == wantY {
			found = true
		}
	}
	if !found {
		t.Errorf("no tile placed at (%d, %d)", wantX, wantY)
	}

	// Collider #2 (16x64 at 16,32) lands at scaled coordinates.
	foundBox := false
	for _, c := range m.Colliders() {
		if c.X == 32 && c.Y == 64 {
			foundBox = true
			if c.W != 32 || c.H != 128 {
				t.Errorf("collider = %vx%v, want 32x128", c.W, c.H)
			}
		}
	}
	if !foundBox {
		t.Error("scaled collider not found")
	}

	// The "Start" point is scaled too: (48,96) -> (96,192).
	p, ok := m.Point("start")
	if !ok {
		t.Fatal(`Point("start") not found`)
	}
	if p.X != 96 || p.Y != 192 {
		t.Errorf("scaled start point = (%v, %v), want (96, 192)", p.X, p.Y)
	}
}
