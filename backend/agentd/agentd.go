package agentd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
)

var (
	// upgrader is safe for concurrent use, and we don't need any particularly
	// specialized configurations for different uses.
	upgrader = &websocket.Upgrader{}
)

// Store specifies storage requirements for Agentd.
type Store interface {
	middlewares.AuthStore
	SessionStore
}

// Agentd is the backend HTTP API.
type Agentd struct {
	stopping   chan struct{}
	running    *atomic.Value
	wg         *sync.WaitGroup
	errChan    chan error
	httpServer *http.Server

	Store      Store
	Host       string
	Port       int
	MessageBus messaging.MessageBus
	TLS        *types.TLSOptions
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

	// TODO: add JWT authentication support
	handler := middlewares.BasicAuthentication(http.HandlerFunc(a.webSocketHandler), a.Store)

	a.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.Host, a.Port),
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("starting agentd on address: ", a.httpServer.Addr)
	a.wg.Add(1)

	go func() {
		defer a.wg.Done()
		var err error
		if a.TLS != nil {
			err = a.httpServer.ListenAndServeTLS(a.TLS.CertFile, a.TLS.KeyFile)
		} else {
			err = a.httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			logger.Errorf("failed to start http/https server %s", err.Error())
		}
	}()

	return nil
}

// Stop Agentd.
func (a *Agentd) Stop() error {
	if err := a.httpServer.Shutdown(context.TODO()); err != nil {
		// failure/timeout shutting down the server gracefully
		logger.Error("failed to shutdown http server gracefully - forcing shutdown")
		if closeErr := a.httpServer.Close(); closeErr != nil {
			logger.Error("failed to shutdown http server forcefully")
		}
	}
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

	cfg := SessionConfig{
		AgentID:       r.Header.Get(transport.HeaderKeyAgentID),
		Environment:   r.Header.Get(transport.HeaderKeyEnvironment),
		Organization:  r.Header.Get(transport.HeaderKeyOrganization),
		User:          r.Header.Get(transport.HeaderKeyUser),
		Subscriptions: strings.Split(r.Header.Get(transport.HeaderKeySubscriptions), ","),
	}

	cfg.Subscriptions = addEntitySubscription(cfg.AgentID, cfg.Subscriptions)

	session, err := NewSession(cfg, transport.NewTransport(conn), a.MessageBus, a.Store)
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
