package dto

import (
	"fmt"
	"sdr/labo1/types"
)

type Participant struct {
	User  *types.User `json:"user"`
	JobId int         `json:"jobId"`
}

type Event struct {
	Id           int           `json:"id"`
	Name         string        `json:"name"`
	Open         bool          `json:"open"`
	Jobs         []types.Job   `json:"jobs"`
	Organizer    *types.User   `json:"organizer"`
	Participants []Participant `json:"participants"`
}

// ToRow gets a representation of an event to a table-printable format
func (event *Event) ToRow() string {
	var openText string
	if event.Open {
		openText = "yes"
	} else {
		openText = "no"
	}
	return fmt.Sprintf("%d\t%s\t%s\t%s", event.Id, event.Name, event.Organizer.Username, openText)
}

type EventRegister struct {
	EventId int `json:"eventId"`
	JobId   int `json:"jobId"`
}

type EventClose struct {
	EventId int `json:"eventId"`
}

type Job struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}

type EventCreate struct {
	Name string `json:"name"`
	Jobs []Job  `json:"jobs"`
}

type EventShow struct {
	EventId int  `json:"eventId"`
	Resume  bool `json:"resume"`
}

// CONVERSIONS

func EventToDTO(event types.Event) Event {
	var jobs []types.Job
	for _, job := range event.Jobs {
		jobs = append(jobs, *job)
	}
	participants := make([]Participant, 0)
	for user, job := range event.Participants {
		participants = append(participants, Participant{
			User:  user,
			JobId: job.Id,
		})
	}
	return Event{
		Id:           event.Id,
		Name:         event.Name,
		Open:         event.Open,
		Jobs:         jobs,
		Organizer:    event.Organizer,
		Participants: participants,
	}
}

func EventsToDTO(events []types.Event) []Event {
	var dtoEvents []Event
	for _, event := range events {
		dtoEvents = append(dtoEvents, EventToDTO(event))
	}
	return dtoEvents
}
