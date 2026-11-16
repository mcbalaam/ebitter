package components

// EnemyController is prototype data for enemy dialogue, actions and attacks.
// Actions and attacks are IDs of other entity prototypes.
type EnemyController struct {
	Lines       []string `yaml:"lines"`
	Actions     []string `yaml:"actions"`
	Attacks     []string `yaml:"attacks"`
	CanBeSpared bool     `yaml:"canBeSpared"`
	CanFlee     bool     `yaml:"canFlee"`
}

func (*EnemyController) Type() string { return "enemyController" }

func init() {
	registerBattleComponent("enemyController", 10, func() *EnemyController { return &EnemyController{} }, func(c *EnemyController) error {
		if err := uniqueNonEmpty("actions", c.Actions); err != nil {
			return err
		}
		return uniqueNonEmpty("attacks", c.Attacks)
	})
}
