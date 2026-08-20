package tiled

import (
	"fmt"
	"path/filepath"

	gotiled "github.com/lafriks/go-tiled"
	"github.com/mcbalaam/ebitter/pkg/embedfs"
)

// LoadMap parses a Tiled .tmx file from the embedded filesystem and builds a
// GameMap with pre-rendered layers, collision boxes and interaction zones.
// The path is relative to the embedded media root, e.g. "media/maps/demo.tmx".
func LoadMap(path string) (*GameMap, error) {
	return LoadMapScaled(path, 1)
}

// LoadMapScaled is LoadMap with the whole world (tiles, colliders, interaction
// zones and map dimensions) scaled by the given factor.
func LoadMapScaled(path string, scale float64) (*GameMap, error) {
	m, err := LoadMapSpecScaled(path, DefaultTilesetCache, scale)
	if err != nil {
		return nil, err
	}
	if err := m.BuildLayers(DefaultTilesetCache); err != nil {
		return nil, err
	}
	return m, nil
}

// LoadMapSpec parses a Tiled .tmx file without pre-rendering tile layers.
// Use this when only the logical contents (collision, interactions, layer
// metadata) are needed.
func LoadMapSpec(path string, cache *TilesetCache) (*GameMap, error) {
	return LoadMapSpecScaled(path, cache, 1)
}

// LoadMapSpecScaled is LoadMapSpec with the world scaled by the given factor.
func LoadMapSpecScaled(path string, cache *TilesetCache, scale float64) (*GameMap, error) {
	if scale <= 0 {
		scale = 1
	}
	spec, err := gotiled.LoadFile(filepath.ToSlash(path), gotiled.WithFileSystem(embedfs.FS))
	if err != nil {
		return nil, fmt.Errorf("tiled: load map %s: %w", path, err)
	}
	if spec.Infinite {
		return nil, fmt.Errorf("tiled: infinite maps are not supported")
	}
	return &GameMap{
		spec:         spec,
		scale:        scale,
		colliders:    scaleColliders(collectColliders(spec.ObjectGroups), scale),
		interactions: scaleInteractions(collectInteractions(spec.ObjectGroups), scale),
		points:       collectPoints(spec.ObjectGroups, scale),
	}, nil
}

func scaleColliders(boxes []*ColliderBox, scale float64) []*ColliderBox {
	if scale == 1 {
		return boxes
	}
	for _, b := range boxes {
		b.X *= scale
		b.Y *= scale
		b.W *= scale
		b.H *= scale
	}
	return boxes
}

func scaleInteractions(zones []*InteractionZone, scale float64) []*InteractionZone {
	if scale == 1 {
		return zones
	}
	for _, z := range zones {
		z.Rect.X *= scale
		z.Rect.Y *= scale
		z.Rect.W *= scale
		z.Rect.H *= scale
	}
	return zones
}

// BuildLayers resolves every tile of every tile layer into cached tileset
// sub-images, ready to be drawn directly at render time. This must be called
// after the Ebiten graphics context has started.
func (m *GameMap) BuildLayers(cache *TilesetCache) error {
	if cache == nil {
		cache = DefaultTilesetCache
	}
	layers, err := buildLayers(cache, m.spec, m.scale)
	if err != nil {
		return err
	}
	m.layers = layers
	return nil
}

// buildLayers resolves each layer's tiles to tileset sub-images, placing them
// in world coordinates scaled by the map's scale factor.
func buildLayers(cache *TilesetCache, spec *gotiled.Map, scale float64) ([]*Layer, error) {
	if cache == nil {
		cache = DefaultTilesetCache
	}
	if scale <= 0 {
		scale = 1
	}

	tw := float64(spec.TileWidth) * scale
	th := float64(spec.TileHeight) * scale

	layers := make([]*Layer, 0, len(spec.Layers))
	for _, gl := range spec.Layers {
		if gl == nil || gl.IsEmpty() {
			continue
		}

		l := &Layer{
			Name:    gl.Name,
			Visible: gl.Visible,
			OffsetX: gl.OffsetX * scale,
			OffsetY: gl.OffsetY * scale,
			Opacity: gl.Opacity,
			Tiles:   make([]*TileAt, 0, len(gl.Tiles)),
		}

		for i, t := range gl.Tiles {
			if t == nil || t.IsNil() {
				continue
			}
			asset, err := cache.Asset(t.Tileset)
			if err != nil {
				return nil, fmt.Errorf("tiled: layer %q: %w", gl.Name, err)
			}
			tileImg, err := asset.Tile(t.ID)
			if err != nil {
				return nil, fmt.Errorf("tiled: layer %q tile %d: %w", gl.Name, i, err)
			}

			col := i % spec.Width
			row := i / spec.Width
			l.Tiles = append(l.Tiles, &TileAt{
				Image:    tileImg,
				X:        float64(col) * tw,
				Y:        float64(row) * th,
				TileW:    int(tw),
				TileH:    int(th),
				Scale:    scale,
				FlipH:    t.HorizontalFlip,
				FlipV:    t.VerticalFlip,
				FlipDiag: t.DiagonalFlip,
			})
		}

		layers = append(layers, l)
	}
	return layers, nil
}
