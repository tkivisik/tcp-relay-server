package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	var network string
	var regPort int
	flag.StringVar(&network, "network", "tcp", "network to use")
	flag.IntVar(&regPort, "regport", 8080, "port for registering relayable apps")
	flag.Parse()

	// Listen to register Apps
	regListener := listenForConnections(network, fmt.Sprintf(":%d", regPort))
	log.Printf("listening for relayable apps to register on port: %d\n", regPort)
	for {
		connRelayToApp := acceptConnections(regListener)
		defer func() {
			log.Printf("%s closed.\n", connRelayToApp.RemoteAddr())
			connRelayToApp.Close()
		}()

		// Listen for clients for a registered App
		go handleClientsForApp(connRelayToApp)
	}
}

// handleClientsForApp sets up an external endpoint and mediates client-app-client communication
func handleClientsForApp(connRelayToApp net.Conn) {
	// :0 will pick an available port
	clientListener := listenForConnections(connRelayToApp.LocalAddr().Network(), ":0")
	relayedPort := clientListener.Addr().(*net.TCPAddr).Port
	log.Printf("listening for clients on port: %d\n", relayedPort)
	fmt.Fprintf(connRelayToApp, fmt.Sprintf("established relay address  port:%d\n", relayedPort))
	for {

		connClientToRelay := acceptConnections(clientListener)
		defer func() {
			log.Printf("%s closed.\n", connClientToRelay.RemoteAddr())
			connClientToRelay.Close()
		}()

		go io.Copy(connRelayToApp, connClientToRelay)
		go io.Copy(connClientToRelay, connRelayToApp)
	}
}

// listenForConnections sets up an endpoint which listens for connections
func listenForConnections(network, port string) net.Listener {
	regListener, err := net.Listen(network, port)
	if err != nil {
		log.Println(err)
	}
	return regListener
}

// acceptConnections accepts remote connections to local net.Listener
func acceptConnections(l net.Listener) net.Conn {
	conn, err := l.Accept()
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("connected from [relay]: %s to: %s\n", conn.LocalAddr(), conn.RemoteAddr())
	return conn
}
