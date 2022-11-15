package lamport

import (
	"math"
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
	Sender   string
	Receiver string
}

type Lamport struct {
	stamp     int
	id        string
	hasAccess bool
	states    map[string]Request
}

// InitLamport inits the needed structure for lamport's algorithm
func InitLamport(numberOfServers int, id string) Lamport {
	var lmp = Lamport{
		stamp:     0,
		id:        id,
		hasAccess: false,
		states:    make(map[string]Request, numberOfServers),
	}

	// Init all to REL, stamp 0
	for key := range lmp.states {
		lmp.states[key] = Request{
			ReqType:  REL,
			Stamp:    lmp.stamp,
			Global:   false,
			Sender:   lmp.id,
			Receiver: lmp.id,
		}
	}

	return lmp
}

func (l *Lamport) setCurrentState(request Request) {
	l.states[l.id] = request
}

func (l *Lamport) currentState() Request {
	return l.states[l.id]
}

// TODO: on peut fusionner les deux functions car au final elle font la même chose (REQ, REL)

// HandleClientAskCriticalSection
func (l *Lamport) HandleClientAskCriticalSection() {
	l.stamp += 1
	l.setCurrentState(Request{
		ReqType: REQ,
		Stamp:   l.stamp,
		Global:  true,
		Sender:  l.id,
	})

	// TODO: Envoi de la requête (currentState - REQ) vers tous (broadcast)
}

func (l *Lamport) HandleClientReleaseCriticalSection() {
	l.stamp += 1
	l.setCurrentState(Request{
		ReqType: REL,
		Stamp:   l.stamp,
		Global:  true,
		Sender:  l.id,
	})

	// TODO: Envoi de la requête (currentState - REL) vers tous (broadcast)
}

// HandleLamportRequest Traiment des messages entre serveurs
func (l *Lamport) HandleLamportRequest(req Request) {
	l.stamp = int(math.Max(float64(l.stamp), float64(req.Stamp)) + 1)

	switch req.ReqType {
	case REQ:
		l.states[req.Sender] = req

		if l.currentState().ReqType != REQ {
			// Build ack to send
			ack := Request{
				ReqType:  ACK,
				Stamp:    l.stamp,
				Global:   false,
				Sender:   l.id,
				Receiver: req.Sender,
			}

			println(ack) // tmp just to use the var
			// TODO: send ack
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

// checkCriticalSectionAccess check if can access to SC and set access if so
func (l *Lamport) checkCriticalSectionAccess() {
	if l.currentState().ReqType != REQ {
		return
	}
	oldest := true

	for i := range l.states {
		if i == l.id {
			continue
		}

		if l.currentState().Stamp > l.states[i].Stamp {
			oldest = false
			break
		}

		if l.currentState().Stamp == l.states[i].Stamp && l.id > i {
			oldest = false
			break
		}
	}

	if oldest {
		l.hasAccess = true
	}
}
