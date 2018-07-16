package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

var (
	port = flag.Int("port", 3333, "port to run tcp server on")
	conn = flag.String("type", "tcp", "type of connection to establish (ex. tcp)")
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
	log.Print(string(buf))
}
