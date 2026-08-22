package main

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
	// "sign_floor": {
	// 	sound: sndText,
	// 	phrases: []string{
	// 		"* (An old, worn sign lies on the floor.)",
	// 		"* (Most of the letters have faded away...)",
	// 		"* (...but you can still make out$n  the word $cffd83a\"RUINS\"$cffffff.)",
	// 	},
	// },
	// "froggit_left": {
	// 	sound: sndText,
	// 	phrases: []string{
	// 		"* Ribbit, ribbit.",
	// 		"* (The froggit stares at you,$n  visibly judging your outfit.)",
	// 		"* (It seems unimpressed.)",
	// 		"* (You feel $cff3a3aDETERMINATION$cffffff anyway.)",
	// 	},
	// },
	// "froggit_right": {
	// 	sound: sndText,
	// 	phrases: []string{
	// 		"* Ribbit! Ribbit!",
	// 		"* (The froggit hops in place,$n  completely unbothered.)",
	// 		"* (What a happy little guy.)",
	// 	},
	// },
	"sign_wall": {
		sound: sndText,
		phrases: []string{
			"* ( An old sign. )",
			"* ( No matter how hard you tried,$p200$n  you couldn't make out any words. )",
		},
	},
}

const sndText = "snd_text"

func initDialogAudio() error {
	player, err := sound.NewSoundPlayer(44100)
	if err != nil {
		return err
	}
	if err := player.RegisterNewSound("media/sound/snd_text.wav", sndText); err != nil {
		return err
	}
	text.DefaultDialog.SoundPlayer = player
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
