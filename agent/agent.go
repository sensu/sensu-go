// Package agent is the running Sensu agent. Agents connect to a Sensu backend,
// register their presence, subscribe to check channels, download relevant
// check packages, execute checks, and send results to the Sensu backend via
// the Event channel.
package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	time "github.com/echlebek/timeproxy"

	"github.com/atlassian/gostatsd/pkg/statsd"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/handler"
	"github.com/sensu/sensu-go/system"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/retry"
	"github.com/sirupsen/logrus"
)

// GetDefaultAgentID returns the default agent ID
func GetDefaultAgentID() string {
	defaultAgentID, err := os.Hostname()
	if err != nil {
		logger.WithError(err).Error("error getting hostname")
		// TODO(greg): wat do?
		defaultAgentID = "unidentified-sensu-agent"
	}
	return defaultAgentID
}

// An Agent receives and acts on messages from a Sensu Backend.
type Agent struct {
	api             *http.Server
	assetGetter     asset.Getter
	backendSelector BackendSelector
	cancel          context.CancelFunc
	config          *Config
	connected       bool
	connectedMu     *sync.RWMutex
	context         context.Context
	entity          *types.Entity
	executor        command.Executor
	handler         *handler.MessageHandler
	header          http.Header
	inProgress      map[string]*types.CheckConfig
	inProgressMu    *sync.Mutex
	statsdServer    *statsd.Server
	sendq           chan *transport.Message
	stopping        chan struct{}
	systemInfo      *types.System
	systemInfoMu    *sync.RWMutex
	wg              *sync.WaitGroup
}

// NewAgent creates a new Agent and returns a pointer to it.
func NewAgent(config *Config) *Agent {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)
	agent := &Agent{
		backendSelector: &RandomBackendSelector{Backends: config.BackendURLs},
		cancel:          cancel,
		context:         ctx,
		connected:       false,
		connectedMu:     &sync.RWMutex{},
		config:          config,
		executor:        command.NewExecutor(),
		handler:         handler.NewMessageHandler(),
		inProgress:      make(map[string]*types.CheckConfig),
		inProgressMu:    &sync.Mutex{},
		stopping:        make(chan struct{}),
		sendq:           make(chan *transport.Message, 10),
		systemInfo:      &types.System{},
		systemInfoMu:    &sync.RWMutex{},
		wg:              &sync.WaitGroup{},
	}

	agent.statsdServer = NewStatsdServer(agent)
	agent.handler.AddHandler(types.CheckRequestType, agent.handleCheck)

	// We don't check for errors here and let the agent get created regardless
	// of system info status.
	_ = agent.refreshSystemInfo()
	return agent
}

func (a *Agent) sendMessage(msgType string, payload []byte) {
	logger.WithFields(logrus.Fields{
		"type":    msgType,
		"payload": string(payload),
	}).Debug("sending message")
	msg := &transport.Message{
		Type:    msgType,
		Payload: payload,
	}
	a.sendq <- msg
}

func (a *Agent) refreshSystemInfo() error {
	info, err := system.Info()
	if err != nil {
		return err
	}

	a.systemInfoMu.Lock()
	a.systemInfo = &info
	a.systemInfoMu.Unlock()

	return nil
}

func (a *Agent) refreshSystemInfoPeriodically() {
	ticker := time.NewTicker(time.Duration(DefaultSystemInfoRefreshInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.refreshSystemInfo(); err != nil {
				logger.WithError(err).Error("failed to refresh system info")
			}
		case <-a.stopping:
			return
		}
	}
}

func (a *Agent) sendKeepalive() error {
	logger.Info("sending keepalive")
	msg := &transport.Message{
		Type: transport.MessageTypeKeepalive,
	}
	keepalive := &types.Event{}

	entity := a.getAgentEntity()

	keepalive.Check = &types.Check{
		ObjectMeta: types.ObjectMeta{
			Name:      "keepalive",
			Namespace: entity.Namespace,
		},
		Interval: a.config.KeepaliveInterval,
		Timeout:  a.config.KeepaliveTimeout,
	}
	keepalive.Entity = a.getAgentEntity()
	keepalive.Timestamp = time.Now().Unix()

	msgBytes, err := json.Marshal(keepalive)
	if err != nil {
		return err
	}
	msg.Payload = msgBytes
	a.sendq <- msg

	return nil
}

func (a *Agent) sendKeepalivePeriodically() {
	ticker := time.NewTicker(time.Duration(a.config.KeepaliveInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.sendKeepalive(); err != nil {
				logger.WithError(err).Error("failed sending keepalive")
			}
		case <-a.stopping:
			return
		}
	}
}

func (a *Agent) buildTransportHeaderMap() http.Header {
	header := http.Header{}
	header.Set(transport.HeaderKeyAgentID, a.config.AgentID)
	header.Set(transport.HeaderKeyNamespace, a.config.Namespace)
	header.Set(transport.HeaderKeyUser, a.config.User)
	header.Set(transport.HeaderKeySubscriptions, strings.Join(a.config.Subscriptions, ","))

	return header
}

// Run starts the Agent.
//
// 1. Start the asset manager.
// 2. Start a statsd server on the agent and logs the received metrics.
// 3. Connect to the backend, return an error if unsuccessful.
// 4. Start the socket listeners, return an error if unsuccessful.
// 5. Start the send/receive pumps.
// 6. Issue a keepalive immediately.
// 7. Start refreshing system info periodically.
// 8. Start sending periodic keepalives.
// 9. Start the API server, shutdown the agent if doing so fails.
func (a *Agent) Run() error {
	var err error
	assetManager := asset.NewManager(a.config.CacheDir, a.getAgentEntity(), a.stopping, a.wg)
	a.assetGetter, err = assetManager.StartAssetManager()
	if err != nil {
		return err
	}

	userCredentials := fmt.Sprintf("%s:%s", a.config.User, a.config.Password)
	userCredentials = base64.StdEncoding.EncodeToString([]byte(userCredentials))
	a.header = a.buildTransportHeaderMap()
	a.header.Set("Authorization", "Basic "+userCredentials)

	// Fail the agent after startup if the id is invalid
	if err := types.ValidateName(a.config.AgentID); err != nil {
		return fmt.Errorf("invalid agent id: %v", err)
	}

	// Start the statsd listener only if the agent configuration has it enabled
	if !a.config.StatsdServer.Disable {
		a.StartStatsd()
	}

	go a.connectionManager()
	go a.refreshSystemInfoPeriodically()
	go a.sendKeepalivePeriodically()

	return nil
}

func (a *Agent) connectionManager() {
	for {
		select {
		case <-a.stopping:
			return
		default:
		}

		a.connectedMu.Lock()
		a.connected = false
		a.connectedMu.Unlock()

		conn := connectWithBackoff(a.backendSelector.Select(), a.config.TLS, a.header)

		a.connectedMu.Lock()
		a.connected = true
		a.connectedMu.Unlock()

		// Send an immediate keepalive once we've connected.
		if err := a.sendKeepalive(); err != nil {
			logger.WithError(err).Error("error sending keepalive")
		}

		done := make(chan struct{})

		go receiveLoop(conn, a.handler, done)
		sendLoop(conn, a.sendq, done, a.stopping)
	}
}

func receiveLoop(conn transport.Transport, handler *handler.MessageHandler, done chan struct{}) {
	defer close(done)

	for {
		m, err := conn.Receive()
		if err != nil {
			logger.WithError(err).Error("transport receive error")
			return
		}

		go func(msg *transport.Message) {
			logger.WithFields(logrus.Fields{
				"type":    msg.Type,
				"payload": string(msg.Payload),
			}).Info("message received")
			err := handler.Handle(msg.Type, msg.Payload)
			if err != nil {
				logger.WithError(err).Error("error handling message")
			}
		}(m)
	}
}

func sendLoop(conn transport.Transport, sendq chan *transport.Message, done, stopping chan struct{}) {
	for {
		select {
		case <-done:
			return
		case <-stopping:
			if err := conn.Close(); err != nil {
				logger.WithError(err).Error("error closing websocket connection")
			}
			return
		case msg := <-sendq:
			if err := conn.Send(msg); err != nil {
				logger.WithError(err).Error("error sending message over websocket")
				return
			}
		}
	}
}

// Connected returns true if the agent is connected to a backend.
func (a *Agent) Connected() bool {
	a.connectedMu.RLock()
	defer a.connectedMu.RUnlock()
	return a.connected
}

// StartAPI starts the Agent HTTP API. After attempting to start the API, if the
// HTTP server encounters a fatal error, it will shutdown the rest of the agent.
func (a *Agent) StartAPI() {
	// Prepare the HTTP API server
	a.api = newServer(a)

	// Start the HTTP API server
	go func() {
		logger.Info("starting api on address: ", a.api.Addr)

		if err := a.api.ListenAndServe(); err != http.ErrServerClosed {
			logger.WithError(err).Fatal("the agent API has crashed")
		}
	}()

	// Allow Stop() to block until the HTTP server shuts down.
	a.wg.Add(1)

	go func() {
		// NOTE: This does not guarantee a clean shutdown of the HTTP API.
		// This is _only_ for the purpose of making Stop() a blocking call.
		// The goroutine running the HTTP Server has to return before Stop()
		// can return, so we use this to signal that goroutine to shutdown.
		<-a.stopping
		logger.Info("API shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err := a.api.Shutdown(ctx); err != nil {
			logger.WithError(err).Error("error shutting down the API server")
		}

		a.wg.Done()
	}()
}

// StartSocketListeners starts the agent's TCP and UDP socket listeners.
func (a *Agent) StartSocketListeners() {
	if _, _, err := a.createListenSockets(); err != nil {
		logger.WithError(err).Error("unable to start socket listeners")
	}
}

// Stop shuts down the agent. It will block until all listening goroutines
// have returned.
func (a *Agent) Stop() {
	a.cancel()
	close(a.stopping)
	a.wg.Wait()
}

// StartStatsd starts up a StatsD listener on the agent, logs an error for any
// failures.
func (a *Agent) StartStatsd() {
	logger.Info("starting statsd server on address: ", a.statsdServer.MetricsAddr)

	go func() {
		if err := a.statsdServer.Run(a.context); err != nil {
			logger.WithError(err).Errorf("error with statsd server on address: %s, statsd listener will not run", a.statsdServer.MetricsAddr)
		}
	}()
}

func connectWithBackoff(url string, tlsOpts *types.TLSOptions, header http.Header) transport.Transport {
	var conn transport.Transport

	backoff := retry.ExponentialBackoff{
		InitialDelayInterval: 500 * time.Millisecond,
		MaxDelayInterval:     10 * time.Second,
		MaxRetryAttempts:     0, // Unlimited attempts
		Multiplier:           1.5,
	}

	if err := backoff.Retry(func(retry int) (bool, error) {
		c, err := transport.Connect(url, tlsOpts, header)
		if err != nil {
			logger.WithError(err).Error("reconnection attempt failed")
			return false, nil
		}

		logger.Info("successfully connected")

		conn = c

		return true, nil
	}); err != nil {
		logger.WithError(err).Fatal("could not connect to transport")
	}

	return conn
}
