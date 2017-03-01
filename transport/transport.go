package transport

import (
	"bytes"
	"errors"
)

var (
	sep = []byte("\n")
)

// Encode a message to be sent over a websocket channel
func Encode(msgType string, payload []byte) []byte {
	buf := []byte(msgType + "\n")
	buf = append(buf, payload...)
	return buf
}

// Decode a message received from a websocket channel.
func Decode(payload []byte) (string, []byte, error) {
	nl := bytes.Index(sep, payload)
	if nl < 0 {
		return "", nil, errors.New("Invalid message.")
	}

	msgType := payload[0:nl]
	msg := payload[nl+1:]
	return string(msgType), msg, nil
}
