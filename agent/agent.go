// Package agent is the running Sensu agent. Agents connect to a Sensu backend,
// register their presence, subscribe to check channels, download relevant
// check packages, execute checks, and send results to the Sensu backend via
// the Event channel.
package agent

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	time "github.com/echlebek/timeproxy"
	"github.com/gogo/protobuf/proto"

	"github.com/atlassian/gostatsd/pkg/statsd"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/agentd"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/handler"
	"github.com/sensu/sensu-go/system"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/util/retry"
	utilstrings "github.com/sensu/sensu-go/util/strings"
	"github.com/sirupsen/logrus"
)

// GetDefaultAgentName returns the default agent name
func GetDefaultAgentName() string {
	defaultAgentName, err := os.Hostname()
	if err != nil {
		logger.WithError(err).Error("error getting hostname")
		// TODO(greg): wat do?
		defaultAgentName = "unidentified-sensu-agent"
	}
	return defaultAgentName
}

// An Agent receives and acts on messages from a Sensu Backend.
type Agent struct {
	allowList       []allowList
	api             *http.Server
	assetGetter     asset.Getter
	backendSelector BackendSelector
	config          *Config
	connected       bool
	connectedMu     sync.RWMutex
	contentType     string
	entity          *corev2.Entity
	executor        command.Executor
	handler         *handler.MessageHandler
	header          http.Header
	inProgress      map[string]*corev2.CheckConfig
	inProgressMu    *sync.Mutex
	statsdServer    *statsd.Server
	sendq           chan *transport.Message
	systemInfo      *corev2.System
	systemInfoMu    sync.RWMutex
	wg              sync.WaitGroup
	apiQueue        queue
	marshal         agentd.MarshalFunc
	unmarshal       agentd.UnmarshalFunc
}

// NewAgent creates a new Agent. It returns non-nil error if there is any error
// when creating the Agent.
func NewAgent(config *Config) (*Agent, error) {
	agent := &Agent{
		backendSelector: &RandomBackendSelector{Backends: config.BackendURLs},
		connected:       false,
		config:          config,
		executor:        command.NewExecutor(),
		handler:         handler.NewMessageHandler(),
		inProgress:      make(map[string]*corev2.CheckConfig),
		inProgressMu:    &sync.Mutex{},
		sendq:           make(chan *transport.Message, 10),
		systemInfo:      &corev2.System{},
		unmarshal:       agentd.UnmarshalJSON,
		marshal:         agentd.MarshalJSON,
	}

	agent.statsdServer = NewStatsdServer(agent)
	agent.handler.AddHandler(corev2.CheckRequestType, agent.handleCheck)

	// We don't check for errors here and let the agent get created regardless
	// of system info status.
	_ = agent.refreshSystemInfo()
	var err error
	agent.apiQueue, err = newQueue(config.CacheDir)
	if err != nil {
		return nil, fmt.Errorf("error creating agent: %s", err)
	}

	allowList, err := readAllowList(config.AllowList, ioutil.ReadFile)
	if err != nil {
		return nil, err
	}
	agent.allowList = allowList

	return agent, nil
}

func (a *Agent) sendMessage(msg *transport.Message) {
	logger.WithFields(logrus.Fields{
		"type":         msg.Type,
		"content_type": a.contentType,
		"payload_size": len(msg.Payload),
	}).Info("sending message")
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

func (a *Agent) refreshSystemInfoPeriodically(ctx context.Context) {
	defer logger.Debug("shutting down system info collector")
	ticker := time.NewTicker(time.Duration(DefaultSystemInfoRefreshInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.refreshSystemInfo(); err != nil {
				logger.WithError(err).Error("failed to refresh system info")
			}
		case <-ctx.Done():
			return
		}
	}
}

func (a *Agent) buildTransportHeaderMap() http.Header {
	header := http.Header{}
	header.Set(transport.HeaderKeyNamespace, a.config.Namespace)
	header.Set(transport.HeaderKeyAgentName, a.config.AgentName)
	if tls := a.config.TLS; tls != nil && len(tls.CertFile) == 0 && len(tls.KeyFile) == 0 {
		logger.Info("using password auth")
		header.Set(transport.HeaderKeyUser, a.config.User)
		userCredentials := fmt.Sprintf("%s:%s", a.config.User, a.config.Password)
		userCredentials = base64.StdEncoding.EncodeToString([]byte(userCredentials))
		header.Set("Authorization", "Basic "+userCredentials)
	} else {
		logger.Info("using tls client auth")
	}
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
func (a *Agent) Run(ctx context.Context) error {
	defer func() {
		if err := a.apiQueue.Close(); err != nil {
			logger.WithError(err).Error("error closing API queue")
		}
	}()
	a.header = a.buildTransportHeaderMap()

	// Fail the agent after startup if the id is invalid
	if err := corev2.ValidateName(a.config.AgentName); err != nil {
		return fmt.Errorf("invalid agent name: %v", err)
	}
	if timeout := a.config.KeepaliveTimeout; timeout < 5 {
		return fmt.Errorf("bad keepalive timeout: %d (minimum value is 5 seconds)", timeout)
	}

	if !a.config.DisableAssets {
		assetManager := asset.NewManager(a.config.CacheDir, a.getAgentEntity(), &a.wg)
		var err error
		a.assetGetter, err = assetManager.StartAssetManager(ctx)
		if err != nil {
			return err
		}
	}

	// Start the statsd listener only if the agent configuration has it enabled
	if !a.config.StatsdServer.Disable {
		a.StartStatsd(ctx)
	}

	go a.connectionManager(ctx)
	go a.refreshSystemInfoPeriodically(ctx)
	go a.handleAPIQueue(ctx)

	a.wg.Wait()
	return nil
}

func (a *Agent) connectionManager(ctx context.Context) {
	defer logger.Debug("shutting down connection manager")
	for {
		a.connectedMu.Lock()
		a.connected = false
		a.connectedMu.Unlock()

		conn, err := a.connectWithBackoff(ctx)
		if err != nil {
			if err == ctx.Err() {
				return
			}
			log.Fatal(err)
		}

		ctx, cancel := context.WithCancel(ctx)

		// Start sending hearbeats to the backend
		conn.Heartbeat(ctx, a.config.BackendHeartbeatInterval, a.config.BackendHeartbeatTimeout)

		a.connectedMu.Lock()
		a.connected = true
		a.connectedMu.Unlock()

		go a.receiveLoop(ctx, cancel, conn)
		if err := a.sendLoop(ctx, cancel, conn); err != nil && err != ctx.Err() {
			logger.WithError(err).Error("error sending messages")
		}
	}
}

func (a *Agent) receiveLoop(ctx context.Context, cancel context.CancelFunc, conn transport.Transport) {
	defer cancel()
	for {
		m, err := conn.Receive()
		if err != nil {
			logger.WithError(err).Error("transport receive error")
			return
		}

		go func(msg *transport.Message) {
			logger.WithFields(logrus.Fields{
				"type":         msg.Type,
				"content_type": a.contentType,
				"payload_size": len(msg.Payload),
			}).Info("message received")
			err := a.handler.Handle(ctx, msg.Type, msg.Payload)
			if err != nil {
				logger.WithError(err).Error("error handling message")
			}
		}(m)
	}
}

func (a *Agent) sendLoop(ctx context.Context, cancel context.CancelFunc, conn transport.Transport) error {
	defer cancel()
	keepalive := time.NewTicker(time.Duration(a.config.KeepaliveInterval) * time.Second)
	defer keepalive.Stop()
	logger.Info("sending keepalive")
	if err := conn.Send(a.newKeepalive()); err != nil {
		logger.WithError(err).Error("error sending message over websocket")
		return err
	}
	for {
		select {
		case <-ctx.Done():
			if err := conn.Close(); err != nil {
				logger.WithError(err).Error("error closing websocket connection")
				return err
			}
			return nil
		case msg := <-a.sendq:
			logger.Info("sending event")
			if err := conn.Send(msg); err != nil {
				logger.WithError(err).Error("error sending message over websocket")
				return err
			}
		case <-keepalive.C:
			logger.Info("sending keepalive")
			if err := conn.Send(a.newKeepalive()); err != nil {
				logger.WithError(err).Error("error sending message over websocket")
				return err
			}
		}
	}
}

func (a *Agent) newKeepalive() *transport.Message {
	msg := &transport.Message{
		Type: transport.MessageTypeKeepalive,
	}
	entity := a.getAgentEntity()

	keepalive := &corev2.Event{
		ObjectMeta: corev2.NewObjectMeta("", entity.Namespace),
	}

	keepalive.Check = &corev2.Check{
		ObjectMeta: corev2.NewObjectMeta("keepalive", entity.Namespace),
		Interval:   a.config.KeepaliveInterval,
		Timeout:    a.config.KeepaliveTimeout,
	}
	keepalive.Entity = a.getAgentEntity()
	keepalive.Timestamp = time.Now().Unix()

	msgBytes, err := a.marshal(keepalive)
	if err != nil {
		// unlikely that this will ever happen
		logger.WithError(err).Error("error sending keepalive")
	}
	msg.Payload = msgBytes

	return msg
}

// Connected returns true if the agent is connected to a backend.
func (a *Agent) Connected() bool {
	a.connectedMu.RLock()
	defer a.connectedMu.RUnlock()
	return a.connected
}

// StartAPI starts the Agent HTTP API. After attempting to start the API, if the
// HTTP server encounters a fatal error, it will shutdown the rest of the agent.
func (a *Agent) StartAPI(ctx context.Context) {
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
		defer a.wg.Done()
		<-ctx.Done()
		logger.Info("API shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err := a.api.Shutdown(ctx); err != nil {
			logger.WithError(err).Error("error shutting down the API server")
		}
	}()
}

// StartSocketListeners starts the agent's TCP and UDP socket listeners.
// Agent TCP/UDP sockets are deprecated in favor of the agent rest api.
func (a *Agent) StartSocketListeners(ctx context.Context) {
	if _, _, err := a.createListenSockets(ctx); err != nil {
		logger.WithError(err).Error("unable to start socket listeners")
	}
}

// StartStatsd starts up a StatsD listener on the agent, logs an error for any
// failures.
func (a *Agent) StartStatsd(ctx context.Context) {
	logger.Info("starting statsd server on address: ", a.statsdServer.MetricsAddr)

	go func() {
		if err := a.statsdServer.Run(ctx); err != nil && err != ctx.Err() {
			logger.WithError(err).Errorf("error with statsd server on address: %s, statsd listener will not run", a.statsdServer.MetricsAddr)
		}
	}()
}

func (a *Agent) connectWithBackoff(ctx context.Context) (transport.Transport, error) {
	var conn transport.Transport

	backoff := retry.ExponentialBackoff{
		InitialDelayInterval: 10 * time.Millisecond,
		MaxDelayInterval:     10 * time.Second,
		Multiplier:           10,
		Ctx:                  ctx,
	}

	err := backoff.Retry(func(retry int) (bool, error) {
		url := a.backendSelector.Select()

		logger.Infof("connecting to backend URL %q", url)
		a.header.Set("Accept", agentd.ProtobufSerializationHeader)
		logger.WithField("header", fmt.Sprintf("Accept: %s", agentd.ProtobufSerializationHeader)).Debug("setting header")
		c, respHeader, err := transport.Connect(url, a.config.TLS, a.header, a.config.BackendHandshakeTimeout)
		if err != nil {
			logger.WithError(err).Error("reconnection attempt failed")
			return false, nil
		}

		logger.Info("successfully connected")

		conn = c

		logger.WithField("header", fmt.Sprintf("Accept: %s", respHeader["Accept"])).Debug("received header")
		if utilstrings.InArray(agentd.ProtobufSerializationHeader, respHeader["Accept"]) {
			a.contentType = agentd.ProtobufSerializationHeader
			a.unmarshal = proto.Unmarshal
			a.marshal = proto.Marshal
			logger.WithField("format", "protobuf").Debug("setting serialization/deserialization")
		} else {
			a.contentType = agentd.JSONSerializationHeader
			a.unmarshal = agentd.UnmarshalJSON
			a.marshal = agentd.MarshalJSON
			logger.WithField("format", "JSON").Debug("setting serialization/deserialization")
		}
		a.header.Set("Content-Type", a.contentType)
		logger.WithField("header", fmt.Sprintf("Content-Type: %s", a.contentType)).Debug("setting header")

		return true, nil
	})

	return conn, err
}
