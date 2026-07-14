package engine

import (
	"image/color"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mcbalaam/ebitter/pkg/engine/queues"
	"github.com/mcbalaam/ebitter/pkg/render"
	"github.com/mcbalaam/ebitter/pkg/sound"
)

type TextEngine struct {
	FontsLoaded map[string]render.AnimatedIcon
}

type Glyph struct {
	Image  *ebiten.Image
	PosX   float64
	PosY   float64
	ScaleX float64
	ScaleY float64
	Tilt   float64
	Color  color.Color
}

func (g *Glyph) Draw(s *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest

	op.GeoM.Scale(g.ScaleX, g.ScaleY)
	op.GeoM.Rotate(g.Tilt)

	op.GeoM.Translate(math.Round(g.PosX), math.Round(g.PosY))

	op.ColorScale.ScaleWithColor(g.Color)
	s.DrawImage(g.Image, op)
}

type TextDisplay struct {
	Font        render.AnimatedIcon
	ScaleX      float64
	ScaleY      float64
	Tilt        float64
	ElapsedTime float64
	IsComplete  bool
	SoundPlayer *sound.SoundPlayer
	Instant     bool

	Commands    []DialogueCommand
	CmdIndex    int
	Displayed   []*Glyph
	OnComplete  func()
	WaitingForZ bool
}

func (t *TextDisplay) Update(deltaTime time.Duration) {
	if t.IsComplete {
		return
	}

	if t.Instant {
		for t.CmdIndex < len(t.Commands) {
			cmd := t.Commands[t.CmdIndex]
			if cmd.Type == CmdEnd {
				t.WaitingForZ = true
				t.CmdIndex++
				return
			}
			if cmd.Type == CmdEndNoWait {
				t.IsComplete = true
				t.CmdIndex++
				if t.OnComplete != nil {
					t.OnComplete()
				}
				return
			}
			if cmd.Type == CmdChar {
				t.Font.SetIconState(cmd.Char)
				glyph := &Glyph{
					Image:  t.Font.CurrentState.CurrentFrameRef.Image,
					PosX:   cmd.X,
					PosY:   cmd.Y,
					ScaleX: t.ScaleX,
					ScaleY: t.ScaleY,
					Tilt:   t.Tilt,
					Color:  cmd.Color,
				}
				t.Displayed = append(t.Displayed, glyph)
			}
			t.CmdIndex++
		}
		t.IsComplete = true
		if t.OnComplete != nil {
			t.OnComplete()
		}
		return
	}

	if t.WaitingForZ {
		if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
			t.WaitingForZ = false
			t.IsComplete = true
			if t.OnComplete != nil {
				t.OnComplete()
			}
		}
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		for i := t.CmdIndex; i < len(t.Commands); i++ {
			if t.Commands[i].Type == CmdChar {
				t.Commands[i].TriggerAt = 0
			}
		}
		t.ElapsedTime = 100000.0
	} else {
		t.ElapsedTime += deltaTime.Seconds()
	}

	for t.CmdIndex < len(t.Commands) {
		cmd := t.Commands[t.CmdIndex]
		if cmd.TriggerAt > t.ElapsedTime {
			break
		}

		if cmd.Type == CmdEnd {
			t.WaitingForZ = true
			t.CmdIndex++
			return
		}

		if cmd.Type == CmdEndNoWait {
			t.IsComplete = true
			t.CmdIndex++
			if t.OnComplete != nil {
				t.OnComplete()
			}
			return
		}

		if cmd.Type == CmdChar {
			t.Font.SetIconState(cmd.Char)

			glyph := &Glyph{
				Image:  t.Font.CurrentState.CurrentFrameRef.Image,
				PosX:   cmd.X,
				PosY:   cmd.Y,
				ScaleX: t.ScaleX,
				ScaleY: t.ScaleY,
				Tilt:   t.Tilt,
				Color:  cmd.Color,
			}

			t.Displayed = append(t.Displayed, glyph)

			if t.SoundPlayer != nil && !inpututil.IsKeyJustPressed(ebiten.KeyX) {
				if err := t.SoundPlayer.PlayVariable("snd_text", 2, 0); err != nil {
					log.Printf("Error playing sound: %v", err)
				}
			}
		}

		t.CmdIndex++
	}

	if t.CmdIndex >= len(t.Commands) && !t.WaitingForZ {
		t.IsComplete = true
		if t.OnComplete != nil {
			t.OnComplete()
		}
	}
}

func (t *TextDisplay) Draw(s *ebiten.Image) {
	for _, glyph := range t.Displayed {
		glyph.Draw(s)
	}
}

func (t *TextDisplay) ForceComplete() {
	if t.IsComplete {
		return
	}
	t.WaitingForZ = false
	t.IsComplete = true
}

func (t *TextDisplay) Destroy() {
	t.Displayed = nil
	t.Commands = nil
}

func (te *TextEngine) DisplayText(style TextStyle, text string,
	soundPlayer *sound.SoundPlayer, onComplete func()) (*TextDisplay, error) {

	if _, exists := te.FontsLoaded[style.FontName]; !exists {
		icon, err := render.NewAnimatedIconFromPath("media/sprites/"+style.FontName, " ")
		if err != nil {
			return nil, err
		}
		te.FontsLoaded[style.FontName] = *icon
	}

	font := te.FontsLoaded[style.FontName]

	textDisplay := &TextDisplay{
		Font:        font,
		ScaleX:      style.ScaleX,
		ScaleY:      style.ScaleY,
		IsComplete:  false,
		SoundPlayer: soundPlayer,
		Displayed:   make([]*Glyph, 0),
		OnComplete:  onComplete,
		Instant:     style.Instant,
	}

	parser := &TextParser{
		Text:         text,
		StartX:       style.StartX,
		StartY:       style.StartY,
		ScaleX:       style.ScaleX,
		ScaleY:       style.ScaleY,
		FontHeight:   style.FontHeight,
		LineSpacing:  style.LineSpacing,
		Delay:        style.DefaultDelay,
		CharWidth:    make(map[string]int),
		CharSpacing:  style.CharSpacing,
		DefaultColor: style.Color,
	}
	textDisplay.Commands = parser.Parse()

	queues.DefaultQueue.ScheduleAt(textDisplay, queues.LayerText)
	queues.DefaultUpdateQueue.Schedule(textDisplay)

	return textDisplay, nil
}
