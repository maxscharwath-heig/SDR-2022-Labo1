package network

import (
	"fmt"
	"net"
	"sdr/labo1/src/types"
	"sdr/labo1/src/utils"
)

type InterServerProtocol struct {
	Connections map[string]*ClientProtocol
}

func (i *InterServerProtocol) BroadcastRequest(endpointId string, data func(auth AuthId) any) {
	for _, c := range i.Connections {
		_, _ = c.SendRequest(endpointId, data)
	}
}

func (i *InterServerProtocol) get(id string) *ClientProtocol {
	return i.Connections[id]
}

func CreateInterServerProtocol() *InterServerProtocol {
	return &InterServerProtocol{
		Connections: make(map[string]*ClientProtocol),
	}
}

func (i *InterServerProtocol) ConnectToServer(url string) bool {
	maxTries := 3
	for try := 0; try < maxTries; try++ {
		if try > 0 {
			utils.LogWarning(false, url, fmt.Sprintf("failed to connect, attempt %d/%d", try+1, maxTries))
		}
		conn, err := net.Dial("tcp", url)
		if err != nil {
			continue
		}
		i.Connections[url] = CreateClientProtocol(conn, func() types.Credentials {
			return types.Credentials{}
		})
		utils.LogSuccess(false, url, "Connected to server")
		return true
	}
	utils.LogError(false, url, "Failed to connect to server")
	return false
}

func (i *InterServerProtocol) ConnectToServers(urls []string) bool {
	successConnected := make(chan bool, 1)
	for _, url := range urls {
		go func(url string) {
			successConnected <- i.ConnectToServer(url)
		}(url)
	}
	for i := 0; i < len(urls); i++ {
		if !<-successConnected {
			return false // if one connection failed, stop
		}
	}
	return true
}
