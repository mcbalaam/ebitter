package scenes

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/mcbalaam/ebitter/sound"
	"github.com/mcbalaam/ebitter/text"
)

type demoDialog struct {
	sound   string
	phrases []string
}

var demoDialogs = map[string]demoDialog{
	"sign_wall": {
		sound: sndText,
		phrases: []string{
			"* ( An old sign. )",
			"* ( No matter how hard you tried,$p300$n  you couldn't make out any words. )",
		},
	},
}

const sndText = "snd_text"

func initDialogAudio(dialog *text.DialogManager) error {
	player, err := sound.NewSoundPlayer(44100)
	if err != nil {
		return err
	}
	if err := player.RegisterNewSound("media/sound/snd_text.ogg", sndText); err != nil {
		return err
	}
	dialog.SoundPlayer = player
	dialog.CharSoundVolume = 0.5
	dialog.CharPitchMin = 1.0
	dialog.CharPitchMax = 1.0
	return nil
}

const (
	textboxX      = 35
	textboxY      = 325
	textboxW      = 570
	textboxH      = 140
	textboxBorder = 5
	textboxPad    = 25
)

var dialogStyle = text.TextStyle{
	FontName:     "determination",
	StartX:       textboxX + textboxPad,
	StartY:       textboxY + textboxPad - 15,
	ScaleX:       0.25,
	ScaleY:       0.25,
	FontHeight:   130,
	LineSpacing:  10,
	DefaultDelay: 0.03,
	Color:        color.White,
	CharWidths:   map[string]int{" ": 60},
}

func drawTextbox(screen *ebiten.Image) {
	vector.FillRect(screen,
		textboxX, textboxY, textboxW, textboxH,
		color.Black, false)
	vector.StrokeRect(screen,
		textboxX, textboxY, textboxW, textboxH,
		textboxBorder, color.White, false)
}
