package components

import (
	"runtime"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/pkg/engine/queues"
	"github.com/mcbalaam/ebitter/pkg/render"
	"github.com/mcbalaam/ebitter/pkg/systems"
)

type Entity struct {
	Transform *Transform
	Velocity  *Velocity
	Sprite    *Sprite
	Collider  *Collider

	Layer int
}

func (e *Entity) IsParallelSafe() bool { return true }

func (e *Entity) Update(dt time.Duration) {
	if e.Velocity != nil && e.Transform != nil {
		seconds := dt.Seconds()
		e.Transform.X += e.Velocity.X * seconds
		e.Transform.Y += e.Velocity.Y * seconds
	}
	if e.Sprite != nil {
		e.Sprite.Update(dt)
	}
	if e.Collider != nil && e.Transform != nil {
		e.Collider.UpdateWorldVerts(e.Transform)
	}
}

func (e *Entity) Draw(s *ebiten.Image) {
	if e.Sprite == nil || e.Transform == nil {
		return
	}
	e.Sprite.Icon.Draw(s, e.Transform.X, e.Transform.Y,
		e.Transform.ScaleX, e.Transform.ScaleY, e.Transform.Rotation)
}

func (e *Entity) Destroy() {
	queues.DefaultQueue.Unschedule(e)
	queues.DefaultUpdateQueue.Unschedule(e)
	e.Transform = nil
	e.Velocity = nil
	e.Sprite = nil
	e.Collider = nil
}

const minParallelCollisions = 64

type collisionPair struct {
	self  *Entity
	other *Entity
}

func (e *Entity) CheckCollisionsWithList(others []*Entity, signalName string) {
	n := len(others)
	if n < minParallelCollisions {
		for _, other := range others {
			if e.Collider != nil && other.Collider != nil {
				if e.Collider.CollidesWith(other.Collider) {
					systems.MasterSignalBus.Emit(signalName, e, other)
				}
			}
		}
		return
	}

	numWorkers := runtime.GOMAXPROCS(0)
	if numWorkers > n {
		numWorkers = n
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	chunkSize := n / numWorkers
	var mu sync.Mutex
	hits := make([]collisionPair, 0, n)

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == numWorkers-1 {
			end = n
		}
		wg.Add(1)
		go func(chunk []*Entity) {
			defer wg.Done()
			for _, other := range chunk {
				if e.Collider != nil && other.Collider != nil && e.Collider.CollidesWith(other.Collider) {
					mu.Lock()
					hits = append(hits, collisionPair{self: e, other: other})
					mu.Unlock()
				}
			}
		}(others[start:end])
	}
	wg.Wait()

	for _, p := range hits {
		systems.MasterSignalBus.Emit(signalName, p.self, p.other)
	}
}

func DrawSpriteOnCollider(screen *ebiten.Image, icon *render.AnimatedIcon, transform *Transform, collider *Collider) {
	if icon == nil || transform == nil || collider == nil {
		return
	}
	offsetVec := Vec{X: collider.OffsetX, Y: collider.OffsetY}
	transformedOffset := TransformPoint(offsetVec, transform.ScaleX, transform.ScaleY, transform.Rotation, 0, 0)
	centerX := transform.X - (collider.Width*transform.ScaleX)/2
	centerY := transform.Y - (collider.Height*transform.ScaleY)/2
	icon.Draw(screen, centerX+transformedOffset.X, centerY+transformedOffset.Y,
		transform.ScaleX, transform.ScaleY, transform.Rotation)
}
