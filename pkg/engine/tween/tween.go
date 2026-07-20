package tween

import (
	"math"
	"time"

	"github.com/mcbalaam/ebitter/pkg/engine/queues"
)

type AnimationMode int

const (
	ModeOnce     AnimationMode = iota
	ModeLoop     AnimationMode = iota
	ModePingPong AnimationMode = iota
)

type Vec2 struct {
	X, Y float64
}

type Tween[T any] struct {
	from   T
	to     T
	setter func(T)
	lerp   func(from, to T, t float64) T

	duration time.Duration
	elapsed  time.Duration
	delay    time.Duration
	delayElapsed time.Duration

	easing  EasingFunc
	mode    AnimationMode
	forward bool

	onComplete func()

	started bool
	done    bool

	repeated int
	repeats  int
}

func New[T any](from, to T, duration time.Duration, setter func(T), lerp func(T, T, float64) T) *Tween[T] {
	return &Tween[T]{
		from: from, to: to, duration: duration,
		setter: setter, lerp: lerp,
		easing: Linear, mode: ModeOnce, forward: true,
	}
}

func NewFloat(from, to float64, duration time.Duration, setter func(float64)) *Tween[float64] {
	return New(from, to, duration, setter,
		func(a, b float64, t float64) float64 { return a + (b-a)*t })
}

func NewVec2(from, to Vec2, duration time.Duration, setter func(Vec2)) *Tween[Vec2] {
	return New(from, to, duration, setter,
		func(a, b Vec2, t float64) Vec2 {
			return Vec2{a.X + (b.X-a.X)*t, a.Y + (b.Y-a.Y)*t}
		})
}

func (t *Tween[T]) SetEasing(easing EasingFunc) *Tween[T] {
	t.easing = easing
	return t
}

func (t *Tween[T]) SetMode(mode AnimationMode) *Tween[T] {
	t.mode = mode
	return t
}

func (t *Tween[T]) SetRepeats(n int) *Tween[T] {
	t.repeats = n
	return t
}

func (t *Tween[T]) Delay(d time.Duration) *Tween[T] {
	t.delay = d
	return t
}

func (t *Tween[T]) OnComplete(fn func()) *Tween[T] {
	t.onComplete = fn
	return t
}

func (t *Tween[T]) Update(dt time.Duration) {
	if t.done {
		return
	}

	if !t.started {
		t.started = true
		t.elapsed = 0
		t.setter(t.from)
	}

	if t.delayElapsed < t.delay {
		t.delayElapsed += dt
		if t.delayElapsed < t.delay {
			return
		}
		dt = t.delayElapsed - t.delay
	}

	t.elapsed += dt

	p := math.Min(float64(t.elapsed)/float64(t.duration), 1.0)
	eased := t.easing(p)

	if t.forward {
		t.setter(t.lerp(t.from, t.to, eased))
	} else {
		t.setter(t.lerp(t.from, t.to, 1-eased))
	}

	if p >= 1.0 {
		t.complete()
	}
}

func (t *Tween[T]) complete() {
	t.repeated++

	stop := false
	if t.repeats > 0 && t.repeated >= t.repeats {
		stop = true
	} else if t.repeats == 0 && t.mode == ModeOnce {
		stop = true
	}

	if stop {
		t.done = true
		if t.onComplete != nil {
			t.onComplete()
		}
		queues.DefaultUpdateQueue.Unschedule(t)
		return
	}

	switch t.mode {
	case ModeLoop:
		t.elapsed = 0
	case ModePingPong:
		t.forward = !t.forward
		t.elapsed = 0
	default:
		t.elapsed = 0
	}
}
