package components

import (
	"fmt"

	"github.com/mcbalaam/ebitter/components"
)

// Item describes an inventory item prototype and its effects.
type Item struct {
	Name       string       `yaml:"name"`
	Target     string       `yaml:"target"`
	Consumable bool         `yaml:"consumable"`
	Effects    []EffectSpec `yaml:"effects"`
}

func (*Item) Type() string { return "item" }

func init() {
	registerBattleComponent("item", 10, func() *Item { return &Item{} }, func(item *Item) error {
		if item.Name == "" {
			return fmt.Errorf("name must not be empty")
		}
		if item.Target == "" {
			return fmt.Errorf("target must not be empty")
		}
		return validateEffects(item.Effects)
	})
}

var _ components.Component = (*Item)(nil)
