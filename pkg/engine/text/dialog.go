package text

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mcbalaam/ebitter/pkg/sound"
)

type DialogManager struct {
	active         *TextDisplay
	queue          []dialogRequest
	ShowAllChecker func() bool
	AdvanceChecker func() bool
	SoundPlayer    *sound.SoundPlayer
}

type dialogRequest struct {
	style     TextStyle
	text      string
	charSound string
}

func defaultShowAll() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyX)
}

func defaultAdvance() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyZ) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace)
}

var DefaultDialog = &DialogManager{
	ShowAllChecker: defaultShowAll,
	AdvanceChecker: defaultAdvance,
}

func (dm *DialogManager) Show(text string, style TextStyle, charSound string) {
	if dm.active == nil {
		dm.start(text, style, charSound)
	} else {
		dm.queue = append(dm.queue, dialogRequest{style, text, charSound})
	}
}

func (dm *DialogManager) Active() bool {
	return dm.active != nil
}

func (dm *DialogManager) RevealedAll() bool {
	return dm.active != nil && dm.active.RevealedAll()
}

func (dm *DialogManager) Waiting() bool {
	return dm.active != nil && dm.active.Waiting()
}

func (dm *DialogManager) Update(dt time.Duration) {
	if dm.active == nil {
		dm.next()
		return
	}

	if dm.ShowAllChecker != nil && dm.ShowAllChecker() {
		if !dm.active.RevealedAll() {
			dm.active.ShowAll()
			return
		}
	}

	if dm.AdvanceChecker != nil && dm.AdvanceChecker() {
		if dm.active.RevealedAll() {
			if dm.active.Advance() {
				dm.active = nil
				dm.next()
				return
			}
		}
	}

	dm.active.Update(dt)

	if dm.active != nil && dm.active.RevealedAll() && dm.AdvanceChecker != nil && dm.AdvanceChecker() {
		if dm.active.Advance() {
			dm.active = nil
			dm.next()
		}
	}
}

func (dm *DialogManager) start(text string, style TextStyle, charSound string) {
	dm.active = NewTextDisplay(text, style)
	dm.active.CharSound = charSound
}

func (dm *DialogManager) next() {
	if len(dm.queue) == 0 {
		return
	}
	req := dm.queue[0]
	dm.queue = dm.queue[1:]
	dm.start(req.text, req.style, req.charSound)
}
