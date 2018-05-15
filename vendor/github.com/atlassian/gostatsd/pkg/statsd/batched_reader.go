package statsd

import (
	"net"
	"runtime"

	"golang.org/x/net/ipv6"
)

type Message struct {
	Buffers [][]byte
	Addr    net.Addr
	N       int
}

type BatchReader interface {
	ReadBatch(ms []Message) (int, error)
}

type V6BatchReader struct {
	conn *ipv6.PacketConn
}

type GenericBatchReader struct {
	conn net.PacketConn
}

func NewBatchReader(conn net.PacketConn) BatchReader {
	if runtime.GOOS == "windows" {
		return &GenericBatchReader{
			conn: conn,
		}
	}
	switch c := conn.(type) {
	case *net.UDPConn:
		return &V6BatchReader{
			conn: ipv6.NewPacketConn(c),
		}
	default:
		return &GenericBatchReader{
			conn: conn,
		}
	}
}

func (br *V6BatchReader) ReadBatch(ms []Message) (int, error) {
	ms6 := make([]ipv6.Message, len(ms))
	for i, m := range ms {
		ms6[i].Buffers = m.Buffers
	}
	count, err := br.conn.ReadBatch(ms6, 0)
	if err != nil {
		return 0, err
	}

	for i := 0; i < count; i++ {
		ms[i].Addr = ms6[i].Addr
		ms[i].N = ms6[i].N
	}

	return count, nil
}

func (gbr *GenericBatchReader) ReadBatch(ms []Message) (int, error) {
	if len(ms) == 0 {
		// This tends to happen in test code when the batch size is not set,
		// if production code starts at all, then a panic won't be a hazard.
		panic("attempt to read 0 packets")
	}
	nbytes, addr, err := gbr.conn.ReadFrom(ms[0].Buffers[0])
	if err != nil {
		return 0, err
	}
	ms[0].Addr = addr
	ms[0].N = nbytes
	return 1, nil
}
