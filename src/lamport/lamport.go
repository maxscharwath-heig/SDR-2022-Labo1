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

func (l *Lamport) currentState() Request {
	return l.states[l.id]
}

func (l *Lamport) handleLamportRequest(req Request) {
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

func (l *Lamport) checkCriticalSectionAccess() {
	if l.currentState().ReqType != REQ {
		return
	}
	oldest := true

	/*
		plusAncienne := vrai
		pour chaque i entre 0 et N sauf moi
			si t[moi].estampille > t[i].estampille alors
				plusAncienne := faux; sortir du pour
			sinon si t[moi].estampille = t[i].estampille et moi > i alors
				plusAncienne := faux; sortir du pour
			fin si
		fin pour
	*/

	if oldest {
		l.hasAccess = true
	}
}
