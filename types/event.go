package types

type Event struct {
	id        int    `json:"id"`
	name      string `json:"name"`
	organizer User   `json:"organizer"`
	jobs      []Job  `json:"jobs"`
}
