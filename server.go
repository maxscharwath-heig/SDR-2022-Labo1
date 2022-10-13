package main

import (
	"fmt"
	"net"
	"os"
	. "sdr/labo1/core"
	"sdr/labo1/network"
	"sdr/labo1/types"
)

type ChanData struct {
}

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

	//init chan data structure
	chanData := ChanData{}

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		fmt.Println("New connexion !")
		go handleRequest(conn, chanData)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn, data ChanData) {
	entryMessages := make(chan network.Message)

	//NEED TO REFACTOR THIS (NOT USE GOROUTINE)
	go network.HandleReceiveData(conn, entryMessages)
	for {
		data := <-entryMessages
		fmt.Println("path: " + data.Path)
		fmt.Println("body: " + data.Body)
	}
}
