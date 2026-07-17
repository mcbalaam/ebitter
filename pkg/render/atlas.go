package render

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/pkg/assets"
	"github.com/mcbalaam/ebitter/pkg/embedfs"
)

// AtlasMeta is a single atlas entity (png + aseprite-style JSON).
type AtlasMeta struct {
	Name   string
	States []assets.IconStateMeta
}

// AtlasManager caches cut icons to reuse them.
type AtlasManager struct {
	mu             sync.RWMutex
	IconStateCache map[string]IconState
	BasePath       string
}

// The main atlas manager instance.
var MasterAtlasManager = &AtlasManager{
	IconStateCache: map[string]IconState{},
	BasePath:       "media/sprites",
}

// Parses an Aseprite format .json file, reads the .png spritesheet from the provided directory and cuts it. Caches the result.
func (m *AtlasManager) CacheIconStates(path string) error {
	entries, err := fs.ReadDir(embedfs.FS, path)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".json") {
			continue
		}

		jsonPath := filepath.Join(path, name)
		b, err := fs.ReadFile(embedfs.FS, jsonPath)
		if err != nil {
			return err
		}

		statesMeta, err := assets.ParseAsepriteJSON(b)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", jsonPath, err)
		}

		base := strings.TrimSuffix(name, filepath.Ext(name))
		pngPath := filepath.Join(path, base+".png")
		if _, err := fs.Stat(embedfs.FS, pngPath); err == nil {
			f, err := embedfs.FS.Open(pngPath)
			if err != nil {
				return fmt.Errorf("open png %s: %w", pngPath, err)
			}

			func() {
				defer f.Close()

				meta := AtlasMeta{
					Name:   base,
					States: statesMeta,
				}

				iconStates, err := m.CutAtlas(f, meta)
				if err != nil {
					err = fmt.Errorf("cut atlas %s: %w", pngPath, err)
					panic(err)
				}

				m.mu.Lock()
				for _, st := range iconStates {
					m.IconStateCache[st.Name] = st
				}
				m.mu.Unlock()
			}()
			if r := recover(); r != nil {
				if errVal, ok := r.(error); ok {
					return errVal
				}
				return fmt.Errorf("unexpected error: %v", r)
			}
		}
	}

	return nil
}

// Cuts the .png spritesheet into `Frame`s and combines them into `IconState`s.
func (m *AtlasManager) CutAtlas(r io.Reader, meta AtlasMeta) ([]IconState, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	states := make([]IconState, 0, len(meta.States))

	for _, s := range meta.States {
		frames := make([]Frame, 0, len(s.Frames))
		for _, fm := range s.Frames {
			rect := fm.Rect.Intersect(bounds)
			if rect.Empty() {
				return nil, fmt.Errorf("frame rect out of bounds: %v", fm.Rect)
			}

			subImg, ok := img.(interface {
				SubImage(r image.Rectangle) image.Image
			})
			if !ok {
				return nil, fmt.Errorf("image does not support SubImage")
			}
			sub := subImg.SubImage(rect)

			eimg := *ebiten.NewImageFromImage(sub)
			frm := Frame{
				Image: &eimg,
				Time:  time.Duration(fm.DurationMS) * time.Millisecond,
			}
			frames = append(frames, frm)
		}

		var mode AnimationMode
		switch strings.ToLower(s.Mode) {
		case "forward", "normal", "fwd":
			mode = AnimationModeLoop
		case "reverse", "backward", "rev":
			mode = AnimationModeLoop
		case "pingpong", "ping-pong":
			mode = AnimationModePingPong
		default:
			mode = AnimationModeOnce
		}

		iconState := IconState{
			Name:   s.Name,
			Frames: frames,
			Mode:   mode,
		}
		if len(frames) > 0 {
			iconState.CurrentFrame = 0
		}

		states = append(states, iconState)
	}

	return states, nil
}

// GetIconState returns a cached icon state by compound key "iconName_iconState".
func (m *AtlasManager) GetIconState(iconName, iconState string) (IconState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := iconName + "_" + iconState
	st, ok := m.IconStateCache[key]
	if ok {
		return st, nil
	}
	return IconState{}, fmt.Errorf("icon state %q not found in cache", key)
}
