// SDR - Labo 2
// Nicolas Crausaz & Maxime Scharwath

package server

import (
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
)

// Data Defines the channels used to access concurrency critical data
type Data struct {
	users  map[int]*types.User
	events []*types.Event
}

var stopServer = make(chan bool)

func Stop() {
	stopServer <- true
}

func Start(serverConfiguration *config.ServerConfiguration) {
	utils.SetEnabled(serverConfiguration.ShowInfosLogs)
	utils.SetCriticDebug(serverConfiguration.Debug)
	utils.LogInfo(true, "debug mode", serverConfiguration.Debug)

	listenerServer, err := net.Listen("tcp", serverConfiguration.GetCurrentUrls().Server)
	if err != nil {
		utils.LogError(true, "Error listening:", err.Error())
		os.Exit(1)
	}

	interServerProtocol := server_server.CreateInterServerProtocol[lamport.Request[[]dto.Event]](serverConfiguration.Id, listenerServer)

	interServerProtocol.ConnectToServers(serverConfiguration.GetOtherServers())

	// [AT THIS POINT, THE SERVER IS CONNECTED TO ALL OTHER SERVERS]

	// init chan data structure
	appData := Data{
		users: make(map[int]*types.User),
	}

	lmpt := lamport.InitLamport[[]dto.Event](interServerProtocol)

	go lmpt.Start() // Start listening to Lamport Messages

	{ // Load configuration
		users, events := serverConfiguration.GetData()
		appData.users = users
		appData.events = events
	}

	listenerClient, err := net.Listen("tcp", serverConfiguration.GetCurrentUrls().Client)
	if err != nil {
		utils.LogError(true, "Error listening:", err.Error())
		os.Exit(1)
	}

	utils.LogSuccess(true, "Server started", serverConfiguration.GetCurrentUrls().Client)

	protocol := client_server.CreateServerProtocol(
		func(credential types.Credentials) (success bool, userId client_server.AuthId) {
			if credential.Username == "" || credential.Password == "" {
				success, userId = false, -1
				return
			}
			for _, user := range appData.users {
				if user.Username == credential.Username && user.Password == credential.Password {
					success, userId = true, user.Id
					return
				}
			}

			return false, -1
		},
		map[string]client_server.ServerEndpoint{
			"create":   createEndpoint(&appData, &lmpt),
			"show":     showEndpoint(&appData),
			"close":    closeEndpoint(&appData, &lmpt),
			"register": registerEndpoint(&appData, &lmpt),
		},
	)

	go func() {
		for {
			conn, err := listenerClient.Accept()
			if err != nil {
				return
			}
			go protocol.HandleConnection(conn)
		}
	}()

	go func() {
		for {
			select {
			case data := <-lmpt.Data:
				protocol.AddPending("UpdateData", func() {
					appData.events = DTOToEvents(data)
					utils.LogInfo(false, "Lamport callback called")
				})
			}
		}
	}()

	go protocol.ProcessRequests()
	<-stopServer
	utils.LogInfo(true, "Stopping server")
	_ = listenerClient.Close()
	_ = listenerServer.Close()
}

type request = network.Request[client_server.HeaderResponse]

// createEndpoint Registers a custom endpoint accessible on the server
func createEndpoint(appData *Data, lmpt *lamport.Lamport[[]dto.Event]) client_server.ServerEndpoint {
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
			events := appData.events
			defer func() {
				lmpt.SendClientReleaseCriticalSection(EventsToDTO(events, appData))
			}()
			select {
			case <-lmpt.SendClientAskCriticalSection():
				event.Id = len(events) + 1
				events = append(events, event)
				return network.CreateResponse(true, EventToDTO(event, appData))
			}
		},
	}
}

// showEndpoint defines an endpoint that displays events
func showEndpoint(appData *Data) client_server.ServerEndpoint {
	return client_server.ServerEndpoint{
		NeedsAuth: false,
		HandlerFunc: func(request request) network.Response[any] {
			data := dto.EventShow{}
			request.GetJson(&data)

			events := appData.events
			if data.EventId != -1 {
				for _, ev := range events {
					if ev.Id == data.EventId {
						return network.CreateResponse(true, EventToDTO(ev, appData))
					}
				}
				return network.CreateResponse(false, "event not found")
			}
			return network.CreateResponse(true, EventsToDTO(events, appData))
		},
	}
}

// closeEndpoint defines an endpoint that closes events
func closeEndpoint(appData *Data, lmpt *lamport.Lamport[[]dto.Event]) client_server.ServerEndpoint {
	return client_server.ServerEndpoint{
		NeedsAuth: true,
		HandlerFunc: func(request request) network.Response[any] {
			data := dto.EventClose{}
			request.GetJson(&data)

			events := appData.events
			defer func() {
				lmpt.SendClientReleaseCriticalSection(EventsToDTO(events, appData))
			}()
			select {
			case <-lmpt.SendClientAskCriticalSection():
				for i, ev := range events {
					if ev.Id == data.EventId {
						if ev.OrganizerId != request.Header.AuthId {
							return network.CreateResponse(false, "you are not the organizer")
						}
						if !ev.Open {
							return network.CreateResponse(false, "event already closed")
						}
						events[i].Open = false
						return network.CreateResponse(true, EventToDTO(events[i], appData))
					}
				}
				return network.CreateResponse(false, "event not found")
			}
		},
	}
}

// registerEndpoint defines an endpoint that register user to events
func registerEndpoint(appData *Data, lmpt *lamport.Lamport[[]dto.Event]) client_server.ServerEndpoint {
	return client_server.ServerEndpoint{
		NeedsAuth: true,
		HandlerFunc: func(request request) network.Response[any] {
			data := dto.EventRegister{}
			request.GetJson(&data)

			events := appData.events
			defer func() {
				lmpt.SendClientReleaseCriticalSection(EventsToDTO(events, appData))
			}()
			select {
			case <-lmpt.SendClientAskCriticalSection():
				for _, ev := range events {
					if ev.Id == data.EventId {
						if err := ev.Register(request.Header.AuthId, data.JobId); err != nil {
							return network.CreateResponse(false, err.Error())
						}
						return network.CreateResponse(true, EventToDTO(ev, appData))
					}
				}
				return network.CreateResponse(false, "event not found")
			}
		},
	}
}

// getUserById find and return and user in the user database
func getUserById(id int, appData *Data) types.User {
	users := appData.users
	if user, ok := users[id]; ok {
		return *user
	}
	return types.User{}
}

// EventToDTO transforms an event to protocol's transmissible data
func EventToDTO(event *types.Event, appData *Data) dto.Event {
	var jobs []types.Job
	for _, job := range event.Jobs {
		jobs = append(jobs, *job)
	}
	participants := make([]dto.Participant, 0)
	for userId, jobId := range event.Participants {
		participants = append(participants, dto.Participant{
			User:  getUserById(userId, appData),
			JobId: jobId,
		})
	}
	return dto.Event{
		Id:           event.Id,
		Name:         event.Name,
		Open:         event.Open,
		Jobs:         jobs,
		Organizer:    getUserById(event.OrganizerId, appData),
		Participants: participants,
	}
}

// EventsToDTO transforms events to protocol's transmissible data
func EventsToDTO(events []*types.Event, appData *Data) []dto.Event {
	var dtoEvents []dto.Event
	for _, event := range events {
		dtoEvents = append(dtoEvents, EventToDTO(event, appData))
	}
	return dtoEvents
}

func DTOToEvent(data dto.Event) *types.Event {
	jobs := make(map[int]*types.Job)

	for _, job := range data.Jobs {
		jobs[job.Id] = &job
	}
	participants := make(map[int]int)
	for _, participant := range data.Participants {
		participants[participant.User.Id] = participant.JobId
	}
	return &types.Event{
		Id:           data.Id,
		Name:         data.Name,
		Open:         data.Open,
		Jobs:         jobs,
		OrganizerId:  data.Organizer.Id,
		Participants: participants,
	}
}

func DTOToEvents(dtoEvents []dto.Event) []*types.Event {
	var events []*types.Event
	for _, dtoEvent := range dtoEvents {
		events = append(events, DTOToEvent(dtoEvent))
	}
	return events
}
