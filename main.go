package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"
	"time"

	echo "github.com/tkivisik/tcp-relay-server/sampleappecho"
)

const (
	timeout          = time.Minute
	RANDOM       int = 0
	INCREMENTONE int = 1
)

var (
	portGlobal int32
)

func main() {
	regPort := flag.Int("regport", 8080, "port for registering relayable apps")
	relayStyle := flag.Int("relaystyle", INCREMENTONE, fmt.Sprintf("%d for random, %d for incrementing one", RANDOM, INCREMENTONE))
	sampleApp := flag.Bool("sampleapp", false, "run a sample echo server app")
	flag.Parse()

	if *sampleApp {
		echo.Run()
	} else {
		RunTCPServer(*regPort, *relayStyle)
	}
}

func RunTCPServer(regPort int, relayStyle int) {
	atomic.AddInt32(&portGlobal, int32(regPort))
	// Listen to register Apps
	regListener := listenForConnections(regPort)
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
		new := atomic.AddInt32(&portGlobal, 1)
		return int(new)
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
	clientListener := listenForConnections(nextPort)
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
			appClientListener := listenForConnections(nextPort)
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
func listenForConnections(port int) net.Listener {
	// Set up endpoint for registering Apps
	regTCPListener, err := net.ListenTCP("tcp", &net.TCPAddr{
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
