package components

import (
	"fmt"

	"github.com/mcbalaam/ebitter/components"
)

// BattleAction describes a selectable action and its resolved effects.
type BattleAction struct {
	Name    string       `yaml:"name"`
	Target  string       `yaml:"target"`
	Effects []EffectSpec `yaml:"effects"`
}

func (*BattleAction) Type() string { return "battleAction" }

func init() {
	registerBattleComponent("battleAction", 10, func() *BattleAction { return &BattleAction{} }, func(action *BattleAction) error {
		if action.Name == "" {
			return fmt.Errorf("name must not be empty")
		}
		if action.Target == "" {
			return fmt.Errorf("target must not be empty")
		}
		return validateEffects(action.Effects)
	})
}

var _ components.Component = (*BattleAction)(nil)
