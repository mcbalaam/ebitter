package components

import (
	"fmt"

	"github.com/mcbalaam/ebitter/components"
)

type AttackEvent struct {
	AtMS       int     `yaml:"atMS"`
	Projectile string  `yaml:"projectile"`
	X          float64 `yaml:"x"`
	Y          float64 `yaml:"y"`
	VelocityX  float64 `yaml:"velocityX"`
	VelocityY  float64 `yaml:"velocityY"`
}

// AttackPattern is a timed list of projectile prototype spawns.
type AttackPattern struct {
	DurationMS int           `yaml:"durationMS"`
	Events     []AttackEvent `yaml:"events"`
}

func (*AttackPattern) Type() string { return "attackPattern" }

func init() {
	registerBattleComponent("attackPattern", 10, func() *AttackPattern { return &AttackPattern{} }, func(pattern *AttackPattern) error {
		if pattern.DurationMS < 0 {
			return fmt.Errorf("durationMS must not be negative")
		}
		lastAt := -1
		for n, event := range pattern.Events {
			if event.AtMS < 0 || event.AtMS < lastAt {
				return fmt.Errorf("events[%d].atMS must be non-negative and events ordered by time", n)
			}
			if pattern.DurationMS > 0 && event.AtMS > pattern.DurationMS {
				return fmt.Errorf("events[%d].atMS exceeds durationMS", n)
			}
			if event.Projectile == "" {
				return fmt.Errorf("events[%d].projectile must not be empty", n)
			}
			lastAt = event.AtMS
		}
		return nil
	})
}

var _ components.Component = (*AttackPattern)(nil)
