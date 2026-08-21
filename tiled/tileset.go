// Package tiled provides loading and basic rendering of Tiled (.tmx) maps,
// including tile layers, collision boxes and interaction zones.
package tiled

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"path/filepath"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	gotiled "github.com/lafriks/go-tiled"
	"github.com/mcbalaam/ebitter/assetfs"
)

// TilesetAsset is a loaded TSX tileset together with its decoded texture.
type TilesetAsset struct {
	Spec    *gotiled.Tileset
	Texture *ebiten.Image

	mu    sync.Mutex
	tiles map[uint32]*ebiten.Image
}

// Tile returns the sub-image for a local tile ID, caching the cut result.
func (t *TilesetAsset) Tile(localID uint32) (*ebiten.Image, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if im, ok := t.tiles[localID]; ok {
		return im, nil
	}

	rect := t.Spec.GetTileRect(localID)
	if !rect.In(t.Texture.Bounds()) {
		return nil, fmt.Errorf("tiled: tile %d rect out of bounds: %v", localID, rect)
	}

	sub, ok := t.Texture.SubImage(rect).(*ebiten.Image)
	if !ok || sub == nil {
		return nil, fmt.Errorf("tiled: tile %d: unexpected sub-image type", localID)
	}
	t.tiles[localID] = sub
	return sub, nil
}

// TilesetCache loads and caches TSX tilesets. Tilesets referenced by multiple
// maps are loaded only once (keyed by their resolved texture path).
type TilesetCache struct {
	mu     sync.Mutex
	byID   map[*gotiled.Tileset]*TilesetAsset
	byPath map[string]*TilesetAsset
}

// NewTilesetCache creates an empty cache.
func NewTilesetCache() *TilesetCache {
	return &TilesetCache{
		byID:   map[*gotiled.Tileset]*TilesetAsset{},
		byPath: map[string]*TilesetAsset{},
	}
}

// DefaultTilesetCache is the shared tileset cache used by LoadMap.
var DefaultTilesetCache = NewTilesetCache()

// Asset returns the cached asset for a tileset or loads it from disk/embed.
func (c *TilesetCache) Asset(ts *gotiled.Tileset) (*TilesetAsset, error) {
	if ts == nil {
		return nil, fmt.Errorf("tiled: nil tileset")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if a, ok := c.byID[ts]; ok {
		return a, nil
	}
	if ts.Image == nil || ts.Image.Source == "" {
		return nil, fmt.Errorf("tiled: tileset %q has no image", ts.Name)
	}

	src := filepath.ToSlash(ts.GetFileFullPath(ts.Image.Source))
	if a, ok := c.byPath[src]; ok {
		c.byID[ts] = a
		return a, nil
	}

	f, err := assetfs.FS.Open(src)
	if err != nil {
		return nil, fmt.Errorf("tiled: open tileset image %s: %w", src, err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("tiled: decode tileset image %s: %w", src, err)
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
	if trans := ts.Image.Trans; trans != nil {
		r, g, b, _ := trans.RGBA()
		target := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 0xff}
		for y := rgba.Bounds().Min.Y; y < rgba.Bounds().Max.Y; y++ {
			for x := rgba.Bounds().Min.X; x < rgba.Bounds().Max.X; x++ {
				c := rgba.RGBAAt(x, y)
				if c.R == target.R && c.G == target.G && c.B == target.B {
					rgba.SetRGBA(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 0})
				}
			}
		}
	}

	a := &TilesetAsset{
		Spec:    ts,
		Texture: ebiten.NewImageFromImage(rgba),
		tiles:   map[uint32]*ebiten.Image{},
	}
	c.byID[ts] = a
	c.byPath[src] = a
	return a, nil
}
