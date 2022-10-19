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
	test        func() bool
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
		Users: []config.UserWithPassword{
			{
				1,
				"user1",
				"pass1",
			},
			{
				2,
				"test",
				"test",
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
			test: func() bool {
				go server.Start(&validSrvConfig)
				conn, err := connect(validClientConfig.FullUrl())
				defer func() {
					conn.Close()
					server.Stop()
				}()

				return err == nil
			},
		},
		{
			description: "Should create event",
			test: func() bool {
				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, func() types.Credentials {
					return types.Credentials{
						Username: "user1",
						Password: "pass1",
					}
				})
				defer func() {
					conn.Close()
					server.Stop()
				}()

				json, _ := cli.SendRequest("create", func(auth network.Auth) any {
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

				event, responseError := network.ParseResponse[*dto.Event](json)

				return event.Name == "Test new event" && event.Jobs[0].Name == "Test" && event.Jobs[0].Capacity == 2 && responseError == nil
			},
		},
		{
			description: "Should close event",
			test: func() bool {
				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, func() types.Credentials {
					return types.Credentials{
						Username: "user1",
						Password: "pass1",
					}
				})
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

				json, _ := cli.SendRequest("close", func(auth network.Auth) any {
					return dto.EventClose{
						EventId: 1,
					}
				})

				event, responseError := network.ParseResponse[*dto.Event](json)

				return event.Open == false && responseError == nil
			},
		},
		{
			description: "Should register to event",
			test: func() bool {
				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, func() types.Credentials {
					return types.Credentials{
						Username: "user1",
						Password: "pass1",
					}
				})
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

				json, _ := cli.SendRequest("register", func(auth network.Auth) any {
					return dto.EventRegister{
						EventId: 1,
						JobId:   1,
					}
				})

				event, responseError := network.ParseResponse[*dto.Event](json)

				return event.Participants[0].User.Id == 1 && event.Participants[0].JobId == 1 && responseError == nil
			},
		},
		{
			description: "Should show all events",
			test: func() bool {
				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, func() types.Credentials {
					return types.Credentials{
						Username: "user1",
						Password: "pass1",
					}
				})
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

				json, _ := cli.SendRequest("show", func(auth network.Auth) any {
					return dto.EventShow{
						EventId: -1,
						Resume:  false,
					}
				})

				event, responseError := network.ParseResponse[[]*dto.Event](json)

				return len(event) == 2 && responseError == nil
			},
		},
		{
			description: "Should show one event",
			test: func() bool {
				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, func() types.Credentials {
					return types.Credentials{
						Username: "user1",
						Password: "pass1",
					}
				})
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

				json, _ := cli.SendRequest("show", func(auth network.Auth) any {
					return dto.EventShow{
						EventId: 1,
						Resume:  false,
					}
				})

				event, responseError := network.ParseResponse[*dto.Event](json)

				return event.Id == 1 && event.Name == "Test new event" && responseError == nil
			},
		},
		{
			description: "Should show one event's resume",
			test: func() bool {
				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, func() types.Credentials {
					return types.Credentials{
						Username: "user1",
						Password: "pass1",
					}
				})
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

				json, _ := cli.SendRequest("show", func(auth network.Auth) any {
					return dto.EventShow{
						EventId: 1,
						Resume:  true,
					}
				})

				event, responseError := network.ParseResponse[*dto.Event](json)
				return event.Id == 1 && event.Name == "Test new event" && responseError == nil
			},
		},
		{
			description: "Should not have duplicate registration",
			test: func() bool {
				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, func() types.Credentials {
					return types.Credentials{
						Username: "user1",
						Password: "pass1",
					}
				})
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
							{
								Name:     "Test 2",
								Capacity: 3,
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

				cli.SendRequest("register", func(auth network.Auth) any {
					return dto.EventRegister{
						EventId: 1,
						JobId:   2,
					}
				})

				json, _ := cli.SendRequest("show", func(auth network.Auth) any {
					return dto.EventShow{
						EventId: 1,
						Resume:  true,
					}
				})

				event, responseError := network.ParseResponse[*dto.Event](json)
				return event.Participants[0].JobId == 2 && responseError == nil
			},
		},
	}

	for _, test := range tests {
		fmt.Printf("TEST: %s: ", test.description)
		if !test.test() {
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
			{
				2,
				"test",
				"test",
			},
		},
	}

	validClientConfig := config.ClientConfiguration{
		Host: "localhost",
		Port: 9001,
	}

	tests = []cliTest{
		{
			description: "Should give error if invalid credentials",
			test: func() bool {

				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())

				defer func() {
					conn.Close()
					server.Stop()
				}()

				cli := network.CreateClientProtocol(conn, func() types.Credentials {
					return types.Credentials{
						Username: "asd",
						Password: "asd",
					}
				})

				_, err := cli.SendRequest("create", func(auth network.Auth) any {
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

				return err.Error() == "invalid credentials"
			},
		},
		{
			description: "Should not register to a closed event",
			test: func() bool {
				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())
				cli := network.CreateClientProtocol(conn, func() types.Credentials {
					return types.Credentials{
						Username: "user1",
						Password: "pass1",
					}
				})
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

				json, _ := cli.SendRequest("register", func(auth network.Auth) any {
					return dto.EventRegister{
						EventId: 1,
					}
				})

				_, responseError := network.ParseResponse[*dto.Event](json)
				return responseError.Error() == "job not found"
			},
		},
		{
			description: "Should not close event if not organizer",
			test: func() bool {

				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())

				defer func() {
					conn.Close()
					server.Stop()
				}()

				cli := network.CreateClientProtocol(conn, func() types.Credentials {
					return types.Credentials{
						Username: "user1",
						Password: "pass1",
					}
				})

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

				cli2 := network.CreateClientProtocol(conn, func() types.Credentials {
					return types.Credentials{
						Username: "test",
						Password: "test",
					}
				})

				json, _ := cli2.SendRequest("close", func(auth network.Auth) any {
					return dto.EventClose{
						EventId: 1,
					}
				})

				_, responseError := network.ParseResponse[*dto.Event](json)

				return responseError.Error() == "you are not the organizer of this event"
			},
		},
		{
			description: "Should not show if event does not exist",
			test: func() bool {

				go server.Start(&validSrvConfig)
				conn, _ := connect(validClientConfig.FullUrl())

				defer func() {
					conn.Close()
					server.Stop()
				}()

				cli := network.CreateClientProtocol(conn, func() types.Credentials {
					return types.Credentials{
						Username: "user1",
						Password: "pass1",
					}
				})

				json, _ := cli.SendRequest("show", func(auth network.Auth) any {
					return dto.EventShow{
						EventId: 1,
						Resume:  false,
					}
				})

				_, responseError := network.ParseResponse[*dto.Event](json)

				return responseError.Error() == "event not found"
			},
		},
	}

	for _, test := range tests {
		fmt.Printf("TEST: %s: ", test.description)
		if !test.test() {
			t.Errorf("ERROR")
		} else {
			fmt.Println("Passed !")
		}
	}
}
