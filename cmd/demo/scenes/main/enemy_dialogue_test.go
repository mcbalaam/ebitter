package mainscene

import (
	"os"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/assetfs"
	democomponents "github.com/mcbalaam/ebitter/cmd/demo/components"
	"github.com/mcbalaam/ebitter/scene"
	"github.com/mcbalaam/ebitter/text"
)

func TestEnemyDialogueWaitsForZBeforeAttack(t *testing.T) {
	previousFS, previousFontRoot := assetfs.FS, text.FontRoot
	assetfs.SetFS(os.DirFS("../.."))
	text.FontRoot = "media/sprites"
	t.Cleanup(func() {
		assetfs.SetFS(previousFS)
		text.FontRoot = previousFontRoot
	})
	s := &demoScene{
		runtime: scene.NewRuntime(),
		battle:  &democomponents.BattleController{Phase: democomponents.BattlePhaseEnemyTurn, Turn: 1},
		input:   newDemoInput(),
		enemies: []*democomponents.EnemyController{
			{Lines: []string{"First enemy."}},
			{Lines: []string{"Second enemy."}},
		},
	}
	s.runtime.Dialog.ShowAllChecker = s.dialogueXPressed
	s.runtime.Dialog.AdvanceChecker = s.dialogueAdvancePressed
	s.updateDialogue(0)
	if !s.enemyDialoguePending || !s.runtime.Dialog.Active() {
		t.Fatal("enemy turn must start a dialogue")
	}
	for phrase := range 2 {
		s.input.justPress = map[ebiten.Key]bool{ebiten.KeyX: true}
		s.updateDialogue(time.Second / 60)
		s.updateDialogue(time.Second / 60)
		if !s.runtime.Dialog.RevealedAll() || s.battle.Phase != democomponents.BattlePhaseEnemyTurn {
			t.Fatal("X must reveal the phrase without advancing into the attack")
		}
		s.input.justPress = map[ebiten.Key]bool{ebiten.KeyZ: true}
		s.updateDialogue(time.Second / 60)
		if phrase == 0 && (!s.runtime.Dialog.Active() || !s.enemyDialoguePending) {
			t.Fatal("Z must advance to the next enemy before starting the attack")
		}
	}
	if s.battle.Phase != democomponents.BattlePhaseAttack || s.enemyDialoguePending || s.runtime.Dialog.Active() {
		t.Fatalf("after last Z: phase=%q, pending=%v, dialogue active=%v", s.battle.Phase, s.enemyDialoguePending, s.runtime.Dialog.Active())
	}
}

func TestEnemyLinesCycleByTurn(t *testing.T) {
	lines := []string{"First", "Second"}
	for turn, want := range []string{"First", "First", "Second", "First"} {
		if got, ok := enemyLineForTurn(lines, turn); !ok || got != want {
			t.Fatalf("turn %d: got %q, available=%v; want %q", turn, got, ok, want)
		}
	}
	if _, ok := enemyLineForTurn(nil, 1); ok {
		t.Fatal("an enemy without lines should not create a dialogue")
	}
}

func TestNarrativeRestartsAfterAttackWithSingleConfiguredLine(t *testing.T) {
	previousFS, previousFontRoot := assetfs.FS, text.FontRoot
	assetfs.SetFS(os.DirFS("../.."))
	text.FontRoot = "media/sprites"
	t.Cleanup(func() {
		assetfs.SetFS(previousFS)
		text.FontRoot = previousFontRoot
	})
	s := &demoScene{
		runtime: scene.NewRuntime(),
		battle: &democomponents.BattleController{
			Phase:        democomponents.BattlePhaseCommandSelect,
			CommandLines: []string{"* The dummy waits."},
		},
		input:                newDemoInput(),
		commandDialogueTurn:  -1,
		enemyAttackDurations: [][]time.Duration{{3 * time.Second}},
	}
	s.runtime.Dialog.ShowAllChecker = s.dialogueXPressed
	s.runtime.Dialog.AdvanceChecker = s.dialogueAdvancePressed
	s.maybeShowCommandDialogue()
	if !s.runtime.Dialog.Active() {
		t.Fatal("the first player turn must start its narrative")
	}
	for turn := range 2 {
		s.input.justPress[ebiten.KeyX] = true
		s.runtime.Dialog.Update(0)
		s.runtime.Dialog.Clear()
		clear(s.input.justPress)
		s.finishPlayerActionDialogue()
		s.finishEnemyDialogue()
		s.updateAttack(3 * time.Second)
		s.maybeShowCommandDialogue()
		if !s.runtime.Dialog.Active() || s.runtime.Dialog.RevealedAll() || s.commandDialogueTurn != turn+1 {
			t.Fatalf("player turn %d must restart narrative typing", turn+1)
		}
	}
}
