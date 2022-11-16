package lamport

import (
	"math"
	"sdr/labo1/src/network/server_server"
	"sdr/labo1/src/utils"
)

// Lamport

type RequestType int

const (
	REQ RequestType = iota
	ACK
	REL
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

	// Init all to REL, stamp 0
	for key := range lmp.states {
		lmp.states[key] = Request{
			ReqType:  REL,
			Stamp:    lmp.stamp,
			Global:   false,
			Sender:   lmp.id(),
			Receiver: lmp.id(),
		}
	}

	return lmp
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

	println("lamport: got", req.ReqType)

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

			// TODO: handle internal client's "request" (demande acccès et attente d'accès)
			utils.LogInfo(false, "Message received from server", request.Sender)
		}
	}
}
