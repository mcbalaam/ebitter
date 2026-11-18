package mainscene

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	democomponents "github.com/mcbalaam/ebitter/cmd/demo/components"
	"github.com/mcbalaam/ebitter/tween"
)

const arenaResizeDuration = 250 * time.Millisecond

func (s *demoScene) syncSoulMode() {
	freeMove := s.battle != nil && (s.battle.Phase == democomponents.BattlePhaseEnemyTurn ||
		s.battle.Phase == democomponents.BattlePhaseAttack)
	if freeMove == s.soulFreeMove {
		return
	}
	s.soulFreeMove = freeMove
	if freeMove {
		s.commandArenaCorners = s.arena.corners
		left, top := s.arena.corners[0].X, s.arena.corners[0].Y
		right, bottom := s.arena.corners[2].X, s.arena.corners[2].Y
		s.soulX, s.soulY = (left+right)/2, (top+bottom)/2
		halfSide := math.Min(right-left, bottom-top) / 2
		s.arena.Resize(
			tween.Vec2{X: s.soulX - halfSide, Y: s.soulY - halfSide},
			tween.Vec2{X: s.soulX + halfSide, Y: s.soulY - halfSide},
			tween.Vec2{X: s.soulX + halfSide, Y: s.soulY + halfSide},
			tween.Vec2{X: s.soulX - halfSide, Y: s.soulY + halfSide},
			arenaResizeDuration,
		)
		s.menuActive = false
		s.closeMenu()
	} else {
		corners := s.commandArenaCorners
		s.arena.Resize(corners[0], corners[1], corners[2], corners[3], arenaResizeDuration)
	}
	s.syncButtonSelection()
}

func (s *demoScene) updateFreeSoul(dt time.Duration) {
	dx, dy := 0.0, 0.0
	if s.input.pressed[ebiten.KeyArrowLeft] {
		dx--
	}
	if s.input.pressed[ebiten.KeyArrowRight] {
		dx++
	}
	if s.input.pressed[ebiten.KeyArrowUp] {
		dy--
	}
	if s.input.pressed[ebiten.KeyArrowDown] {
		dy++
	}
	distance := s.soulController.Speed * dt.Seconds()
	if dx != 0 && dy != 0 {
		distance /= math.Sqrt2
	}
	s.soulX += dx * distance
	s.soulY += dy * distance
	s.clampSoulToArena()
}

func (s *demoScene) clampSoulToArena() {
	// Soul coordinates are its center; keep the whole 16x16 sprite inside the stroke.
	margin := float64(s.arena.stroke) + 8
	s.soulX = min(max(s.soulX, s.arena.corners[0].X+margin), s.arena.corners[2].X-margin)
	s.soulY = min(max(s.soulY, s.arena.corners[0].Y+margin), s.arena.corners[2].Y-margin)
}
