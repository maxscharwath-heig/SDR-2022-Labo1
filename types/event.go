package types

import "fmt"

type Event struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Jobs      []Job  `json:"jobs"`
	Open      bool   `json:"open"`
	Organizer *User  `json:"organizer"`
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
