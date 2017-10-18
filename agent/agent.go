// Package agent is the running Sensu agent. Agents connect to a Sensu backend,
// register their presence, subscribe to check channels, download relevant
// check packages, execute checks, and send results to the Sensu backend via
// the Event channel.
package agent

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
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
	// API contains the Sensu client HTTP API configuration
	API *APIConfig
	// Socket contains the Sensu client socket configuration
	Socket *SocketConfig
}

// SocketConfig contains the Socket configuration
type SocketConfig struct {
	Host string
	Port int
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
		Socket: &SocketConfig{
			Host: "127.0.0.1",
			Port: 3030,
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

// createListenSockets UDP and TCP socket listeners on port 3030 for external check
// events.
func (a *Agent) createListenSockets() error {
	addr := fmt.Sprintf("%s:%d", a.config.Socket.Host, a.config.Socket.Port)

	// Setup UDP socket listener
	UDPServerAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	udpListen, err := net.ListenUDP("udp", UDPServerAddr)
	if err != nil {
		return err
	}
	logger.Infof("starting UDP listener on %s", addr)
	go a.handleUDPMessages(udpListen)

	// Setup TCP socket listener
	TCPServerAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	logger.Infof("starting TCP listener on %s", addr)
	tcpListen, err := net.ListenTCP("tcp", TCPServerAddr)
	if err != nil {
		return err
	}

	go func() {
		defer tcpListen.Close()

		for {
			conn, err := tcpListen.Accept()
			select {
			case <-a.stopping:
				return
			default:
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

// Streams can be of any length. The socket protocol does not require
// any headers, instead the socket tries to parse everything it has
// been sent each time a chunk of data arrives. Once the JSON parses
// successfully, the Sensu agent publishes the result. After
// timeout (default is 500 msec) since the most recent chunk
// of data was received, the agent will give up on the sender, and
// instead respond "invalid" and close the connection.
func (a *Agent) handleTCPMessages(c net.Conn) {
	defer c.Close()
	var buf []byte
	messageBuffer := bytes.NewBuffer(buf)
	connReader := bufio.NewReader(c)
	waitTime := 10 * time.Millisecond

	// Read incoming tcp messages from client until we hit a valid JSON message.
	// If we don't get valid JSON or a ping request after 500ms, close the
	// connection (timeout).
	for i := 0; i < 50; i++ {
		// Read for up to 10ms on each loop iteration
		err := c.SetReadDeadline(time.Now().Add(waitTime))
		if err != nil {
			logger.Debugf("Error setting readDeadline: %s", err)
		}
		_, err = connReader.WriteTo(messageBuffer)
		// Check error condition. If it's a timeout error, continue so we can read
		// any remaining partial packets. Any other error type returns.
		if err != nil {
			if opError, ok := err.(*net.OpError); ok && !opError.Timeout() {
				logger.Debugf("error reading message from tcp socket %s", err.Error())
				return
			}
		}

		match, _ := regexp.MatchString("\\s+ping\\s+", string(messageBuffer.Bytes()))
		if match {
			logger.Debug("tcp socket received ping")
			_, err = c.Write([]byte("pong"))
			if err != nil {
				logger.Errorf("could not write response to tcp socket %s", err)
			}
			return
		}
		// Check our received data for valid JSON. If we get invalid JSON at this point,
		// read again from client, add any new message to the buffer, and parse
		// again.
		var event types.Event
		if err = json.Unmarshal(messageBuffer.Bytes(), &event); err != nil {
			continue
		}

		if event.Entity == nil {
			event.Entity = a.entity
		}

		// At this point, should receive valid JSON, so send it along to the
		// message sender.
		payload, err := json.Marshal(event)
		if err != nil {
			logger.Errorf("could not marshal json payload")
			return
		}

		a.sendMessage(transport.MessageTypeEvent, payload)
		c.Write([]byte("ok"))
		return
	}
	c.Write([]byte("invalid"))
}

// If the socket receives a message containing whitespace and the
// string "ping", it will ignore it.
//
// The socket assumes all other messages will contain a single,
// complete, JSON hash. The hash must be a valid JSON check result.
// Deserialization failures will be logged at the ERROR level by the
// Sensu agent, but the sender of the invalid data will not be
// notified.

func (a *Agent) handleUDPMessages(c net.PacketConn) {
	var buf [1500]byte

	go func() {
		<-a.stopping
		c.Close()
	}()
	// Read everything sent from the connection to the message buffer. Any error
	// will return. If the buffer is zero bytes, close the connection and return.
	for {
		bytesRead, _, err := c.ReadFrom(buf[0:])
		select {
		case <-a.stopping:
			logger.Debug("UDP listener stopped")
			return
		default:
			if err != nil {
				logger.Error(err)
				c.Close()
				return
			} else if bytesRead == 0 {
				c.Close()
				return
			}
			// If the message is a ping, return without notifying sender.
			match, _ := regexp.MatchString("\\s+ping\\s+", string(buf[:bytesRead]))
			if match {
				return
			}

			// Check the message for valid JSON. Valid JSON payloads are passed to the
			// message sender with the addition of the agent's entity if it is not
			// included in the message. Any JSON errors are logged, and we return.
			var event types.Event
			if err = json.Unmarshal(buf[:bytesRead], &event); err != nil {
				logger.Errorf("UDP Invalid event data: %s", err)
			}

			if event.Entity == nil {
				event.Entity = a.entity
			}
			payload, err := json.Marshal(event)
			if err != nil {
				return
			}
			a.sendMessage(transport.MessageTypeEvent, payload)
		}

	}
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

func (a *Agent) receivePump(conn transport.Transport) {
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

func (a *Agent) sendPump(conn transport.Transport) {
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
		Type: transport.MessageTypeKeepalive,
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

// Run starts the Agent.
//
// 1. Connect to the backend, return an error if unsuccessful.
// 2. Start the socket listeners, return an error if unsuccessful.
// 3. Start the send/receive pumps.
// 4. Start sending keepalives.
// 5. Start the API server, shutdown the agent if doing so fails.
func (a *Agent) Run() error {
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

	if err := a.createListenSockets(); err != nil {
		return err
	}

	// These are in separate goroutines so that they can, theoretically, be executing
	// concurrently.
	go a.sendPump(conn)
	go a.receivePump(conn)

	go func() {
		keepaliveTicker := time.NewTicker(time.Duration(a.config.KeepaliveInterval) * time.Second)
		for {
			select {
			case <-keepaliveTicker.C:
				if err := a.sendKeepalive(); err != nil {
					logger.WithError(err).Error("failed sending keepalive")
				}
			case <-a.stopping:
				return
			}

		}
	}()

	// Prepare the HTTP API server
	a.api = newServer(a)

	// Start the HTTP API server
	go func() {
		logger.Info("starting api on address: ", a.api.Addr)

		if err := a.api.ListenAndServe(); err != http.ErrServerClosed {
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
