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
	users  chan []types.User
	events chan []types.Event
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
	chanData := ChanData{
		users:  make(chan []types.User, 1),
		events: make(chan []types.Event, 1),
	}
	chanData.users <- config.Users
	chanData.events <- config.Events

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
		msg := <-entryMessages
		switch msg.Path {
		case "create":
			create(conn, msg, data)
		case "show":
			show(conn, msg, data)
		}
	}
}

func handleAuth(credential types.Credentials, data ChanData) (types.User, error) {
	users := <-data.users
	defer func() {
		data.users <- users
	}()
	if credential.Username == "" || credential.Password == "" {
		return types.User{}, fmt.Errorf("invalid credentials")
	}
	for _, user := range users {
		if user.Username == credential.Username && user.Password == credential.Password {
			return user, nil
		}
	}
	return types.User{}, fmt.Errorf("user not found")
}

// Handles creating of new Events.
func create(conn net.Conn, message network.Message, data ChanData) {
	request := network.RequestFromJson[types.Event](message.Body)
	user, err := handleAuth(request.Credentials, data)
	if err != nil {
		network.SendData(conn, message.Path, err.Error())
		return
	}
	network.SendResponse(conn, message.Path, network.Response[types.User]{true, user})
}

func show(conn net.Conn, message network.Message, data ChanData) {
	// request := network.FromJson[types.Event](message.Body)

	// fmt.Println(request)

	events := <-data.events
	defer func() {
		data.events <- events
	}()

	network.SendRequest(conn, message.Path, network.Request[any]{
		Data: events,
	})
}
