// Package agent is the running Sensu agent. Agents connect to a Sensu backend,
// register their presence, subscribe to check channels, download relevant
// check packages, execute checks, and send results to the Sensu backend via
// the Event channel.
package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/sensu/sensu-go/handler"
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
	// BackendURL is the URL to the Sensu Backend. Default: ws://127.0.0.1:8080
	BackendURL string
	// Subscriptions is an array of subscription names. Default: empty array.
	Subscriptions []string
	// KeepaliveInterval is the interval, in seconds, when agents will send a
	// keepalive to sensu-backend. Default: 60
	KeepaliveInterval int
}

// NewConfig provides a new Config object initialized with defaults.
func NewConfig() *Config {
	c := &Config{
		BackendURL:        "ws://127.0.0.1:8080/agents/ws",
		Subscriptions:     []string{},
		KeepaliveInterval: 60,
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Println("error getting hostname: ", err.Error())
		// TODO(greg): wat do?
		c.AgentID = "unidentified-sensu-agent"
	}
	c.AgentID = hostname

	return c
}

// An Agent receives and acts on messages from a Sensu Backend.
type Agent struct {
	config       *Config
	backendURL   string
	handler      *handler.MessageHandler
	conn         *transport.Transport
	sendq        chan *transport.Message
	disconnected bool
	stopping     chan struct{}
	stopped      chan struct{}
	entity       *types.Entity
}

// NewAgent creates a new Agent and returns a pointer to it.
func NewAgent(config *Config) *Agent {
	return &Agent{
		config:       config,
		backendURL:   config.BackendURL,
		handler:      handler.NewMessageHandler(),
		disconnected: true,
		stopping:     make(chan struct{}),
		stopped:      make(chan struct{}),
		sendq:        make(chan *transport.Message, 10),
	}
}

func (a *Agent) receivePump(wg *sync.WaitGroup, conn *transport.Transport) {
	wg.Add(1)
	defer wg.Done()

	log.Println("connected - starting receivePump")
	for {
		if a.disconnected {
			log.Println("disconnected - stopping receivePump")
			return
		}

		m, err := conn.Receive()
		if err != nil {
			switch err := err.(type) {
			case transport.ConnectionError:
				a.disconnected = true
			case transport.ClosedError:
				a.disconnected = true
			default:
				log.Println("recv error:", err.Error())
			}
			continue
		}
		log.Println("message received - type: ", m.Type, " message: ", m.Payload)

		err = a.handler.Handle(m.Type, m.Payload)
		if err != nil {
			log.Println("error handling message:", err.Error())
		}
	}
}

func (a *Agent) sendPump(wg *sync.WaitGroup, conn *transport.Transport) {
	wg.Add(1)
	defer wg.Done()

	// The sendPump is actually responsible for shutting down the transport
	// to prevent a race condition between it and something else trying
	// to close the transport (which actually causes a write to the websocket
	// connection.)
	defer a.conn.Close()

	log.Println("connected - starting sendPump")
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case msg := <-a.sendq:
			err := conn.Send(msg)
			if err != nil {
				switch err := err.(type) {
				case transport.ConnectionError:
					a.disconnected = true
				case transport.ClosedError:
					a.disconnected = true
				default:
					log.Println("send error:", err.Error())
				}
			}
		case <-ticker.C:
			if a.disconnected {
				log.Println("disconnected - stopping sendPump")
				return
			}
		case <-a.stopping:
			// we abandon anything left in the queue and exit the sendpump
			// TODO(greg): leave no messages behind! if we can.
			return
		}
	}
}

func (a *Agent) sendKeepalive() error {
	log.Println("sending keepalive")
	msg := &transport.Message{
		Type: types.KeepaliveType,
	}
	keepalive := &types.Event{}
	keepalive.Entity = a.getDefaultEntity()
	keepalive.Timestamp = time.Now().Unix()
	msgBytes, err := json.Marshal(keepalive)
	if err != nil {
		return err
	}
	msg.Payload = msgBytes

	a.sendq <- msg

	return nil
}

func (a *Agent) keepaliveLoop() {
	log.Println("starting keepalive loop")

	ticker := time.NewTicker(time.Duration(a.config.KeepaliveInterval) * time.Second)
	defer ticker.Stop()
	// when we're disconnected, we want to pause sending keepalives so that we
	// can queue up things in the send queue that are important.
	for {
		select {
		case <-ticker.C:
			if !a.disconnected {
				if err := a.sendKeepalive(); err != nil {
					log.Println("error marshaling keepalive: ", err.Error())
				}
			}
		case <-a.stopping:
			log.Println("stopping keepalive loop")
			return
		}
	}
}

func (a *Agent) getDefaultEntity() *types.Entity {
	if a.entity == nil {
		e := &types.Entity{
			ID: a.config.AgentID,
		}
		a.entity = e
	}

	return a.entity
}

func (a *Agent) handshake() error {
	handshake := &types.AgentHandshake{
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
	conn, err := transport.Connect(a.config.BackendURL)
	if err != nil {
		return err
	}
	a.conn = conn
	wg := &sync.WaitGroup{}
	err = a.handshake()
	if err != nil {
		return err
	}

	a.disconnected = false
	go a.sendPump(wg, conn)
	go a.receivePump(wg, conn)
	go a.keepaliveLoop()

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
			case <-ticker.C:
				if a.disconnected {
					log.Println("disconnected - attempting to reconnect: ", a.backendURL)
					wg.Wait()
					conn, err := transport.Connect(a.backendURL)
					if err != nil {
						log.Println("connection error:", err.Error())
						// TODO(greg): exponential backoff
						time.Sleep(1 * time.Second)
						retries++
						// TODO(greg): Figure out a max backoff / max retries thing
						// before we fail over to the configured backend url
						if retries >= 30 {
							a.backendURL = a.config.BackendURL
						}
						continue
					}
					a.conn = conn
					log.Println("reconnected: ", a.backendURL)
					wg.Add(2)
					err = a.handshake()
					if err != nil {
						log.Println("handshake error: ", err.Error())
						continue
					}
					go a.sendPump(wg, conn)
					go a.receivePump(wg, conn)
					a.disconnected = false
				}
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

func (a *Agent) sendMessage(msgType string, payload []byte) {
	// blocks until message can be enqueued.
	// TODO(greg): ring buffer?
	msg := &transport.Message{
		Type:    msgType,
		Payload: payload,
	}
	a.sendq <- msg
}

func (a *Agent) addHandler(msgType string, handlerFunc handler.MessageHandlerFunc) {
	a.handler.AddHandler(msgType, handlerFunc)
}
