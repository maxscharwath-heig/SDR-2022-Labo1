package types

type Job struct {
	id       int    `json:"id"`
	name     string `json:"name"`
	capacity int    `json:"capacity"`
}
