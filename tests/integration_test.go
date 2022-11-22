// SDR - Labo 2
// Nicolas Crausaz & Maxime Scharwath

package tests

import (
	"net"
	server "sdr/labo1/src"
	"sdr/labo1/src/config"
	"sdr/labo1/src/dto"
	"sdr/labo1/src/network"
	"sdr/labo1/src/network/client_server"
	"sdr/labo1/src/types"
	"testing"
	"time"
)

func connect(addr string) (*net.TCPConn, error) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	return conn, err
}

func expect(t *testing.T, value any, expected any) {
	if value != expected {
		t.Errorf("Expected %v, got %v", expected, value)
	}
}

func expectError(t *testing.T, err error, expected string) {
	if err == nil {
		t.Errorf("Expected error %v, got nil", expected)
	} else if err.Error() != expected {
		t.Errorf("Expected error %v, got %v", expected, err.Error())
	}
}

func TestSuccess(t *testing.T) {
	validServerConfigs := []*config.ServerConfiguration{
		{
			Id: 0,
			Servers: []config.ServerUrl{
				{
					Client: "localhost:10000",
					Server: "localhost:11000",
				},
				{
					Client: "localhost:10001",
					Server: "localhost:11001",
				},
				{
					Client: "localhost:10002",
					Server: "localhost:11002",
				},
			},
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
			Debug:         false,
			ShowInfosLogs: false,
		},
		{
			Id: 1,
			Servers: []config.ServerUrl{
				{
					Client: "localhost:10000",
					Server: "localhost:11000",
				},
				{
					Client: "localhost:10001",
					Server: "localhost:11001",
				},
				{
					Client: "localhost:10002",
					Server: "localhost:11002",
				},
			},
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
			Debug:         false,
			ShowInfosLogs: false,
		},
		{
			Id: 2,
			Servers: []config.ServerUrl{
				{
					Client: "localhost:10000",
					Server: "localhost:11000",
				},
				{
					Client: "localhost:10001",
					Server: "localhost:11001",
				},
				{
					Client: "localhost:10002",
					Server: "localhost:11002",
				},
			},
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
			Debug:         false,
			ShowInfosLogs: false,
		},
	}

	validClientConfig := config.ClientConfiguration{
		Servers: []string{
			"localhost:10000",
			"localhost:10001",
			"localhost:10002",
		},
	}

	t.Run("should connect to server", func(t *testing.T) {
		startServers(validServerConfigs)

		conn, err := connect(validClientConfig.Servers[0])
		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})
		expect(t, err, nil)
	})

	t.Run("should create event", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])
		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})
		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		json, _ := cli.SendRequest("create", func(auth client_server.AuthId) any {
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
		expect(t, responseError, nil)
		expect(t, event.Name, "Test new event")
		expect(t, event.Organizer.Id, 1)
		expect(t, event.Jobs[0].Name, "Test")
		expect(t, event.Jobs[0].Capacity, 2)
		expect(t, event.Open, true)
	})

	t.Run("should close event", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])
		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})
		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		_, _ = cli.SendRequest("create", func(auth client_server.AuthId) any {
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

		json, _ := cli.SendRequest("close", func(auth client_server.AuthId) any {
			return dto.EventClose{
				EventId: 1,
			}
		})

		event, responseError := network.ParseResponse[*dto.Event](json)

		expect(t, responseError, nil)
		expect(t, event.Open, false)
	})

	t.Run("should register to event", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])
		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})
		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		_, _ = cli.SendRequest("create", func(auth client_server.AuthId) any {
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

		json, _ := cli.SendRequest("register", func(auth client_server.AuthId) any {
			return dto.EventRegister{
				EventId: 1,
				JobId:   1,
			}
		})

		event, responseError := network.ParseResponse[*dto.Event](json)

		expect(t, responseError, nil)
		expect(t, event.Jobs[0].Capacity, 2)
		expect(t, event.Jobs[0].Count, 1)
		expect(t, event.Participants[0].JobId, 1)
	})

	t.Run("should show all events", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])
		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})
		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		_, _ = cli.SendRequest("create", func(auth client_server.AuthId) any {
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
		_, _ = cli.SendRequest("create", func(auth client_server.AuthId) any {
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

		json, _ := cli.SendRequest("show", func(auth client_server.AuthId) any {
			return dto.EventShow{
				EventId: -1,
				Resume:  false,
			}
		})

		event, responseError := network.ParseResponse[[]*dto.Event](json)

		expect(t, responseError, nil)
		expect(t, len(event), 2)
	})

	t.Run("should show one event", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])
		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})
		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		_, _ = cli.SendRequest("create", func(auth client_server.AuthId) any {
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

		json, _ := cli.SendRequest("show", func(auth client_server.AuthId) any {
			return dto.EventShow{
				EventId: 1,
				Resume:  false,
			}
		})

		event, responseError := network.ParseResponse[*dto.Event](json)

		expect(t, responseError, nil)
		expect(t, event.Id, 1)
		expect(t, event.Name, "Test new event")
	})

	t.Run("should show one event resume", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])
		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})
		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		_, _ = cli.SendRequest("create", func(auth client_server.AuthId) any {
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

		_, _ = cli.SendRequest("register", func(auth client_server.AuthId) any {
			return dto.EventRegister{
				EventId: 1,
				JobId:   1,
			}
		})

		json, _ := cli.SendRequest("show", func(auth client_server.AuthId) any {
			return dto.EventShow{
				EventId: 1,
				Resume:  true,
			}
		})

		event, responseError := network.ParseResponse[*dto.Event](json)

		expect(t, responseError, nil)
		expect(t, event.Id, 1)
		expect(t, event.Name, "Test new event")
	})

	t.Run("should not have duplicate registration", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])
		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})
		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		_, _ = cli.SendRequest("create", func(auth client_server.AuthId) any {
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

		_, _ = cli.SendRequest("register", func(auth client_server.AuthId) any {
			return dto.EventRegister{
				EventId: 1,
				JobId:   1,
			}
		})

		_, _ = cli.SendRequest("register", func(auth client_server.AuthId) any {
			return dto.EventRegister{
				EventId: 1,
				JobId:   2,
			}
		})

		json, _ := cli.SendRequest("show", func(auth client_server.AuthId) any {
			return dto.EventShow{
				EventId: 1,
				Resume:  true,
			}
		})

		event, responseError := network.ParseResponse[*dto.Event](json)

		expect(t, responseError, nil)
		expect(t, event.Participants[0].JobId, 2)
	})
}

func TestErrors(t *testing.T) {
	validServerConfigs := []*config.ServerConfiguration{
		{
			Id: 0,
			Servers: []config.ServerUrl{
				{
					Client: "localhost:10000",
					Server: "localhost:11000",
				},
				{
					Client: "localhost:10001",
					Server: "localhost:11001",
				},
				{
					Client: "localhost:10002",
					Server: "localhost:11002",
				},
			},
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
			Debug:         false,
			ShowInfosLogs: false,
		},
		{
			Id: 1,
			Servers: []config.ServerUrl{
				{
					Client: "localhost:10000",
					Server: "localhost:11000",
				},
				{
					Client: "localhost:10001",
					Server: "localhost:11001",
				},
				{
					Client: "localhost:10002",
					Server: "localhost:11002",
				},
			},
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
			Debug:         false,
			ShowInfosLogs: false,
		},
		{
			Id: 2,
			Servers: []config.ServerUrl{
				{
					Client: "localhost:10000",
					Server: "localhost:11000",
				},
				{
					Client: "localhost:10001",
					Server: "localhost:11001",
				},
				{
					Client: "localhost:10002",
					Server: "localhost:11002",
				},
			},
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
			Debug:         false,
			ShowInfosLogs: false,
		},
	}

	validClientConfig := config.ClientConfiguration{
		Servers: []string{
			"localhost:10000",
			"localhost:10001",
			"localhost:10002",
		},
	}

	t.Run("should give error if invalid credentials", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])

		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "asd",
				Password: "asd",
			}
		})

		_, err := cli.SendRequest("create", func(auth client_server.AuthId) any {
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

		expectError(t, err, "invalid credentials")
	})

	t.Run("should not register to a closed event", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])
		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})

		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		_, _ = cli.SendRequest("create", func(auth client_server.AuthId) any {
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

		_, _ = cli.SendRequest("close", func(auth client_server.AuthId) any {
			return dto.EventClose{
				EventId: 1,
			}
		})

		json, _ := cli.SendRequest("register", func(auth client_server.AuthId) any {
			return dto.EventRegister{
				EventId: 1,
			}
		})

		_, responseError := network.ParseResponse[*dto.Event](json)
		expectError(t, responseError, "job not found")
	})

	t.Run("should not close event if not organizer", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])

		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})

		_, _ = cli.SendRequest("create", func(auth client_server.AuthId) any {
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

		cli2 := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "test",
				Password: "test",
			}
		})

		json, _ := cli2.SendRequest("close", func(auth client_server.AuthId) any {
			return dto.EventClose{
				EventId: 1,
			}
		})

		_, responseError := network.ParseResponse[*dto.Event](json)

		expectError(t, responseError, "you are not the organizer")
	})

	t.Run("should not close event if already closed", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])
		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})

		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		_, _ = cli.SendRequest("create", func(auth client_server.AuthId) any {
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

		_, _ = cli.SendRequest("close", func(auth client_server.AuthId) any {
			return dto.EventClose{
				EventId: 1,
			}
		})

		json, _ := cli.SendRequest("close", func(auth client_server.AuthId) any {
			return dto.EventClose{
				EventId: 1,
			}
		})

		_, responseError := network.ParseResponse[*dto.Event](json)

		expectError(t, responseError, "event already closed")
	})

	t.Run("should not show if event does not exist", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])

		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})

		json, _ := cli.SendRequest("show", func(auth client_server.AuthId) any {
			return dto.EventShow{
				EventId: 1,
				Resume:  false,
			}
		})

		_, responseError := network.ParseResponse[*dto.Event](json)

		expectError(t, responseError, "event not found")
	})

	t.Run("should have error if empty event name", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])

		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})

		json, _ := cli.SendRequest("create", func(auth client_server.AuthId) any {
			return dto.EventCreate{
				Name: "",
				Jobs: []dto.Job{
					{
						Name:     "Test",
						Capacity: 2,
					},
				},
			}
		})

		_, responseError := network.ParseResponse[*dto.Event](json)

		expectError(t, responseError, "name is required")
	})

	t.Run("should have error if empty job name", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])

		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})

		json, _ := cli.SendRequest("create", func(auth client_server.AuthId) any {
			return dto.EventCreate{
				Name: "Event",
				Jobs: []dto.Job{
					{
						Name:     "",
						Capacity: 2,
					},
				},
			}
		})

		_, responseError := network.ParseResponse[*dto.Event](json)

		expectError(t, responseError, "name is required")
	})

	t.Run("should have error if empty job capacity", func(t *testing.T) {
		startServers(validServerConfigs)
		conn, _ := connect(validClientConfig.Servers[0])

		t.Cleanup(func() {
			_ = conn.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		cli := client_server.CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})

		json, _ := cli.SendRequest("create", func(auth client_server.AuthId) any {
			return dto.EventCreate{
				Name: "Event",
				Jobs: []dto.Job{
					{
						Name:     "test",
						Capacity: -1,
					},
				},
			}
		})

		_, responseError := network.ParseResponse[*dto.Event](json)

		expectError(t, responseError, "capacity must be greater than 0")
	})

	// Test replication of data
	t.Run("should replicate changes when creating event", func(t *testing.T) {
		startServers(validServerConfigs)

		conn1, _ := connect(validClientConfig.Servers[0])
		conn2, _ := connect(validClientConfig.Servers[1])
		conn3, _ := connect(validClientConfig.Servers[2])

		cli1 := client_server.CreateClientProtocol(conn1, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})

		cli2 := client_server.CreateClientProtocol(conn2, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})

		cli3 := client_server.CreateClientProtocol(conn3, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})

		t.Cleanup(func() {
			_ = conn1.Close()
			_ = conn2.Close()
			_ = conn3.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		_, _ = cli1.SendRequest("create", func(auth client_server.AuthId) any {
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

		jsonCli2, _ := cli2.SendRequest("show", func(auth client_server.AuthId) any {
			return dto.EventShow{
				EventId: 1,
				Resume:  false,
			}
		})

		jsonCli3, _ := cli3.SendRequest("show", func(auth client_server.AuthId) any {
			return dto.EventShow{
				EventId: 1,
				Resume:  false,
			}
		})

		expect(t, jsonCli2, "{\"success\":true,\"data\":{\"id\":1,\"name\":\"Test new event\",\"open\":true,\"jobs\":[{\"id\":1,\"name\":\"Test\",\"capacity\":2,\"count\":0}],\"organizer\":{\"id\":1,\"username\":\"user1\"},\"participants\":[]}}")
		expect(t, jsonCli3, "{\"success\":true,\"data\":{\"id\":1,\"name\":\"Test new event\",\"open\":true,\"jobs\":[{\"id\":1,\"name\":\"Test\",\"capacity\":2,\"count\":0}],\"organizer\":{\"id\":1,\"username\":\"user1\"},\"participants\":[]}}")
	})

	t.Run("should replicate changes when closing event", func(t *testing.T) {
		startServers(validServerConfigs)

		conn1, _ := connect(validClientConfig.Servers[0])
		conn2, _ := connect(validClientConfig.Servers[1])
		conn3, _ := connect(validClientConfig.Servers[2])

		cli1 := client_server.CreateClientProtocol(conn1, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})

		cli2 := client_server.CreateClientProtocol(conn2, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})

		cli3 := client_server.CreateClientProtocol(conn3, func() types.Credentials {
			return types.Credentials{
				Username: "user1",
				Password: "pass1",
			}
		})

		t.Cleanup(func() {
			_ = conn1.Close()
			_ = conn2.Close()
			_ = conn3.Close()
			for range validServerConfigs {
				server.Stop()
			}
		})

		_, _ = cli1.SendRequest("create", func(auth client_server.AuthId) any {
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

		_, _ = cli2.SendRequest("close", func(auth client_server.AuthId) any {
			return dto.EventClose{
				EventId: 1,
			}
		})

		jsonCli2, _ := cli2.SendRequest("show", func(auth client_server.AuthId) any {
			return dto.EventShow{
				EventId: 1,
				Resume:  false,
			}
		})

		jsonCli3, _ := cli3.SendRequest("show", func(auth client_server.AuthId) any {
			return dto.EventShow{
				EventId: 1,
				Resume:  false,
			}
		})

		expect(t, jsonCli2, "{\"success\":true,\"data\":{\"id\":1,\"name\":\"Test new event\",\"open\":false,\"jobs\":[{\"id\":1,\"name\":\"Test\",\"capacity\":2,\"count\":0}],\"organizer\":{\"id\":1,\"username\":\"user1\"},\"participants\":[]}}")
		expect(t, jsonCli3, "{\"success\":true,\"data\":{\"id\":1,\"name\":\"Test new event\",\"open\":false,\"jobs\":[{\"id\":1,\"name\":\"Test\",\"capacity\":2,\"count\":0}],\"organizer\":{\"id\":1,\"username\":\"user1\"},\"participants\":[]}}")
	})

}

func startServers(servers []*config.ServerConfiguration) {
	for _, conf := range servers {
		go server.Start(conf)
		time.Sleep(30 * time.Millisecond) // Time for the previous server to boot up
	}
}
