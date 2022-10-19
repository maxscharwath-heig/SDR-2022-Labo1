package server

import (
	"fmt"
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

type ChanData struct {
	users  chan map[int]*types.User
	events chan []*types.Event
}

var enableCriticDebug = false

var stopServer = make(chan bool)

func Stop() {
	stopServer <- true
}

func Start(serverConfiguration *config.ServerConfiguration) {
	utils.SetEnabled(serverConfiguration.ShowInfosLogs)
	enableCriticDebug = serverConfiguration.Debug
	if enableCriticDebug {
		utils.LogInfo("Debug mode enabled")
	}

	l, err := net.Listen("tcp", serverConfiguration.FullUrl())
	if err != nil {
		utils.LogError("Error listening:", err.Error())
		os.Exit(1)
	}

	utils.LogSuccess("Listening on " + serverConfiguration.FullUrl())

	//init chan data structure
	chanData := ChanData{
		users:  make(chan map[int]*types.User, 1),
		events: make(chan []*types.Event, 1),
	}

	{ // LOAD DATA FROM CONFIG
		users, events := serverConfiguration.GetData()
		chanData.users <- users
		chanData.events <- events
	}

	protocol := network.ServerProtocol{
		AuthFunc: func(credential types.Credentials) (bool, network.AuthId) {
			users := <-chanData.users
			startCriticSection("AuthFunc")
			defer func() {
				endCriticSection("AuthFunc")
				chanData.users <- users
			}()
			if credential.Username == "" || credential.Password == "" {
				return false, -1
			}
			for _, user := range users {
				if user.Username == credential.Username && user.Password == credential.Password {
					return true, -1
				}
			}
			return false, -1
		},
		Endpoints: map[string]network.Endpoint{
			"create":   createEndpoint(&chanData),
			"show":     showEndpoint(&chanData),
			"close":    closeEndpoint(&chanData),
			"register": registerEndpoint(&chanData),
		},
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go protocol.Process(conn)
		}
	}()
	<-stopServer
	utils.LogInfo("Stopping server")
	_ = l.Close()
}

func createEndpoint(chanData *ChanData) network.Endpoint {
	return network.Endpoint{
		NeedsAuth: true,
		HandlerFunc: func(request network.Request) network.Response[any] {
			events := <-chanData.events
			startCriticSection("HandlerFunc(create)")
			defer func() {
				endCriticSection("HandlerFunc(create)")
				chanData.events <- events
			}()

			data := dto.EventCreate{}
			request.GetJson(&data)

			if data.Name == "" {
				return network.CreateResponse(false, "name is required")
			}

			event := &types.Event{
				Id:           len(events) + 1,
				Name:         data.Name,
				Open:         true,
				OrganizerId:  request.AuthId,
				Jobs:         make(map[int]*types.Job),
				Participants: make(map[int]int),
			}
			for i, job := range data.Jobs {
				id := i + 1

				if job.Capacity < 1 {
					return network.CreateResponse(false, "capacity must be greater than 0")
				}

				if job.Name == "" {
					return network.CreateResponse(false, "name is required")
				}

				event.Jobs[id] = &types.Job{
					Id:       id,
					Name:     job.Name,
					Capacity: job.Capacity,
				}
			}
			events = append(events, event)
			return network.CreateResponse(true, EventToDTO(event, chanData))
		},
	}
}

func showEndpoint(chanData *ChanData) network.Endpoint {
	return network.Endpoint{
		NeedsAuth: false,
		HandlerFunc: func(request network.Request) network.Response[any] {
			events := <-chanData.events
			startCriticSection("HandlerFunc(show)")
			defer func() {
				endCriticSection("HandlerFunc(show)")
				chanData.events <- events
			}()

			data := dto.EventShow{}
			request.GetJson(&data)

			if data.EventId != -1 {
				for _, ev := range events {
					if ev.Id == data.EventId {
						return network.CreateResponse(true, EventToDTO(ev, chanData))
					}
				}
				return network.CreateResponse(false, "event not found")
			}
			return network.CreateResponse(true, EventsToDTO(events, chanData))
		},
	}
}

func closeEndpoint(chanData *ChanData) network.Endpoint {
	return network.Endpoint{
		NeedsAuth: true,
		HandlerFunc: func(request network.Request) network.Response[any] {
			events := <-chanData.events
			startCriticSection("HandlerFunc(close)")
			defer func() {
				endCriticSection("HandlerFunc(close)")
				chanData.events <- events
			}()

			data := dto.EventClose{}
			request.GetJson(&data)

			for i, ev := range events {
				if ev.Id == data.EventId {
					if ev.OrganizerId != request.AuthId {
						return network.CreateResponse(false, "you are not the organizer")
					}
					events[i].Open = false
					return network.CreateResponse(true, EventToDTO(events[i], chanData))
				}
			}
			return network.CreateResponse(false, "event not found")
		},
	}
}

func registerEndpoint(chanData *ChanData) network.Endpoint {
	return network.Endpoint{
		NeedsAuth: true,
		HandlerFunc: func(request network.Request) network.Response[any] {
			data := dto.EventRegister{}
			request.GetJson(&data)

			events := <-chanData.events
			startCriticSection("HandlerFunc(register)")
			defer func() {
				endCriticSection("HandlerFunc(register)")
				chanData.events <- events
			}()

			for _, ev := range events {
				if ev.Id == data.EventId {
					if err := ev.Register(request.AuthId, data.JobId); err != nil {
						return network.CreateResponse(false, err.Error())
					}
					return network.CreateResponse(true, EventToDTO(ev, chanData))
				}
			}
			return network.CreateResponse(false, "event not found")
		},
	}
}

func delayer(sec time.Duration) {
	time.Sleep(time.Second * sec)
}

func startCriticSection(section string) {
	if !enableCriticDebug {
		return
	}
	utils.Log(true, "START CRITICAL SECTION", colors.BackgroundRed, section)
	delayer(5)
}

func endCriticSection(section string) {
	if !enableCriticDebug {
		return
	}
	utils.Log(true, "END CRITICAL SECTION", colors.BackgroundRed, section)
}

func getUserById(id int, chanData *ChanData) types.User {
	users := <-chanData.users
	startCriticSection(fmt.Sprintf("getUserById(%d)", id))
	defer func() {
		endCriticSection(fmt.Sprintf("getUserById(%d)", id))
		chanData.users <- users
	}()
	return *users[id]
}

// CONVERSIONS

func EventToDTO(event *types.Event, chanData *ChanData) dto.Event {
	var jobs []types.Job
	for _, job := range event.Jobs {
		jobs = append(jobs, *job)
	}
	participants := make([]dto.Participant, 0)
	for userId, jobId := range event.Participants {
		participants = append(participants, dto.Participant{
			User:  getUserById(userId, chanData),
			JobId: jobId,
		})
	}
	return dto.Event{
		Id:           event.Id,
		Name:         event.Name,
		Open:         event.Open,
		Jobs:         jobs,
		Organizer:    getUserById(event.OrganizerId, chanData),
		Participants: participants,
	}
}

func EventsToDTO(events []*types.Event, chanData *ChanData) []dto.Event {
	var dtoEvents []dto.Event
	for _, event := range events {
		dtoEvents = append(dtoEvents, EventToDTO(event, chanData))
	}
	return dtoEvents
}
