package agentd

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/transport"
)

var (
	// upgrader is safe for concurrent use, and we don't need any particularly
	// specialized configurations for different uses.
	upgrader = &websocket.Upgrader{}
)

// Agentd is the backend HTTP API.
type Agentd struct {
	stopping chan struct{}
	running  *atomic.Value
	wg       *sync.WaitGroup
	errChan  chan error

	Store      store.Store
	Port       int
	MessageBus messaging.MessageBus
}

// Start Agentd.
func (a *Agentd) Start() error {
	if a.Store == nil {
		return errors.New("no store found")
	}

	a.stopping = make(chan struct{}, 1)
	a.running = &atomic.Value{}
	a.wg = &sync.WaitGroup{}

	a.errChan = make(chan error, 1)

	handler := http.HandlerFunc(a.webSocketHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", a.Port),
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("starting agentd on address: ", server.Addr)

	go func() {
		defer a.wg.Done()
		server.ListenAndServe()
	}()

	return nil
}

// Stop Agentd.
func (a *Agentd) Stop() error {
	a.running.Store(false)
	close(a.stopping)
	a.wg.Wait()
	close(a.errChan)

	return nil
}

// Status returns an error if Agentd is unhealthy.
func (a *Agentd) Status() error {
	return nil
}

// Err returns a channel to listen for terminal errors on.
func (a *Agentd) Err() <-chan error {
	return a.errChan
}

func (a *Agentd) webSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("transport error on websocket upgrade: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := NewSession(transport.NewTransport(conn), a.MessageBus, a.Store)
	if err != nil {
		logger.Error("failed to create session: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = session.Start()
	if err != nil {
		logger.Error("failed to start session: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
