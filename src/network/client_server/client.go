package client_server

import (
	"fmt"
	"net"
	"sdr/labo1/src/network"
	"sdr/labo1/src/types"
)

// ClientProtocol
// is the protocol that is used to handle the client side of the protocol.
// - Conn: the connection that is used to communicate with the server.
// - AuthFunc: the function that is called to authenticate the user. Need to return the credentials.
type ClientProtocol struct {
	Conn     net.Conn
	AuthFunc func() types.Credentials
	conn     network.Connection
}

// CreateClientProtocol Constructor
func CreateClientProtocol(conn net.Conn, authFunc func() types.Credentials) *ClientProtocol {
	return &ClientProtocol{
		Conn:     conn,
		AuthFunc: authFunc,
		conn:     *network.CreateConnection(conn),
	}
}

// SendRequest
// Send a request to the server
//   - endpointId: the endpointId of the endpoint that should be called
//   - data: the function that is called after the response is received and the authentication is done
//     The function returns the response of the endpoint.
func (p ClientProtocol) SendRequest(endpointId string, data func(auth AuthId) any) (response string, err error) {
	err = p.conn.SendData(endpointId)

	if err != nil {
		return
	}

	var header HeaderResponse
	err = p.conn.GetJson(&header)

	if err != nil {
		return
	}

	if !header.Valid {
		return "", fmt.Errorf("invalid endpoint")
	}
	authResponse := AuthResponse{}
	if header.NeedsAuth {
		err = p.conn.SendJSON(p.AuthFunc())
		if err != nil {
			return
		}

		err = p.conn.GetJson(&authResponse)

		if err != nil {
			return
		}

		if !authResponse.Success {
			return "", fmt.Errorf("invalid credentials")
		}
	}
	err = p.conn.SendJSON(data(authResponse.Auth))
	if err != nil {
		return
	}

	return p.conn.GetLine()
}

// Close close client connexion
func (p ClientProtocol) Close() error {
	return p.Conn.Close()
}

// IsClosed check client's connection status
func (p ClientProtocol) IsClosed() bool {
	return p.conn.IsClosed()
}

// OnClose Execute handler on client connexion close
func (p ClientProtocol) OnClose(handler func()) {
	go func() {
		for {
			if p.IsClosed() {
				handler()
				break
			}
		}
	}()
}

func (p ClientProtocol) Handshake() bool {
	return p.conn.SendHandshake("client_server")
}
