package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
)

// TestListenForConnections tests for network and port
func TestListenForConnections(t *testing.T) {
	tables := []struct {
		network string
		port    string
	}{
		{"tcp", ":8080"},
		{"tcp", ":9090"},
	}
	for _, table := range tables {
		listener := listenForConnections(table.network, table.port)
		defer listener.Close()

		if network := listener.Addr().Network(); network != table.network {
			t.Errorf("Network not as expected, got: %s, want: %s.", network, table.network)
		}
		if address := listener.Addr().String(); strings.HasSuffix(address, table.port) == false {
			t.Errorf("Port not as expected, got: %s, want: %s.", address, table.port)
		}
	}
}

// TestListenForConnectionsPanic tests for panic if port in use
func TestListenForConnectionsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	listener1 := listenForConnections("tcp", ":9990")
	listener2 := listenForConnections("tcp", ":9990")
	listener1.Close()
	listener2.Close()

}

// TestAcceptConnections tests for multiple connections per listener
// Currently the worst test :)
func TestAcceptConnections(t *testing.T) {
	wg := sync.WaitGroup{}
	listener := listenForConnections("tcp", ":31331")
	defer listener.Close()
	go func() {
		conn := acceptConnections(listener)
		defer conn.Close()
		fmt.Println(conn.RemoteAddr())
	}()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			conn, err := net.Dial("tcp", ":31331")
			defer conn.Close()
			if err != nil {
				t.Errorf("Was expecting no errors in accepting multiple dials")
			}
			//			fmt.Printf("connected from [relay]: %s to: %s\n", conn.LocalAddr(), conn.RemoteAddr())
			wg.Done()
		}()
	}
	wg.Wait()
}
