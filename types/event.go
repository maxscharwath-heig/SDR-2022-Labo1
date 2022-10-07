package types

type Event struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Organizer User   `json:"organizer"`
	Jobs      []Job  `json:"jobs"`
	Open      bool   `json:"open"`
}
