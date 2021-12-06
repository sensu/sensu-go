package opampd

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sensu/sensu-go/api/core/v3/protobufs"
	"google.golang.org/protobuf/proto"
)

type Config struct {
	Host string
	Port int
}

type OpAMPD struct {
	host        string
	port        int
	upgrader    *websocket.Upgrader
	connections map[string]*websocket.Conn
}

// New creates and bind the OpAMP server to the specified port.
func New(config *Config) (*OpAMPD, error) {
	d := &OpAMPD{
		host: config.Host,
		port: config.Port,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
		connections: make(map[string]*websocket.Conn),
	}

	router := mux.NewRouter()
	router.HandleFunc("/ws", d.handleWS)
	return d, nil
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
		logger.Errorf("error parsing protobuf message: %v\n", err)
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
