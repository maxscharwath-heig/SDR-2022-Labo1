package server_server

import (
	"fmt"
	"net"
	"sdr/labo1/src/network"
	"sdr/labo1/src/utils"
)

type InterServerProtocol[T any] struct {
	serverId    int
	listener    net.Listener
	connections map[int]*network.Connection
	chanMessage chan T
}

func CreateInterServerProtocol[T any](serverId int, listener net.Listener) *InterServerProtocol[T] {
	return &InterServerProtocol[T]{
		serverId:    serverId,
		listener:    listener,
		connections: make(map[int]*network.Connection),
		chanMessage: make(chan T),
	}
}

type serverConnection struct {
	serverId int
	conn     *network.Connection
}

func (p *InterServerProtocol[T]) ConnectToServers(urls []string) {
	ready := make(chan serverConnection) // Channel to wait for all connections to be ready
	useListener := make(chan bool)       // Channel to wait if the listener is used
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
			if value, e := network.GetJson[int](*conn); e == nil {
				// Send the serverId to the server
				_ = conn.SendJSON(p.serverId)
				ready <- serverConnection{value, conn}
			} else {
				utils.LogError(false, "Error accepting: ", e.Error())
				_ = conn.Close()
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
			conn.SendJSON(p.serverId)
			if value, e := network.GetJson[int](*conn); e == nil {
				ready <- serverConnection{value, conn}
			}
		}(url)
	}
	utils.LogInfo(true, "Waiting for all servers to connect...")
	for range urls {
		server := <-ready
		utils.LogSuccess(true, "Server", server.serverId, "connected")
		p.connections[server.serverId] = server.conn
	}
	useListener <- false // stop the listener
	utils.LogSuccess(true, "All servers connected!")

	go p.listenMessages() // Start listening messages
}

func (p *InterServerProtocol[T]) SendTo(serverId int, data T) error {
	if conn, ok := p.connections[serverId]; ok {
		return conn.SendJSON(data)
	}
	return fmt.Errorf("server %d is not connected", serverId)
}

func (p *InterServerProtocol[T]) SendToAll(data T) error {
	for _, conn := range p.connections {
		if err := conn.SendJSON(data); err != nil {
			return err
		}
	}
	return nil
}

func (p *InterServerProtocol[T]) listenMessages() {
	for serverId := range p.connections {
		go func(serverId int) {
			for {
				data, err := network.GetJson[T](*p.connections[serverId])
				if err != nil {
					utils.LogError(false, "Error receiving message from server", serverId, ":", err.Error())
					continue
				}
				p.chanMessage <- data
			}
		}(serverId)
	}
}

func (p *InterServerProtocol[T]) GetMessageChan() chan T {
	return p.chanMessage
}

/*
func (p *InterServerProtocol[T]) GetMessage() {
	for {
		select {
		case msg := <-p.chanMessage:
			//TODO
			//msg.serverId is the serverId of the sender
			//msg.data is the data received (LamportStruct)

			//Example
			p.SendTo(msg.serverId, msg.data) // Send the data back to the sender
			p.SendToAll(msg.data) // Send the data to all the other servers
			utils.LogInfo(false, "Message received from server", msg.serverId)
		}
	}
}
*/
