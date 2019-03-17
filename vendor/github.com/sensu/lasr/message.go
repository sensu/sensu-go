package lasr

import (
	"bytes"
	"encoding"
	"encoding/binary"
)

// ID is used for uniquely identifying messages in a Q.
type ID interface {
	encoding.BinaryMarshaler
}

// Uint64ID is the default ID used by lasr.
type Uint64ID uint64

func (id Uint64ID) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	err := binary.Write(buf, binary.BigEndian, id)
	return buf.Bytes(), err
}

func (id *Uint64ID) UnmarshalBinary(b []byte) error {
	return binary.Read(bytes.NewReader(b), binary.BigEndian, id)
}

// Message is a messaged returned from Q on Receive.
//
// Message contains a Body and an ID. The ID will be equal to the ID that was
// returned on Send, Delay or Wait for this message.
type Message struct {
	Body []byte
	ID   []byte
	q    *Q
	once int32
	err  error
}
