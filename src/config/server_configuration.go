// SDR - Labo 1
// Nicolas Crausaz & Maxime Scharwath

package config

import (
	"fmt"
	"sdr/labo1/src/dto"
	"sdr/labo1/src/types"
)

// UserWithPassword
type UserWithPassword struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ServerConfiguration contains the information
type ServerConfiguration struct {
	Host          string             `json:"host"`
	Port          int                `json:"port"`
	Users         []UserWithPassword `json:"users"`
	Events        []dto.Event        `json:"events"`
	Debug         bool               `json:"debug"`
	ShowInfosLogs bool               `json:"showInfosLogs"`
}

func (config ServerConfiguration) FullUrl() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}

func (config ServerConfiguration) GetData() (users map[int]*types.User, events []*types.Event) {
	users = make(map[int]*types.User)
	for _, user := range config.Users {
		users[user.Id] = &types.User{
			Id:       user.Id,
			Username: user.Username,
			Password: user.Password,
		}
	}

	for _, event := range config.Events {
		e := &types.Event{
			Id:           event.Id,
			Name:         event.Name,
			Open:         event.Open,
			OrganizerId:  event.Organizer.Id,
			Jobs:         make(map[int]*types.Job),
			Participants: make(map[int]int),
		}
		for _, job := range event.Jobs {
			e.Jobs[job.Id] = &types.Job{
				Id:       job.Id,
				Name:     job.Name,
				Capacity: job.Capacity,
			}
		}
		for _, participant := range event.Participants {
			e.Register(participant.User.Id, participant.JobId)
		}
		events = append(events, e)
	}
	return
}
