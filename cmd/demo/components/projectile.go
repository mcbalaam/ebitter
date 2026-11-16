package components

import (
	"fmt"

	"github.com/mcbalaam/ebitter/components"
)

// Projectile configures damage and lifetime for a projectile entity.
type Projectile struct {
	Damage     int    `yaml:"damage"`
	Owner      string `yaml:"owner"`
	LifetimeMS int    `yaml:"lifetimeMS"`
}

func (*Projectile) Type() string { return "projectile" }

func init() {
	registerBattleComponent("projectile", 10, func() *Projectile { return &Projectile{} }, func(projectile *Projectile) error {
		if projectile.Damage < 0 || projectile.LifetimeMS < 0 {
			return fmt.Errorf("damage and lifetimeMS must not be negative")
		}
		return nil
	})
}

var _ components.Component = (*Projectile)(nil)
