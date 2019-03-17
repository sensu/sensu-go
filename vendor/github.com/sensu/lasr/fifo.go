package lasr

import "sync"

// fifo is for buffering received messages
type fifo struct {
	data []*Message
	sync.Mutex
}

func newFifo(size int) *fifo {
	return &fifo{
		data: make([]*Message, 0, size),
	}
}

func (f *fifo) Pop() *Message {
	msg := f.data[0]
	f.data = append(f.data[0:0], f.data[1:]...)
	return msg
}

func (f *fifo) Push(m *Message) {
	if len(f.data) == cap(f.data) {
		panic("push to full buffer")
	}
	f.data = append(f.data, m)
}

func (f *fifo) Len() int {
	return len(f.data)
}

func (f *fifo) Cap() int {
	return cap(f.data)
}

func (f *fifo) SetError(err error) {
	for i := range f.data {
		f.data[i].err = err
	}
}

func (f *fifo) Drain() {
	f.data = f.data[0:0]
}
