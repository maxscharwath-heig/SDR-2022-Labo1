package types

import "fmt"

type Job struct {
	Id           int     `json:"id"`
	Name         string  `json:"name"`
	Capacity     int     `json:"capacity"`
	Participants []*User `json:"participants"`
}

func (job *Job) ToRow() string {
	return fmt.Sprintf("%d\t%s\t%d", job.Id, job.Name, job.Capacity)
}
