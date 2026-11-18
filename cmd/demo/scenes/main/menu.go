package mainscene

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	democomponents "github.com/mcbalaam/ebitter/cmd/demo/components"
	"github.com/mcbalaam/ebitter/text"
)

type menuEntry struct {
	option  menuOption
	text    *text.TextString
	anchorX float64
	anchorY float64
}

func (s *demoScene) openMenu() {
	if s.selectedButton < 0 || s.selectedButton >= len(s.buttonNames) {
		return
	}
	if s.battle != nil && s.battle.Phase == democomponents.BattlePhaseCommandSelect {
		s.runtime.Dialog.ShowAll()
		s.runtime.Dialog.SetVisible(false)
	}
	options := s.menuOptions[s.buttonNames[s.selectedButton]]
	s.closeMenu()
	s.menuIndex = 0
	style := text.TextStyle{
		FontName: "determination", ScaleX: 0.25, ScaleY: 0.25,
		FontHeight: 130, Color: color.White,
		CharSpacing: 4,
	}
	if len(options) == 0 {
		options = []menuOption{{label: "(No items)"}}
	}
	for i, option := range options {
		y := 287.0 + float64(i)*32
		entry := &menuEntry{
			option:  option,
			anchorX: 72, anchorY: y,
			text: text.NewTextString("* "+option.label, 90, y-27, style),
		}
		entry.text.Queues = s.runtime.Queues
		s.menuEntries = append(s.menuEntries, entry)
	}
	s.menuActive = len(s.menuEntries) > 0
	if s.menuActive {
		s.moveSoulToMenuEntry()
	}
}

func (s *demoScene) closeMenu() {
	for _, entry := range s.menuEntries {
		entry.text.Destroy()
	}
	s.menuEntries = nil
}

func (s *demoScene) moveSoulToMenuEntry() {
	if s.menuIndex < 0 || s.menuIndex >= len(s.menuEntries) {
		return
	}
	entry := s.menuEntries[s.menuIndex]
	s.soulX, s.soulY = entry.anchorX, entry.anchorY
}

func (s *demoScene) updateMenu() {
	if s.input.justPressed(ebiten.KeyArrowUp) || s.input.justPressed(ebiten.KeyArrowLeft) {
		previousIndex := s.menuIndex
		s.menuIndex--
		if s.menuIndex < 0 {
			s.menuIndex = len(s.menuEntries) - 1
		}
		if s.menuIndex != previousIndex {
			s.playSFX(sndSqueak)
		}
		s.moveSoulToMenuEntry()
	}
	if s.input.justPressed(ebiten.KeyArrowDown) || s.input.justPressed(ebiten.KeyArrowRight) {
		previousIndex := s.menuIndex
		s.menuIndex++
		if s.menuIndex >= len(s.menuEntries) {
			s.menuIndex = 0
		}
		if s.menuIndex != previousIndex {
			s.playSFX(sndSqueak)
		}
		s.moveSoulToMenuEntry()
	}
	if s.input.justPressed(ebiten.KeyZ) {
		s.confirmMenuChoice()
	}
	if s.input.justPressed(ebiten.KeyX) {
		s.menuActive = false
		s.menuConfirmed = false
		s.playSFX(sndSqueak)
		s.closeMenu()
		s.syncButtonSelection()
	}
}

func (s *demoScene) drawMenu(screen *ebiten.Image) {
	if !s.menuActive {
		return
	}
	for _, entry := range s.menuEntries {
		entry.text.Draw(screen)
	}
}
