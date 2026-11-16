package components

import (
	"fmt"

	"github.com/mcbalaam/ebitter/components"
)

type BattlePhase string

const (
	BattlePhaseCommandSelect BattlePhase = "commandSelect"
	BattlePhaseTargetSelect  BattlePhase = "targetSelect"
	BattlePhasePlayerAction  BattlePhase = "playerAction"
	BattlePhaseEnemyTurn     BattlePhase = "enemyTurn"
	BattlePhaseAttack        BattlePhase = "attack"
	BattlePhaseVictory       BattlePhase = "victory"
	BattlePhaseDefeat        BattlePhase = "defeat"
	BattlePhaseFled          BattlePhase = "fled"
)

// BattleController holds current battle phase and selection indices.
type BattleController struct {
	Phase          BattlePhase `yaml:"phase"`
	Turn           int         `yaml:"turn"`
	SelectedTarget int         `yaml:"selectedTarget"`
}

func (*BattleController) Type() string { return "battleController" }

func (b *BattleController) Init(*components.Entity) {
	if b.Phase == "" {
		b.Phase = BattlePhaseCommandSelect
	}
}

func (b *BattleController) SetPhase(phase BattlePhase) { b.Phase = phase }

func init() {
	registerBattleComponent("battleController", -100, func() *BattleController { return &BattleController{} }, func(controller *BattleController) error {
		if controller.Turn < 0 || controller.SelectedTarget < 0 {
			return fmt.Errorf("turn and selectedTarget must not be negative")
		}
		switch controller.Phase {
		case "", BattlePhaseCommandSelect, BattlePhaseTargetSelect, BattlePhasePlayerAction,
			BattlePhaseEnemyTurn, BattlePhaseAttack, BattlePhaseVictory, BattlePhaseDefeat, BattlePhaseFled:
			return nil
		default:
			return fmt.Errorf("unknown phase %q", controller.Phase)
		}
	})
}
