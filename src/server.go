// SDR - Labo 2
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
	"sdr/labo1/src/network/client_server"
	"sdr/labo1/src/network/lamport"
	"sdr/labo1/src/network/server_server"
	"sdr/labo1/src/types"
	"sdr/labo1/src/utils"
	"sdr/labo1/src/utils/colors"
	"time"
)

// ChanData Defines the channels used to access concurrency critical data
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
	utils.LogInfo(true, "debug mode", enableCriticDebug)

	listenerServer, err := net.Listen("tcp", serverConfiguration.GetCurrentUrls().Server)
	if err != nil {
		utils.LogError(true, "Error listening:", err.Error())
		os.Exit(1)
	}

	interServerProtocol := server_server.CreateInterServerProtocol[lamport.Request](serverConfiguration.Id, listenerServer)

	interServerProtocol.ConnectToServers(serverConfiguration.GetOtherServers())

	// [AT THIS POINT, THE SERVER IS CONNECTED TO ALL OTHER SERVERS]

	lmpt := lamport.InitLamport(interServerProtocol)

	go lmpt.Start() // Start listening to Lamport Messages

	// init chan data structure
	chanData := ChanData{
		users:  make(chan map[int]*types.User, 1),
		events: make(chan []*types.Event, 1),
	}

	{ // Load configuration
		users, events := serverConfiguration.GetData()
		chanData.users <- users
		chanData.events <- events
	}

	listenerClient, err := net.Listen("tcp", serverConfiguration.GetCurrentUrls().Client)
	if err != nil {
		utils.LogError(true, "Error listening:", err.Error())
		os.Exit(1)
	}

	utils.LogSuccess(true, "Server started", serverConfiguration.GetCurrentUrls().Client)

	protocol := client_server.ServerProtocol{
		AuthFunc: func(credential types.Credentials) (success bool, userId client_server.AuthId) {
			start, end := createCriticalSection("users", "AuthFunc")

			select {
			case users := <-chanData.users:
				start()

				defer func() {
					end()
					chanData.users <- users
				}()

				if credential.Username == "" || credential.Password == "" {
					success, userId = false, -1
					return
				}
				for _, user := range users {
					if user.Username == credential.Username && user.Password == credential.Password {
						success, userId = true, user.Id
						return
					}
				}

				return false, -1
			}
		},
		Endpoints: map[string]client_server.ServerEndpoint{
			"create":   createEndpoint(&chanData, &lmpt),
			"show":     showEndpoint(&chanData, &lmpt),
			"close":    closeEndpoint(&chanData, &lmpt),
			"register": registerEndpoint(&chanData, &lmpt),
		},
	}

	go func() {
		for {
			conn, err := listenerClient.Accept()
			if err != nil {
				return
			}
			go protocol.Process(conn)
		}
	}()
	<-stopServer
	utils.LogInfo(true, "Stopping server")
	_ = listenerClient.Close()
}

type request = network.Request[client_server.HeaderResponse]

// createEndpoint Registers a custom endpoint accessible on the server
func createEndpoint(chanData *ChanData, lmpt *lamport.Lamport) client_server.ServerEndpoint {
	return client_server.ServerEndpoint{
		NeedsAuth: true,
		HandlerFunc: func(request request) network.Response[any] {
			data := dto.EventCreate{}
			request.GetJson(&data)

			if data.Name == "" {
				return network.CreateResponse(false, "name is required")
			}

			event := &types.Event{
				Name:         data.Name,
				Open:         true,
				OrganizerId:  request.Header.AuthId,
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
			<-lmpt.SendClientAskCriticalSection()
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

// showEndpoint defines an endpoint that displays events
func showEndpoint(chanData *ChanData, lmpt *lamport.Lamport) client_server.ServerEndpoint {
	return client_server.ServerEndpoint{
		NeedsAuth: false,
		HandlerFunc: func(request request) (res network.Response[any]) {
			data := dto.EventShow{}
			request.GetJson(&data)

			start, end := createCriticalSection("events", "HandlerFunc(show)")

			select {
			case events := <-chanData.events:
				start()
				<-lmpt.SendClientAskCriticalSection()
				if data.EventId != -1 {
					for _, ev := range events {
						if ev.Id == data.EventId {
							end()
							chanData.events <- events
							return network.CreateResponse(true, EventToDTO(ev, chanData))
						}
					}
					end()
					chanData.events <- events
					return network.CreateResponse(false, "event not found")
				}
				end()
				chanData.events <- events
				return network.CreateResponse(true, EventsToDTO(events, chanData))

			case <-time.After(2 * time.Second):
				return network.CreateResponse(false, "timeout")
			}
		},
	}
}

// closeEndpoint defines an endpoint that closes events
func closeEndpoint(chanData *ChanData, lmpt *lamport.Lamport) client_server.ServerEndpoint {
	return client_server.ServerEndpoint{
		NeedsAuth: true,
		HandlerFunc: func(request request) (res network.Response[any]) {
			data := dto.EventClose{}
			request.GetJson(&data)

			start, end := createCriticalSection("events", "HandlerFunc(close)")

			select {
			case events := <-chanData.events:
				start()
				<-lmpt.SendClientAskCriticalSection()
				for i, ev := range events {
					if ev.Id == data.EventId {
						if ev.OrganizerId != request.Header.AuthId {
							res = network.CreateResponse(false, "you are not the organizer")
							break
						}
						if !ev.Open {
							res = network.CreateResponse(false, "event already closed")
							break
						}
						events[i].Open = false
						res = network.CreateResponse(true, EventToDTO(events[i], chanData))
					}
				}
				end()
				chanData.events <- events
				if &res == nil {
					res = network.CreateResponse(false, "event not found")
				}
				return
			case <-time.After(80 * time.Millisecond):
				return network.CreateResponse(false, "timeout")
			}
		},
	}
}

// registerEndpoint defines an endpoint that register user to events
func registerEndpoint(chanData *ChanData, lmpt *lamport.Lamport) client_server.ServerEndpoint {
	return client_server.ServerEndpoint{
		NeedsAuth: true,
		HandlerFunc: func(request request) (res network.Response[any]) {
			data := dto.EventRegister{}
			request.GetJson(&data)

			start, end := createCriticalSection("events", "HandlerFunc(register)")

			select {
			case events := <-chanData.events:
				start()
				<-lmpt.SendClientAskCriticalSection()
				for _, ev := range events {
					if ev.Id == data.EventId {
						if err := ev.Register(request.Header.AuthId, data.JobId); err != nil {
							res = network.CreateResponse(false, err.Error())
							break
						}
						res = network.CreateResponse(true, EventToDTO(ev, chanData))
						break
					}
				}
				end()
				chanData.events <- events
				if &res == nil {
					res = network.CreateResponse(false, "event not found")
				}
				return

			case <-time.After(80 * time.Millisecond):
				return network.CreateResponse(false, "timeout")
			}
		},
	}
}

// delayer delay execution by seconds
func delayer(sec time.Duration) {
	time.Sleep(time.Second * sec)
}

// createCriticalSection access a critical section (for debug)
func createCriticalSection(chanName string, name string) (start func(), end func()) {
	if !enableCriticDebug {
		return func() {}, func() {}
	}

	// Generate an identifier the critical section
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	id := hex.EncodeToString(b)

	start = func() {
		utils.Log(true, fmt.Sprintf("CRITIC START [%s]", id), colors.BackgroundRed, fmt.Sprintf("🔒%s\t%s", chanName, name))
		delayer(5)
	}
	end = func() {
		utils.Log(true, fmt.Sprintf("CRITIC END   [%s]", id), colors.BackgroundRed, fmt.Sprintf("🔓%s\t%s", chanName, name))
	}
	return
}

// getUserById find and return and user in the user database
func getUserById(id int, chanData *ChanData) types.User {
	start, end := createCriticalSection("users", fmt.Sprintf("getUserById(%d)", id))

	select {
	case users := <-chanData.users:
		start()
		chanData.users <- users
		end()
		if user, ok := users[id]; ok {
			return *user
		}
	}
	return types.User{}
}

// EventToDTO transforms an event to protocol's transmissible data
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

// EventsToDTO transforms events to protocol's transmissible data
func EventsToDTO(events []*types.Event, chanData *ChanData) []dto.Event {
	var dtoEvents []dto.Event
	for _, event := range events {
		dtoEvents = append(dtoEvents, EventToDTO(event, chanData))
	}
	return dtoEvents
}
