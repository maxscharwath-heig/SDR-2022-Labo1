package types

type Event struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	OrganizerId int    `json:"organizer_id"`
	Jobs        []Job  `json:"jobs"`
	Open        bool   `json:"open"`
	Organizer   User   `json:"organizer"`
}

//setOrganizer

func (event *Event) SetOrganizer(organizer User) {
	event.Organizer = organizer
	event.OrganizerId = organizer.Id
}
