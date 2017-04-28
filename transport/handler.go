package transport

import (
	"fmt"
	"sync"
)

// MessageHandlerFunc is a function accepting a byte array message payload
// that returns an optional error.
type MessageHandlerFunc func(payload []byte) error

// MessageHandler manages mappings of message types to their associated
// message handling functions.
type MessageHandler interface {
	// AddHandler adds a message handler function for a given message type.
	AddHandler(messageType string, handlerFunc MessageHandlerFunc)

	// Handle routes a given message to its associated message handler.
	Handle(msg *Message) error
}

// A MessageHandler is responsible for routing messages of a set of types to
// their associated handler functions.
type mapHandler struct {
	handlerMap     map[string]MessageHandlerFunc
	handlerMapLock *sync.RWMutex
}

// NewMessageHandler initializes and returns a pointer to a new MessageHandler.
func NewMessageHandler() MessageHandler {
	return &mapHandler{
		handlerMap:     map[string]MessageHandlerFunc{},
		handlerMapLock: &sync.RWMutex{},
	}
}

func (h *mapHandler) getHandlerFor(msgType string) (MessageHandlerFunc, error) {
	h.handlerMapLock.RLock()
	defer h.handlerMapLock.RUnlock()

	handlerFunc, ok := h.handlerMap[msgType]
	if !ok {
		return nil, fmt.Errorf("unknown message type: %s", msgType)
	}
	return handlerFunc, nil
}

// AddHandler is used to register a MessageHandlerFunc for a given message type.
// This currently on supports a single handler for each message type. Subsequent
// calls to AddHandler will replace the current handler for a given message
// type. Last write wins.
func (h *mapHandler) AddHandler(msgType string, handlerFunc MessageHandlerFunc) {
	h.handlerMapLock.Lock()
	defer h.handlerMapLock.Unlock()

	h.handlerMap[msgType] = handlerFunc
}

// Handle is used to dispatch a message of msgType type with a byte-array
// payload. This will return an error if the handler function returns an error
// or if there is no handler for a given message type.
func (h *mapHandler) Handle(msg *Message) error {
	handleFunc, err := h.getHandlerFor(msg.Type)
	if err != nil {
		return err
	}

	return handleFunc(msg.Payload)
}
