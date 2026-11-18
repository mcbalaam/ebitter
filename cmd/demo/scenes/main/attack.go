package mainscene

import (
	"fmt"
	"time"

	democomponents "github.com/mcbalaam/ebitter/cmd/demo/components"
	"github.com/mcbalaam/ebitter/components"
)

func (s *demoScene) initEnemyAttacks() error {
	for _, enemy := range s.enemies {
		durations := make([]time.Duration, 0, len(enemy.Attacks))
		for _, id := range enemy.Attacks {
			blueprint, err := s.prototypes.Load(id)
			if err != nil {
				return fmt.Errorf("load enemy attack %q: %w", id, err)
			}
			entity, err := blueprint.Instantiate()
			if err != nil {
				return fmt.Errorf("instantiate enemy attack %q: %w", id, err)
			}
			pattern, ok := components.ComponentOf[*democomponents.AttackPattern](entity, "attackPattern")
			if !ok {
				return fmt.Errorf("enemy attack %q has no attackPattern", id)
			}
			durations = append(durations, time.Duration(pattern.DurationMS)*time.Millisecond)
		}
		s.enemyAttackDurations = append(s.enemyAttackDurations, durations)
	}
	return nil
}

func (s *demoScene) startEnemyAttack() {
	s.attackRemaining = 0
	turn := max(s.battle.Turn-1, 0)
	for _, durations := range s.enemyAttackDurations {
		if len(durations) > 0 {
			s.attackRemaining = max(s.attackRemaining, durations[turn%len(durations)])
		}
	}
	s.battle.SetPhase(democomponents.BattlePhaseAttack)
}

func (s *demoScene) updateAttack(dt time.Duration) {
	if s.battle == nil || s.battle.Phase != democomponents.BattlePhaseAttack {
		return
	}
	s.attackRemaining -= dt
	if s.attackRemaining > 0 {
		return
	}
	s.attackRemaining = 0
	s.battle.SetPhase(democomponents.BattlePhaseCommandSelect)
}
