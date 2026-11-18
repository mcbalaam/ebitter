package mainscene

import (
	"os"
	"testing"
	"time"

	democomponents "github.com/mcbalaam/ebitter/cmd/demo/components"
	"github.com/mcbalaam/ebitter/components"
)

func TestDummyAttackWaitsThreeSecondsThenReturnsPlayerControl(t *testing.T) {
	s := newArenaTestScene()
	s.prototypes = components.NewBlueprintLibrary(os.DirFS("../../prototypes"), ".")
	s.enemies = []*democomponents.EnemyController{{Attacks: []string{"trainingDummy_slowShot"}}}
	s.battle.Turn = 1
	if err := s.initEnemyAttacks(); err != nil {
		t.Fatal(err)
	}
	s.buttonNames = []string{"fightButton"}
	button := components.NewEntity("fightButton")
	button.Add(&components.Transform{X: 32, Y: 431})
	s.buttons = map[string]*components.Entity{"fightButton": button}
	original := s.arena.corners
	s.finishEnemyDialogue()
	s.updateSoul(0)
	s.runtime.Queues.ExecuteUpdate(arenaResizeDuration)

	s.updateAttack(2999 * time.Millisecond)
	if s.battle.Phase != democomponents.BattlePhaseAttack || !s.soulFreeMove {
		t.Fatal("attack must keep free movement until all three seconds elapse")
	}
	s.updateAttack(time.Millisecond)
	s.updateSoul(0)
	s.runtime.Queues.ExecuteUpdate(arenaResizeDuration)
	if s.battle.Phase != democomponents.BattlePhaseCommandSelect || s.soulFreeMove {
		t.Fatal("at three seconds, control must return to the command buttons")
	}
	if s.arena.corners != original || s.soulX != 48 || s.soulY != 453 || !s.buttonActive(0) {
		t.Fatalf("player UI not restored: arena=%v, soul=(%v,%v), button active=%v", s.arena.corners, s.soulX, s.soulY, s.buttonActive(0))
	}
	if s.battle.Turn != 1 {
		t.Fatal("attack completion must not increment the turn a second time")
	}

	// Every attack gets its own full duration, regardless of the previous timer.
	s.finishEnemyDialogue()
	s.updateAttack(2 * time.Second)
	if s.battle.Phase != democomponents.BattlePhaseAttack {
		t.Fatal("a later attack must restart its three-second wait")
	}
	s.updateAttack(2 * time.Second)
	if s.battle.Phase != democomponents.BattlePhaseCommandSelect {
		t.Fatal("a large frame delta must also finish the wait")
	}
}
