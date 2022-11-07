// SDR - Labo 1
// Nicolas Crausaz & Maxime Scharwath

// Package client_server
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
package client_server

import (
	"sdr/labo1/src/network"
	"sdr/labo1/src/types"
)

// AuthResponse
// is the response of the server after the authentication.
// - Success: true if the authentication was successful
// - AuthId: the authentication data ( see: type AuthId )
type AuthResponse struct {
	Success bool   `json:"success"`
	Auth    AuthId `json:"auth"`
}

// AuthFunc
// is the function that is called to authenticate the user.
// Returns true if the authentication was successful and the authentication data. (see: type AuthId)
type AuthFunc func(credentials types.Credentials) (bool, AuthId)

// Endpoint
// is the endpoint struct that is used to register an endpoint.
//   - NeedsAuth: true if the endpoint needs authentication
//   - HandlerFunc: the function that is called after the request is received and the authentication is done.
//     The function returns the response of the endpoint.
type Endpoint[T any] struct {
	NeedsAuth   bool
	HandlerFunc func(request network.Request[T]) network.Response[any]
}
