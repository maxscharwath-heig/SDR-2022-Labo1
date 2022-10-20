// SDR - Labo 1
// Nicolas Crausaz & Maxime Scharwath

package server

import (
	"encoding/hex"
	"fmt"
	"math/rand"
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
	utils.LogInfo("debug mode", enableCriticDebug)

	l, err := net.Listen("tcp", serverConfiguration.FullUrl())
	if err != nil {
		utils.LogError("Error listening:", err.Error())
		os.Exit(1)
	}

	utils.LogSuccess("Server started", serverConfiguration.FullUrl())

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
			start, end := createCriticalSection("users", "AuthFunc")
			users := <-chanData.users
			start()
			defer func() {
				end()
				chanData.users <- users
			}()
			if credential.Username == "" || credential.Password == "" {
				return false, -1
			}
			for _, user := range users {
				if user.Username == credential.Username && user.Password == credential.Password {
					return true, user.Id
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
			data := dto.EventCreate{}
			request.GetJson(&data)

			if data.Name == "" {
				return network.CreateResponse(false, "name is required")
			}

			event := &types.Event{
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
			start, end := createCriticalSection("events", "HandlerFunc(create)")
			events := <-chanData.events
			start()
			defer func() {
				end()
				chanData.events <- events
			}()
			event.Id = len(events) + 1
			events = append(events, event)
			return network.CreateResponse(true, EventToDTO(event, chanData))
		},
	}
}

func showEndpoint(chanData *ChanData) network.Endpoint {
	return network.Endpoint{
		NeedsAuth: false,
		HandlerFunc: func(request network.Request) network.Response[any] {
			data := dto.EventShow{}
			request.GetJson(&data)

			start, end := createCriticalSection("events", "HandlerFunc(show)")
			events := <-chanData.events
			start()
			defer func() {
				end()
				chanData.events <- events
			}()

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
			data := dto.EventClose{}
			request.GetJson(&data)

			start, end := createCriticalSection("events", "HandlerFunc(close)")
			events := <-chanData.events
			start()
			defer func() {
				end()
				chanData.events <- events
			}()

			for i, ev := range events {
				if ev.Id == data.EventId {
					if ev.OrganizerId != request.AuthId {
						return network.CreateResponse(false, "you are not the organizer")
					}
					if !ev.Open {
						return network.CreateResponse(false, "event already closed")
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

			start, end := createCriticalSection("events", "HandlerFunc(register)")
			events := <-chanData.events
			start()
			defer func() {
				end()
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

func createCriticalSection(chanName string, name string) (start func(), end func()) {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	id := hex.EncodeToString(b)

	start = func() {
		if !enableCriticDebug {
			return
		}
		utils.Log(true, fmt.Sprintf("CRITIC START [%s]", id), colors.BackgroundRed, fmt.Sprintf("🔒%s\t%s", chanName, name))
		delayer(5)
	}
	end = func() {
		if !enableCriticDebug {
			return
		}
		utils.Log(true, fmt.Sprintf("CRITIC END   [%s]", id), colors.BackgroundRed, fmt.Sprintf("🔓%s\t%s", chanName, name))
	}
	return
}

func getUserById(id int, chanData *ChanData) types.User {
	start, end := createCriticalSection("users", fmt.Sprintf("getUserById(%d)", id))
	users := <-chanData.users
	start()
	defer func() {
		end()
		chanData.users <- users
	}()
	if user, ok := users[id]; ok {
		return *user
	}
	return types.User{}
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
