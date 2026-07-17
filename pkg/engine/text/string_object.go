package text

import (
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/pkg/engine/queues"
	"github.com/mcbalaam/ebitter/pkg/render"
)

var textStringFontCache sync.Map

type TextString struct {
	X, Y       float64
	Rotation   float64
	Alpha      float64
	Text       string
	Style      TextStyle
	Layer      int
	Visible    bool
	Centered   bool
	UpdateFunc func(ts *TextString, dt time.Duration)

	font        *render.AnimatedIcon
	image       *ebiten.Image
	totalWidth  float64
	totalHeight float64
	loaded      bool
	destroyed   bool
	mu          sync.Mutex
}

func NewTextString(text string, x, y float64, style TextStyle) *TextString {
	return &TextString{
		Text:    text,
		X:       x,
		Y:       y,
		Style:   style,
		Alpha:   1,
		Layer:   queues.LayerText,
		Visible: true,
	}
}

func (ts *TextString) Show() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.destroyed {
		return
	}
	ts.Visible = true
	queues.DefaultQueue.ScheduleAt(ts, ts.Layer)
	queues.DefaultUpdateQueue.Schedule(ts)
}

func (ts *TextString) Hide() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.Visible = false
}

func (ts *TextString) Destroy() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.destroyed = true
	ts.Visible = false
	ts.font = nil
}

func (ts *TextString) SetText(text string) {
	ts.mu.Lock()
	ts.Text = text
	ts.loaded = false
	ts.mu.Unlock()
}

func (ts *TextString) SetPosition(x, y float64) {
	ts.mu.Lock()
	ts.X = x
	ts.Y = y
	ts.mu.Unlock()
}

func (ts *TextString) ensureFont() {
	if ts.loaded {
		return
	}

	fontName := ts.Style.FontName
	if cached, ok := textStringFontCache.Load(fontName); ok {
		ts.font = cached.(*render.AnimatedIcon)
	} else {
		font, err := render.NewAnimatedIconFromPath("media/sprites/"+fontName, " ")
		if err != nil {
			return
		}
		textStringFontCache.Store(fontName, font)
		ts.font = font
	}

	charSpacing := ts.Style.CharSpacing
	lineH := ts.Style.FontHeight + ts.Style.LineSpacing

	defaultCharWidth := ts.font.CurrentState.CurrentFrameRef.Image.Bounds().Dx()

	lines := strings.Split(ts.Text, "\n")
	cx := 0.0
	cy := 0.0
	maxLineW := 0.0
	maxH := 0

	type placedGlyph struct {
		img  *ebiten.Image
		x, y float64
	}
	var placed []placedGlyph

	for li, line := range lines {
		if li > 0 {
			cy += lineH
			cx = 0
		}
		for _, r := range line {
			char := string(r)
			ts.font.SetIconState(char)
			frame := ts.font.CurrentState.CurrentFrameRef
			if frame == nil {
				cx += float64(defaultCharWidth) + charSpacing
				continue
			}
			w := float64(frame.Image.Bounds().Dx())
			h := frame.Image.Bounds().Dy()
			if h > maxH {
				maxH = h
			}

			placed = append(placed, placedGlyph{
				img: frame.Image,
				x:   cx,
				y:   cy,
			})

			cx += w + charSpacing
		}
		if cx > maxLineW {
			maxLineW = cx
		}
	}

	ts.totalWidth = maxLineW
	ts.totalHeight = cy + float64(maxH)

	if ts.image == nil || ts.image.Bounds().Dx() < int(maxLineW) || ts.image.Bounds().Dy() < int(ts.totalHeight) {
		ts.image = ebiten.NewImage(int(maxLineW), int(ts.totalHeight))
	} else {
		ts.image.Clear()
	}

	col := ts.Style.Color
	for _, p := range placed {
		op := &ebiten.DrawImageOptions{}
		if col != nil {
			op.ColorScale.ScaleWithColor(col)
		}
		op.GeoM.Translate(p.x, p.y)
		ts.image.DrawImage(p.img, op)
	}

	ts.loaded = true
}

func (ts *TextString) Update(deltaTime time.Duration) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.destroyed || !ts.Visible {
		return
	}
	if ts.UpdateFunc != nil {
		ts.UpdateFunc(ts, deltaTime)
	}
}

func (ts *TextString) Draw(screen *ebiten.Image) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.destroyed || !ts.Visible {
		return
	}
	ts.ensureFont()

	if ts.image == nil {
		return
	}

	sx := ts.Style.ScaleX
	sy := ts.Style.ScaleY

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest

	w := float64(ts.image.Bounds().Dx())
	h := float64(ts.image.Bounds().Dy())

	if ts.Centered {
		op.GeoM.Translate(-w/2, -h/2)
	}
	op.GeoM.Scale(sx, sy)
	op.GeoM.Rotate(ts.Rotation)
	op.GeoM.Translate(ts.X, ts.Y)

	if ts.Alpha < 0.999 {
		op.ColorScale.ScaleAlpha(float32(ts.Alpha))
	}
	screen.DrawImage(ts.image, op)
}
