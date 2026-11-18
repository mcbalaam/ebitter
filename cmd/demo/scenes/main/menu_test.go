package mainscene

import (
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/assetfs"
	democomponents "github.com/mcbalaam/ebitter/cmd/demo/components"
	"github.com/mcbalaam/ebitter/components"
	"github.com/mcbalaam/ebitter/render"
	gametext "github.com/mcbalaam/ebitter/text"
)

func TestMenuOptionsComeFromEncounterAndInventoryPrototypes(t *testing.T) {
	previousFS, previousSpriteRoot := assetfs.FS, components.SpriteRoot
	assetfs.SetFS(os.DirFS("../.."))
	components.SpriteRoot = "media/sprites"
	t.Cleanup(func() {
		assetfs.SetFS(previousFS)
		components.SpriteRoot = previousSpriteRoot
	})
	library := components.NewBlueprintLibrary(os.DirFS("../../prototypes"), ".")
	blueprint, err := library.Load("trainingEncounter")
	if err != nil {
		t.Fatal(err)
	}
	encounter, err := blueprint.Instantiate()
	if err != nil {
		t.Fatal(err)
	}
	battle, ok := components.ComponentOf[*democomponents.BattleController](encounter, "battleController")
	if !ok {
		t.Fatal("training encounter has no battleController")
	}
	inventory, ok := components.ComponentOf[*democomponents.Inventory](encounter, "inventory")
	if !ok {
		t.Fatal("training encounter has no inventory")
	}

	options, err := buildMenuOptions(library, battle, inventory)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"fightButton": "Training Dummy",
		"actButton":   "Check",
		"itemButton":  "Spider Donut",
		"mercyButton": "Spare",
	}
	for button, label := range want {
		if len(options[button]) != 1 || options[button][0].label != label {
			t.Errorf("%s menu options = %#v, want one entry %q", button, options[button], label)
		}
	}
	if got := options["actButton"][0].dialogue; len(got) != 1 || got[0] != "* Training Dummy - ATK 0 DEF 0$p500$n* Text description." {
		t.Fatalf("action dialogue = %v, want prototype dialogue", got)
	}
}
func TestActionDialogueCompletionAdvancesBattlePhase(t *testing.T) {
	battle := &democomponents.BattleController{Phase: democomponents.BattlePhasePlayerAction}
	s := &demoScene{battle: battle, actionDialoguePending: true}

	s.finishPlayerActionDialogue()
	if battle.Phase != democomponents.BattlePhaseEnemyTurn || battle.Turn != 1 {
		t.Fatalf("after action dialogue: phase=%q turn=%d, want enemyTurn and turn 1", battle.Phase, battle.Turn)
	}
	if s.actionDialoguePending {
		t.Fatal("action dialogue should no longer be pending")
	}
}

func TestMoveSoulToMenuEntryUsesEntryAnchor(t *testing.T) {
	s := &demoScene{
		menuIndex: 1,
		menuEntries: []*menuEntry{
			{anchorX: 72, anchorY: 286},
			{anchorX: 72, anchorY: 318},
		},
	}

	s.moveSoulToMenuEntry()
	if s.soulX != 72 || s.soulY != 318 {
		t.Fatalf("soul position = (%v, %v), want anchor (72, 318)", s.soulX, s.soulY)
	}
}

func TestSoulHiddenOutsidePlayerSelectionPhases(t *testing.T) {
	s := &demoScene{
		soul:   &render.AnimatedIcon{},
		battle: &democomponents.BattleController{Phase: democomponents.BattlePhasePlayerAction},
	}
	if s.soulVisible() {
		t.Fatal("soul should be hidden during player action")
	}
	s.battle.SetPhase(democomponents.BattlePhaseCommandSelect)
	if !s.soulVisible() {
		t.Fatal("soul should be visible during command selection")
	}
}

func TestPhaseLabelTracksBattleController(t *testing.T) {
	s := &demoScene{
		battle:     &democomponents.BattleController{Phase: democomponents.BattlePhaseEnemyTurn},
		phaseLabel: gametext.NewTextString("commandSelect", 8, 8, gametext.TextStyle{}),
	}
	s.updatePhaseLabel()
	if s.phaseLabel.Text != string(democomponents.BattlePhaseEnemyTurn) {
		t.Fatalf("phase label = %q, want %q", s.phaseLabel.Text, democomponents.BattlePhaseEnemyTurn)
	}
}

func TestButtonsReturnToNormalDuringPlayerAction(t *testing.T) {
	s := &demoScene{
		selectedButton: 0,
		battle: &democomponents.BattleController{
			Phase: democomponents.BattlePhasePlayerAction,
		},
	}
	if s.buttonActive(0) {
		t.Fatal("selected button should be normal during player action")
	}
	s.battle.SetPhase(democomponents.BattlePhaseCommandSelect)
	if !s.buttonActive(0) {
		t.Fatal("selected button should be active during command selection")
	}
}

func TestCommandDialogueUsesTurnIndex(t *testing.T) {
	lines := []string{"first", "second"}
	for turn, want := range lines {
		got, ok := commandLineForTurn(lines, turn)
		if !ok || got != want {
			t.Fatalf("turn %d: got %q, available=%v; want %q", turn, got, ok, want)
		}
	}
	if got, ok := commandLineForTurn(lines, len(lines)); !ok || got != lines[0] {
		t.Fatalf("expected command lines to cycle, got %q, available=%v", got, ok)
	}
	if _, ok := commandLineForTurn(nil, 0); ok {
		t.Fatal("empty command lines should not start a dialogue")
	}
	if _, ok := commandLineForTurn(lines, -1); ok {
		t.Fatal("negative turns should not select a command line")
	}
}

func TestDialogueXDoesNotDismissWhileSubmenuIsOpen(t *testing.T) {
	s := &demoScene{input: &demoInput{
		justPress: map[ebiten.Key]bool{ebiten.KeyX: true},
	}}
	if !s.dialogueXPressed() {
		t.Fatal("X should control the dialog while the main menu is open")
	}
	s.battle = &democomponents.BattleController{Phase: democomponents.BattlePhaseCommandSelect}
	if s.dialogueAdvancePressed() {
		t.Fatal("X should not dismiss the command dialogue while the main menu is open")
	}
	s.menuActive = true
	if s.dialogueXPressed() {
		t.Fatal("X should cancel the submenu without dismissing the background dialog")
	}
}

func TestActionDialogueUsesZToAdvanceAndXOnlyToReveal(t *testing.T) {
	s := &demoScene{
		battle: &democomponents.BattleController{Phase: democomponents.BattlePhasePlayerAction},
		input:  &demoInput{justPress: map[ebiten.Key]bool{ebiten.KeyX: true}},
	}
	if s.dialogueAdvancePressed() {
		t.Fatal("X should reveal text, not advance an action dialogue")
	}
	s.input.justPress = map[ebiten.Key]bool{ebiten.KeyZ: true}
	if !s.dialogueAdvancePressed() {
		t.Fatal("Z should advance an action dialogue")
	}
}

func TestHPBarWidthScalesWithMaxHP(t *testing.T) {
	if got := hpBarWidth(10); got != 15 {
		t.Fatalf("width for maxHP 10 = %v, want 15", got)
	}
	if got := hpBarWidth(20); got != 30 {
		t.Fatalf("width for maxHP 20 = %v, want 30", got)
	}
	if got := hpBarWidth(100); got != 150 {
		t.Fatalf("width for maxHP 100 = %v, want capped width 150", got)
	}
}
