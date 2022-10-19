package types

import "fmt"

type Event struct {
	Id           int
	Name         string
	Jobs         map[int]*Job
	Open         bool
	Organizer    *User
	Participants map[*User]*Job
}

func (event *Event) Unregister(user *User) {
	if job := event.Participants[user]; job != nil {
		job.Count--
	}
	delete(event.Participants, user)
}

func (event *Event) Register(user *User, jobId int) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}
	if job, ok := event.Jobs[jobId]; ok {
		if job.Count < job.Capacity {
			event.Unregister(user)
			event.Participants[user] = job
			job.Count++
			return nil
		}
	}
	return fmt.Errorf("job not found")
}
