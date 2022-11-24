package lamport

import (
	"fmt"
	"math"
	"sdr/labo1/src/network/server_server"
	"sdr/labo1/src/utils"
	"strings"
)

// Lamport

type RequestType int

const (
	REQ RequestType = 0
	ACK RequestType = 1
	REL RequestType = 2
)

type Request[T any] struct {
	ReqType  RequestType `json:"req_type"`
	Stamp    int         `json:"stamp"`
	Data     T           `json:"data"`
	Sender   int         `json:"sender"`
	Global   bool        `json:"global"`
	Receiver int         `json:"receiver"`
}

type Lamport[T any] struct {
	stamp         int
	protocol      *server_server.InterServerProtocol[Request[T]]
	hasAccess     bool
	states        map[int]Request[T]
	waitForAccess chan bool
	onData        func(data T)
}

func (l *Lamport[T]) id() int {
	return l.protocol.GetServerId()
}

func (l *Lamport[T]) setCurrentState(request Request[T]) {
	l.setLamportState(l.id(), request)
}

func (l *Lamport[T]) currentState() Request[T] {
	return l.states[l.id()]
}

func (l *Lamport[T]) sendRequest(request Request[T]) {
	if request.Global {
		_ = l.protocol.SendToAll(request)
	} else {
		_ = l.protocol.SendTo(request.Receiver, request)
	}
}

func (l *Lamport[T]) setLamportState(serverId int, request Request[T]) {
	l.states[serverId] = request
	l.debug()
	tmp := l.checkCriticalSectionAccess()
	if tmp && !l.hasAccess {
		l.hasAccess = true
		l.waitForAccess <- true
	}
}

// InitLamport inits the needed structure for lamport's algorithm
func InitLamport[T any](p *server_server.InterServerProtocol[Request[T]], onData func(data T)) Lamport[T] {
	var lmp = Lamport[T]{
		stamp:         0,
		protocol:      p,
		hasAccess:     false,
		states:        make(map[int]Request[T], p.GetNumberOfServers()),
		waitForAccess: make(chan bool, 1),
		onData:        onData,
	}

	for i := 0; i < p.GetNumberOfServers(); i++ {
		lmp.states[i] = Request[T]{
			ReqType: REL,
			Stamp:   0,
		}
	}
	return lmp
}

func (l *Lamport[T]) debug() {
	var str = map[RequestType]string{
		REQ: "REQ",
		ACK: "ACK",
		REL: "REL",
	}
	headers := []string{"Servers"}
	data := []string{fmt.Sprintf("T:%d SC:%t", l.stamp, l.hasAccess)}
	for key, state := range l.states {
		headers = append(headers, fmt.Sprintf("Server %d", key))
		data = append(data, fmt.Sprintf("%s(%d)", str[state.ReqType], state.Stamp))
	}
	utils.PrintTable(headers, []string{strings.Join(data, "\t")})
}

// SendClientAskCriticalSection indique que le client souhaite l'accès
func (l *Lamport[T]) SendClientAskCriticalSection() chan bool {
	l.stamp += 1
	l.setCurrentState(Request[T]{
		ReqType: REQ,
		Stamp:   l.stamp,
		Sender:  l.id(),
		Global:  true,
	})

	l.sendRequest(l.currentState())
	return l.waitForAccess
}

// SendClientReleaseCriticalSection indique que le client sort de SC
func (l *Lamport[T]) SendClientReleaseCriticalSection(data T) {
	l.hasAccess = false
	l.stamp += 1
	l.setCurrentState(Request[T]{
		ReqType: REL,
		Stamp:   l.stamp,
		Sender:  l.id(),
		Data:    data,
		Global:  true,
	})

	l.sendRequest(l.currentState())
	l.onData(data)
}

// handleLamportRequest Traitment des messages entre serveurs
func (l *Lamport[T]) handleLamportRequest(req Request[T]) {
	l.stamp = int(math.Max(float64(l.stamp), float64(req.Stamp)) + 1)

	switch req.ReqType {
	case REQ:
		l.setLamportState(req.Sender, req)

		if l.currentState().ReqType != REQ {
			ack := Request[T]{
				ReqType:  ACK,
				Stamp:    l.stamp,
				Sender:   l.id(),
				Receiver: req.Sender,
			}
			l.sendRequest(ack)
		}

	case ACK:
		if l.states[req.Sender].ReqType != REQ {
			l.setLamportState(req.Sender, req)
		}
	case REL:
		l.onData(req.Data)
		l.setLamportState(req.Sender, req)
	}
}

// checkCriticalSectionAccess check if process can access to SC
func (l *Lamport[T]) checkCriticalSectionAccess() bool {
	if l.currentState().ReqType != REQ {
		return false
	}
	for i := range l.states {
		if i == l.id() {
			continue
		}

		if l.currentState().Stamp > l.states[i].Stamp {
			return false
		}

		if l.currentState().Stamp == l.states[i].Stamp && l.id() > i {
			return false
		}
	}

	return true
}

func (l *Lamport[T]) Start() {
	utils.LogInfo(true, "Lamport:", "started")
	for {
		select {
		// REQ, ACK, REL
		case request := <-l.protocol.GetMessageChan():
			l.handleLamportRequest(request)
		}
	}
}
