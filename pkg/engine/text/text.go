package text

import (
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/pkg/engine/queues"
	"github.com/mcbalaam/ebitter/pkg/render"
)

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
	textStr     *TextString
	Commands    []DialogueCommand
	CmdIndex    int
	Displayed   []*Glyph
	IsComplete  bool
	OnComplete  func()
	CharSound   string
	elapsed     float64
	waiting     bool
	skipSound   bool
	revealedAll bool
}

func NewTextDisplay(text string, style TextStyle) *TextDisplay {
	font := loadFont(style.FontName)
	if font == nil {
		return nil
	}

	td := &TextDisplay{
		textStr:   NewTextString(text, style.StartX, style.StartY, style),
		Displayed: make([]*Glyph, 0),
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
	td.Commands = parser.Parse()

	if style.Instant {
		td.skipSound = true
		td.revealAll()
	}

	queues.DefaultQueue.ScheduleAt(td, queues.LayerText)
	queues.DefaultUpdateQueue.Schedule(td)

	return td
}

func loadFont(fontName string) *render.AnimatedIcon {
	if cached, ok := textStringFontCache.Load(fontName); ok {
		return cached.(*render.AnimatedIcon)
	}
	icon, err := render.NewAnimatedIconFromPath("media/sprites/"+fontName, " ")
	if err != nil {
		return nil
	}
	textStringFontCache.Store(fontName, icon)
	return icon
}

func (t *TextDisplay) revealAll() {
	for _, cmd := range t.Commands {
		if cmd.Type == CmdChar {
			t.revealChar(cmd)
		}
	}
	t.IsComplete = true
}

func (t *TextDisplay) skipAll() {
	t.elapsed = 1e6
	t.Advance()
}

func (t *TextDisplay) revealChar(cmd DialogueCommand) {
	if !t.skipSound && t.CharSound != "" && DefaultDialog.SoundPlayer != nil {
		DefaultDialog.SoundPlayer.PlaySound(t.CharSound, 100)
	}
	font := loadFont(t.textStr.Style.FontName)
	if font == nil {
		return
	}
	font.SetIconState(cmd.Char)
	frame := font.CurrentState.CurrentFrameRef
	if frame == nil {
		return
	}
	glyph := &Glyph{
		Image:  frame.Image,
		PosX:   cmd.X,
		PosY:   cmd.Y,
		ScaleX: t.textStr.Style.ScaleX,
		ScaleY: t.textStr.Style.ScaleY,
		Color:  cmd.Color,
	}
	t.Displayed = append(t.Displayed, glyph)
}

func (t *TextDisplay) ShowAll() {
	if t.IsComplete || t.revealedAll {
		return
	}
	t.skipSound = true
	for t.CmdIndex < len(t.Commands) {
		cmd := t.Commands[t.CmdIndex]
		if cmd.Type == CmdEnd || cmd.Type == CmdEndNoWait {
			t.CmdIndex++
			continue
		}
		if cmd.Type == CmdChar {
			t.revealChar(cmd)
		}
		t.CmdIndex++
	}
	t.skipSound = false
	t.revealedAll = true
	t.waiting = false
}

func (t *TextDisplay) Advance() bool {
	if t.IsComplete {
		return true
	}
	if !t.revealedAll {
		return false
	}
	t.Displayed = t.Displayed[:0]
	t.IsComplete = true
	if t.OnComplete != nil {
		t.OnComplete()
	}
	return true
}

func (t *TextDisplay) Update(deltaTime time.Duration) {
	if t.IsComplete {
		return
	}

	if t.revealedAll || t.waiting {
		return
	}

	t.elapsed += deltaTime.Seconds()

	for t.CmdIndex < len(t.Commands) {
		cmd := t.Commands[t.CmdIndex]
		if cmd.TriggerAt > t.elapsed {
			break
		}

		if cmd.Type == CmdEnd {
			t.CmdIndex++
			if t.CmdIndex >= len(t.Commands) {
				t.revealedAll = true
				t.waiting = false
			} else {
				t.waiting = true
			}
			return
		}

		if cmd.Type == CmdEndNoWait {
			t.CmdIndex++
			t.Displayed = t.Displayed[:0]
			t.IsComplete = true
			if t.OnComplete != nil {
				t.OnComplete()
			}
			return
		}

		if cmd.Type == CmdChar {
			t.revealChar(cmd)
		}

		t.CmdIndex++
	}

	if t.CmdIndex >= len(t.Commands) && !t.revealedAll {
		t.revealedAll = true
		t.waiting = false
	}
}

func (t *TextDisplay) Draw(s *ebiten.Image) {
	for _, glyph := range t.Displayed {
		glyph.Draw(s)
	}
}

func (t *TextDisplay) RevealedAll() bool { return t.revealedAll }
func (t *TextDisplay) Waiting() bool { return t.waiting }

func (t *TextDisplay) Destroy() {
	t.Displayed = nil
	t.Commands = nil
}
