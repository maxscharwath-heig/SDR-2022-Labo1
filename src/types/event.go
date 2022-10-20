// SDR - Labo 1
// Nicolas Crausaz & Maxime Scharwath

package types

import "fmt"

// Event contains all the data of an event
type Event struct {
	Id           int
	Name         string
	Jobs         map[int]*Job
	Open         bool
	OrganizerId  int
	Participants map[int]int
}

// Unregister removes a user from a job that was previously registered
func (event *Event) Unregister(userId int) {
	if jobId, ok := event.Participants[userId]; ok {
		event.Jobs[jobId].Count--
	}
	delete(event.Participants, userId)
}

// Register adds an user to a job
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
