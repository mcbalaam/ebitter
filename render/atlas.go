// Package render provides tools for managing and rendering animated icons from spritesheets.
package render

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"io/fs"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/assetfs"
)

// AtlasMeta is a single atlas entity (png + aseprite-style JSON).
type AtlasMeta struct {
	Name   string
	States []IconStateMeta
}

// AtlasManager caches cut icons to reuse them.
type AtlasManager struct {
	mu             sync.RWMutex
	IconStateCache map[string]IconState
	processed      map[string]bool
	dirIcons       map[string][]string
	// FS is the filesystem icons are read from. Nil falls back to assetfs.FS.
	FS fs.FS
}

// The main atlas manager instance.
var MasterAtlasManager = &AtlasManager{
	IconStateCache: map[string]IconState{},
	processed:      map[string]bool{},
	dirIcons:       map[string][]string{},
}

func (m *AtlasManager) fsys() fs.FS {
	if m.FS != nil {
		return m.FS
	}
	return assetfs.FS
}

// CacheIconStates parses every Aseprite .json/.png pair in dir, cuts the
// spritesheets and caches the resulting icon states. Re-scanning a directory is
// skipped once processed.
func (m *AtlasManager) CacheIconStates(dir string) error {
	m.mu.RLock()
	done := m.processed[dir]
	m.mu.RUnlock()
	if done {
		return nil
	}

	entries, err := fs.ReadDir(m.fsys(), dir)
	if err != nil {
		return err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".json") {
			continue
		}

		jsonPath := path.Join(dir, name)
		b, err := fs.ReadFile(m.fsys(), jsonPath)
		if err != nil {
			return err
		}

		statesMeta, err := ParseAsepriteJSON(b)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", jsonPath, err)
		}

		base := strings.TrimSuffix(name, path.Ext(name))
		pngPath := path.Join(dir, base+".png")
		if _, err := fs.Stat(m.fsys(), pngPath); err != nil {
			continue
		}

		f, err := m.fsys().Open(pngPath)
		if err != nil {
			return fmt.Errorf("open png %s: %w", pngPath, err)
		}

		meta := AtlasMeta{
			Name:   base,
			States: statesMeta,
		}

		iconStates, err := m.CutAtlas(f, meta)
		f.Close()
		if err != nil {
			return fmt.Errorf("cut atlas %s: %w", pngPath, err)
		}

		m.mu.Lock()
		for _, st := range iconStates {
			m.IconStateCache[base+"/"+st.Name] = st
		}
		m.mu.Unlock()

		names = append(names, base)
	}

	m.mu.Lock()
	m.dirIcons[dir] = names
	m.processed[dir] = true
	m.mu.Unlock()

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
				Image:   &eimg,
				Time:    time.Duration(fm.DurationMS) * time.Millisecond,
				Advance: float64(fm.Advance),
			}
			frames = append(frames, frm)
		}

		iconState := IconState{
			Name:   s.Name,
			Frames: frames,
			Mode:   animationModeFromAseprite(s.Mode),
		}
		if len(frames) > 0 {
			iconState.CurrentFrame = 0
		}

		states = append(states, iconState)
	}

	return states, nil
}

// AnimatedIcon builds an icon from the assets in dir, starting in stateKey.
// Icon names are tracked per directory, so the directory name need not match
// the .json/.png basename.
func (m *AtlasManager) AnimatedIcon(dir string, stateKey string) (*AnimatedIcon, error) {
	if err := m.CacheIconStates(dir); err != nil {
		return nil, fmt.Errorf("cache icon states: %w", err)
	}

	m.mu.RLock()
	names := m.dirIcons[dir]
	nameSet := make(map[string]struct{}, len(names))
	for _, n := range names {
		nameSet[n] = struct{}{}
	}
	iconStates := make(map[string]*IconState)
	for k, v := range m.IconStateCache {
		name, state, ok := strings.Cut(k, "/")
		if !ok {
			continue
		}
		if _, ok := nameSet[name]; !ok {
			continue
		}
		st := v
		iconStates[state] = &st
	}
	m.mu.RUnlock()

	initial, ok := iconStates[stateKey]
	if !ok {
		return nil, fmt.Errorf("get icon state %q: not found", stateKey)
	}
	if len(initial.Frames) == 0 {
		return nil, fmt.Errorf("icon state %q: no frames", stateKey)
	}

	initial.CurrentFrame = initialFrame(initial.Mode, len(initial.Frames))
	initial.CurrentFrameRef = &initial.Frames[initial.CurrentFrame]
	initial.elapsed = 0
	initial.dir = initialDir(initial.Mode)

	return &AnimatedIcon{
		CurrentState: initial,
		IconStates:   iconStates,
	}, nil
}

// initialDir returns the starting ping-pong direction for a mode.
func initialDir(mode AnimationMode) int {
	if mode == AnimationModePingPongReverse {
		return -1
	}
	return 1
}

// initialFrame returns the frame a state starts on: the last frame for a
// reverse ping-pong (which plays backward first), the first otherwise.
func initialFrame(mode AnimationMode, n int) int {
	if mode == AnimationModePingPongReverse && n > 0 {
		return n - 1
	}
	return 0
}

// GetIconState returns a cached icon state by compound key "iconName/iconState".
func (m *AtlasManager) GetIconState(iconName, iconState string) (IconState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := iconName + "/" + iconState
	st, ok := m.IconStateCache[key]
	if ok {
		return st, nil
	}
	return IconState{}, fmt.Errorf("icon state %q not found in cache", key)
}
