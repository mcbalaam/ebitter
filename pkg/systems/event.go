package systems

import "sync"

type SignalInteractor interface{}

type Signal struct {
	Name    string
	Emitent *SignalInteractor
	Data    interface{}
}

type SignalHandler func(Signal)

type SignalSubscription struct {
	Name      string
	Recepient *SignalInteractor
	Handler   SignalHandler
}

type SignalBus struct {
	mu            sync.RWMutex
	Subscriptions []SignalSubscription
}

var MasterSignalBus = SignalBus{}

func (b *SignalBus) Subscribe(name string, recepient SignalInteractor, handler SignalHandler) {
	b.mu.Lock()
	b.Subscriptions = append(b.Subscriptions, SignalSubscription{
		Name:      name,
		Recepient: &recepient,
		Handler:   handler,
	})
	b.mu.Unlock()
}

func (b *SignalBus) Emit(name string, source SignalInteractor, data ...interface{}) {
	signal := Signal{
		Name:    name,
		Emitent: &source,
	}

	if len(data) > 0 {
		signal.Data = data[0]
	}

	b.mu.RLock()
	for _, sub := range b.Subscriptions {
		if sub.Name == name {
			sub.Handler(signal)
		}
	}
	b.mu.RUnlock()
}
