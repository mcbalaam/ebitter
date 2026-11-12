package queues

import (
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// QueueSet groups the queues owned by one game or scene. Pausing a set stops
// its update, draw, and delete queues from executing without discarding work.
type QueueSet struct {
	Update *UpdateQueue
	Draw   *DrawQueue
	Delete *DeleteQueue
	paused atomic.Bool
}

// NewQueueSet creates an independent set of queues.
func NewQueueSet() *QueueSet {
	return &QueueSet{
		Update: &UpdateQueue{},
		Draw:   &DrawQueue{},
		Delete: &DeleteQueue{},
	}
}

// DefaultQueueSet adapts the package's legacy process-wide queues.
var DefaultQueueSet = &QueueSet{
	Update: DefaultUpdateQueue,
	Draw:   DefaultQueue,
	Delete: DefaultDeleteQueue,
}

func (s *QueueSet) Pause()       { s.paused.Store(true) }
func (s *QueueSet) Resume()      { s.paused.Store(false) }
func (s *QueueSet) Paused() bool { return s.paused.Load() }

func (s *QueueSet) ExecuteUpdate(dt time.Duration) {
	if s != nil && !s.Paused() && s.Update != nil {
		s.Update.Execute(dt)
	}
}

func (s *QueueSet) ExecuteDraw(screen *ebiten.Image) {
	if s != nil && !s.Paused() && s.Draw != nil {
		s.Draw.Execute(screen)
	}
}

func (s *QueueSet) ExecuteDelete() {
	if s != nil && !s.Paused() && s.Delete != nil {
		s.Delete.Execute()
	}
}

// QDelOn schedules obj on the specified queue set. A nil set uses the legacy
// default delete queue.
func QDelOn(set *QueueSet, obj Destroyable) {
	if set == nil || set.Delete == nil {
		QDel(obj)
		return
	}
	set.Delete.Schedule(obj)
}
