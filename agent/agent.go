// Package agent is the running Sensu agent. Agents connect to a Sensu backend,
// register their presence, subscribe to check channels, download relevant
// check packages, execute checks, and send results to the Sensu backend via
// the Event channel.
package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/handler"
	"github.com/sensu/sensu-go/system"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
)

const (
	// MaxMessageBufferSize specifies the maximum number of messages of a given
	// type that an agent will queue before rejecting messages.
	MaxMessageBufferSize = 10
)

// A Config specifies Agent configuration.
type Config struct {
	// AgentID is the entity ID for the running agent. Default is hostname.
	AgentID string
	// BackendURLs is a list of URLs for the Sensu Backend. Default: ws://127.0.0.1:8080
	BackendURLs []string
	// Subscriptions is an array of subscription names. Default: empty array.
	Subscriptions []string
	// KeepaliveInterval is the interval, in seconds, when agents will send a
	// keepalive to sensu-backend. Default: 60
	KeepaliveInterval int
	// Deregister indicates whether the entity is ephemeral
	Deregister bool
	// DeregistrationHandler specifies a single deregistration handler
	DeregistrationHandler string
	// CacheDir path where cached data is stored
	CacheDir string
}

var logger *logrus.Entry

func init() {
	logger = logrus.WithFields(logrus.Fields{
		"component": "agent",
	})
}

// NewConfig provides a new Config object initialized with defaults.
func NewConfig() *Config {
	c := &Config{
		BackendURLs:       []string{"ws://127.0.0.1:8081"},
		Subscriptions:     []string{},
		KeepaliveInterval: 20,
		CacheDir:          "/var/cache/sensu",
	}
	hostname, err := os.Hostname()
	if err != nil {
		logger.Error("error getting hostname: ", err.Error())
		// TODO(greg): wat do?
		c.AgentID = "unidentified-sensu-agent"
	}
	c.AgentID = hostname

	return c
}

// An Agent receives and acts on messages from a Sensu Backend.
type Agent struct {
	config          *Config
	backendSelector BackendSelector
	handler         *handler.MessageHandler
	conn            transport.Transport
	sendq           chan *transport.Message
	stopping        chan struct{}
	stopped         chan struct{}
	entity          *types.Entity
	assetManager    *AssetManager
}

// NewAgent creates a new Agent and returns a pointer to it.
func NewAgent(config *Config) *Agent {
	agent := &Agent{
		config:          config,
		backendSelector: &RandomBackendSelector{Backends: config.BackendURLs},
		handler:         handler.NewMessageHandler(),
		stopping:        make(chan struct{}),
		stopped:         make(chan struct{}),
		sendq:           make(chan *transport.Message, 10),
	}

	agent.handler.AddHandler(types.CheckConfigType, agent.handleCheck)
	agent.assetManager = NewAssetManager(config.CacheDir)

	return agent
}

func (a *Agent) receiveMessages(out chan *transport.Message) {
	defer close(out)
	for {
		m, err := a.conn.Receive()
		if err != nil {
			switch err := err.(type) {
			case transport.ConnectionError, transport.ClosedError:
				logger.Error("recv error: ", err.Error())
				return
			default:
				logger.Error("recv error: ", err.Error())
				continue
			}
		}
		out <- m
	}
}

func (a *Agent) receivePump(wg *sync.WaitGroup, conn transport.Transport) {
	defer func() {
		wg.Done()
		logger.Info("recv pump shutting down")
	}()

	logger.Info("connected - starting receivePump")

	recvChan := make(chan *transport.Message)
	go a.receiveMessages(recvChan)

	for {
		select {
		case <-a.stopping:
			return
		case msg, ok := <-recvChan:
			if !ok {
				return
			}

			logger.Info("message received - type: ", msg.Type, " message: ", string(msg.Payload))
			err := a.handler.Handle(msg.Type, msg.Payload)
			if err != nil {
				logger.Error("error handling message:", err.Error())
			}
		}
	}
}

func (a *Agent) sendMessage(msgType string, payload []byte) {
	// blocks until message can be enqueued.
	// TODO(greg): ring buffer?
	msg := &transport.Message{
		Type:    msgType,
		Payload: payload,
	}
	a.sendq <- msg
}

func (a *Agent) sendPump(wg *sync.WaitGroup, conn transport.Transport) {
	defer func() {
		wg.Done()
		logger.Info("sendpump shutting down")
	}()

	// The sendPump is actually responsible for shutting down the transport
	// to prevent a race condition between it and something else trying
	// to close the transport (which actually causes a write to the websocket
	// connection.)
	defer a.conn.Close()

	logger.Info("connected - starting sendPump")
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	keepaliveTicker := time.NewTicker(time.Duration(a.config.KeepaliveInterval) * time.Second)

	for {
		select {
		case <-keepaliveTicker.C:
			a.sendKeepalive()
		default:
		}

		select {
		case msg := <-a.sendq:
			err := conn.Send(msg)
			if err != nil {
				switch err := err.(type) {
				case transport.ConnectionError, transport.ClosedError:
					logger.Error("send error: ", err.Error())
					return
				default:
					logger.Error("send error:", err.Error())
				}
			}
		case <-a.stopping:
			return
		default:
			if a.conn.Closed() {
				return
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func (a *Agent) sendKeepalive() error {
	logger.Info("sending keepalive")
	msg := &transport.Message{
		Type: types.KeepaliveType,
	}
	keepalive := &types.Event{}
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

func (a *Agent) getAgentEntity() *types.Entity {
	if a.entity == nil {
		e := &types.Entity{
			ID:            a.config.AgentID,
			Class:         "agent",
			Subscriptions: a.config.Subscriptions,
			Deregister:    a.config.Deregister,
		}

		if a.config.DeregistrationHandler != "" {
			e.Deregistration = types.Deregistration{
				Handler: a.config.DeregistrationHandler,
			}
		}

		s, err := system.Info()
		if err == nil {
			e.System = s
		}

		a.entity = e
	}

	return a.entity
}

func (a *Agent) handshake() error {
	handshake := &types.AgentHandshake{
		ID:            a.config.AgentID,
		Subscriptions: a.config.Subscriptions,
	}
	msgBytes, err := json.Marshal(handshake)
	if err != nil {
		return err
	}

	// shoot first, ask questions later.
	agentHandshakeMsg := &transport.Message{
		Type:    types.AgentHandshakeType,
		Payload: msgBytes,
	}
	err = a.conn.Send(agentHandshakeMsg)
	if err != nil {
		return err
	}

	m, err := a.conn.Receive()
	if err != nil {
		return err
	}

	if m.Type != types.BackendHandshakeType {
		return errors.New("backend did not send handshake")
	}

	response := types.BackendHandshake{}
	err = json.Unmarshal(m.Payload, &response)
	if err != nil {
		return fmt.Errorf("error unmarshaling backend handshake: %s", err.Error())
	}

	err = a.sendKeepalive()
	if err != nil {
		return err
	}

	return nil
}

// Run starts the Agent's connection manager which handles connecting and
// reconnecting to the Sensu Backend. It also handles coordination of the
// agent's read and write pumps.
//
// If Run cannot establish an initial connection to the specified Backend
// URL, Run will return an error.
func (a *Agent) Run() error {
	// TODO(greg): this whole thing reeks. i want to be able to return an error
	// if we can't connect, but maybe we do the channel w/ terminal errors thing
	// here as well. yeah. i think we should do that instead.
	conn, err := transport.Connect(a.backendSelector.Select())
	if err != nil {
		return err
	}
	a.conn = conn
	err = a.handshake()
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go a.sendPump(wg, conn)
	go a.receivePump(wg, conn)

	pumpsReturned := make(chan struct{})
	go func() {
		wg.Wait()
		close(pumpsReturned)
	}()

	go func(wg *sync.WaitGroup) {
		retries := 0
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-a.stopping:
				wg.Wait()
				close(a.sendq)
				close(a.stopped)
				return

			case <-pumpsReturned:
				nextBackend := a.backendSelector.Select()
				logger.Info("disconnected - attempting to reconnect: ", nextBackend)
				conn, err := transport.Connect(nextBackend)
				if err != nil {
					logger.Error("connection error:", err.Error())
					// TODO(greg): exponential backoff
					time.Sleep(1 * time.Second)
					retries++
					continue
				}

				a.conn = conn
				logger.Info("reconnected: ", nextBackend)

				err = a.handshake()
				if err != nil {
					logger.Error("handshake error: ", err.Error())
					continue
				}

				wg.Add(2)
				go a.sendPump(wg, conn)
				go a.receivePump(wg, conn)
				pumpsReturned = make(chan struct{})
				go func() {
					wg.Wait()
					close(pumpsReturned)
				}()
			}
		}
	}(wg)

	return nil
}

// Stop will cause the Agent to finish processing requests and then cleanly
// shutdown.
func (a *Agent) Stop() {
	close(a.stopping)
	select {
	case <-a.stopped:
		return
	case <-time.After(1 * time.Second):
		return
	}
}

func (a *Agent) addHandler(msgType string, handlerFunc handler.MessageHandlerFunc) {
	a.handler.AddHandler(msgType, handlerFunc)
}
