package tiled

import "strings"

// Schema describes the naming conventions a map uses for object groups, custom
// properties and trigger values. DefaultSchema matches the conventions
// documented in QUICKSTART.md, so existing maps need no changes.
type Schema struct {
	// CollisionKeywords match (case-insensitively, by substring) the name or
	// class of an object group that contributes collision boxes.
	CollisionKeywords []string
	// InteractionKeywords match the name or class of an interaction group.
	InteractionKeywords []string

	// CollisionProperty is a boolean property that marks a group as collision.
	CollisionProperty string
	// SignalProperty names the property holding an interaction's signal.
	SignalProperty string
	// InteractionProperty names the property giving an interaction its name.
	InteractionProperty string
	// TriggerProperty names the property selecting the trigger mode.
	TriggerProperty string
	// NameProperty is the fallback property for an object's name.
	NameProperty string
	// BlueprintProperty names the property that places an entity prefab.
	BlueprintProperty string
	// MinZoomProperty / MaxZoomProperty name the camera zoom limit properties.
	MinZoomProperty string
	MaxZoomProperty string

	// DefaultSignal is emitted by a zone that declares no signal property.
	DefaultSignal string
	// ButtonTriggers lists TriggerProperty values that select button mode.
	ButtonTriggers []string
}

// DefaultSchema returns the conventions used by the engine by default.
func DefaultSchema() Schema {
	return Schema{
		CollisionKeywords:   []string{"collision", "hitbox"},
		InteractionKeywords: []string{"interact"},
		CollisionProperty:   "collision",
		SignalProperty:      "signal",
		InteractionProperty: "interaction",
		TriggerProperty:     "trigger",
		NameProperty:        "name",
		BlueprintProperty:   "blueprint",
		MinZoomProperty:     "min_zoom",
		MaxZoomProperty:     "max_zoom",
		DefaultSignal:       "interact",
		ButtonTriggers:      []string{"button"},
	}
}

// hasKeyword reports whether name or class contains any of the keywords
// (case-insensitive substring match).
func (s Schema) hasKeyword(keywords []string, name, class string) bool {
	ln, lc := strings.ToLower(name), strings.ToLower(class)
	for _, kw := range keywords {
		kw = strings.ToLower(kw)
		if kw == "" {
			continue
		}
		if strings.Contains(ln, kw) || strings.Contains(lc, kw) {
			return true
		}
	}
	return false
}

// isButtonTrigger reports whether value selects the button trigger mode.
func (s Schema) isButtonTrigger(value string) bool {
	v := strings.ToLower(value)
	for _, t := range s.ButtonTriggers {
		if strings.ToLower(t) == v {
			return true
		}
	}
	return false
}
