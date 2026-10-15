package components

import "time"

// TriggerMode selects how a Trigger fires.
type TriggerMode int

const (
	// TriggerTouch fires once when a box enters the zone.
	TriggerTouch TriggerMode = iota
	// TriggerButton fires when a box is inside the zone and the interact
	// button is pressed; it re-arms only after the box leaves.
	TriggerButton
)

// Trigger is a sensor zone attached to an entity. It emits its Signal on the
// entity's bus when activated by Check.
type Trigger struct {
	W       float64 `yaml:"w"`
	H       float64 `yaml:"h"`
	OffsetX float64 `yaml:"offsetX"`
	OffsetY float64 `yaml:"offsetY"`
	Signal  string  `yaml:"signal"`
	Mode    string  `yaml:"mode"`
	Once    bool    `yaml:"once"`

	active bool
	fired  bool
}

func (t *Trigger) Type() string { return "trigger" }

func (t *Trigger) Update(e *Entity, dt time.Duration) {}

// Rect returns the trigger's world-space box for the given entity.
func (t *Trigger) Rect(e *Entity) AABB {
	tf := e.Transform()
	if tf == nil {
		return AABB{}
	}
	cx := tf.X + t.OffsetX
	cy := tf.Y + t.OffsetY
	return AABB{X: cx - t.W/2, Y: cy - t.H/2, W: t.W, H: t.H}
}

// TriggerMode resolves the component's string mode.
func (t *Trigger) TriggerMode() TriggerMode {
	if isButtonMode(t.Mode) {
		return TriggerButton
	}
	return TriggerTouch
}

// Check evaluates the trigger against box and, on activation, emits its signal
// (Source = source, Data = e). It reports whether it fired on this call.
// interactPressed only matters for TriggerButton.
func (t *Trigger) Check(e *Entity, box AABB, interactPressed bool, source any) bool {
	if t.Once && t.fired {
		return false
	}
	overlapping := t.Rect(e).Overlaps(box)
	switch t.TriggerMode() {
	case TriggerButton:
		if overlapping && interactPressed && !t.active {
			t.active = true
			return t.emit(e, source)
		}
	default:
		if overlapping && !t.active {
			t.active = true
			return t.emit(e, source)
		}
	}
	if !overlapping {
		t.active = false
	}
	return false
}

func (t *Trigger) emit(e *Entity, source any) bool {
	t.fired = true
	if t.Signal != "" {
		e.signalBus().Emit(t.Signal, source, e)
	}
	return true
}

func isButtonMode(mode string) bool {
	return mode == "button" || mode == "Button" || mode == "BUTTON"
}
