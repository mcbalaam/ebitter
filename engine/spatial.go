package engine

import (
	"math"

	"github.com/mcbalaam/ebitter/pkg/engine/components"
)

type SpatialGrid struct {
	cellSize float64
	cells    map[[2]int][]*components.Entity
}

func NewSpatialGrid(cellSize float64) *SpatialGrid {
	return &SpatialGrid{
		cellSize: cellSize,
		cells:    make(map[[2]int][]*components.Entity),
	}
}

func (g *SpatialGrid) Clear() {
	for k := range g.cells {
		delete(g.cells, k)
	}
}

func (g *SpatialGrid) cellKey(x, y float64) [2]int {
	return [2]int{int(math.Floor(x / g.cellSize)), int(math.Floor(y / g.cellSize))}
}

func (g *SpatialGrid) Insert(e *components.Entity) {
	if e.Transform == nil {
		return
	}
	key := g.cellKey(e.Transform.X, e.Transform.Y)
	g.cells[key] = append(g.cells[key], e)
}

func (g *SpatialGrid) Query(e *components.Entity) []*components.Entity {
	if e.Transform == nil {
		return nil
	}

	key := g.cellKey(e.Transform.X, e.Transform.Y)
	seen := make(map[*components.Entity]struct{})
	var result []*components.Entity

	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			k := [2]int{key[0] + dx, key[1] + dy}
			for _, other := range g.cells[k] {
				if other != e {
					if _, ok := seen[other]; !ok {
						seen[other] = struct{}{}
						result = append(result, other)
					}
				}
			}
		}
	}
	return result
}
