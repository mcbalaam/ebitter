package mainscene

import (
	"os"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/assetfs"
	democomponents "github.com/mcbalaam/ebitter/cmd/demo/components"
	"github.com/mcbalaam/ebitter/components"
	"github.com/mcbalaam/ebitter/scene"
	"github.com/mcbalaam/ebitter/text"
)

func configureMenuTestAssets(t *testing.T) {
	t.Helper()
	previousFS, previousSprites, previousFonts := assetfs.FS, components.SpriteRoot, text.FontRoot
	assetfs.SetFS(os.DirFS("../.."))
	components.SpriteRoot, text.FontRoot = "media/sprites", "media/sprites"
	t.Cleanup(func() {
		assetfs.SetFS(previousFS)
		components.SpriteRoot, text.FontRoot = previousSprites, previousFonts
	})
}

func TestSubmenuCompletesNarrativeAndCancelKeepsItRevealed(t *testing.T) {
	configureMenuTestAssets(t)
	s := &demoScene{
		runtime: scene.NewRuntime(),
		battle: &democomponents.BattleController{
			Phase:        democomponents.BattlePhaseCommandSelect,
			CommandLines: []string{"* Beginning.$p5000$n* Ending.$e"},
		},
		input:               newDemoInput(),
		buttonNames:         []string{"actButton"},
		commandDialogueTurn: -1,
		menuOptions:         map[string][]menuOption{"actButton": {{label: "Check", prototype: "check"}}},
	}
	s.runtime.Dialog.ShowAllChecker = s.dialogueXPressed
	s.runtime.Dialog.AdvanceChecker = s.dialogueAdvancePressed
	s.updateDialogue(0)
	if s.runtime.Dialog.RevealedAll() {
		t.Fatal("narrative must still be typing before entering the submenu")
	}
	s.openMenu()
	if !s.menuActive || !s.runtime.Dialog.Active() || !s.runtime.Dialog.RevealedAll() {
		t.Fatal("opening a submenu must instantly finish the narrative, including pauses")
	}
	s.updateDialogue(10 * time.Second)
	s.input.justPress[ebiten.KeyX] = true
	s.updateMenu()
	s.updateDialogue(0)
	if s.menuActive || !s.runtime.Dialog.Active() || !s.runtime.Dialog.RevealedAll() {
		t.Fatal("cancel must return to the fully revealed narrative without dismissing or restarting it")
	}
}

func TestUsedConsumablesDisappearAndStackCountsRefresh(t *testing.T) {
	configureMenuTestAssets(t)
	s := &demoScene{
		runtime:              scene.NewRuntime(),
		battle:               &democomponents.BattleController{Phase: democomponents.BattlePhaseCommandSelect},
		inventory:            &democomponents.Inventory{Items: []democomponents.InventoryEntry{{Prototype: "spiderDonut", Count: 2}}},
		prototypes:           components.NewBlueprintLibrary(os.DirFS("../../prototypes"), "."),
		input:                newDemoInput(),
		buttonNames:          []string{"itemButton"},
		enemyAttackDurations: [][]time.Duration{{3 * time.Second}},
	}
	var err error
	s.menuOptions, err = buildMenuOptions(s.prototypes, s.battle, s.inventory)
	if err != nil {
		t.Fatal(err)
	}
	s.openMenu()
	s.input.justPress[ebiten.KeyX] = true
	s.updateMenu()
	if s.inventory.Count("spiderDonut") != 2 {
		t.Fatal("cancelling the item menu must not consume an item")
	}
	clear(s.input.justPress)
	for remaining := 1; remaining >= 0; remaining-- {
		s.openMenu()
		s.confirmMenuChoice()
		if s.inventory.Count("spiderDonut") != remaining || s.battle.Phase != democomponents.BattlePhasePlayerAction || !s.runtime.Dialog.Active() {
			t.Fatal("confirming a consumable must consume exactly one and begin its action dialogue")
		}
		options := s.menuOptions["itemButton"]
		if remaining == 1 && (len(options) != 1 || options[0].label != "Spider Donut") {
			t.Fatalf("remaining stack must update its label: %v", options)
		}
		if remaining == 0 && len(options) != 0 {
			t.Fatal("the last consumed item must disappear from the item menu")
		}
		s.runtime.Dialog.Clear()
		s.finishPlayerActionDialogue()
		s.finishEnemyDialogue()
		s.updateAttack(3 * time.Second)
	}
	s.openMenu()
	s.confirmMenuChoice()
	if s.battle.Phase != democomponents.BattlePhaseCommandSelect || s.inventory.Count("spiderDonut") != 0 {
		t.Fatal("an exhausted item must not be usable on a later turn")
	}
}

func TestNonConsumableItemRemainsAfterUse(t *testing.T) {
	configureMenuTestAssets(t)
	option := menuOption{
		label: "Key", prototype: "key", dialogue: []string{"* You use the key."},
		item: &democomponents.Item{Name: "Key", Consumable: false},
	}
	s := &demoScene{
		runtime:     scene.NewRuntime(),
		battle:      &democomponents.BattleController{Phase: democomponents.BattlePhaseCommandSelect},
		inventory:   &democomponents.Inventory{Items: []democomponents.InventoryEntry{{Prototype: "key", Count: 1}}},
		buttonNames: []string{"itemButton"},
		menuOptions: map[string][]menuOption{"itemButton": {option}},
	}
	s.openMenu()
	s.confirmMenuChoice()
	if s.inventory.Count("key") != 1 || len(s.menuOptions["itemButton"]) != 1 || s.battle.Phase != democomponents.BattlePhasePlayerAction {
		t.Fatal("non-consumable items must stay available after a successful use")
	}
}
