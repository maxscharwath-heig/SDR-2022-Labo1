package dto

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
	EventId int
	Resume  bool
}
