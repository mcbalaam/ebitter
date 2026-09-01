package tween

import (
	"math"
	"time"

	"github.com/mcbalaam/ebitter/queues"
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
	from    T
	to      T
	current T
	setter  func(T)
	lerp    func(from, to T, t float64) T

	duration     time.Duration
	elapsed      time.Duration
	delay        time.Duration
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
		from: from, to: to, current: from, duration: duration,
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

// SetDuration overrides the tween length; used when reviving a follow tween
// with a fresh duration.
func (t *Tween[T]) SetDuration(d time.Duration) *Tween[T] {
	t.duration = d
	return t
}

func (t *Tween[T]) OnComplete(fn func()) *Tween[T] {
	t.onComplete = fn
	return t
}

// Schedule adds the tween to the default update queue if it is not already
// there. Calling it repeatedly is safe.
func (t *Tween[T]) Schedule() *Tween[T] {
	if !queues.DefaultUpdateQueue.IsScheduled(t) {
		queues.DefaultUpdateQueue.Schedule(t)
	}
	return t
}

// Value returns the tween's most recently applied value.
func (t *Tween[T]) Value() T {
	return t.current
}

// Stop halts a tween and removes it from the default update queue without
// firing its completion callback.
func Stop[T any](t *Tween[T]) {
	if t == nil {
		return
	}
	t.done = true
	queues.DefaultUpdateQueue.Unschedule(t)
}

// Retarget redirects a running (or finished) tween toward a new destination,
// restarting from its current value. It is meant for follow-style tweens whose
// target moves every frame (e.g. a camera chasing a player). Retargeting a
// finished tween revives it and re-schedules it on the default queue.
func (t *Tween[T]) Retarget(to T) {
	t.from = t.current
	t.to = to
	t.elapsed = 0
	t.forward = true
	if t.done {
		t.done = false
		t.repeated = 0
		t.started = false
		t.Schedule()
	}
}

func (t *Tween[T]) apply(v T) {
	t.current = v
	t.setter(v)
}

func (t *Tween[T]) Update(dt time.Duration) {
	if t.done {
		return
	}

	if !t.started {
		t.started = true
		t.elapsed = 0
		t.apply(t.from)
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
		t.apply(t.lerp(t.from, t.to, eased))
	} else {
		t.apply(t.lerp(t.from, t.to, 1-eased))
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
