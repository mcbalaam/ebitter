package mainscene

import (
	"math"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	democomponents "github.com/mcbalaam/ebitter/cmd/demo/components"
	"github.com/mcbalaam/ebitter/render"
	"github.com/mcbalaam/ebitter/scene"
)

func newArenaTestScene() *demoScene {
	runtime := scene.NewRuntime()
	return &demoScene{
		runtime:        runtime,
		battle:         &democomponents.BattleController{Phase: democomponents.BattlePhaseEnemyTurn},
		soul:           &render.AnimatedIcon{},
		soulController: &democomponents.SoulController{Speed: 180},
		arena:          NewBattleBox(32, 250, 576, 140, runtime.Queues.Update),
		input:          newDemoInput(),
		soulX:          72, soulY: 453,
	}
}

func TestEnemyTurnReturnsSoulToSquareArena(t *testing.T) {
	s := newArenaTestScene()
	s.updateSoul(0)
	if !s.soulFreeMove || !s.soulVisible() || s.soulX != 320 || s.soulY != 320 {
		t.Fatalf("enemy turn entry: free=%v, visible=%v, position=(%v,%v)", s.soulFreeMove, s.soulVisible(), s.soulX, s.soulY)
	}
	// Move against a wall while it is still shrinking.
	s.input.pressed[ebiten.KeyArrowLeft] = true
	for range 4 {
		s.updateSoul(time.Second)
		s.runtime.Queues.ExecuteUpdate(arenaResizeDuration / 4)
		s.clampSoulToArena()
		if s.soulX-8 < s.arena.corners[0].X+float64(s.arena.stroke) {
			t.Fatal("shrinking arena must keep the entire soul inside its border")
		}
	}
	width := s.arena.corners[2].X - s.arena.corners[0].X
	height := s.arena.corners[2].Y - s.arena.corners[0].Y
	if width != 140 || height != width {
		t.Fatalf("enemy arena = %vx%v, want 140x140", width, height)
	}
}

func TestFreeSoulMovesImmediatelyStopsAndStaysInside(t *testing.T) {
	s := newArenaTestScene()
	s.updateSoul(0)
	s.runtime.Queues.ExecuteUpdate(arenaResizeDuration)
	s.input.pressed[ebiten.KeyArrowRight] = true
	s.updateSoul(100 * time.Millisecond)
	if s.soulX != 338 || s.soulY != 320 {
		t.Fatalf("held Right must immediately move at configured speed: (%v,%v)", s.soulX, s.soulY)
	}
	clear(s.input.pressed)
	s.updateSoul(time.Second)
	if s.soulX != 338 {
		t.Fatal("releasing a key must stop movement without inertia")
	}
	s.input.pressed[ebiten.KeyArrowLeft] = true
	s.input.pressed[ebiten.KeyArrowUp] = true
	x, y := s.soulX, s.soulY
	s.updateSoul(100 * time.Millisecond)
	if math.Abs(math.Hypot(s.soulX-x, s.soulY-y)-18) > 1e-9 {
		t.Fatal("diagonal movement must not be faster than axial movement")
	}
	s.updateSoul(10 * time.Second)
	if s.soulX != 263 || s.soulY != 263 {
		t.Fatalf("soul must stop at inner walls: (%v,%v), want (263,263)", s.soulX, s.soulY)
	}
}

func TestAttackPreservesSoulPositionAndCommandSelectRestoresBox(t *testing.T) {
	s := newArenaTestScene()
	original := s.arena.corners
	s.updateSoul(0)
	s.runtime.Queues.ExecuteUpdate(arenaResizeDuration)
	s.soulX, s.soulY = 330, 310
	s.battle.SetPhase(democomponents.BattlePhaseAttack)
	s.updateSoul(0)
	if !s.soulFreeMove || !s.soulVisible() || s.soulX != 330 || s.soulY != 310 {
		t.Fatal("attack must preserve free movement and the enemy-turn soul position")
	}
	s.battle.SetPhase(democomponents.BattlePhaseCommandSelect)
	s.syncSoulMode()
	s.runtime.Queues.ExecuteUpdate(arenaResizeDuration)
	if s.soulFreeMove || s.arena.corners != original {
		t.Fatal("command selection must leave free movement and restore the menu box")
	}
}
