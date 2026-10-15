package text

import (
	"image/color"
	"math"
	"math/rand"
	"path"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/queues"
	"github.com/mcbalaam/ebitter/render"
	"github.com/mcbalaam/ebitter/sound"
)

var (
	// FontRoot is the assetfs directory fonts are loaded from. Each font is a
	// subdirectory in Aseprite format; override for a different layout.
	FontRoot = "media/sprites"
	// Filter is the draw filter used for glyphs and baked text strings.
	Filter = ebiten.FilterNearest
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
	op.Filter = Filter

	op.GeoM.Scale(g.ScaleX, g.ScaleY)
	op.GeoM.Rotate(g.Tilt)
	op.GeoM.Translate(math.Round(g.PosX), math.Round(g.PosY))

	if g.Color != nil {
		op.ColorScale.ScaleWithColor(g.Color)
	}
	s.DrawImage(g.Image, op)
}

type TextDisplay struct {
	textStr    *TextString
	Commands   []DialogueCommand
	CmdIndex   int
	Displayed  []*Glyph
	IsComplete bool
	OnComplete func()
	CharSound  string
	// CharSoundVolume is the playback volume of CharSound. Values <= 0 use 0.5.
	CharSoundVolume float64
	// CharPitchMin/Max bound a random pitch multiplier applied to each
	// character sound. When CharPitchMax <= CharPitchMin (or Min <= 0) the
	// sound plays at its original pitch.
	CharPitchMin float64
	CharPitchMax float64
	SoundPlayer  *sound.SoundPlayer
	elapsed      float64
	waiting      bool
	skipSound    bool
	revealedAll  bool
}

// NewTextDisplay parses text into a timed reveal command list and schedules
// the display on the draw queue. Updates must be driven by the owner —
// DialogManager does this for dialog; standalone users should call Update(dt)
// (or schedule it on queues.DefaultUpdateQueue) themselves.
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
		Font:         font,
		CharWidth:    style.CharWidths,
		CharSpacing:  style.CharSpacing,
		DefaultColor: style.Color,
	}
	td.Commands = parser.Parse()

	if style.Instant {
		td.skipSound = true
		td.revealAll()
	}

	queues.DefaultQueue.ScheduleAt(td, queues.LayerText)

	return td
}

func loadFont(fontName string) *render.AnimatedIcon {
	if cached, ok := textStringFontCache.Load(fontName); ok {
		return cached.(*render.AnimatedIcon)
	}
	icon, err := render.NewAnimatedIconFromPath(path.Join(FontRoot, fontName), " ")
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
	t.revealedAll = true
	t.IsComplete = true
}

func (t *TextDisplay) skipAll() {
	t.elapsed = 1e6
	t.Advance()
}

// playCharSound plays the per-character typing sound, optionally with a random
// pitch in [CharPitchMin, CharPitchMax].
func (t *TextDisplay) playCharSound() {
	volume := t.CharSoundVolume
	if volume <= 0 {
		volume = 0.5
	}
	if t.CharPitchMin > 0 && t.CharPitchMax > t.CharPitchMin {
		pitch := t.CharPitchMin + rand.Float64()*(t.CharPitchMax-t.CharPitchMin)
		_ = t.SoundPlayer.PlayVariable(t.CharSound, volume, pitch)
		return
	}
	_ = t.SoundPlayer.PlaySound(t.CharSound, volume)
}

func (t *TextDisplay) revealChar(cmd DialogueCommand) {
	if !t.skipSound && t.CharSound != "" && t.SoundPlayer != nil {
		t.playCharSound()
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
			t.revealedAll = true
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
func (t *TextDisplay) Waiting() bool     { return t.waiting }

func (t *TextDisplay) Destroy() {
	queues.DefaultQueue.Unschedule(t)
	queues.DefaultUpdateQueue.Unschedule(t)
	t.Displayed = nil
	t.Commands = nil
}
