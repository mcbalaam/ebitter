// Package events provides a signal bus for subscribing to and emitting
// game events (collisions, interactions, etc.).
package events

import "sync"

type SignalInteractor interface{}

type Signal struct {
	Name   string
	Source SignalInteractor
	Data   interface{}
}

type SignalHandler func(Signal)

type SignalSubscription struct {
	Name      string
	Recipient SignalInteractor
	Handler   SignalHandler
}

type SignalBus struct {
	mu            sync.RWMutex
	Subscriptions []SignalSubscription
}

var MasterSignalBus = SignalBus{}

func (b *SignalBus) Subscribe(name string, recipient SignalInteractor, handler SignalHandler) {
	b.mu.Lock()
	b.Subscriptions = append(b.Subscriptions, SignalSubscription{
		Name:      name,
		Recipient: recipient,
		Handler:   handler,
	})
	b.mu.Unlock()
}

func (b *SignalBus) Emit(name string, source SignalInteractor, data ...interface{}) {
	signal := Signal{
		Name:   name,
		Source: source,
	}

	if len(data) > 0 {
		signal.Data = data[0]
	}

	b.mu.RLock()
	subs := append([]SignalSubscription{}, b.Subscriptions...)
	b.mu.RUnlock()

	for _, sub := range subs {
		if sub.Name == name {
			sub.Handler(signal)
		}
	}
}
