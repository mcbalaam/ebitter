package mainscene

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	democomponents "github.com/mcbalaam/ebitter/cmd/demo/components"
)

func (s *demoScene) updateDialogue(dt time.Duration) {
	s.maybeShowCommandDialogue()
	if s.menuActive && s.battle != nil && s.battle.Phase == democomponents.BattlePhaseCommandSelect {
		return
	}
	if s.battle != nil && s.battle.Phase == democomponents.BattlePhaseEnemyTurn && !s.enemyDialoguePending {
		s.startEnemyDialogue()
		return
	}
	if !s.runtime.Dialog.Active() {
		return
	}
	s.runtime.Dialog.Update(dt)
	if !s.runtime.Dialog.Active() {
		if s.actionDialoguePending {
			s.finishPlayerActionDialogue()
		} else if s.enemyDialoguePending {
			s.finishEnemyDialogue()
		}
	}
}

func (s *demoScene) dialogueXPressed() bool {
	return !s.menuActive && s.input.justPressed(ebiten.KeyX)
}

func (s *demoScene) dialogueAdvancePressed() bool {
	return !s.menuActive &&
		s.battle != nil &&
		(s.battle.Phase == democomponents.BattlePhasePlayerAction ||
			s.battle.Phase == democomponents.BattlePhaseEnemyTurn) &&
		s.input.justPressed(ebiten.KeyZ)
}

func commandLineForTurn(lines []string, turn int) (string, bool) {
	if turn < 0 || len(lines) == 0 {
		return "", false
	}
	return lines[turn%len(lines)], true
}

func (s *demoScene) maybeShowCommandDialogue() {
	if s.battle == nil || s.battle.Phase != democomponents.BattlePhaseCommandSelect ||
		s.commandDialogueTurn == s.battle.Turn || s.runtime.Dialog.Active() {
		return
	}
	s.commandDialogueTurn = s.battle.Turn
	s.runtime.Dialog.SetVisible(true)
	line, ok := commandLineForTurn(s.battle.CommandLines, s.battle.Turn)
	if ok {
		s.runtime.Dialog.Show(line, battleDialogStyle(), sndText)
	}
}

func (s *demoScene) confirmMenuChoice() {
	if s.menuIndex < 0 || s.menuIndex >= len(s.menuEntries) || s.battle == nil {
		return
	}
	selected := s.menuEntries[s.menuIndex].option
	if selected.prototype == "" {
		return
	}
	if !s.consumeMenuItem(selected) {
		return
	}
	s.playSFX(sndSelect)
	s.selectedMenuOption = selected
	s.menuConfirmed = true
	s.menuActive = false
	s.closeMenu()
	s.battle.SetPhase(democomponents.BattlePhasePlayerAction)
	s.syncButtonSelection()
	s.runtime.Dialog.Clear()
	s.actionDialoguePending = true
	dialogue := s.selectedMenuOption.dialogue
	if len(dialogue) == 0 {
		dialogue = []string{"* " + s.selectedMenuOption.label + "."}
	}
	for _, line := range dialogue {
		s.runtime.Dialog.Show(line, battleDialogStyle(), sndText)
	}
	if !s.runtime.Dialog.Active() {
		s.finishPlayerActionDialogue()
	}
}

func (s *demoScene) finishPlayerActionDialogue() {
	if s.battle != nil {
		s.battle.Turn++
		s.battle.SetPhase(democomponents.BattlePhaseEnemyTurn)
	}
	s.actionDialoguePending = false
}

func (s *demoScene) playSFX(name string) {
	if player := s.runtime.Dialog.SoundPlayer; player != nil {
		_ = player.PlaySound(name, 0.5)
	}
}
