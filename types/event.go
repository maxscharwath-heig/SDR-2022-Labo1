package types

import "fmt"

type Event struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	OrganizerId int    `json:"organizer_id"`
	Jobs        []Job  `json:"jobs"`
	Open        bool   `json:"open"`
	Organizer   User   `json:"organizer"`
}

// SetOrganizer set the organizer of the event
func (event *Event) SetOrganizer(organizer User) {
	event.Organizer = organizer
	event.OrganizerId = organizer.Id
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
