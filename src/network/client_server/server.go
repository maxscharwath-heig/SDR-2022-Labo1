// SDR - Labo 2
// Nicolas Crausaz & Maxime Scharwath

package client_server

import (
	"fmt"
	"io"
	"net"
	"sdr/labo1/src/network"
	"sdr/labo1/src/types"
	"sdr/labo1/src/utils"
)

// ServerEndpoint extends the Endpoint struct with a function that is called when the endpoint is called.
type ServerEndpoint = Endpoint[HeaderResponse]

// AuthId
// The type of the auth object
type AuthId = int

// HeaderResponse
// is the first response of the server.
// - Valid: true if the endpoint is valid
// - NeedsAuth: true if the endpoint needs authentication
type HeaderResponse struct {
	Valid     bool   `json:"valid"`
	NeedsAuth bool   `json:"needsAuth"`
	AuthId    AuthId `json:"-"`
}

type pendingRequest struct {
	name     string
	callback func()
}

// ServerProtocol
// is the protocol that is used to handle the server side of the protocol.
// - AuthFunc: the function that is called to authenticate the user.
// - Endpoints: the endpoints that are registered. It is a map of the endpointId and the endpoint.
type ServerProtocol struct {
	AuthFunc               AuthFunc
	Endpoints              map[string]ServerEndpoint
	pendingRequest         chan pendingRequest
	pendingPriorityRequest chan pendingRequest
}

func CreateServerProtocol(authFunc AuthFunc) ServerProtocol {
	return ServerProtocol{
		AuthFunc:               authFunc,
		Endpoints:              make(map[string]ServerEndpoint),
		pendingRequest:         make(chan pendingRequest),
		pendingPriorityRequest: make(chan pendingRequest),
	}
}

func (p ServerProtocol) AddEndpoint(endpointId string, endpoint ServerEndpoint) {
	p.Endpoints[endpointId] = endpoint
}

func (p ServerProtocol) ProcessPriorityRequests() {
	for {
		select {
		case pending := <-p.pendingPriorityRequest:
			utils.CreateCriticalSection(fmt.Sprintf("sync priority %s", pending.name), pending.callback)
		default:
			return
		}
	}
}

func (p ServerProtocol) ProcessRequests() {
	for {
		select {
		case pending := <-p.pendingRequest:
			p.ProcessPriorityRequests()
			utils.CreateCriticalSection(fmt.Sprintf("sync %s", pending.name), pending.callback)
		default:
			p.ProcessPriorityRequests()
		}
	}
}

func (p ServerProtocol) AddPending(name string, priority bool, callback func()) {
	if priority {
		p.pendingPriorityRequest <- pendingRequest{name, callback}
	} else {
		p.pendingRequest <- pendingRequest{name, callback}
	}
}

// HandleConnection is the function that is called to process the connection. It is called in a go routine.
func (p ServerProtocol) HandleConnection(c net.Conn) {
	utils.LogInfo(false, "new connection", c.RemoteAddr())
	defer func() {
		utils.LogInfo(true, "close connection", c.RemoteAddr())
		_ = c.Close()
	}()

	conn := network.CreateConnection(c)
	var err error
	ready := make(chan struct{}, 1)
	ready <- struct{}{} // A connection can handle one request at a time
	for {
		if conn.IsClosed() || err == io.EOF {
			utils.LogInfo(false, "connection closed", c.RemoteAddr())
			break
		}

		select {
		case <-ready:

			request := network.Request[HeaderResponse]{Conn: c}

			request.EndpointId, err = conn.GetLine()
			if err != nil {
				utils.LogInfo(false, "error while receiving endpointId", err)
				continue
			}

			endpoint, ok := p.Endpoints[request.EndpointId]
			if ok {
				request.Header.Valid = true
				request.Header.NeedsAuth = endpoint.NeedsAuth
			}

			err = conn.SendJSON(request.Header)
			if err != nil {
				utils.LogWarning(false, "error while sending header", err)
				continue
			}

			if !request.Header.Valid {
				utils.LogWarning(false, "invalid endpoint, canceling request")
				continue
			}

			go p.AddPending(fmt.Sprintf("Request %s (auth)", request.EndpointId), false, func() {
				if request.Header.NeedsAuth {
					var credentials types.Credentials

					if e := conn.GetJson(&credentials); e != nil {
						utils.LogWarning(false, "error while receiving credentials", e)
						ready <- struct{}{}
						return
					}

					isValid, auth := p.AuthFunc(credentials)

					if e := conn.SendJSON(AuthResponse{Success: isValid, Auth: auth}); e != nil {
						utils.LogWarning(false, "error while sending auth response", e)
						ready <- struct{}{}
						return
					}

					request.Header.AuthId = auth
					if !isValid {
						utils.LogWarning(false, "invalid credentials, canceling request")
						ready <- struct{}{}
						return
					}
				}
				go p.AddPending(fmt.Sprintf("Request %s (data)", request.EndpointId), false, func() {
					defer func() {
						ready <- struct{}{}
					}()
					if data, e := conn.GetLine(); e != nil {
						utils.LogWarning(false, "error while receiving data", e)
						return
					} else {
						request.Data = data
					}

					response := endpoint.HandlerFunc(request)

					if e := conn.SendJSON(response); e != nil {
						utils.LogWarning(false, "error while sending response", e)
						return
					}
				})
			})
		}
	}
}
