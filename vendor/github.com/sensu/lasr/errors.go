package lasr

import "errors"

var (
	// ErrAckNack is returned by Ack and Nack when of them has been called already.
	ErrAckNack = errors.New("lasr: Ack or Nack already called")

	// ErrQClosed is returned by Send, Receive and Close when the Q has already
	// been closed.
	ErrQClosed = errors.New("lasr: Q is closed")

	// ErrOptionsApplied is called when an Option is applied to a Q after NewQ
	// has already returned.
	ErrOptionsApplied = errors.New("lasr: options cannot be applied after New")
)
