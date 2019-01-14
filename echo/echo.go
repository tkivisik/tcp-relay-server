package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/pkg/errors"
)

var network, iPv4 string
var regPort int

func init() {
	flag.StringVar(&network, "network", "tcp", "network to use")
	flag.StringVar(&iPv4, "ip", "localhost", "relay server IPv4")
	flag.IntVar(&regPort, "regport", 8080, "port for registering relayable apps")
}

func main() {
	flag.Parse()
	iPAndPort := fmt.Sprintf("%s:%d", iPv4, regPort)
	conn, err := Open(network, iPAndPort)
	if err != nil {
		fmt.Println(errors.Wrap(err, "Client: Failed to open connection to "+iPAndPort))
	}
	go serveForever(conn)
	select {} // block forever
}

// Open makes a net.Conn connection with a relay server.
func Open(network, addr string) (net.Conn, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	log.Printf("dialing %s %s succeeded\n", network, addr)
	return conn, nil //bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}

// serveForever wraps the relayed app to notify it's external ip:port
func serveForever(connToRelay net.Conn) {
	rw := bufio.NewReadWriter(bufio.NewReader(connToRelay), bufio.NewWriter(connToRelay))
	// Notify relayed app of external IP:PORT
	line, err := rw.ReadBytes('\n')
	if err != nil {
		log.Println(err)
	}
	log.Printf("EXTERNAL PORT INFO: %s", string(line))
	customFunction(rw)
}

// customFunction is an app to be relayed
func customFunction(rw *bufio.ReadWriter) {
	log.Println("customFunction() started")
	for {
		line, err := rw.ReadBytes('\n')
		if err != nil {
			log.Println(err)
			break
		}

		log.Printf("CUSTOM MESSAGE TO BE ECHOED: %s", string(line))
		_, err = rw.Write(line)
		if err != nil {
			log.Println(err)
		}
		rw.Flush()
	}

	log.Println("customFunction() finished")
}
