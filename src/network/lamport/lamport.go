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
	ReqType  RequestType
	Stamp    int
	Global   bool
	Sender   int
	Receiver int
}

type Lamport struct {
	stamp     int
	protocol  *server_server.InterServerProtocol[Request]
	hasAccess bool
	states    map[int]Request
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
		stamp:     0,
		protocol:  p,
		hasAccess: false,
		states:    make(map[int]Request, p.GetNumberOfServers()),
	}

	for i := 0; i < p.GetNumberOfServers(); i++ {
		lmp.states[i] = Request{REL, 0, false, -1, -1}
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

// HandleClientAskCriticalSection indique que le client souhaite l'accès
func (l *Lamport) HandleClientAskCriticalSection() {
	l.stamp += 1
	l.setCurrentState(Request{
		ReqType: REQ,
		Stamp:   l.stamp,
		Global:  true,
		Sender:  l.id(),
	})

	l.protocol.SendToAll(l.currentState())
}

// HandleClientReleaseCriticalSection indique que le client sort de SC
func (l *Lamport) HandleClientReleaseCriticalSection() {
	l.stamp += 1
	l.setCurrentState(Request{
		ReqType: REL,
		Stamp:   l.stamp,
		Global:  true,
		Sender:  l.id(),
	})

	l.protocol.SendToAll(l.currentState())
}

// HandleLamportRequest Traiment des messages entre serveurs
func (l *Lamport) HandleLamportRequest(req Request) {
	l.stamp = int(math.Max(float64(l.stamp), float64(req.Stamp)) + 1)

	switch req.ReqType {
	case REQ:
		l.states[req.Sender] = req

		if l.currentState().ReqType != REQ {
			ack := Request{
				ReqType:  ACK,
				Stamp:    l.stamp,
				Global:   false,
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
	}
	l.checkCriticalSectionAccess()
}

// checkCriticalSectionAccess check if process can access to SC and set access if so
func (l *Lamport) checkCriticalSectionAccess() {
	if l.currentState().ReqType != REQ {
		return
	}
	oldest := true

	for i := range l.states {
		if i == l.id() {
			continue
		}

		if l.currentState().Stamp > l.states[i].Stamp {
			oldest = false
			break
		}

		if l.currentState().Stamp == l.states[i].Stamp && l.id() > i {
			oldest = false
			break
		}
	}

	if oldest {
		l.hasAccess = true
	}
}

func (l *Lamport) Start() {
	utils.LogInfo(true, "Lamport:", "started")
	for {
		select {
		// REQ, ACK, REL
		case request := <-l.protocol.GetMessageChan():
			l.HandleLamportRequest(request)
			l.debug()
			// TODO: handle internal client's "request" (demande acccès et attente d'accès)
		}
	}
}
