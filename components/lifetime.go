package components

import (
	"time"

	"github.com/mcbalaam/ebitter/queues"
)

// Lifetime destroys the entity after TTL seconds. A non-positive TTL means the
// entity lives forever.
type Lifetime struct {
	TTL float64 `yaml:"ttl"`

	elapsed float64
}

func (l *Lifetime) Type() string { return "lifetime" }

func (l *Lifetime) Update(e *Entity, dt time.Duration) {
	if l.TTL <= 0 {
		return
	}
	l.elapsed += dt.Seconds()
	if l.elapsed >= l.TTL {
		queues.QDelOn(e.Queues, e)
	}
}
