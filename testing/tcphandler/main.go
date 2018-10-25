package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/sensu/sensu-go/types"
)

var (
	port = flag.Int("port", 3333, "port to run tcp server on")
	conn = flag.String("type", "tcp", "type of connection to establish (ex. tcp)")
	file = flag.String("file", "/tmp/tcp-handler.json", "optionally write json to file (ex. /tmp/tcp-handler.json")
)

func main() {
	flag.Parse()
	l, err := net.Listen(*conn, fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	log.Print(fmt.Sprintf("Listening on %s 127.0.0.1:%d", *conn, *port))
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 4096)
	_, err := conn.Read(buf)
	if err != nil {
		log.Print("Error reading: ", err.Error())
		return
	}
	if *file != "" {
		event := &types.Event{}
		err = json.Unmarshal(bytes.Trim(buf, "\x00"), &event)
		if err != nil {
			log.Print("Recieved non-event: ", err.Error())
			return
		}
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Print("Error marshalling event: ", err.Error())
			return
		}
		err = ioutil.WriteFile(*file, eventJSON, 0644)
		if err != nil {
			log.Print("Error writing to file: ", err.Error())
			return
		}
	}
	log.Print(string(buf))
}
