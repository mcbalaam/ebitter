package queues

import "sync"

type Destroyable interface {
	Destroy()
}

type DeleteQueue struct {
	mu      sync.Mutex
	objects []Destroyable
}

var DefaultDeleteQueue = &DeleteQueue{}

func QDel(obj Destroyable) {
	DefaultDeleteQueue.Schedule(obj)
}

func (q *DeleteQueue) Schedule(obj Destroyable) {
	q.mu.Lock()
	q.objects = append(q.objects, obj)
	q.mu.Unlock()
}

func (q *DeleteQueue) Execute() {
	q.mu.Lock()
	all := append([]Destroyable{}, q.objects...)
	q.objects = q.objects[:0]
	q.mu.Unlock()

	for _, obj := range all {
		obj.Destroy()
	}
}
