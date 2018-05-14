package pool

import (
	"sync"
)

// DatabramBufferPool is a strongly typed wrapper around a sync.Pool for *[][]byte
type DatagramBufferPool struct {
	p sync.Pool
}

func NewDatagramBufferPool(bufferSize int) *DatagramBufferPool {
	return &DatagramBufferPool{
		p: sync.Pool{
			New: func() interface{} {
				b := [][]byte{make([]byte, bufferSize)}
				return &b
			},
		},
	}
}

func (p *DatagramBufferPool) Get() *[][]byte {
	return p.p.Get().(*[][]byte)
}

func (p *DatagramBufferPool) Put(b *[][]byte) {
	p.p.Put(b)
}
