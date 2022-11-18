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

type Request struct {
	ReqType  RequestType `json:"req_type"`
	Stamp    int         `json:"stamp"`
	Data     any         `json:"data"`
	Sender   int         `json:"sender"`
	Receiver int         `json:"receiver"`
}

type Lamport struct {
	stamp         int
	protocol      *server_server.InterServerProtocol[Request]
	hasAccess     bool
	states        map[int]Request
	waitForAccess chan bool
}

func (l *Lamport) id() int {
	return l.protocol.GetServerId()
}

func (l *Lamport) setCurrentState(request Request) {
	l.states[l.id()] = request
}

func (l *Lamport) currentState() Request {
	return l.states[l.id()]
}

// InitLamport inits the needed structure for lamport's algorithm
func InitLamport(p *server_server.InterServerProtocol[Request]) Lamport {
	var lmp = Lamport{
		stamp:         0,
		protocol:      p,
		hasAccess:     false,
		states:        make(map[int]Request, p.GetNumberOfServers()),
		waitForAccess: make(chan bool, 1),
	}

	for i := 0; i < p.GetNumberOfServers(); i++ {
		lmp.states[i] = Request{
			ReqType: REL,
			Stamp:   0,
		}
	}
	return lmp
}

func (l *Lamport) debug() {
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

// TODO: on peut fusionner les deux functions car au final elle font la même chose (REQ, REL)

// SendClientAskCriticalSection indique que le client souhaite l'accès
func (l *Lamport) SendClientAskCriticalSection() chan bool {
	l.stamp += 1
	l.setCurrentState(Request{
		ReqType: REQ,
		Stamp:   l.stamp,
		Sender:  l.id(),
	})

	l.protocol.SendToAll(l.currentState())
	return l.waitForAccess
}

// SendClientReleaseCriticalSection indique que le client sort de SC
func (l *Lamport) SendClientReleaseCriticalSection(data any) {
	l.stamp += 1
	l.setCurrentState(Request{
		ReqType: REL,
		Stamp:   l.stamp,
		Sender:  l.id(),
		Data:    data,
	})

	l.protocol.SendToAll(l.currentState())
}

// handleLamportRequest Traitment des messages entre serveurs
func (l *Lamport) handleLamportRequest(req Request) {
	l.stamp = int(math.Max(float64(l.stamp), float64(req.Stamp)) + 1)

	switch req.ReqType {
	case REQ:
		l.states[req.Sender] = req

		if l.currentState().ReqType != REQ {
			ack := Request{
				ReqType:  ACK,
				Stamp:    l.stamp,
				Sender:   l.id(),
				Receiver: req.Sender,
			}
			l.protocol.SendTo(ack.Receiver, ack)
		}

	case ACK:
		if l.states[req.Sender].ReqType != REQ {
			l.states[req.Sender] = req
		}
	case REL:
		l.states[req.Sender] = req

		//UPDATE DATA
	}
	tmp := l.checkCriticalSectionAccess()
	if tmp && !l.hasAccess {
		l.waitForAccess <- true
	}
	l.hasAccess = tmp
}

// checkCriticalSectionAccess check if process can access to SC and set access if so
func (l *Lamport) checkCriticalSectionAccess() bool {
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

func (l *Lamport) Start() {
	utils.LogInfo(true, "Lamport:", "started")
	for {
		select {
		// REQ, ACK, REL
		case request := <-l.protocol.GetMessageChan():
			l.handleLamportRequest(request)
			l.debug()
			// TODO: handle internal client's "request" (demande acccès et attente d'accès)
		}
	}
}
