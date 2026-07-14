package queues

import (
	"runtime"
	"sync"
	"time"
)

type Updateable interface {
	Update(dt time.Duration)
}

type ParallelSafe interface {
	IsParallelSafe() bool
}

type UpdateQueue struct {
	mu      sync.Mutex
	objects []Updateable
}

var DefaultUpdateQueue = &UpdateQueue{}

func (q *UpdateQueue) Schedule(o Updateable) {
	q.mu.Lock()
	q.objects = append(q.objects, o)
	q.mu.Unlock()
}

func (q *UpdateQueue) Unschedule(o Updateable) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, obj := range q.objects {
		if obj == o {
			q.objects = append(q.objects[:i], q.objects[i+1:]...)
			return
		}
	}
}

const minParallelItems = 64

func (q *UpdateQueue) Execute(dt time.Duration) {
	q.mu.Lock()
	all := append([]Updateable{}, q.objects...)
	q.mu.Unlock()

	var safe, sequential []Updateable
	for _, obj := range all {
		if s, ok := obj.(ParallelSafe); ok && s.IsParallelSafe() {
			safe = append(safe, obj)
		} else {
			sequential = append(sequential, obj)
		}
	}

	for _, obj := range sequential {
		obj.Update(dt)
	}

	if len(safe) < minParallelItems {
		for _, obj := range safe {
			obj.Update(dt)
		}
		return
	}

	n := runtime.GOMAXPROCS(0)
	if n > len(safe) {
		n = len(safe)
	}
	if n < 1 {
		n = 1
	}

	chunkSize := len(safe) / n
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == n-1 {
			end = len(safe)
		}
		wg.Add(1)
		go func(batch []Updateable) {
			defer wg.Done()
			for _, obj := range batch {
				obj.Update(dt)
			}
		}(safe[start:end])
	}
	wg.Wait()
}
