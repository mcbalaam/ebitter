package scenes

import (
	"fmt"

	democomponents "github.com/mcbalaam/ebitter/cmd/demo/components"
	"github.com/mcbalaam/ebitter/components"
)

type menuOption struct {
	label     string
	prototype string
	dialogue  []string
}

func buildMenuOptions(library *components.BlueprintLibrary, battle *democomponents.BattleController, inventory *democomponents.Inventory) (map[string][]menuOption, error) {
	options := map[string][]menuOption{
		"fightButton": {},
		"actButton":   {},
		"itemButton":  {},
		"mercyButton": {},
	}

	for _, enemyID := range battle.Enemies {
		enemy, err := loadMenuEntity(library, enemyID)
		if err != nil {
			return nil, fmt.Errorf("load enemy %q: %w", enemyID, err)
		}
		controller, ok := components.ComponentOf[*democomponents.EnemyController](enemy, "enemyController")
		if !ok {
			return nil, fmt.Errorf("enemy prototype %q has no enemyController", enemyID)
		}
		name := controller.Name
		if name == "" {
			name = enemyID
		}
		options["fightButton"] = append(options["fightButton"], menuOption{
			label: name, prototype: enemyID, dialogue: controller.FightDialogue,
		})

		for _, actionID := range controller.Actions {
			actionEntity, err := loadMenuEntity(library, actionID)
			if err != nil {
				return nil, fmt.Errorf("load action %q for enemy %q: %w", actionID, enemyID, err)
			}
			action, ok := components.ComponentOf[*democomponents.BattleAction](actionEntity, "battleAction")
			if !ok {
				return nil, fmt.Errorf("action prototype %q has no battleAction", actionID)
			}
			options["actButton"] = append(options["actButton"], menuOption{
				label: action.Name, prototype: actionID, dialogue: action.Dialogue,
			})
		}

		if controller.CanBeSpared {
			label := "Spare"
			if len(battle.Enemies) > 1 {
				label += " " + name
			}
			options["mercyButton"] = append(options["mercyButton"], menuOption{
				label: label, prototype: enemyID, dialogue: controller.SpareDialogue,
			})
		}
		if controller.CanFlee {
			label := "Flee"
			if len(battle.Enemies) > 1 {
				label += " " + name
			}
			options["mercyButton"] = append(options["mercyButton"], menuOption{
				label: label, prototype: enemyID, dialogue: controller.FleeDialogue,
			})
		}
	}

	for _, entry := range inventory.Items {
		itemEntity, err := loadMenuEntity(library, entry.Prototype)
		if err != nil {
			return nil, fmt.Errorf("load item %q: %w", entry.Prototype, err)
		}
		item, ok := components.ComponentOf[*democomponents.Item](itemEntity, "item")
		if !ok {
			return nil, fmt.Errorf("item prototype %q has no item component", entry.Prototype)
		}
		label := item.Name
		if entry.Count > 1 {
			label = fmt.Sprintf("%s x%d", label, entry.Count)
		}
		options["itemButton"] = append(options["itemButton"], menuOption{
			label: label, prototype: entry.Prototype, dialogue: item.Dialogue,
		})
	}

	return options, nil
}

func loadMenuEntity(library *components.BlueprintLibrary, id string) (*components.Entity, error) {
	blueprint, err := library.Load(id)
	if err != nil {
		return nil, err
	}
	entity, err := blueprint.Instantiate()
	if err != nil {
		return nil, err
	}
	return entity, nil
}
