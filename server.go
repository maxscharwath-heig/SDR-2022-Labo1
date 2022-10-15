package main

import (
	"fmt"
	"net"
	"os"
	"sdr/labo1/config"
	. "sdr/labo1/core"
	"sdr/labo1/dto"
	"sdr/labo1/network"
	"sdr/labo1/types"
)

type ChanData struct {
	users  chan []types.User
	events chan []types.Event
}

func main() {
	// Listen for incoming connections.
	config := ReadConfig("config/server.json", &config.ServerConfiguration{})

	l, err := net.Listen("tcp", config.FullUrl())
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

	{ // LOAD DATA FROM CONFIG
		users, events := config.GetData()
		chanData.users <- users
		chanData.events <- events
	}

	protocol := network.ServerProtocol{
		AuthFunc: func(credential types.Credentials) (bool, any) {
			users := <-chanData.users
			defer func() {
				chanData.users <- users
			}()
			if credential.Username == "" || credential.Password == "" {
				return false, nil
			}
			for _, user := range users {
				if user.Username == credential.Username && user.Password == credential.Password {
					return true, &user
				}
			}
			return false, nil
		},
		Endpoints: map[string]network.Endpoint{
			"create":   createEndpoint(&chanData),
			"show":     showEndpoint(&chanData),
			"close":    closeEndpoint(&chanData),
			"register": registerEndpoint(&chanData),
		},
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go protocol.Process(conn)
	}
}

func createEndpoint(chanData *ChanData) network.Endpoint {
	return network.Endpoint{
		NeedsAuth: true,
		HandlerFunc: func(request network.Request) any {
			events := <-chanData.events
			defer func() {
				chanData.events <- events
			}()

			data := dto.EventCreate{}
			request.GetJson(&data)

			event := types.Event{
				Id:        len(events) + 1,
				Name:      data.Name,
				Open:      true,
				Organizer: request.Auth.(*types.User),
			}
			for i, job := range data.Jobs {
				id := i + 1
				event.Jobs[id] = types.Job{
					Id:       id,
					Name:     job.Name,
					Capacity: job.Capacity,
				}
			}
			events = append(events, event)
			return dto.EventToDTO(event)
		},
	}
}

func showEndpoint(chanData *ChanData) network.Endpoint {
	return network.Endpoint{
		NeedsAuth: false,
		HandlerFunc: func(request network.Request) any {
			events := <-chanData.events
			defer func() {
				chanData.events <- events
			}()

			data := dto.EventShow{}
			request.GetJson(&data)

			if data.EventId != -1 {
				for _, ev := range events {
					if ev.Id == data.EventId {
						return dto.EventToDTO(ev)
					}
				}
				return nil
			}
			return dto.EventsToDTO(events)
		},
	}
}

func closeEndpoint(chanData *ChanData) network.Endpoint {
	return network.Endpoint{
		NeedsAuth: true,
		HandlerFunc: func(request network.Request) any {
			events := <-chanData.events
			defer func() {
				chanData.events <- events
			}()

			data := dto.EventClose{}
			request.GetJson(&data)

			for i, ev := range events {
				if ev.Id == data.EventId {
					if ev.Organizer.Id != request.Auth.(*types.User).Id {
						return nil
					}
					events[i].Open = false
					return dto.EventToDTO(events[i])
				}
			}
			return nil
		},
	}
}

func registerEndpoint(chanData *ChanData) network.Endpoint {
	return network.Endpoint{
		NeedsAuth: true,
		HandlerFunc: func(request network.Request) any {
			data := dto.EventRegister{}
			request.GetJson(&data)

			events := <-chanData.events
			defer func() {
				chanData.events <- events
			}()

			for _, ev := range events {
				if ev.Id == data.EventId {
					return ev.Register(request.Auth.(*types.User), data.JobId)
				}
			}
			return false
		},
	}
}
