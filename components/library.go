package components

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
	"sync"

	"github.com/mcbalaam/ebitter/assetfs"
	"github.com/mcbalaam/ebitter/events"
	"github.com/mcbalaam/ebitter/queues"
)

// BlueprintLibrary loads and caches entity blueprints from a directory of YAML
// files, one entity per file named after its id.
type BlueprintLibrary struct {
	mu   sync.RWMutex
	fsys fs.FS
	dir  string
	byID map[string]*Blueprint
}

// NewBlueprintLibrary returns a library reading from dir. A nil filesystem
// falls back to assetfs.FS.
func NewBlueprintLibrary(fsys fs.FS, dir string) *BlueprintLibrary {
	if fsys == nil {
		fsys = assetfs.FS
	}
	return &BlueprintLibrary{
		fsys: fsys,
		dir:  dir,
		byID: map[string]*Blueprint{},
	}
}

// Load returns the blueprint with the given id, reading dir/<id>.yml on first
// use.
func (l *BlueprintLibrary) Load(id string) (*Blueprint, error) {
	l.mu.RLock()
	bp := l.byID[id]
	l.mu.RUnlock()
	if bp != nil {
		return bp, nil
	}

	f, err := l.fsys.Open(path.Join(l.dir, id+".yml"))
	if err != nil {
		return nil, fmt.Errorf("blueprint %q: %w", id, err)
	}
	defer f.Close()

	bp, err = LoadBlueprint(f)
	if err != nil {
		return nil, fmt.Errorf("blueprint %q: %w", id, err)
	}
	if bp.ID == "" {
		bp.ID = id
	}

	l.mu.Lock()
	l.byID[bp.ID] = bp
	l.mu.Unlock()
	return bp, nil
}

// LoadAll reads every .yml/.yaml file in the library directory.
func (l *BlueprintLibrary) LoadAll() error {
	entries, err := fs.ReadDir(l.fsys, l.dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		lower := strings.ToLower(name)
		if !strings.HasSuffix(lower, ".yml") && !strings.HasSuffix(lower, ".yaml") {
			continue
		}
		id := strings.TrimSuffix(strings.TrimSuffix(name, ".yml"), ".yaml")
		id = strings.TrimSuffix(id, ".YML")
		id = strings.TrimSuffix(id, ".YAML")
		if _, err := l.Load(id); err != nil {
			return err
		}
	}
	return nil
}

// Spawn instantiates the blueprint id and places its transform at (x, y).
func (l *BlueprintLibrary) Spawn(id string, x, y float64) (*Entity, error) {
	return l.SpawnWith(id, x, y, nil, nil)
}

// SpawnWith instantiates a prototype at (x, y) with an optional scene bus and
// queue set. Nil dependencies use package defaults.
func (l *BlueprintLibrary) SpawnWith(id string, x, y float64, bus *events.SignalBus, q *queues.QueueSet) (*Entity, error) {
	bp, err := l.Load(id)
	if err != nil {
		return nil, err
	}
	e, err := bp.InstantiateWith(bus, q)
	if err != nil {
		return nil, err
	}
	if t := e.Transform(); t != nil {
		t.X, t.Y = x, y
	}
	return e, nil
}
