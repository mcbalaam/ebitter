package tiled

import (
	"strings"

	gotiled "github.com/lafriks/go-tiled"
)

// Point is a named location on the map (e.g. a spawn marker), in world
// coordinates.
type Point struct {
	X, Y float64
}

// collectPoints gathers named point objects (zero-size marker objects such as
// Tiled point objects) from all object groups, keyed by lower-cased name.
// Coordinates are scaled by the map's scale factor.
func collectPoints(groups []*gotiled.ObjectGroup, scale float64) map[string]Point {
	out := map[string]Point{}
	for _, g := range groups {
		if g == nil {
			continue
		}
		for _, obj := range g.Objects {
			if !isPointObject(obj) {
				continue
			}
			name := obj.Name
			if name == "" {
				name = obj.Properties.GetString("name")
			}
			if name == "" {
				continue
			}
			out[strings.ToLower(name)] = Point{X: obj.X * scale, Y: obj.Y * scale}
		}
	}
	return out
}

// isPointObject reports whether the object is a point-like marker: Tiled point
// objects carry no size and no shape. (go-tiled does not expose the <point/>
// marker explicitly.)
func isPointObject(obj *gotiled.Object) bool {
	return obj.GID == 0 && obj.Width == 0 && obj.Height == 0 &&
		len(obj.Ellipses) == 0 && len(obj.Polygons) == 0 &&
		len(obj.PolyLines) == 0 && obj.Text == nil
}
