// Package agent is the running Sensu agent. Agents connect to a Sensu backend,
// register their presence, subscribe to check channels, download relevant
// check packages, execute checks, and send results to the Sensu backend via
// the Event channel.
package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sensu/sensu-go/agent/assetmanager"
	"github.com/sensu/sensu-go/handler"
	"github.com/sensu/sensu-go/system"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
)

const (
	// MaxMessageBufferSize specifies the maximum number of messages of a given
	// type that an agent will queue before rejecting messages.
	MaxMessageBufferSize = 10
	ListenPort           = ":3030"
)

// A Config specifies Agent configuration.
type Config struct {
	// AgentID is the entity ID for the running agent. Default is hostname.
	AgentID string
	// BackendURLs is a list of URLs for the Sensu Backend. Default: ws://127.0.0.1:8081
	BackendURLs []string
	// Subscriptions is an array of subscription names. Default: empty array.
	Subscriptions []string
	// KeepaliveInterval is the interval, in seconds, when agents will send a
	// keepalive to sensu-backend. Default: 60
	KeepaliveInterval int
	// KeepaliveTimeout is the time after which a sensu-agent is considered dead
	// back the backend.
	KeepaliveTimeout uint
	// Deregister indicates whether the entity is ephemeral
	Deregister bool
	// DeregistrationHandler specifies a single deregistration handler
	DeregistrationHandler string
	// CacheDir path where cached data is stored
	CacheDir string
	// Environment sets the Agent's RBAC environment identifier
	Environment string
	// Organization sets the Agent's RBAC organization identifier
	Organization string
	// User sets the Agent's username
	User string
	// Password sets Agent's password
	Password string
	// TLS sets the TLSConfig for agent TLS options
	TLS *types.TLSOptions
	// API contains the HTTP API configuration
	API *APIConfig
}

// NewConfig provides a new Config object initialized with defaults.
func NewConfig() *Config {
	c := &Config{
		BackendURLs:       []string{"ws://127.0.0.1:8081"},
		Subscriptions:     []string{},
		KeepaliveInterval: 20,
		KeepaliveTimeout:  120,
		CacheDir:          "/var/cache/sensu",
		Environment:       "default",
		Organization:      "default",
		User:              "agent",
		Password:          "P@ssw0rd!",
		API: &APIConfig{
			Host: "127.0.0.1",
			Port: 3031,
		},
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
	assetManager    *assetmanager.Manager
	api             *http.Server
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

	agent.handler.AddHandler(types.CheckRequestType, agent.handleCheck)
	agent.assetManager = assetmanager.New(config.CacheDir, agent.getAgentEntity())

	return agent
}

func (a *Agent) tcpSocket() error {
	logger.Debug("starting TCP server")
	listen, err := net.Listen("tcp", ListenPort)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-a.stopped:
				listen.Close()
				return
			default:
				conn, err := listen.Accept()
				logger.Debug("accepted new TCP connection")
				if err != nil {
					logger.Error(err)
					return
				}
				go a.handleTCPMessages(conn)
			}
		}
	}()
	return err
}

func (a *Agent) udpSocket() error {
	logger.Debug("starting UDP socket")
	UDPServerAddr, err := net.ResolveUDPAddr("udp", ListenPort)
	if err != nil {
		return err
	}

	listen, err := net.ListenUDP("udp", UDPServerAddr)
	if err == nil {
		go a.handleUDPMessages(listen)
	}

	return err
}

func (a *Agent) handleTCPMessages(c net.Conn) {
	// check for entity - if no entity present, use agent's entity
	// send data: call a.sendMessage(msgType string, payload []byte)
	defer c.Close()
	var buf [1500]byte

	for {
		select {
		case <-a.stopped:
			return
		default:
			readLen, err := c.Read(buf[0:])
			if err == io.EOF {
				continue
			} else if err != nil {
				logger.Error(err)
			}

			if err := a.parseBuffer(buf, readLen); err != nil {
				logger.Errorf("Invalid event data: %s", err)
				c.Write([]byte("invalid event data\n"))
			} else {
				logger.Info("Received message on TCP socket")
				a.sendMessage(types.EventType, buf[:readLen])
			}
		}
	}
}

func (a *Agent) handleUDPMessages(c net.PacketConn) {
	defer c.Close()
	var buf [1500]byte

	for {
		select {
		case <-a.stopped:
			return
		default:
			readLen, _, err := c.ReadFrom(buf[0:])
			if err != nil {
				logger.Error(err)
				continue
			}

			if err := a.parseBuffer(buf, readLen); err != nil {
				logger.Errorf("UDP Invalid event data: %s", err)
			} else {
				a.sendMessage(types.EventType, buf[:readLen])
			}
		}
	}
}

// check for presense of entity here and add if it doesn't exist in data
func (a *Agent) parseBuffer(buffer [1500]byte, length int) (err error) {
	readString := buffer[:length]

	var event map[string]interface{}
	if err = json.Unmarshal(readString, &event); err != nil {
		return err
	}
	return nil
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

	for {
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
			ID:               a.config.AgentID,
			Class:            "agent",
			Subscriptions:    a.config.Subscriptions,
			Deregister:       a.config.Deregister,
			KeepaliveTimeout: a.config.KeepaliveTimeout,
			Environment:      a.config.Environment,
			Organization:     a.config.Organization,
			User:             a.config.User,
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
		Environment:   a.config.Environment,
		Organization:  a.config.Organization,
		User:          a.config.User,
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
	userCredentials := fmt.Sprintf("%s:%s", a.config.User, a.config.Password)
	userCredentials = base64.StdEncoding.EncodeToString([]byte(userCredentials))
	header := http.Header{"Authorization": {"Basic " + userCredentials}}

	conn, err := transport.Connect(a.backendSelector.Select(), a.config.TLS, header)
	if err != nil {
		return err
	}
	a.conn = conn
	err = a.handshake()
	if err != nil {
		return err
	}

	if err := a.udpSocket(); err != nil {
		return err
	}

	err = a.tcpSocket()
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

	go func() {
		keepaliveTicker := time.NewTicker(time.Duration(a.config.KeepaliveInterval) * time.Second)
		for {
			select {
			case <-keepaliveTicker.C:
				a.sendKeepalive()
			case <-a.stopping:
				return
			}

		}
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
				conn, err := transport.Connect(nextBackend, a.config.TLS, header)
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

	// Start the API
	go func() {
		a.api = newServer(a)
		logger.Info("starting api on address: ", a.api.Addr)

		if err := a.api.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}()

	// Gracefully shutdown the API when the agent stops
	go func() {
		<-a.stopping
		logger.Info("api shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err := a.api.Shutdown(ctx); err != nil {
			logger.Fatal(err)
		}
	}()

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
