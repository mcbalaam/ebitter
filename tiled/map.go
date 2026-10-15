package tiled

import (
	"strconv"
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
	spawns       []Spawn
	schema       Schema
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

// Spawns returns the entity placement requests read from the map (objects with
// a "blueprint" property). The game instantiates them, e.g. through
// components.BlueprintLibrary.Spawn.
func (m *GameMap) Spawns() []Spawn {
	return m.spawns
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

// mapProperties returns the map-level custom properties, or nil if none.
func (m *GameMap) mapProperties() gotiled.Properties {
	if m.spec == nil || m.spec.Properties == nil {
		return nil
	}
	return *m.spec.Properties
}

// PropertyFloat returns the float map property with the given name. Both float-
// and int-typed properties are accepted; a value that cannot be parsed yields
// def. The lookup is case-sensitive to match Tiled.
func (m *GameMap) PropertyFloat(name string, def float64) float64 {
	props := m.mapProperties()
	p := props.Get(name)
	if p == nil {
		return def
	}
	switch p.Type {
	case "float", "int", "":
		if v, err := strconv.ParseFloat(p.Value, 64); err == nil {
			return v
		}
	}
	return def
}

// MinZoom returns the map's min-zoom property (default 1).
func (m *GameMap) MinZoom() float64 { return m.PropertyFloat(m.schema.MinZoomProperty, 1) }

// MaxZoom returns the map's max-zoom property (default 1).
func (m *GameMap) MaxZoom() float64 { return m.PropertyFloat(m.schema.MaxZoomProperty, 1) }

// Schema returns the naming conventions the map was loaded with.
func (m *GameMap) Schema() Schema { return m.schema }

// DrawLayers draws all visible tile layers onto dst, offset by (ox, oy).
func (m *GameMap) DrawLayers(dst *ebiten.Image, ox, oy float64) {
	for _, l := range m.layers {
		l.Draw(dst, ox+l.OffsetX, oy+l.OffsetY)
	}
}
