package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

const (
	RANDOM int = iota
	INCREMENTONE
)

func init() {
	flag.StringVar(&network, "network", "tcp", "network to use")
	flag.IntVar(&regPort, "regport", 8080, "port for registering relayable apps")
	flag.IntVar(&relayStyle, "relaystyle", INCREMENTONE, fmt.Sprintf("%d for random, %d for incrementing one", RANDOM, INCREMENTONE))
}

var network string
var regPort, relayStyle int

func main() {
	flag.Parse()

	// Listen to register Apps
	regListener := listenForConnections(network, regPort)
	log.Printf("listening for relayable apps to register on port: %d\n", regPort)
	for {
		// Accept a connection from an App
		connAppCommand := acceptConnections(regListener)
		defer func() {
			log.Printf("%s closed.\n", connAppCommand.RemoteAddr())
			connAppCommand.Close()
		}()

		// Listen for clients for a registered App
		go handleClientsForApp(connAppCommand, regPort, relayStyle)
	}
}

var port int
var mux sync.Mutex

func newPort(regPort, relayStyle int) int {
	if relayStyle == INCREMENTONE {
		mux.Lock()
		if port == 0 {
			port = regPort
		}
		port++
		mux.Unlock()
		return port
	}
	return 0
}

// handleClientsForApp sets up an external endpoint and mediates client-app-client communication
func handleClientsForApp(connAppCommand net.Conn, regPort, relayStyle int) {
	nextPort := newPort(regPort, relayStyle)

	// port=0 will pick an available port
	clientListener := listenForConnections(connAppCommand.LocalAddr().Network(), nextPort)
	relayedPort := clientListener.Addr().(*net.TCPAddr).Port
	log.Printf("listening for clients on port: %d\n", relayedPort)
	// Notify App of it's external relayed port
	fmt.Fprintf(connAppCommand, fmt.Sprintf("established relay address port:%d\n", relayedPort))
	for {
		connClientToRelay := acceptConnections(clientListener)
		defer func() {
			log.Printf("%s closed.\n", connClientToRelay.RemoteAddr())
			connClientToRelay.Close()
		}()

		// Ask an App to make a new connection to a new socket for each client.
		nextPort := newPort(regPort, RANDOM)
		appClientListener := listenForConnections(connAppCommand.LocalAddr().Network(), nextPort)
		appClientPort := appClientListener.Addr().(*net.TCPAddr).Port //clientListener.Addr().(*net.TCPAddr).Port
		fmt.Fprintf(connAppCommand, fmt.Sprintf("%d\n", appClientPort))
		log.Printf("listening for AppClient connections on port: %d\n", appClientPort)
		connAppClient := acceptConnections(appClientListener)
		defer func() {
			log.Printf("%s closed.\n", connClientToRelay.RemoteAddr())
			connClientToRelay.Close()
		}()

		go io.Copy(connAppClient, connClientToRelay)
		go io.Copy(connClientToRelay, connAppClient)
	}
}

// listenForConnections sets up an endpoint which listens for connections
func listenForConnections(network string, port int) net.Listener {
	// Set up endpoint for registering Apps
	regTCPListener, err := net.ListenTCP(network, &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: port,
		Zone: "",
	})
	if err != nil {
		log.Panic(err)
	}
	return regTCPListener
}

// acceptConnections accepts remote connections to local net.Listener
func acceptConnections(l net.Listener) net.Conn {
	conn, err := l.Accept()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("connected from [relay]: %s to: %s\n", conn.LocalAddr(), conn.RemoteAddr())
	return conn
}
