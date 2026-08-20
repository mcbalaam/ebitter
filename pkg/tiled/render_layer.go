package tiled

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// TileAt describes a single tile placement inside a layer. X, Y, TileW and
// TileH are in (already scaled) world coordinates; Scale is the factor the
// tile image itself is magnified by when drawn.
type TileAt struct {
	Image   *ebiten.Image
	X, Y    float64
	TileW   int
	TileH   int
	Scale   float64
	FlipH   bool
	FlipV   bool
	FlipDiag bool
}

// drawTile draws a single tile from a tileset texture onto target.
//
// The tile image is placed into the grid cell at (tx, ty) (top-left in world
// coordinates), bottom-aligned when the tile is smaller than the cell.
// Flipping and the diagonal (anti-diagonal) flip follow Tiled's cell renderer:
// the diagonal flip rotates the image 90 degrees clockwise and swaps the
// horizontal/vertical flip flags.
func drawTile(target, tileImage *ebiten.Image, tx, ty float64, tw, th int, scale float64, horizontalFlip, verticalFlip, diagonalFlip bool, alpha float32) {
	if scale <= 0 {
		scale = 1
	}
	iw := float64(tileImage.Bounds().Dx())
	ih := float64(tileImage.Bounds().Dy())
	w := iw * scale
	h := ih * scale

	// Tiled anchors tile images to the bottom-left of their cell.
	fx := tx
	fy := ty + float64(th) - h
	cx := fx + w/2
	cy := fy + h/2

	hf, vf, df := horizontalFlip, verticalFlip, diagonalFlip
	sx, sy := scale, scale
	rot := 0.0
	if df {
		rot = math.Pi / 2
		hf, vf = vf, !hf
	}
	if hf {
		sx = -sx
	}
	if vf {
		sy = -sy
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-iw/2, -ih/2)
	op.GeoM.Scale(sx, sy)
	if rot != 0 {
		op.GeoM.Rotate(rot)
	}
	op.GeoM.Translate(cx, cy)
	if alpha > 0 && alpha < 1 {
		op.ColorScale.ScaleAlpha(alpha)
	}
	target.DrawImage(tileImage, op)
}