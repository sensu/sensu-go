// Package agent is the running Sensu agent. Agents connect to a Sensu backend,
// register their presence, subscribe to check channels, download relevant
// check packages, execute checks, and send results to the Sensu backend via
// the Event channel.
package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sensu/sensu-go/agent/assetmanager"
	"github.com/sensu/sensu-go/handler"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
)

const (
	// MaxMessageBufferSize specifies the maximum number of messages of a given
	// type that an agent will queue before rejecting messages.
	MaxMessageBufferSize = 10

	// TCPSocketReadDeadline specifies the maximum time the TCP socket will wait
	// to receive data.
	TCPSocketReadDeadline = 500 * time.Millisecond
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
	KeepaliveTimeout uint32
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
		logger.WithError(err).Error("error getting hostname")
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
	wg              *sync.WaitGroup
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
		wg:              &sync.WaitGroup{},
	}

	agent.handler.AddHandler(types.CheckRequestType, agent.handleCheck)
	agent.assetManager = assetmanager.New(config.CacheDir, agent.getAgentEntity())

	return agent
}

// createListenSockets UDP and TCP socket listeners on port 3030 for external check
// events.
func (a *Agent) createListenSockets() (string, string, error) {
	// we have two listeners that we want to shut down before agent.Stop() returns.
	a.wg.Add(2)

	addr := fmt.Sprintf("%s:%d", a.config.Socket.Host, a.config.Socket.Port)

	// Setup UDP socket listener
	UDPServerAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return "", "", err
	}

	udpListen, err := net.ListenUDP("udp", UDPServerAddr)
	if err != nil {
		return "", "", err
	}
	logger.Infof("starting UDP listener on %s", addr)
	go a.handleUDPMessages(udpListen)

	// Setup TCP socket listener
	TCPServerAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return "", "", err
	}

	logger.Infof("starting TCP listener on %s", addr)
	tcpListen, err := net.ListenTCP("tcp", TCPServerAddr)
	if err != nil {
		return "", "", err
	}

	// we have to monitor the stopping channel out of band, otherwise
	// the tcpListen.Accept() loop will never return.
	var isListenerClosed bool
	go func() {
		<-a.stopping
		logger.Debug("TCP listener stopped")
		isListenerClosed = true
		tcpListen.Close()
	}()

	go func() {
		// Actually block the waitgroup until the last call to Accept()
		// returns.
		defer a.wg.Done()

		for {
			conn, err := tcpListen.Accept()
			if err != nil {
				// Only log the error if the listener was not properly stopped by us
				if !isListenerClosed {
					logger.WithError(err).Error("error accepting TCP connection")
				}
				tcpListen.Close()
				return
			}
			go a.handleTCPMessages(conn)
		}
	}()

	return tcpListen.Addr().String(), udpListen.LocalAddr().String(), err
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

	// Read incoming tcp messages from client until we hit a valid JSON message.
	// If we don't get valid JSON or a ping request after 500ms, close the
	// connection (timeout).
	readDeadline := time.Now().Add(TCPSocketReadDeadline)

	// Only allow 500ms of IO. After this time, all IO calls on the connection
	// will fail.
	if err := c.SetReadDeadline(readDeadline); err != nil {
		logger.WithError(err).Error("error setting read deadline")
		return
	}

	// It is possible that our buffered readers/writers will cause us
	// to iterate.
	for time.Now().Before(readDeadline) {
		_, err := connReader.WriteTo(messageBuffer)
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
				logger.WithError(err).Error("could not write response to tcp socket")
			}
			return
		}
		// Check our received data for valid JSON. If we get invalid JSON at this point,
		// read again from client, add any new message to the buffer, and parse
		// again.
		var payload map[string]interface{}
		var event types.Event
		if err = json.Unmarshal(messageBuffer.Bytes(), &payload); err != nil {
			continue
		}

		if err = translateToEvent(payload, &event); err != nil {
			logger.WithError(err).Error("1.x returns \"invalid\"")
			return
		}

		// Prepare the event by mutating it as required so it passes validation
		if err = prepareEvent(a, &event); err != nil {
			logger.WithError(err).Error("invalid event")
			return
		}

		// At this point, should receive valid JSON, so send it along to the
		// message sender.
		message, err := json.Marshal(event)
		if err != nil {
			logger.WithError(err).Error("could not marshal json payload")
			return
		}

		a.sendMessage(transport.MessageTypeEvent, message)
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
		a.wg.Done()
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
				logger.WithError(err).Error("Error reading from UDP socket")
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
				logger.WithError(err).Error("UDP Invalid event data")
				return
			}

			// Prepare the event by mutating it as required so it passes validation
			if err = prepareEvent(a, &event); err != nil {
				logger.WithError(err).Error("invalid event")
				return
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
			logger.WithError(err).Error("transport receive error")
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
				logger.WithError(err).Error("error handling message")
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
				logger.WithError(err).Fatal("transport send error")
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

func (a *Agent) buildTransportHeaderMap() http.Header {
	header := http.Header{}
	header.Set(transport.HeaderKeyAgentID, a.config.AgentID)
	header.Set(transport.HeaderKeyEnvironment, a.config.Environment)
	header.Set(transport.HeaderKeyOrganization, a.config.Organization)
	header.Set(transport.HeaderKeyUser, a.config.User)
	header.Set(transport.HeaderKeySubscriptions, strings.Join(a.config.Subscriptions, ","))

	return header
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
	header := a.buildTransportHeaderMap()
	header.Set("Authorization", "Basic "+userCredentials)

	conn, err := transport.Connect(a.backendSelector.Select(), a.config.TLS, header)
	if err != nil {
		return err
	}
	a.conn = conn

	if _, _, err := a.createListenSockets(); err != nil {
		return err
	}

	// These are in separate goroutines so that they can, theoretically, be executing
	// concurrently.
	go a.sendPump(conn)
	go a.receivePump(conn)

	// Send an immediate keepalive once we've connected.
	a.sendKeepalive()

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

	// Allow Stop() to block until the HTTP server shuts down.
	a.wg.Add(1)
	go func() {
		// NOTE: This does not guarantee a clean shutdown of the HTTP API.
		// This is _only_ for the purpose of making Stop() a blocking call.
		// The goroutine running the HTTP Server has to return before Stop()
		// can return, so we use this to signal that goroutine to shutdown.
		<-a.stopping
		logger.Info("api shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		a.api.Shutdown(ctx)
		a.wg.Done()
	}()

	return nil
}

// Stop shuts down the agent. It will block until all listening goroutines
// have returned.
func (a *Agent) Stop() {
	close(a.stopping)
	a.wg.Wait()
}

func (a *Agent) addHandler(msgType string, handlerFunc handler.MessageHandlerFunc) {
	a.handler.AddHandler(msgType, handlerFunc)
}
