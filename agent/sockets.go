package agent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/v1"
)

var (
	pingRe = regexp.MustCompile(`\s+ping\s+`)
)

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
		if err := tcpListen.Close(); err != nil {
			logger.Debug(err)
		}
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
				if err := tcpListen.Close(); err != nil {
					logger.Debug(err)
				}
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
	defer func() {
		if err := c.Close(); err != nil {
			logger.Debug(err)
		}
	}()
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

		if match := pingRe.Match(messageBuffer.Bytes()); match {
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
		var event types.Event
		var result v1.CheckResult
		if err = json.Unmarshal(messageBuffer.Bytes(), &result); err != nil {
			continue
		}

		if err = translateToEvent(a, result, &event); err != nil {
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
		payload, err := json.Marshal(event)
		if err != nil {
			logger.WithError(err).Error("could not marshal json payload")
			return
		}

		a.sendMessage(transport.MessageTypeEvent, payload)
		_, _ = c.Write([]byte("ok"))
		return
	}
	_, _ = c.Write([]byte("invalid"))
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
		if err := c.Close(); err != nil {
			logger.Debug(err)
		}
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
				if err := c.Close(); err != nil {
					logger.Debug(err)
				}
				return
			} else if bytesRead == 0 {
				if err := c.Close(); err != nil {
					logger.Debug(err)
				}
				return
			}
			// If the message is a ping, return without notifying sender.
			if match := pingRe.Match(buf[:bytesRead]); match {
				return
			}

			// Check the message for valid JSON. Valid JSON payloads are passed to the
			// message sender with the addition of the agent's entity if it is not
			// included in the message. Any JSON errors are logged, and we return.
			var event types.Event
			var result v1.CheckResult
			if err = json.Unmarshal(buf[:bytesRead], &result); err != nil {
				logger.WithError(err).Error("UDP Invalid event data")
				return
			}

			if err = translateToEvent(a, result, &event); err != nil {
				logger.WithError(err).Error("1.x returns \"invalid\"")
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
