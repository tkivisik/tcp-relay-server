package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	echo "github.com/tkivisik/tcp-relay-server/sampleappecho"
)

const (
	RANDOM int = iota
	INCREMENTONE
	timeout = time.Minute
)

var (
	network             string
	regPort, relayStyle int
	sampleApp           bool
	port                int
	mux                 sync.Mutex
)

func init() {
	flag.StringVar(&network, "network", "tcp", "network to use")
	flag.IntVar(&regPort, "regport", 8080, "port for registering relayable apps")
	flag.IntVar(&relayStyle, "relaystyle", INCREMENTONE, fmt.Sprintf("%d for random, %d for incrementing one", RANDOM, INCREMENTONE))
	flag.BoolVar(&sampleApp, "sampleapp", false, "run a sample echo server app")
}

func main() {
	flag.Parse()

	if sampleApp {
		echo.Run()
	} else {
		RunTCPServer()
	}
}

func RunTCPServer() {
	// Listen to register Apps
	regListener := listenForConnections(network, regPort)
	log.Printf("listening for relayable apps to register on port: %d\n", regPort)
	for {
		// Accept a connection from an App
		connAppCommand := acceptConnections(regListener)

		// Listen for clients for a registered App
		go handleClientsForApp(connAppCommand, regPort, relayStyle)
	}
}

func newPort(regPort, relayStyle int) int {
	if relayStyle == INCREMENTONE {
		mux.Lock()
		defer mux.Unlock()
		if port == 0 {
			port = regPort
		}
		port++
		return port
	}
	return 0
}

// handleClientsForApp sets up an external endpoint and mediates client-app-client communication
func handleClientsForApp(connAppCommand net.Conn, regPort, relayStyle int) {
	defer func() {
		log.Printf("%s (connAppCommand) closed.\n", connAppCommand.RemoteAddr())
		connAppCommand.Close()
	}()

	nextPort := newPort(regPort, relayStyle)

	// port=0 will pick an available port
	clientListener := listenForConnections(connAppCommand.LocalAddr().Network(), nextPort)
	relayedPort := clientListener.Addr().(*net.TCPAddr).Port
	log.Printf("listening for clients on port: %d\n", relayedPort)
	// Notify App of it's external relayed port
	fmt.Fprintf(connAppCommand, "established relay address port:%d\n", relayedPort)
	for {
		connClientToRelay := acceptConnections(clientListener)

		go func(connClientToRelay net.Conn) {
			defer func() {
				log.Printf("%s (connClientToRelay) closed.\n", connClientToRelay.RemoteAddr())
				connClientToRelay.Close()
			}()

			// Ask an App to make a new connection to a new socket for each client.
			nextPort := newPort(regPort, RANDOM)
			appClientListener := listenForConnections(connAppCommand.LocalAddr().Network(), nextPort)
			appClientPort := appClientListener.Addr().(*net.TCPAddr).Port //clientListener.Addr().(*net.TCPAddr).Port
			fmt.Fprintf(connAppCommand, "%d\n", appClientPort)
			log.Printf("listening for AppClient connections on port: %d\n", appClientPort)
			connAppClient := acceptConnections(appClientListener)
			defer func() {
				log.Printf("%s (connAppClinet) closed.\n", connAppClient.RemoteAddr())
				connAppClient.Close()
			}()

			go io.Copy(connAppClient, connClientToRelay)
			go io.Copy(connClientToRelay, connAppClient)

			time.Sleep(timeout) // TODO - rewrite to close whenever any connections are lost

		}(connClientToRelay)
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
