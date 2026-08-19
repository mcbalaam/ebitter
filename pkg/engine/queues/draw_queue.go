package queues

import (
	"sort"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	LayerBackground = -200
	LayerArena      = -100
	LayerEntity     = 0
	LayerUI         = 100
	LayerText       = 200
	LayerOverlay    = 300
)

type Drawable interface {
	Draw(s *ebiten.Image)
}

type drawEntry struct {
	layer int
	obj   Drawable
}

type DrawQueue struct {
	mu      sync.Mutex
	entries []drawEntry
	dirty   bool
}

var DefaultQueue = &DrawQueue{}

func (d *DrawQueue) Schedule(o Drawable) {
	d.ScheduleAt(o, 0)
}

func (d *DrawQueue) ScheduleAt(o Drawable, layer int) {
	d.mu.Lock()
	d.entries = append(d.entries, drawEntry{layer: layer, obj: o})
	d.dirty = true
	d.mu.Unlock()
}

func (d *DrawQueue) Unschedule(o Drawable) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, e := range d.entries {
		if e.obj == o {
			d.entries = append(d.entries[:i], d.entries[i+1:]...)
			return
		}
	}
}

func (d *DrawQueue) Execute(s *ebiten.Image) {
	d.mu.Lock()
	if d.dirty {
		d.sort()
		d.dirty = false
	}
	entries := append([]drawEntry{}, d.entries...)
	d.mu.Unlock()

	for _, e := range entries {
		e.obj.Draw(s)
	}
}

func (d *DrawQueue) sort() {
	sort.SliceStable(d.entries, func(i, j int) bool {
		return d.entries[i].layer < d.entries[j].layer
	})
}
