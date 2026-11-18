package mainscene

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	democomponents "github.com/mcbalaam/ebitter/cmd/demo/components"
	"github.com/mcbalaam/ebitter/components"
	"github.com/mcbalaam/ebitter/render"
	"github.com/mcbalaam/ebitter/text"
)

const (
	enemyBubbleX = 350
	enemyBubbleY = 100
)

func (s *demoScene) initEnemyDialogue() error {
	bubble, err := render.NewAnimatedIconFromPath("media/sprites/bubble", "idle")
	if err != nil {
		return fmt.Errorf("load enemy speech bubble: %w", err)
	}
	s.enemyBubble = bubble
	for _, id := range s.battle.Enemies {
		blueprint, err := s.prototypes.Load(id)
		if err != nil {
			return fmt.Errorf("load speaking enemy %q: %w", id, err)
		}
		entity, err := blueprint.InstantiateWith(s.runtime.Signals, s.runtime.Queues)
		if err != nil {
			return fmt.Errorf("instantiate enemy %q: %w", id, err)
		}
		controller, ok := components.ComponentOf[*democomponents.EnemyController](entity, "enemyController")
		if !ok {
			return fmt.Errorf("enemy prototype %q has no enemyController", id)
		}
		s.enemies = append(s.enemies, controller)
		entity.Schedule()
	}
	return nil
}

func enemyDialogStyle() text.TextStyle {
	return text.TextStyle{
		FontName: "greater-determination-sb",
		StartX:   enemyBubbleX + 40, StartY: enemyBubbleY + 4,
		ScaleX: 0.15, ScaleY: 0.15,
		FontHeight: 29, LineSpacing: 95,
		CharSpacing:  3,
		DefaultDelay: 0.03, Color: color.Black,
	}
}

// Each enemy speaks one configured line per turn, cycling when exhausted.
func enemyLineForTurn(lines []string, turn int) (string, bool) {
	if len(lines) == 0 {
		return "", false
	}
	index := turn - 1 // Turn advances after the player's action dialogue.
	if index < 0 {
		index = 0
	}
	return lines[index%len(lines)], true
}

func (s *demoScene) startEnemyDialogue() {
	s.runtime.Dialog.Clear()
	for _, enemy := range s.enemies {
		if line, ok := enemyLineForTurn(enemy.Lines, s.battle.Turn); ok {
			s.runtime.Dialog.Show(line, enemyDialogStyle(), sndEnemyText)
		}
	}
	s.enemyDialoguePending = s.runtime.Dialog.Active()
	if !s.enemyDialoguePending {
		s.finishEnemyDialogue()
	}
}

func (s *demoScene) finishEnemyDialogue() {
	s.enemyDialoguePending = false
	s.startEnemyAttack()
}

func (s *demoScene) drawEnemyBubble(screen *ebiten.Image) {
	if s.enemyDialoguePending && s.enemyBubble != nil {
		s.enemyBubble.Draw(screen, enemyBubbleX, enemyBubbleY, 1, 1, 0)
	}
}
