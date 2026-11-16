package components

import (
	"fmt"

	enginecomponents "github.com/mcbalaam/ebitter/components"
	"github.com/mcbalaam/ebitter/events"
	"gopkg.in/yaml.v3"
)

const SignalBattleButtonActive = "battleButton.active"

// BattleButton maps an activity signal addressed to its entity to the sprite's
// on/off state.
type BattleButton struct {
	active bool
}

func (*BattleButton) Type() string { return "battleButton" }

func (b *BattleButton) Init(e *enginecomponents.Entity) {
	e.Signals().Subscribe(SignalBattleButtonActive, e, func(sig events.Signal) {
		if sig.Source != e {
			return
		}
		active, ok := sig.Data.(bool)
		if !ok {
			return
		}
		b.active = active
		if sprite := e.Sprite(); sprite != nil && sprite.Icon != nil {
			state := "off"
			if active {
				state = "on"
			}
			_ = sprite.Icon.SetIconState(state)
		}
	})
}

func (b *BattleButton) Active() bool { return b.active }

// SetBattleButtonActive emits an activity change addressed to e.
func SetBattleButtonActive(e *enginecomponents.Entity, active bool) {
	if e != nil {
		e.Signals().Emit(SignalBattleButtonActive, e, active)
	}
}

func init() {
	err := enginecomponents.Register("battleButton", 30,
		func(*yaml.Node) (enginecomponents.Component, error) {
			return &BattleButton{}, nil
		}, func(node *yaml.Node) error {
			if node == nil {
				return nil
			}
			if node.Kind != yaml.MappingNode {
				return fmt.Errorf("expected mapping")
			}
			for i := 0; i+1 < len(node.Content); i += 2 {
				if node.Content[i].Value != "order" {
					return fmt.Errorf("unknown field %q", node.Content[i].Value)
				}
			}
			return nil
		})
	if err != nil {
		panic(err)
	}
}
