package components

// EffectSpec is a serializable instruction for the battle effect runner.
// Unique mechanics can refer to a named handler instead of embedding code.
type EffectSpec struct {
	Type    string `yaml:"type"`
	Target  string `yaml:"target,omitempty"`
	Amount  int    `yaml:"amount,omitempty"`
	Status  string `yaml:"status,omitempty"`
	Item    string `yaml:"item,omitempty"`
	Handler string `yaml:"handler,omitempty"`
}
