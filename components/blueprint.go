package components

import (
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mcbalaam/ebitter/queues"
)

// Blueprint is a named entity template parsed from a YAML document.
type Blueprint struct {
	ID    string
	Layer int
	Tags  []string
	State map[string]any
	comps []blueprintComponent
}

type blueprintComponent struct {
	name  string
	order int
	node  *yaml.Node
}

// blueprintDoc mirrors the on-disk YAML shape. Nodes are kept raw so component
// order and per-component structure survive.
type blueprintDoc struct {
	ID         string         `yaml:"id"`
	Layer      yaml.Node      `yaml:"layer"`
	Tags       []string       `yaml:"tags"`
	State      map[string]any `yaml:"state"`
	Components yaml.Node      `yaml:"components"`
}

// LoadBlueprint parses a single-entity YAML document.
func LoadBlueprint(r io.Reader) (*Blueprint, error) {
	var doc blueprintDoc
	dec := yaml.NewDecoder(r)
	var root yaml.Node
	if err := dec.Decode(&root); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("components: empty blueprint")
		}
		return nil, err
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) != 1 || root.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("components: blueprint must be a mapping")
	}
	seenFields := map[string]bool{}
	for i := 0; i+1 < len(root.Content[0].Content); i += 2 {
		key := root.Content[0].Content[i].Value
		switch key {
		case "id", "layer", "tags", "state", "components":
		default:
			return nil, fmt.Errorf("components: blueprint: unknown field %q", key)
		}
		if seenFields[key] {
			return nil, fmt.Errorf("components: blueprint: duplicate field %q", key)
		}
		seenFields[key] = true
	}
	if err := root.Decode(&doc); err != nil {
		return nil, err
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); err == nil {
		return nil, fmt.Errorf("components: blueprint: multiple YAML documents are not supported")
	} else if err != io.EOF {
		return nil, err
	}
	if doc.Layer.Kind != 0 && !validLayer(doc.Layer) {
		return nil, fmt.Errorf("components: blueprint %q: invalid layer %q", doc.ID, doc.Layer.Value)
	}

	bp := &Blueprint{
		ID:    doc.ID,
		Layer: decodeLayer(doc.Layer),
		Tags:  doc.Tags,
		State: doc.State,
	}

	if doc.Components.Kind != 0 && doc.Components.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("components: blueprint %q: components must be a mapping", doc.ID)
	}
	if doc.Components.Kind == yaml.MappingNode {
		content := doc.Components.Content
		seen := make(map[string]bool, len(content)/2)
		for i := 0; i+1 < len(content); i += 2 {
			name := content[i].Value
			node := normalizeNode(content[i+1])
			if seen[name] {
				return nil, fmt.Errorf("components: blueprint %q: duplicate component %q", doc.ID, name)
			}
			seen[name] = true
			if err := ValidateComponent(name, node); err != nil {
				return nil, fmt.Errorf("components: blueprint %q, components.%s: %w", doc.ID, name, err)
			}
			order, err := componentOrder(name, node)
			if err != nil {
				return nil, fmt.Errorf("components: blueprint %q, components.%s: %w", doc.ID, name, err)
			}
			bp.comps = append(bp.comps, blueprintComponent{
				name:  name,
				order: order,
				node:  node,
			})
		}
	}

	return bp, nil
}

// Instantiate builds a fresh entity from the blueprint. Components are created
// through the registry, so custom components registered by the game work too.
func (b *Blueprint) Instantiate() (*Entity, error) {
	e := NewEntity(b.ID)
	e.Layer = b.Layer
	if len(b.Tags) > 0 {
		e.Tags = append([]string(nil), b.Tags...)
	}
	for k, v := range b.State {
		e.State[k] = v
	}
	for _, bc := range b.comps {
		c, err := NewComponent(bc.name, bc.node)
		if err != nil {
			return nil, fmt.Errorf("components: blueprint %q: %w", b.ID, err)
		}
		e.AddWithOrder(c, bc.order)
	}
	return e, nil
}

// normalizeNode turns an explicit null node into nil so factories see "no
// parameters".
func normalizeNode(n *yaml.Node) *yaml.Node {
	if n == nil {
		return nil
	}
	if n.Kind == yaml.ScalarNode && n.Tag == "!!null" {
		return nil
	}
	return n
}

// componentOrder returns the order declared inside a component node, falling
// back to the registry default.
func componentOrder(name string, node *yaml.Node) (int, error) {
	if node != nil && node.Kind == yaml.MappingNode {
		content := node.Content
		for i := 0; i+1 < len(content); i += 2 {
			if content[i].Value == "order" {
				var o int
				if err := content[i+1].Decode(&o); err != nil {
					return 0, fmt.Errorf("order must be an integer: %w", err)
				}
				return o, nil
			}
		}
	}
	if o, ok := DefaultOrder(name); ok {
		return o, nil
	}
	return 0, nil
}

func validLayer(n yaml.Node) bool {
	if n.Tag == "!!null" {
		return true
	}
	if n.Kind != yaml.ScalarNode {
		return false
	}
	var i int
	if n.Decode(&i) == nil {
		return true
	}
	switch strings.ToLower(n.Value) {
	case "background", "arena", "entity", "ui", "text", "overlay":
		return true
	default:
		return false
	}
}

// decodeLayer maps a layer name or integer to a draw-queue layer.
func decodeLayer(n yaml.Node) int {
	if n.Kind != yaml.ScalarNode {
		return 0
	}
	var i int
	if err := n.Decode(&i); err == nil {
		return i
	}
	switch strings.ToLower(n.Value) {
	case "background":
		return queues.LayerBackground
	case "arena":
		return queues.LayerArena
	case "entity":
		return queues.LayerEntity
	case "ui":
		return queues.LayerUI
	case "text":
		return queues.LayerText
	case "overlay":
		return queues.LayerOverlay
	}
	return 0
}
