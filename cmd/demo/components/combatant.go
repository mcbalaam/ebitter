package components

import (
	"fmt"

	enginecomponents "github.com/mcbalaam/ebitter/components"
)

// Combatant contains mutable battle stats shared by the player and enemies.
type Combatant struct {
	HP      int `yaml:"hp"`
	MaxHP   int `yaml:"maxHP"`
	Defense int `yaml:"defense"`
}

func (*Combatant) Type() string { return "combatant" }

func (c *Combatant) Init(*enginecomponents.Entity) {
	if c.HP <= 0 && c.MaxHP > 0 {
		c.HP = c.MaxHP
	}
}

func (c *Combatant) Damage(amount int) int {
	if amount <= 0 || c.HP <= 0 {
		return 0
	}
	dealt := amount - c.Defense
	if dealt < 1 {
		dealt = 1
	}
	if dealt > c.HP {
		dealt = c.HP
	}
	c.HP -= dealt
	return dealt
}

func (c *Combatant) Heal(amount int) int {
	if amount <= 0 || c.HP <= 0 {
		return 0
	}
	old := c.HP
	c.HP += amount
	if c.HP > c.MaxHP {
		c.HP = c.MaxHP
	}
	return c.HP - old
}

func (c *Combatant) Defeated() bool { return c == nil || c.HP <= 0 }

func init() {
	registerBattleComponent("combatant", -5, func() *Combatant { return &Combatant{} }, func(c *Combatant) error {
		if c.MaxHP <= 0 {
			return fmt.Errorf("maxHP must be greater than zero")
		}
		if c.HP < 0 || c.HP > c.MaxHP {
			return fmt.Errorf("hp must be between zero and maxHP")
		}
		if c.Defense < 0 {
			return fmt.Errorf("defense must not be negative")
		}
		return nil
	})
}
