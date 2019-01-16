package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/pkg/errors"
)

var networkEcho, iPv4Echo string
var regPortEcho int

func main() {
	Run()
}

func Run() {
	flag.StringVar(&networkEcho, "networkecho", "tcp", "networkEcho to use")
	flag.StringVar(&iPv4Echo, "ip", "localhost", "relay server iPv4Echo")
	flag.IntVar(&regPortEcho, "regportecho", 8080, "port for registering relayable apps")
	flag.Parse()
	socket := fmt.Sprintf("%s:%d", iPv4Echo, regPortEcho)
	// Register the new app, be informed about clients through it.
	conn, err := Open(networkEcho, iPv4Echo, regPortEcho)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Client: Failed to open connection to "+socket))
	}

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	// Notify relayed app of external IP:PORT
	line, err := rw.ReadBytes('\n')
	if err != nil {
		log.Println(err)
	}
	log.Printf("EXTERNAL PORT INFO: %s", string(line))
	for {
		// Whenever a client connects to that external port, Relay server
		// opens a new port for the app to connect to, to mediate client-
		// app communication. 'line' contains that new port number.
		line, err := rw.ReadBytes('\n')
		if err != nil {
			log.Println(err)
			break
		}
		clientPort, err := strconv.Atoi(fmt.Sprintf("%s", line[:len(line)-1]))
		if err != nil {
			log.Fatal(err)
		}
		// Connect to that new port
		connClientAtRelay, err := Open("tcp", iPv4Echo, int(clientPort))
		if err != nil {
			log.Fatal(err)
		}
		go customFunction(connClientAtRelay)
	}
}

// Open makes a net.Conn connection with a relay server.
func Open(networkEcho, ip string, port int) (net.Conn, error) {
	if ip == "localhost" {
		ip = "127.0.0.1"
	}
	conn, err := net.DialTCP(networkEcho, nil, &net.TCPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
		Zone: "",
	})
	if err != nil {
		return nil, err
	}
	log.Printf("dialing %s %s:%d succeeded\n", networkEcho, ip, port)
	return conn, nil
}

// customFunction is an app to be relayed (currently just an echo)
// It receives a net.Conn connection (more precisely net.TCPConn),
// which is able to receive []byte and write []byte
func customFunction(conn net.Conn) {
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	go io.Copy(rw, rw)
}
