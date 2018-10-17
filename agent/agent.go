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
	"github.com/sensu/sensu-go/agent/assetmanager"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/handler"
	"github.com/sensu/sensu-go/system"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/path"
	"github.com/sensu/sensu-go/util/retry"
	"github.com/sirupsen/logrus"
)

const (
	// MaxMessageBufferSize specifies the maximum number of messages of a given
	// type that an agent will queue before rejecting messages.
	MaxMessageBufferSize = 10

	// TCPSocketReadDeadline specifies the maximum time the TCP socket will wait
	// to receive data.
	TCPSocketReadDeadline = 500 * time.Millisecond

	// DefaultAPIHost specifies the default API Host
	DefaultAPIHost = "127.0.0.1"
	// DefaultAPIPort specifies the default API Port
	DefaultAPIPort = 3031
	// DefaultBackendURL specifies the default backend URL
	DefaultBackendURL = "ws://127.0.0.1:8081"
	// DefaultKeepaliveInterval specifies the default keepalive interval
	DefaultKeepaliveInterval = 20
	// DefaultNamespace specifies the default namespace
	DefaultNamespace = "default"
	// DefaultPassword specifies the default password
	DefaultPassword = "P@ssw0rd!"
	// DefaultSocketHost specifies the default socket host
	DefaultSocketHost = "127.0.0.1"
	// DefaultSocketPort specifies the default socket port
	DefaultSocketPort = 3030
	// DefaultStatsdDisable specifies if the statsd listener is disabled
	DefaultStatsdDisable = false
	// DefaultStatsdFlushInterval specifies the default flush interval for statsd
	DefaultStatsdFlushInterval = 10
	// DefaultStatsdMetricsHost specifies the default metrics host for statsd server
	DefaultStatsdMetricsHost = "127.0.0.1"
	// DefaultStatsdMetricsPort specifies the default metrics port for statsd server
	DefaultStatsdMetricsPort = 8125
	// DefaultSystemInfoRefreshInterval specifies the default refresh interval
	// (in seconds) for the agent's cached system information.
	DefaultSystemInfoRefreshInterval = 20
	// DefaultUser specifies the default user
	DefaultUser = "agent"
)

// A Config specifies Agent configuration.
type Config struct {
	// AgentID is the entity ID for the running agent. Default is hostname.
	AgentID string
	// API contains the Sensu client HTTP API configuration
	API *APIConfig
	// BackendURLs is a list of URLs for the Sensu Backend. Default:
	// ws://127.0.0.1:8081
	BackendURLs []string
	// CacheDir path where cached data is stored
	CacheDir string
	// Deregister indicates whether the entity is ephemeral
	Deregister bool
	// DeregistrationHandler specifies a single deregistration handler
	DeregistrationHandler string
	// ExtendedAttributes contains any extended attributes passed to the agent on
	// start
	ExtendedAttributes []byte
	// KeepaliveInterval is the interval, in seconds, when agents will send a
	// keepalive to sensu-backend.
	KeepaliveInterval uint32
	// KeepaliveTimeout is the time after which a sensu-agent is considered dead
	// by the backend. See DefaultKeepaliveTimeout in types package for default
	// value.
	KeepaliveTimeout uint32
	// Namespace sets the Agent's RBAC namespace identifier
	Namespace string
	// Password sets Agent's password
	Password string
	// Redact contains the fields to redact when marshalling the agent's entity
	Redact []string
	// Socket contains the Sensu client socket configuration
	Socket *SocketConfig
	// StatsdServer contains the statsd server configuration
	StatsdServer *StatsdServerConfig
	// Subscriptions is an array of subscription names. Default: empty array.
	Subscriptions []string
	// TLS sets the TLSConfig for agent TLS options
	TLS *types.TLSOptions
	// User sets the Agent's username
	User string
}

// StatsdServerConfig contains the statsd server configuration
type StatsdServerConfig struct {
	Host          string
	Port          int
	FlushInterval int
	Handlers      []string
	Disable       bool
}

// SocketConfig contains the Socket configuration
type SocketConfig struct {
	Host string
	Port int
}

// FixtureConfig provides a new Config object initialized with defaults for use
// in tests
func FixtureConfig() *Config {
	c := &Config{
		AgentID: GetDefaultAgentID(),
		API: &APIConfig{
			Host: DefaultAPIHost,
			Port: DefaultAPIPort,
		},
		BackendURLs:       []string{},
		CacheDir:          path.SystemCacheDir("sensu-agent"),
		KeepaliveInterval: DefaultKeepaliveInterval,
		KeepaliveTimeout:  types.DefaultKeepaliveTimeout,
		Namespace:         DefaultNamespace,
		Password:          DefaultPassword,
		Socket: &SocketConfig{
			Host: DefaultSocketHost,
			Port: DefaultSocketPort,
		},
		StatsdServer: &StatsdServerConfig{
			Host:          DefaultStatsdMetricsHost,
			Port:          DefaultStatsdMetricsPort,
			FlushInterval: DefaultStatsdFlushInterval,
			Handlers:      []string{},
			Disable:       DefaultStatsdDisable,
		},
		User: DefaultUser,
	}
	return c
}

// NewConfig provides a new empty Config object
func NewConfig() *Config {
	c := &Config{
		API:          &APIConfig{},
		Socket:       &SocketConfig{},
		StatsdServer: &StatsdServerConfig{},
	}
	return c
}

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
	assetManager    *assetmanager.Manager
	backendSelector BackendSelector
	cancel          context.CancelFunc
	config          *Config
	conn            transport.Transport
	context         context.Context
	entity          *types.Entity
	executor        command.Executor
	handler         *handler.MessageHandler
	header          http.Header
	inProgress      map[string]*types.CheckConfig
	inProgressMu    *sync.Mutex
	statsdServer    *statsd.Server
	sendq           chan *transport.Message
	stopped         chan struct{}
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
		config:          config,
		handler:         handler.NewMessageHandler(),
		inProgress:      make(map[string]*types.CheckConfig),
		inProgressMu:    &sync.Mutex{},
		executor:        command.NewExecutor(),
		stopping:        make(chan struct{}),
		stopped:         make(chan struct{}),
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
	agent.assetManager = assetmanager.New(config.CacheDir, agent.getAgentEntity())

	return agent
}

func (a *Agent) receiveMessages(out chan *transport.Message) {
	defer close(out)
	for {
		m, err := a.conn.Receive()
		if err != nil {
			logger.WithError(err).Error("transport receive error")

			// If we encountered a connection error, try to reconnect
			if _, ok := err.(transport.ConnectionError); ok {
				// The first step is to close the current websocket connection, which is
				// no longer useful
				if err := a.conn.Close(); err != nil {
					logger.Debug(err)
				}

				// Now, we must attempt to reconnect to the backend, with exponential
				// backoff
				backoff := retry.ExponentialBackoff{
					InitialDelayInterval: 500 * time.Millisecond,
					MaxDelayInterval:     10 * time.Second,
					MaxRetryAttempts:     0, // Unlimited attempts
					Multiplier:           1.5,
				}
				if err := backoff.Retry(func(retry int) (bool, error) {
					if err = a.conn.Reconnect(a.backendSelector.Select(), a.config.TLS, a.header); err != nil {
						logger.WithError(err).Error("reconnection attempt failed")
						return false, nil
					}

					// At this point, the attempt was successful
					logger.Info("successfully reconnected")
					return true, nil
				}); err != nil {
					logger.WithError(err).Fatal("could not reconnect to transport")
				}
			}

		}
		out <- m
	}
}

func (a *Agent) receivePump() {
	logger.Info("connected - starting receivePump")

	recvChan := make(chan *transport.Message)
	go a.receiveMessages(recvChan)

	for {
		select {
		case <-a.stopping:
			return
		case msg, ok := <-recvChan:
			if msg == nil || !ok {
				continue
			}

			logger.WithFields(logrus.Fields{
				"type":    msg.Type,
				"payload": string(msg.Payload),
			}).Info("message received")
			err := a.handler.Handle(msg.Type, msg.Payload)
			if err != nil {
				logger.WithError(err).Error("error handling message")
			}
		}
	}
}

func (a *Agent) sendMessage(msgType string, payload []byte) {
	logger.WithFields(logrus.Fields{
		"type":    msgType,
		"payload": string(payload),
	}).Debug("sending message")
	// blocks until message can be enqueued.
	// TODO(greg): ring buffer?
	msg := &transport.Message{
		Type:    msgType,
		Payload: payload,
	}
	a.sendq <- msg
}

func (a *Agent) sendPump() {
	// The sendPump is actually responsible for shutting down the transport
	// to prevent a race condition between it and something else trying
	// to close the transport (which actually causes a write to the websocket
	// connection.)
	defer func() {
		if err := a.conn.Close(); err != nil {
			logger.Debug(err)
		}
	}()

	logger.Info("connected - starting sendPump")

	for {
		select {
		case msg := <-a.sendq:
			if err := a.conn.Send(msg); err != nil {
				logger.WithError(err).Warning("transport send error")
			}
		case <-a.stopping:
			return
		}
	}
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

	keepalive.Check = &types.Check{
		Name:     "keepalive",
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
// 1. Start a statsd server on the agent and logs the received metrics.
// 2. Connect to the backend, return an error if unsuccessful.
// 3. Start the socket listeners, return an error if unsuccessful.
// 4. Start the send/receive pumps.
// 5. Issue a keepalive immediately.
// 6. Start refreshing system info periodically.
// 7. Start sending periodic keepalives.
// 8. Start the API server, shutdown the agent if doing so fails.
func (a *Agent) Run() error {
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

	conn, err := transport.Connect(a.backendSelector.Select(), a.config.TLS, a.header)
	if err != nil {
		return err
	}

	a.conn = conn

	// These are in separate goroutines so that they can, theoretically, be executing
	// concurrently.
	go a.sendPump()
	go a.receivePump()

	// Send an immediate keepalive once we've connected.
	if err := a.sendKeepalive(); err != nil {
		logger.WithError(err).Error("error sending keepalive")
	}

	go a.refreshSystemInfoPeriodically()
	go a.sendKeepalivePeriodically()

	return nil
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
