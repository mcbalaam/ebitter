package tiled

import (
	gotiled "github.com/lafriks/go-tiled"
)

// Spawn is an entity placement request read from a Tiled object that carries a
// non-empty string "blueprint" property. X, Y are in scaled world pixels and
// are the object's origin as authored in Tiled. Properties holds every custom
// property of the object so the game can apply per-instance overrides.
type Spawn struct {
	Blueprint  string
	Name       string
	X, Y       float64
	Properties map[string]string
}

// Property returns a custom property string, or "" when absent.
func (s Spawn) Property(name string) string {
	return s.Properties[name]
}

// collectSpawns gathers blueprint placement requests from every object group.
func collectSpawns(groups []*gotiled.ObjectGroup, scale float64, schema Schema) []Spawn {
	var out []Spawn
	for _, g := range groups {
		if g == nil {
			continue
		}
		for _, obj := range g.Objects {
			if obj == nil {
				continue
			}
			bp := obj.Properties.GetString(schema.BlueprintProperty)
			if bp == "" {
				continue
			}
			out = append(out, Spawn{
				Blueprint:  bp,
				Name:       obj.Name,
				X:          obj.X * scale,
				Y:          obj.Y * scale,
				Properties: propStrings(obj.Properties),
			})
		}
	}
	return out
}

func propStrings(p gotiled.Properties) map[string]string {
	if len(p) == 0 {
		return nil
	}
	m := make(map[string]string, len(p))
	for _, prop := range p {
		if prop == nil {
			continue
		}
		m[prop.Name] = prop.Value
	}
	return m
}
