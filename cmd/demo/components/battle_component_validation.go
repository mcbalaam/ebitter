package components

import "fmt"

func uniqueNonEmpty(field string, values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			return fmt.Errorf("%s must not contain empty prototype IDs", field)
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("%s contains duplicate prototype %q", field, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateEffects(effects []EffectSpec) error {
	for n, effect := range effects {
		switch effect.Type {
		case "damage", "heal", "setMercy", "addStatus", "removeStatus", "removeItem", "flee", "say":
		case "":
			return fmt.Errorf("effects[%d].type must not be empty", n)
		default:
			if effect.Handler == "" {
				return fmt.Errorf("effects[%d]: unsupported type %q (custom effects need a handler)", n, effect.Type)
			}
		}
		if (effect.Type == "damage" || effect.Type == "heal") && effect.Amount < 0 {
			return fmt.Errorf("effects[%d].amount must not be negative", n)
		}
		if (effect.Type == "addStatus" || effect.Type == "removeStatus") && effect.Status == "" {
			return fmt.Errorf("effects[%d].status is required", n)
		}
		if effect.Type == "removeItem" && effect.Item == "" {
			return fmt.Errorf("effects[%d].item is required", n)
		}
	}
	return nil
}
