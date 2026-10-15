package components

import (
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

// Factory builds a component from its YAML node. node may be nil for a
// component declared with no parameters.
type Factory func(node *yaml.Node) (Component, error)

type registration struct {
	order   int
	factory Factory
}

var (
	registryMu sync.RWMutex
	registry   = map[string]registration{}
)

// Register adds (or replaces) a component type under name with its default
// update/draw order and a YAML factory. The order is used when a blueprint
// omits an explicit order.
func Register(name string, order int, factory Factory) {
	registryMu.Lock()
	registry[name] = registration{order: order, factory: factory}
	registryMu.Unlock()
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
	return r.factory(node)
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
