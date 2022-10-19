// Package network
// This package contains the network protocol for the client and server
// The protocol is a simple JSON based protocol
// The protocol is used to send requests to the server and receive responses
// Example of how the protocol works:
// - Client sends a endpointId to the server
// - Server sends a HeaderResponse to the client with information about the endpoint
// - If the endpoint needs authentication, the client sends the credentials to the server
// - The server sends a AuthResponse to the client with information about the authentication
// - If the authentication is successful or the endpoint doesn't need authentication, the client sends the data to the server
// - The server sends the response to the client
package network

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sdr/labo1/src/types"
	"sdr/labo1/src/utils"
	"strings"
)

// Auth
// The type of the auth object
type Auth = *types.User

// HeaderResponse
// is the first response of the server.
// - Valid: true if the endpoint is valid
// - NeedsAuth: true if the endpoint needs authentication
type HeaderResponse struct {
	Valid     bool
	NeedsAuth bool
}

// AuthResponse
// is the response of the server after the authentication.
// - Success: true if the authentication was successful
// - Auth: the authentication data ( see: type Auth )
type AuthResponse struct {
	Success bool
	Auth    Auth
}

// AuthFunc
// is the function that is called to authenticate the user.
// Returns true if the authentication was successful and the authentication data. ( see: type Auth )
type AuthFunc func(credentials types.Credentials) (bool, Auth)

// Request is the request struct
type Request struct {
	EndpointId string
	Header     HeaderResponse
	Auth       Auth
	Data       string
}

func (r Request) GetJson(data any) {
	_ = json.Unmarshal([]byte(r.Data), data)
}

// Endpoint
// is the endpoint struct that is used to register an endpoint.
//   - NeedsAuth: true if the endpoint needs authentication
//   - HandlerFunc: the function that is called after the request is received and the authentication is done.
//     The function returns the response of the endpoint.
type Endpoint struct {
	NeedsAuth   bool
	HandlerFunc func(request Request) any
}

// connection
// is used to handle the connection and create a wrapper around it.
type connection struct {
	conn   net.Conn
	reader *bufio.Reader
}

// Determine if the connection is closed
func (c connection) isClosed() bool {
	_, err := c.conn.Read(make([]byte, 0))
	return err != nil
}

// Send raw data to the connection
func (c connection) sendData(data string) error {
	utils.LogInfo("send", data)
	_, err := fmt.Fprintln(c.conn, data)
	return err
}

// Send data as JSON to the connection
func (c connection) sendJSON(data any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.sendData(string(bytes))
}

// Get a line from the connection ( without the newline )
func (c connection) getLine() (string, error) {
	data, err := c.reader.ReadString('\n')
	data = strings.Trim(data, "\n")
	utils.LogInfo("recv", data, err)
	return data, err
}

// Get a line from the connection and parse it as JSON
func (c connection) getJson(data any) error {
	jsonString, err := c.getLine()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(jsonString), data)
}

// get HeaderResponse from the connection
func (c connection) getHeader() (HeaderResponse, error) {
	var header HeaderResponse
	err := c.getJson(&header)
	return header, err
}

// ServerProtocol
// is the protocol that is used to handle the server side of the protocol.
// - AuthFunc: the function that is called to authenticate the user.
// - Endpoints: the endpoints that are registered. It is a map of the endpointId and the endpoint.
type ServerProtocol struct {
	AuthFunc  AuthFunc
	Endpoints map[string]Endpoint
}

// Process
// is the function that is called to process the connection. It is called in a go routine.
func (p ServerProtocol) Process(c net.Conn) {
	utils.LogInfo("new connection", c.RemoteAddr())
	defer func() {
		utils.LogInfo("close connection", c.RemoteAddr())
		_ = c.Close()
	}()

	conn := connection{
		conn:   c,
		reader: bufio.NewReader(c),
	}
	for {
		if conn.isClosed() {
			break
		}
		var err error
		request := Request{}
		request.EndpointId, err = conn.getLine()
		if err != nil {
			utils.LogError("error while receiving endpointId", err)
			continue
		}

		endpoint, ok := p.Endpoints[request.EndpointId]
		if ok {
			request.Header.Valid = true
			request.Header.NeedsAuth = endpoint.NeedsAuth
		}

		err = conn.sendJSON(request.Header)
		if err != nil {
			utils.LogError("error while sending header", err)
			continue
		}

		if !request.Header.Valid {
			utils.LogError("invalid endpoint, canceling request")
			continue
		}

		if request.Header.NeedsAuth {
			var credentials types.Credentials

			err = conn.getJson(&credentials)
			if err != nil {
				utils.LogError("error while receiving credentials", err)
				continue
			}

			isValid, auth := p.AuthFunc(credentials)

			err = conn.sendJSON(AuthResponse{isValid, auth})
			if err != nil {
				utils.LogError("error while sending auth response", err)
				continue
			}

			request.Auth = auth
			if !isValid {
				utils.LogError("invalid credentials, canceling request")
				continue
			}
		}
		request.Data, err = conn.getLine()
		if err != nil {
			utils.LogError("error while receiving data", err)
			continue
		}

		response := endpoint.HandlerFunc(request)
		err = conn.sendJSON(response)
		if err != nil {
			utils.LogError("error while sending response", err)
			continue
		}
	}
}

// ClientProtocol
// is the protocol that is used to handle the client side of the protocol.
// - Conn: the connection that is used to communicate with the server.
// - AuthFunc: the function that is called to authenticate the user. Need to return the credentials.
type ClientProtocol struct {
	Conn     net.Conn
	AuthFunc func() types.Credentials
	conn     connection
}

// CreateClientProtocol Constructor
func CreateClientProtocol(conn net.Conn, authFunc func() types.Credentials) *ClientProtocol {
	c := connection{conn: conn, reader: bufio.NewReader(conn)}
	return &ClientProtocol{
		Conn:     conn,
		AuthFunc: authFunc,
		conn:     c,
	}
}

// SendRequest
// Send a request to the server
// - endpointId: the endpointId of the endpoint that should be called
// - data: the function that is called after the response is received and the authentication is done.
func (p ClientProtocol) SendRequest(endpointId string, data func(auth Auth) any) (response string, err error) {
	err = p.conn.sendData(endpointId)

	if err != nil {
		return
	}

	header, err := p.conn.getHeader()

	if err != nil {
		return
	}

	if !header.Valid {
		return "", fmt.Errorf("invalid endpoint")
	}
	authResponse := AuthResponse{}
	if header.NeedsAuth {
		err = p.conn.sendJSON(p.AuthFunc())
		if err != nil {
			return
		}

		err = p.conn.getJson(&authResponse)

		if err != nil {
			return
		}

		if !authResponse.Success {
			return "", fmt.Errorf("invalid credentials")
		}
	}
	err = p.conn.sendJSON(data(authResponse.Auth))
	if err != nil {
		return
	}

	return p.conn.getLine()
}

func (p ClientProtocol) Close() error {
	return p.Conn.Close()
}

func (p ClientProtocol) IsClosed() bool {
	return p.conn.isClosed()
}
