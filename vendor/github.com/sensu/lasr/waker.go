package lasr

import (
	"container/heap"
	"sync"
	"time"
)

type timeHeap []time.Time

func (h timeHeap) Len() int {
	return len(h)
}

func (h timeHeap) Less(i, j int) bool {
	return h[j].After(h[i])
}

func (h timeHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *timeHeap) Push(x interface{}) {
	*h = append(*h, x.(time.Time))
}

func (h *timeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *timeHeap) PushTime(t time.Time) {
	heap.Push(h, t)
}

func (h *timeHeap) PopTime() time.Time {
	if len(*h) == 0 {
		never := time.Unix(0, 1<<63-1)
		return never
	}
	return heap.Pop(h).(time.Time)
}

// waker wakes up when told to, or in the future according to an ordered
// list of times (stored by timeHeap).
type waker struct {
	C        chan struct{}
	closed   chan struct{}
	reset    chan struct{}
	nextWake time.Time
	wakes    timeHeap
	sync.Mutex
}

func newWaker(closed chan struct{}) *waker {
	w := &waker{
		C:      make(chan struct{}, 1),
		reset:  make(chan struct{}, 1),
		closed: closed,
	}
	go func() {
		for {
			w.Lock()
			nextWake := w.wakes.PopTime()
			w.Unlock()
			at := nextWake.Sub(time.Now())
			timer := time.NewTimer(at)
			select {
			case <-timer.C:
				w.Wake()
				timer.Stop()
			case <-w.closed:
				timer.Stop()
				return
			case <-w.reset:
				w.Lock()
				w.wakes.PushTime(nextWake)
				w.Unlock()
				timer.Stop()
			}
		}
	}()
	return w
}

func (w *waker) Wake() {
	select {
	case <-w.closed:
		panic("Wake() on closed waker")
	case w.C <- struct{}{}:
	default:
	}
}

func (w *waker) WakeAt(t time.Time) {
	select {
	case <-w.closed:
		panic("WakeAt() on closed waker")
	default:
	}
	if time.Now().After(t) {
		w.Wake()
	}
	reset := false
	w.Lock()
	defer func() {
		w.Unlock()
		if reset {
			w.reset <- struct{}{}
		}
	}()
	if w.nextWake.IsZero() || w.nextWake.After(t) {
		reset = true
		w.nextWake = t
	}
	w.wakes.PushTime(t)
}
