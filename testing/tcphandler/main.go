package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var (
	port = flag.Int("port", 3333, "port to run extension server on")
	conn = flag.String("type", "tcp", "type of connection to establish (ex. tcp)")
)

func main() {
	flag.Parse()
	l, err := net.Listen(*conn, fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer l.Close()
	log.Print(fmt.Sprintf("Listening on %s 127.0.0.1:%d", *conn, *port))
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	buf := make([]byte, 4096)
	_, err := conn.Read(buf)
	if err != nil {
		log.Print("Error reading: ", err.Error())
	}
	log.Print(string(buf))
	conn.Close()
}
