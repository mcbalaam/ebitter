package components

// Interactable is a Trigger that also carries a UI prompt. It defaults to
// button activation; rendering the prompt is left to the game.
type Interactable struct {
	*Trigger
	Prompt string `yaml:"prompt"`
}

// NewInteractable builds an interactable from a trigger and a prompt.
func NewInteractable(t *Trigger, prompt string) *Interactable {
	if t == nil {
		t = &Trigger{}
	}
	return &Interactable{Trigger: t, Prompt: prompt}
}

func (i *Interactable) Type() string { return "interactable" }
