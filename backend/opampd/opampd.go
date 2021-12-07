package opampd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sensu/sensu-go/api/core/v3/protobufs"
	"google.golang.org/protobuf/proto"
)

type Config struct {
	Host string
	Port int
	Path string
}

type OpAMPD struct {
	host        string
	port        int
	path        string
	upgrader    *websocket.Upgrader
	connections map[string]*websocket.Conn
	err         error
	httpServer  *http.Server
	wg          *sync.WaitGroup
	errChan     chan error
}

// New creates and bind the OpAMP server to the specified port.
func New(config *Config) (*OpAMPD, error) {
	d := &OpAMPD{
		host: config.Host,
		port: config.Port,
		path: config.Path,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
		connections: make(map[string]*websocket.Conn),
		wg:          &sync.WaitGroup{},
		errChan:     make(chan error, 1),
	}

	router := mux.NewRouter()
	router.HandleFunc(d.path, d.handleWS)

	d.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", d.host, d.port),
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		//TLSConfig:    tlsServerConfig,
		// Capture the log entries from agentd's HTTP server
		//ErrorLog: logger,
		ConnState: func(c net.Conn, cs http.ConnState) {
			if cs != http.StateClosed {
				var msg []byte
				if _, err := c.Read(msg); err != nil {
					logger.WithError(err).Error("websocket connection error")
				}
			}
		},
	}

	return d, nil
}

func (d *OpAMPD) Start() error {
	logger.Warn("starting opampd on address: ", d.httpServer.Addr)
	ln, err := net.Listen("tcp", d.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("failed to start agentd: %s", err)
	}

	d.wg.Add(1)

	go func() {
		defer d.wg.Done()
		err := d.httpServer.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			d.errChan <- fmt.Errorf("opampd failed while serving: %s", err)
		}
	}()

	return nil
}

func (d *OpAMPD) Stop() error {
	if err := d.httpServer.Shutdown(context.TODO()); err != nil {
		// failure/timeout shutting down the server gracefully
		logger.Error("failed to shutdown http server gracefully - forcing shutdown")
		if closeErr := d.httpServer.Close(); closeErr != nil {
			logger.Error("failed to shutdown http server forcefully")
		}
	}
	d.wg.Wait()

	return nil
}

func (d *OpAMPD) Err() <-chan error {
	return d.errChan
}

func (d *OpAMPD) Name() string {
	return "opampd"
}

// handleWS upgrades an incoming http message into a WebSocket connection.
func (d *OpAMPD) handleWS(response http.ResponseWriter, request *http.Request) {
	remoteAddr := request.RemoteAddr
	connection, err := d.upgrader.Upgrade(response, request, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			logger.Error(err)
		}
		return
	}
	logger.Infof("WebSocket client connected from %s\n", remoteAddr)
	d.connections[remoteAddr] = connection
	go d.messageReader(connection)
}

// messageReader  is a goroutine that reads messages from a websocket
///connection. There is one goroutine per client connection. The received
// messages are published to the inMessages channel to notify the listeners
func (d *OpAMPD) messageReader(connection *websocket.Conn) {
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("error: %v", err.Error())
			}
			break
		}
		go d.handleMessage(message)
	}
}

// handleMessage parses a protobuf message received from the agent and calls the
// appropriate handler.
func (d *OpAMPD) handleMessage(message []byte) {
	a2s := protobufs.AgentToServer{}
	err := proto.Unmarshal(message, &a2s)
	if err != nil {
		logger.Errorf("error parsing protobuf message: %v", err)
		return
	}

	logger.Infof("OpAMP message received from %s\n", a2s.InstanceUid)

	if a2s.StatusReport != nil {
		logger.Infoln("status report")
	} else if a2s.AddonStatuses != nil {
		logger.Infoln("addon statuses")
	} else if a2s.AgentInstallStatus != nil {
		logger.Infoln("agent install status")
	} else if a2s.AgentDisconnect != nil {
		logger.Infoln("agent disconnect")
	} else {
		// invalid message
	}
}
