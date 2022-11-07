package server_server

import (
	"net"
	"sdr/labo1/src/network"
	"sdr/labo1/src/utils"
)

type InterServerProtocol struct {
	listener    net.Listener
	connections map[string]*network.Connection
}

func CreateInterServerProtocol(listener net.Listener) *InterServerProtocol {
	return &InterServerProtocol{
		listener:    listener,
		connections: make(map[string]*network.Connection),
	}
}

func (p *InterServerProtocol) ConnectToServers(urls []string) {
	ready := make(chan string, len(urls)) // Channel to wait for all connections to be ready
	useListener := make(chan bool, 1)     // Channel to wait if the listener is used
	access := make(chan bool, 1)          // Channel to wait for access to the connections map
	access <- true                        // Give access to the connections map
	go func() {
		for {
			if !<-useListener { // Close the listener if no more connections are needed
				utils.LogInfo(false, "inter server protocol", "listener closed")
				return
			}
			c, err := p.listener.Accept()
			if err != nil {
				utils.LogError(false, "Error accepting: ", err.Error())
				useListener <- true // The listener is still needed
				continue
			}
			conn := network.CreateConnection(c)
			if value, e := network.GetResponse[string](*conn, "serverHandshake"); e == nil {

				<-access
				p.connections[value] = conn
				access <- true

				ready <- value
			} else {
				utils.LogError(false, "Error accepting: ", e.Error())
				conn.Close()
				useListener <- true // The listener is still needed
				continue
			}
		}
	}()

	for _, url := range urls {
		go func(url string) {
			c, err := net.Dial("tcp", url)
			if err != nil {
				utils.LogError(false, url, "is not reachable, waiting connection...")
				useListener <- true // Tell to use the listener method
				return
			}
			//DIRECT CONNECTION
			conn := network.CreateConnection(c)

			<-access
			p.connections[conn.RemoteAddr().String()] = conn
			access <- true

			conn.SendResponse("serverHandshake", true, p.listener.Addr().String())
			ready <- conn.RemoteAddr().String()
		}(url)
	}
	utils.LogInfo(true, "Waiting for all servers to connect...")
	for range urls {
		utils.LogInfo(true, "Server", <-ready, "connected")
	}
	useListener <- false // stop the listener
	utils.LogSuccess(true, "All servers connected!")
}
