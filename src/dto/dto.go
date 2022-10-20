// SDR - Labo 1
// Nicolas Crausaz & Maxime Scharwath

package dto

import (
	"fmt"
	"sdr/labo1/src/types"
)

// Participant represents a registration of a user in an event's job
type Participant struct {
	User  types.User `json:"user"`
	JobId int        `json:"jobId"`
}

// Event contains all the data of an event
type Event struct {
	Id           int           `json:"id"`
	Name         string        `json:"name"`
	Open         bool          `json:"open"`
	Jobs         []types.Job   `json:"jobs"`
	Organizer    types.User    `json:"organizer"`
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

// EventRegister defines required data for a register request
type EventRegister struct {
	EventId int `json:"eventId"`
	JobId   int `json:"jobId"`
}

// EventClose defines required data for a close request
type EventClose struct {
	EventId int `json:"eventId"`
}

// Job defines required data for a job in a create request
type Job struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}

// EventCreate defines required data for a create request
type EventCreate struct {
	Name string `json:"name"`
	Jobs []Job  `json:"jobs"`
}

// EventShow defines required data for a show request
type EventShow struct {
	EventId int  `json:"eventId"`
	Resume  bool `json:"resume"`
}
