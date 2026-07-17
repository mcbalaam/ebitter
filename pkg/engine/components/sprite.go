package components

import (
	"time"

	"github.com/mcbalaam/ebitter/pkg/render"
)

type Sprite struct {
	Icon *render.AnimatedIcon
}

func (s *Sprite) Update(dt time.Duration) {
	if s.Icon != nil {
		s.Icon.Update(dt)
	}
}
