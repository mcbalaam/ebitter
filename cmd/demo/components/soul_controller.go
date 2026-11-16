package components

import (
	"fmt"

	"github.com/mcbalaam/ebitter/components"
)

// SoulController configures soul movement and hit invulnerability.
type SoulController struct {
	Speed             float64 `yaml:"speed"`
	InvulnerabilityMS int     `yaml:"invulnerabilityMS"`
}

func (*SoulController) Type() string { return "soulController" }

func init() {
	registerBattleComponent("soulController", 10, func() *SoulController { return &SoulController{} }, func(soul *SoulController) error {
		if soul.Speed < 0 || soul.InvulnerabilityMS < 0 {
			return fmt.Errorf("speed and invulnerabilityMS must not be negative")
		}
		return nil
	})
}

var _ components.Component = (*SoulController)(nil)
