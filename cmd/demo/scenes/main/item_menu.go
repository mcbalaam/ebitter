package mainscene

import (
	"fmt"
	"slices"
)

func itemMenuLabel(name string, count int) string {
	if count > 1 {
		return fmt.Sprintf("%s x%d", name, count)
	}
	return name
}

func (s *demoScene) consumeMenuItem(option menuOption) bool {
	if option.item == nil || !option.item.Consumable {
		return true
	}
	if !s.inventory.Remove(option.prototype, 1) {
		return false
	}
	count := s.inventory.Count(option.prototype)
	options := s.menuOptions["itemButton"]
	for index, entry := range options {
		if entry.prototype != option.prototype {
			continue
		}
		if count == 0 {
			s.menuOptions["itemButton"] = slices.Delete(options, index, index+1)
		} else {
			options[index].label = itemMenuLabel(option.item.Name, count)
		}
		break
	}
	return true
}
