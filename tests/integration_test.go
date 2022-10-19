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

// TODO: encore un test pour voir que si on s'incrit deux fois ça garde que la derniere

func TestSuccess(t *testing.T) {
	validSrvConfig := config.ServerConfiguration{
		Host: "localhost",
		Port: 9001,
		Users: []config.UserWithPassword{
			{
				1,
				"user1",
				"pass1",
			},
		},
	}

	validClientConfig := config.ClientConfiguration{
		Host: "localhost",
		Port: 9001,
	}

	tests = []cliTest{
		{
			description: "Should connect to server",
			test: func(creds func() types.Credentials) bool {

				go server.Start(&validSrvConfig)
				conn, err := connect(validClientConfig.FullUrl())
				defer func() {
					conn.Close()
					server.Stop()
				}()

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
			description: "Should create event",
			test: func(creds func() types.Credentials) bool {
				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, creds)
				defer func() {
					conn.Close()
					server.Stop()
				}()

				response, err := cli.SendRequest("create", func(auth network.Auth) any {
					return dto.EventCreate{
						Name: "Test new event",
						Jobs: []dto.Job{
							{
								Name:     "Test",
								Capacity: 2,
							},
						},
					}
				})

				expectedResponse := `{"id":1,"name":"Test new event","open":true,"jobs":[{"id":1,"name":"Test","capacity":2,"count":0}],"organizer":{"id":1,"username":"user1"},"participants":[]}`

				return response == expectedResponse && err == nil
			},
			credentials: func() types.Credentials {
				return types.Credentials{
					Username: "user1",
					Password: "pass1",
				}
			},
		},
		{
			description: "Should close event",
			test: func(creds func() types.Credentials) bool {
				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, creds)
				defer func() {
					conn.Close()
					server.Stop()
				}()

				cli.SendRequest("create", func(auth network.Auth) any {
					return dto.EventCreate{
						Name: "Test new event",
						Jobs: []dto.Job{
							{
								Name:     "Test",
								Capacity: 2,
							},
						},
					}
				})

				response, err := cli.SendRequest("close", func(auth network.Auth) any {
					return dto.EventClose{
						EventId: 1,
					}
				})

				expectedResponse := `{"id":1,"name":"Test new event","open":false,"jobs":[{"id":1,"name":"Test","capacity":2,"count":0}],"organizer":{"id":1,"username":"user1"},"participants":[]}`

				return response == expectedResponse && err == nil
			},
			credentials: func() types.Credentials {
				return types.Credentials{
					Username: "user1",
					Password: "pass1",
				}
			},
		},
		{
			description: "Should register to event",
			test: func(creds func() types.Credentials) bool {
				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, creds)
				defer func() {
					conn.Close()
					server.Stop()
				}()

				cli.SendRequest("create", func(auth network.Auth) any {
					return dto.EventCreate{
						Name: "Test new event",
						Jobs: []dto.Job{
							{
								Name:     "Test",
								Capacity: 2,
							},
						},
					}
				})

				response, err := cli.SendRequest("register", func(auth network.Auth) any {
					return dto.EventRegister{
						EventId: 1,
						JobId:   1,
					}
				})

				expectedResponse := `true`
				return response == expectedResponse && err == nil
			},
			credentials: func() types.Credentials {
				return types.Credentials{
					Username: "user1",
					Password: "pass1",
				}
			},
		},
		{
			description: "Should show all events",
			test: func(creds func() types.Credentials) bool {
				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, creds)
				defer func() {
					conn.Close()
					server.Stop()
				}()

				cli.SendRequest("create", func(auth network.Auth) any {
					return dto.EventCreate{
						Name: "Test new event",
						Jobs: []dto.Job{
							{
								Name:     "Test",
								Capacity: 2,
							},
						},
					}
				})
				cli.SendRequest("create", func(auth network.Auth) any {
					return dto.EventCreate{
						Name: "Test new event 2",
						Jobs: []dto.Job{
							{
								Name:     "Test 2",
								Capacity: 2,
							},
						},
					}
				})

				response, err := cli.SendRequest("show", func(auth network.Auth) any {
					return dto.EventShow{
						EventId: -1,
						Resume:  false,
					}
				})

				expectedResponse := `[{"id":1,"name":"Test new event","open":true,"jobs":[{"id":1,"name":"Test","capacity":2,"count":0}],"organizer":{"id":1,"username":"user1"},"participants":[]},{"id":2,"name":"Test new event 2","open":true,"jobs":[{"id":1,"name":"Test 2","capacity":2,"count":0}],"organizer":{"id":1,"username":"user1"},"participants":[]}]`
				return response == expectedResponse && err == nil
			},
			credentials: func() types.Credentials {
				return types.Credentials{
					Username: "user1",
					Password: "pass1",
				}
			},
		},
		{
			description: "Should show one event",
			test: func(creds func() types.Credentials) bool {
				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, creds)
				defer func() {
					conn.Close()
					server.Stop()
				}()

				cli.SendRequest("create", func(auth network.Auth) any {
					return dto.EventCreate{
						Name: "Test new event",
						Jobs: []dto.Job{
							{
								Name:     "Test",
								Capacity: 2,
							},
						},
					}
				})

				response, err := cli.SendRequest("show", func(auth network.Auth) any {
					return dto.EventShow{
						EventId: 1,
						Resume:  false,
					}
				})

				expectedResponse := `{"id":1,"name":"Test new event","open":true,"jobs":[{"id":1,"name":"Test","capacity":2,"count":0}],"organizer":{"id":1,"username":"user1"},"participants":[]}`

				return response == expectedResponse && err == nil
			},
			credentials: func() types.Credentials {
				return types.Credentials{
					Username: "user1",
					Password: "pass1",
				}
			},
		},
		{
			description: "Should show one event's resume",
			test: func(creds func() types.Credentials) bool {
				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, creds)
				defer func() {
					conn.Close()
					server.Stop()
				}()

				cli.SendRequest("create", func(auth network.Auth) any {
					return dto.EventCreate{
						Name: "Test new event",
						Jobs: []dto.Job{
							{
								Name:     "Test",
								Capacity: 2,
							},
						},
					}
				})

				cli.SendRequest("register", func(auth network.Auth) any {
					return dto.EventRegister{
						EventId: 1,
						JobId:   1,
					}
				})

				response, err := cli.SendRequest("show", func(auth network.Auth) any {
					return dto.EventShow{
						EventId: 1,
						Resume:  true,
					}
				})

				expectedResponse := `{"id":1,"name":"Test new event","open":true,"jobs":[{"id":1,"name":"Test","capacity":2,"count":1}],"organizer":{"id":1,"username":"user1"},"participants":[{"user":{"id":1,"username":"user1"},"jobId":1}]}`

				return response == expectedResponse && err == nil
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
		fmt.Printf("TEST: %s: ", test.description)
		if !test.test(test.credentials) {
			t.Errorf("ERROR")
		} else {
			fmt.Println("Passed !")
		}
	}
}

func TestErrors(t *testing.T) {
	validSrvConfig := config.ServerConfiguration{
		Host: "localhost",
		Port: 9001,
		Users: []config.UserWithPassword{
			{
				1,
				"user1",
				"pass1",
			},
		},
	}

	validClientConfig := config.ClientConfiguration{
		Host: "localhost",
		Port: 9001,
	}

	tests = []cliTest{
		{
			description: "Should give error if invalid auth",
			test: func(creds func() types.Credentials) bool {

				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, creds)
				defer func() {
					conn.Close()
					server.Stop()
				}()

				response, err := cli.SendRequest("create", func(auth network.Auth) any {
					return dto.EventCreate{
						Name: "Test new event",
						Jobs: []dto.Job{
							{
								Name:     "Test",
								Capacity: 2,
							},
						},
					}
				})

				fmt.Println("")

				expectedResponse := `{"Success":false,"Auth":null}`

				return response == expectedResponse && err == nil
			},
			credentials: func() types.Credentials {
				return types.Credentials{
					Username: "coasceec",
					Password: "psfasdadfasdf",
				}
			},
		},
		{
			description: "Should not register to a closed event",
			test: func(creds func() types.Credentials) bool {

				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, creds)
				defer func() {
					conn.Close()
					server.Stop()
				}()

				cli.SendRequest("create", func(auth network.Auth) any {
					return dto.EventCreate{
						Name: "Test new event",
						Jobs: []dto.Job{
							{
								Name:     "Test",
								Capacity: 2,
							},
						},
					}
				})

				cli.SendRequest("close", func(auth network.Auth) any {
					return dto.EventClose{
						EventId: 1,
					}
				})

				response, _ := cli.SendRequest("register", func(auth network.Auth) any {
					return dto.EventRegister{
						EventId: 1,
					}
				})

				expectedResponse := `false`

				return response == expectedResponse
			},
			credentials: func() types.Credentials {
				return types.Credentials{
					Username: "user1",
					Password: "pass1",
				}
			},
		},
		{
			description: "Should not close event if not organizer",
			test: func(creds func() types.Credentials) bool {

				// TODO: comment changer de creds ?

				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, creds)
				defer func() {
					conn.Close()
					server.Stop()
				}()

				cli.SendRequest("create", func(auth network.Auth) any {
					return dto.EventCreate{
						Name: "Test new event",
						Jobs: []dto.Job{
							{
								Name:     "Test",
								Capacity: 2,
							},
						},
					}
				})

				cli.SendRequest("close", func(auth network.Auth) any {
					return dto.EventClose{
						EventId: 1,
					}
				})

				response, _ := cli.SendRequest("register", func(auth network.Auth) any {
					return dto.EventRegister{
						EventId: 1,
					}
				})

				expectedResponse := `false`

				return response == expectedResponse
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
		fmt.Printf("TEST: %s: ", test.description)
		if !test.test(test.credentials) {
			t.Errorf("ERROR")
		} else {
			fmt.Println("Passed !")
		}
	}
}
