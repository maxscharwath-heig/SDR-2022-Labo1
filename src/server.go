package server

import (
	"net"
	"os"
	"sdr/labo1/src/config"
	"sdr/labo1/src/dto"
	"sdr/labo1/src/network"
	"sdr/labo1/src/types"
	"sdr/labo1/src/utils"
	"sdr/labo1/src/utils/colors"
	"time"
)

type chanData struct {
	users  chan []*types.User
	events chan []*types.Event
}

func Start(serverConfiguration *config.ServerConfiguration) {
	utils.SetEnabled(serverConfiguration.ShowInfosLogs)

	l, err := net.Listen("tcp", serverConfiguration.FullUrl())
	if err != nil {
		utils.LogError("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()

	utils.LogSuccess("Listening on " + serverConfiguration.FullUrl())

	//init chan data structure
	chanData := chanData{
		users:  make(chan []*types.User, 1),
		events: make(chan []*types.Event, 1),
	}

	{ // LOAD DATA FROM CONFIG
		users, events := serverConfiguration.GetData()
		chanData.users <- users
		chanData.events <- events
	}

	protocol := network.ServerProtocol{
		AuthFunc: func(credential types.Credentials) (bool, network.Auth) {
			users := <-chanData.users
			utils.Log("START CRITICAL SECTION", colors.Red, "AuthFunc")
			delayer(5)
			defer func() {
				utils.Log("END CRITICAL SECTION", colors.Red, "AuthFunc")
				chanData.users <- users
			}()
			if credential.Username == "" || credential.Password == "" {
				return false, nil
			}
			for _, user := range users {
				if user.Username == credential.Username && user.Password == credential.Password {
					return true, user
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
			utils.LogError("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go protocol.Process(conn)
	}
}

func createEndpoint(chanData *chanData) network.Endpoint {
	return network.Endpoint{
		NeedsAuth: true,
		HandlerFunc: func(request network.Request) any {
			events := <-chanData.events
			utils.Log("START CRITICAL SECTION", colors.Red, "HandlerFunc(create)")
			delayer(5)
			defer func() {
				utils.Log("END CRITICAL SECTION", colors.Red, "HandlerFunc(create)")
				chanData.events <- events
			}()

			data := dto.EventCreate{}
			request.GetJson(&data)

			event := &types.Event{
				Id:           len(events) + 1,
				Name:         data.Name,
				Open:         true,
				Organizer:    request.Auth,
				Jobs:         make(map[int]*types.Job),
				Participants: make(map[*types.User]*types.Job),
			}
			for i, job := range data.Jobs {
				id := i + 1
				event.Jobs[id] = &types.Job{
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

func showEndpoint(chanData *chanData) network.Endpoint {
	return network.Endpoint{
		NeedsAuth: false,
		HandlerFunc: func(request network.Request) any {
			events := <-chanData.events
			utils.Log("START CRITICAL SECTION", colors.Red, "HandlerFunc(show)")
			delayer(5)
			defer func() {
				utils.Log("END CRITICAL SECTION", colors.Red, "HandlerFunc(show)")
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

func closeEndpoint(chanData *chanData) network.Endpoint {
	return network.Endpoint{
		NeedsAuth: true,
		HandlerFunc: func(request network.Request) any {
			events := <-chanData.events
			utils.Log("START CRITICAL SECTION", colors.Red, "HandlerFunc(close)")
			delayer(5)
			defer func() {
				utils.Log("END CRITICAL SECTION", colors.Red, "HandlerFunc(close)")
				chanData.events <- events
			}()

			data := dto.EventClose{}
			request.GetJson(&data)

			for i, ev := range events {
				if ev.Id == data.EventId {
					if ev.Organizer.Id != request.Auth.Id {
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

func registerEndpoint(chanData *chanData) network.Endpoint {
	return network.Endpoint{
		NeedsAuth: true,
		HandlerFunc: func(request network.Request) any {
			data := dto.EventRegister{}
			request.GetJson(&data)

			events := <-chanData.events
			utils.Log("START CRITICAL SECTION", colors.Red, "HandlerFunc(register)")
			delayer(5)
			defer func() {
				utils.Log("END CRITICAL SECTION", colors.Red, "HandlerFunc(register)")
				chanData.events <- events
			}()

			for _, ev := range events {
				if ev.Id == data.EventId {
					return ev.Register(request.Auth, data.JobId)
				}
			}
			return false
		},
	}
}

func delayer(sec time.Duration) {
	time.Sleep(time.Second * sec)
}
