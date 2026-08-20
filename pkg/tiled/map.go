package tiled

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	gotiled "github.com/lafriks/go-tiled"
)

// Layer is a tile layer expressed as a list of tile placements. Tiles are
// drawn directly from the cached tileset sub-images at render time.
type Layer struct {
	Name    string
	Visible bool
	OffsetX float64
	OffsetY float64
	Opacity float32
	Tiles   []*TileAt
}

// Draw renders the layer onto dst, applying the given scroll offset.
func (l *Layer) Draw(dst *ebiten.Image, sx, sy float64) {
	if !l.Visible || len(l.Tiles) == 0 {
		return
	}
	for _, t := range l.Tiles {
		drawTile(dst, t.Image, t.X+sx, t.Y+sy,
			t.TileW, t.TileH, t.Scale,
			t.FlipH, t.FlipV, t.FlipDiag,
			l.Opacity)
	}
}
// GameMap is a parsed Tiled map with tile layers, collision boxes and
// interaction zones. All world coordinates and dimensions are multiplied by
// the map's scale factor.
type GameMap struct {
	spec         *gotiled.Map
	scale        float64
	layers       []*Layer
	colliders    []*ColliderBox
	interactions []*InteractionZone
	points       map[string]Point
}

// Scale returns the world scale factor the map was loaded with.
func (m *GameMap) Scale() float64 {
	return m.scale
}

// Width returns the map width in (scaled) pixels.
func (m *GameMap) Width() int {
	return int(float64(m.spec.Width*m.spec.TileWidth) * m.scale)
}

// Height returns the map height in (scaled) pixels.
func (m *GameMap) Height() int {
	return int(float64(m.spec.Height*m.spec.TileHeight) * m.scale)
}

// TileWidth returns the width of a single tile in (scaled) pixels.
func (m *GameMap) TileWidth() int {
	return int(float64(m.spec.TileWidth) * m.scale)
}

// TileHeight returns the height of a single tile in (scaled) pixels.
func (m *GameMap) TileHeight() int {
	return int(float64(m.spec.TileHeight) * m.scale)
}

// Layers returns the pre-rendered tile layers in draw order.
func (m *GameMap) Layers() []*Layer {
	return m.layers
}

// Colliders returns the static collision boxes of the map.
func (m *GameMap) Colliders() []*ColliderBox {
	return m.colliders
}

// Interactions returns the interaction zones of the map.
func (m *GameMap) Interactions() []*InteractionZone {
	return m.interactions
}

// Point returns the named point object (e.g. a "start" spawn marker) and
// whether it exists. The lookup is case-insensitive.
func (m *GameMap) Point(name string) (Point, bool) {
	p, ok := m.points[strings.ToLower(name)]
	return p, ok
}

// Spec returns the underlying go-tiled map structure.
func (m *GameMap) Spec() *gotiled.Map {
	return m.spec
}

// DrawLayers draws all visible tile layers onto dst, offset by (ox, oy).
func (m *GameMap) DrawLayers(dst *ebiten.Image, ox, oy float64) {
	for _, l := range m.layers {
		l.Draw(dst, ox+l.OffsetX, oy+l.OffsetY)
	}
}
