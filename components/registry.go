package components

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Factory builds a component from its YAML node. node may be nil for a
// component declared with no parameters.
type Factory func(node *yaml.Node) (Component, error)

// Validator checks a component's YAML configuration without constructing it.
type Validator func(node *yaml.Node) error

type registration struct {
	order    int
	factory  Factory
	validate Validator
}

var (
	registryMu sync.RWMutex
	registry   = map[string]registration{}
)

// Register adds a component type under name with its default update/draw order,
// YAML factory, and YAML validator. The order is used when a blueprint omits an
// explicit order. The validator must check the component node without creating
// runtime resources; component-level "order" is reserved by the blueprint.
// Registration fails if name is empty, factory is nil, or name is already in
// use; registrations cannot be silently replaced.
func Register(name string, order int, factory Factory, validate Validator) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("components: component name must not be empty")
	}
	if factory == nil {
		return fmt.Errorf("components: component %q has no factory", name)
	}
	if validate == nil {
		return fmt.Errorf("components: component %q has no validator", name)
	}

	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := registry[name]; exists {
		return fmt.Errorf("components: component %q is already registered", name)
	}
	registry[name] = registration{order: order, factory: factory, validate: validate}
	return nil
}

// ValidateComponent checks a YAML node with the registered component validator.
func ValidateComponent(name string, node *yaml.Node) error {
	registryMu.RLock()
	r, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return fmt.Errorf("components: unknown component %q", name)
	}
	if err := r.validate(node); err != nil {
		return fmt.Errorf("components: validate component %q: %w", name, err)
	}
	return nil
}

func validateKeys(allowed ...string) Validator {
	keys := make(map[string]bool, len(allowed)+1)
	for _, key := range allowed {
		keys[key] = true
	}
	keys["order"] = true
	return func(node *yaml.Node) error {
		if node == nil {
			return nil
		}
		if node.Kind != yaml.MappingNode {
			return fmt.Errorf("expected mapping node (kind %d)", node.Kind)
		}
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i].Value
			if !keys[key] {
				return fmt.Errorf("unknown field %q", key)
			}
		}
		return nil
	}
}

func strictConfig[T any](newT func() T) Validator {
	return func(node *yaml.Node) error {
		if node == nil {
			return nil
		}
		if node.Kind != yaml.MappingNode {
			return fmt.Errorf("expected mapping node (kind %d)", node.Kind)
		}
		// order is blueprint metadata rather than a component config field.
		copyNode := *node
		copyNode.Content = make([]*yaml.Node, 0, len(node.Content))
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value == "order" {
				continue
			}
			copyNode.Content = append(copyNode.Content, node.Content[i], node.Content[i+1])
		}
		data, err := yaml.Marshal(&copyNode)
		if err != nil {
			return err
		}
		dec := yaml.NewDecoder(bytes.NewReader(data))
		dec.KnownFields(true)
		return dec.Decode(newT())
	}
}

func vectorFields(validator Validator, fields ...string) Validator {
	return func(node *yaml.Node) error {
		if err := validator(node); err != nil {
			return err
		}
		if node == nil {
			return nil
		}
		for i := 0; i+1 < len(node.Content); i += 2 {
			for _, field := range fields {
				if node.Content[i].Value != field || node.Content[i+1].Tag == "!!null" {
					continue
				}
				if node.Content[i+1].Kind != yaml.SequenceNode || len(node.Content[i+1].Content) != 2 {
					return fmt.Errorf("%s must contain exactly two values", field)
				}
			}
		}
		return nil
	}
}

// DefaultOrder returns the registered default order for a component type.
func DefaultOrder(name string) (int, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	r, ok := registry[name]
	return r.order, ok
}

// Registered reports whether a component type has been registered.
func Registered(name string) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	_, ok := registry[name]
	return ok
}

// NewComponent builds the component registered under name from a YAML node.
func NewComponent(name string, node *yaml.Node) (Component, error) {
	registryMu.RLock()
	r, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("components: unknown component %q", name)
	}
	if r.factory == nil {
		return nil, fmt.Errorf("components: component %q has no factory", name)
	}
	c, err := r.factory(node)
	if err != nil {
		return nil, fmt.Errorf("components: create component %q: %w", name, err)
	}
	if c == nil {
		return nil, fmt.Errorf("components: factory for %q returned nil", name)
	}
	if got := c.Type(); got != name {
		return nil, fmt.Errorf("components: factory for %q returned component type %q", name, got)
	}
	return c, nil
}

// decode builds a Factory that decodes a YAML node into a new T. The component
// struct is expected to carry yaml tags for any non-lowercase field names.
func decode[T Component](newT func() T) Factory {
	return func(node *yaml.Node) (Component, error) {
		v := newT()
		if node != nil {
			if err := node.Decode(v); err != nil {
				return nil, err
			}
		}
		return v, nil
	}
}

// decodeYAML decodes node into v, tolerating a nil node (leaving v untouched).
func decodeYAML(node *yaml.Node, v any) error {
	if node == nil {
		return nil
	}
	return node.Decode(v)
}
