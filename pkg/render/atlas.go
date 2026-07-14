package render

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/pkg/assets"
)

// AtlasMeta is a single atlas entity (png + aseprite-style JSON)
type AtlasMeta struct {
	Name   string
	States []assets.IconStateMeta
}

type AtlasManager struct {
	mu             sync.RWMutex
	IconStateCache map[string]IconState
	BasePath       string
}

var DefaultManager = &AtlasManager{
	IconStateCache: map[string]IconState{},
	BasePath:       "media/sprites",
}

func (m *AtlasManager) CacheIconStates(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	entries, err := os.ReadDir(path)
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
		b, err := os.ReadFile(jsonPath)
		if err != nil {
			return err
		}

		statesMeta, err := assets.ParseAsepriteJSON(b)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", jsonPath, err)
		}

		base := strings.TrimSuffix(name, filepath.Ext(name))
		pngPath := filepath.Join(path, base+".png")
		if _, err := os.Stat(pngPath); err == nil {
			f, err := os.Open(pngPath)
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

func (m *AtlasManager) LoadIconState(key string) (IconState, error) {
	parts := strings.SplitN(key, "_", 2)
	if len(parts) != 2 {
		return IconState{}, fmt.Errorf("invalid key format: expected atlasname_statename")
	}
	atlasName := parts[0]

	spritesDir := filepath.Join(m.BasePath, atlasName)
	jsonPath := filepath.Join(spritesDir, atlasName+".json")
	pngPath := filepath.Join(spritesDir, atlasName+".png")

	if _, err := os.Stat(jsonPath); err != nil {
		return IconState{}, fmt.Errorf("json not found for atlas %q: %w", atlasName, err)
	}
	if _, err := os.Stat(pngPath); err != nil {
		return IconState{}, fmt.Errorf("png not found for atlas %q: %w", atlasName, err)
	}

	b, err := os.ReadFile(jsonPath)
	if err != nil {
		return IconState{}, err
	}
	statesMeta, err := assets.ParseAsepriteJSON(b)
	if err != nil {
		return IconState{}, fmt.Errorf("assets json: %w", err)
	}

	f, err := os.Open(pngPath)
	if err != nil {
		return IconState{}, err
	}
	defer f.Close()

	meta := AtlasMeta{Name: atlasName, States: statesMeta}
	iconStates, err := m.CutAtlas(f, meta)
	if err != nil {
		return IconState{}, fmt.Errorf("cut atlas: %w", err)
	}

	m.mu.Lock()
	for _, st := range iconStates {
		m.IconStateCache[st.Name] = st
	}
	m.mu.Unlock()

	st, err := m.GetIconState(key)
	if err != nil {
		return st, nil
	}
	return IconState{}, fmt.Errorf("state %q not found in atlas %q", parts[1], atlasName)
}

func (m *AtlasManager) GetIconState(key string) (IconState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	st, ok := m.IconStateCache[key]
	if ok {
		return st, nil
	} else {
		return m.LoadIconState(key)
	}
}
