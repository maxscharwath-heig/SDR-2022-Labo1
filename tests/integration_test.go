package tests

import (
	"fmt"
	"net"
	server "sdr/labo1/src"
	"sdr/labo1/src/config"
	"sdr/labo1/src/dto"
	"sdr/labo1/src/network"
	"sdr/labo1/src/types"
	"testing"
)

type cliTest struct {
	description string
	test        func(func() types.Credentials) bool
	credentials func() types.Credentials
}

var tests []cliTest

func connect(addr string) (*net.TCPConn, error) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)

	return conn, err
}

func TestSuccess(t *testing.T) {
	validSrvConfig := config.ServerConfiguration{
		Host: "localhost",
		Port: 9001,
	}

	validClientConfig := config.ClientConfiguration{
		Host: "localhost",
		Port: 9001,
	}

	tests = []cliTest{
		{
			description: "Should connect to server",
			test: func(creds func() types.Credentials) bool {

				server.Start(&validSrvConfig)
				conn, err := connect(validClientConfig.FullUrl())
				conn.Close()

				return err == nil
			},
			credentials: func() types.Credentials {
				return types.Credentials{
					Username: "user1",
					Password: "pass1",
				}
			},
		},
		{
			description: "Should connect create event",
			test: func(creds func() types.Credentials) bool {
				server.Start(&validSrvConfig)
				conn, err := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, creds)

				response, err := cli.SendRequest("create", func(auth network.Auth) any {
					return dto.EventCreate{
						Name: "Test new event",
						Jobs: []dto.Job{
							{},
						},
					}
				})

				fmt.Println(response)

				err1 := conn.Close()
				if err1 != nil {
					fmt.Println("error on close")
					return false
				}

				return err == nil
			},
			credentials: func() types.Credentials {
				return types.Credentials{
					Username: "user1",
					Password: "pass1",
				}
			},
		},
	}

	for _, test := range tests {
		fmt.Println("TEST:", test.description)
		fmt.Println("Passed:", test.test(test.credentials) == true)
	}

}

func TestErrors(t *testing.T) {
	// Ne pas se connecter si mauvaise config

	// Ne pas pouvoir rejoindre une manif fermée

	// ne pas pouvoir cloture une manif si on est pas le créateur
}
