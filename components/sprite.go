package components

import (
	"path"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/render"
)

// SpriteRoot is the assetfs directory sprite icons are loaded from when a
// blueprint refers to a sprite by name. The application must configure it.
var SpriteRoot string

// Sprite draws the entity's animated icon at the transform position.
type Sprite struct {
	Icon    *render.AnimatedIcon
	OffsetX float64
	OffsetY float64
	// AnchorX/AnchorY shift the drawn image by a fraction of the frame size,
	// e.g. 0.5 centers horizontally and 1.0 bottom-aligns vertically.
	AnchorX float64
	AnchorY float64
	FlipH   bool
	FlipV   bool
}

func (s *Sprite) Type() string { return "sprite" }

func (s *Sprite) Update(e *Entity, dt time.Duration) {
	if s.Icon != nil {
		s.Icon.Update(dt)
	}
}

func (s *Sprite) Draw(e *Entity, screen *ebiten.Image) {
	if s.Icon == nil {
		return
	}
	t := e.Transform()
	if t == nil {
		return
	}

	sx, sy := t.ScaleX, t.ScaleY
	if sx == 0 {
		sx = 1
	}
	if sy == 0 {
		sy = 1
	}

	x := t.X + s.OffsetX
	y := t.Y + s.OffsetY

	if (s.AnchorX != 0 || s.AnchorY != 0) && s.Icon.CurrentState != nil {
		if f := s.Icon.CurrentState.CurrentFrameRef; f != nil && f.Image != nil {
			b := f.Image.Bounds()
			x -= s.AnchorX * float64(b.Dx()) * sx
			y -= s.AnchorY * float64(b.Dy()) * sy
		}
	}

	if s.FlipH {
		sx = -sx
	}
	if s.FlipV {
		sy = -sy
	}

	s.Icon.Draw(screen, x, y, sx, sy, t.Rotation)
}

// spritePath resolves a sprite name to its assetfs directory.
func spritePath(name string) string { return path.Join(SpriteRoot, name) }
