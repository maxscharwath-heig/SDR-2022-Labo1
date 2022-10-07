package main

import (
	"fmt"
	"net"
	"os"
	. "sdr/labo1/core"
	"sdr/labo1/types"
)

func main() {
	// Listen for incoming connections.
	config := ReadConfig("config/server.json", &types.ServerConfiguration{})

	l, err := net.Listen(config.Type, config.FullUrl())
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()

	fmt.Println("Listening on " + config.FullUrl())
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		fmt.Println("New connexion !")
		go handleRequest(conn)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	// Send a response back to person contacting us.
	conn.Write([]byte("Welcome to the server !!"))
	// Close the connection when you're done with it.
	conn.Close()
}
