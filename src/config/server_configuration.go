package config

import (
	"fmt"
	"sdr/labo1/src/dto"
	"sdr/labo1/src/types"
)

type UserWithPassword struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

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

func (config ServerConfiguration) GetData() (users []*types.User, events []*types.Event) {
	for _, user := range config.Users {
		users = append(users, &types.User{
			Id:       user.Id,
			Username: user.Username,
			Password: user.Password,
		})
	}

	for _, event := range config.Events {
		e := &types.Event{
			Id:           event.Id,
			Name:         event.Name,
			Open:         event.Open,
			Organizer:    types.FindUser(users, event.Organizer.Id),
			Jobs:         make(map[int]*types.Job),
			Participants: make(map[*types.User]*types.Job),
		}
		for _, job := range event.Jobs {
			e.Jobs[job.Id] = &types.Job{
				Id:       job.Id,
				Name:     job.Name,
				Capacity: job.Capacity,
			}
		}
		if e.Organizer == nil {
			panic("Organizer not found")
		}
		for _, participant := range event.Participants {
			e.Register(types.FindUser(users, participant.User.Id), participant.JobId)
		}
		events = append(events, e)
	}
	return
}
