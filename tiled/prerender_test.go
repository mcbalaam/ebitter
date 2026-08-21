package tiled

import (
	"os"
	"testing"

	"github.com/mcbalaam/ebitter/pkg/embedfs"
)

func TestLoadMapPrerender(t *testing.T) {
	embedfs.SetFS(os.DirFS("../.."))

	m, err := LoadMap("pkg/tiled/testdata/testmap.tmx")
	if err != nil {
		t.Fatalf("LoadMap: %v", err)
	}
	layers := m.Layers()
	if len(layers) != 2 {
		t.Fatalf("layers = %d, want 2", len(layers))
	}
	for _, l := range layers {
		if l.Name != "ground" && l.Name != "deco" {
			t.Errorf("unexpected layer %q", l.Name)
		}
		if len(l.Tiles) == 0 {
			t.Errorf("layer %q has no tiles", l.Name)
			continue
		}
		if !l.Visible {
			t.Errorf("layer %q not visible", l.Name)
		}
		t.Logf("layer %q tiles=%d opacity=%v", l.Name, len(l.Tiles), l.Opacity)
	}
}
