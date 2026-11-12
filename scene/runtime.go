package scene

import (
	"github.com/mcbalaam/ebitter/events"
	"github.com/mcbalaam/ebitter/queues"
	"github.com/mcbalaam/ebitter/text"
)

// Runtime contains mutable services owned by one scene.
type Runtime struct {
	Queues  *queues.QueueSet
	Signals *events.SignalBus
	Dialog  *text.DialogManager
}

// NewRuntime creates scene-local queues, signal bus, and dialog state.
func NewRuntime() *Runtime {
	q := queues.NewQueueSet()
	dialog := text.NewDialogManager()
	dialog.Queues = q
	return &Runtime{
		Queues:  q,
		Signals: &events.SignalBus{},
		Dialog:  dialog,
	}
}

// RuntimeProvider is implemented by scenes that own scene-scoped services.
type RuntimeProvider interface {
	Runtime() *Runtime
}

// LifecycleScene receives stack activation changes from SceneManager.
type LifecycleScene interface {
	Pause()
	Resume()
}
