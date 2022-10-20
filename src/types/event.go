// SDR - Labo 1
// Nicolas Crausaz & Maxime Scharwath

package types

import "fmt"

type Event struct {
	Id           int
	Name         string
	Jobs         map[int]*Job
	Open         bool
	OrganizerId  int
	Participants map[int]int
}

func (event *Event) Unregister(userId int) {
	if jobId, ok := event.Participants[userId]; ok {
		event.Jobs[jobId].Count--
	}
	delete(event.Participants, userId)
}

func (event *Event) Register(userId int, jobId int) error {
	if job, ok := event.Jobs[jobId]; ok {
		if job.Count < job.Capacity {
			event.Unregister(userId)
			event.Participants[userId] = jobId
			job.Count++
			return nil
		}
	}
	return fmt.Errorf("job not found")
}
