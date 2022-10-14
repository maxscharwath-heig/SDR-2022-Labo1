package types

import "fmt"

type UserWithPassword struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ServerConfiguration struct {
	Type   string             `json:"type"`
	Host   string             `json:"host"`
	Port   int                `json:"port"`
	Users  []UserWithPassword `json:"users"`
	Events []Event            `json:"events"`
}

func (config ServerConfiguration) FullUrl() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}

func (config ServerConfiguration) GetUsers() []User {
	var users []User
	for _, user := range config.Users {
		users = append(users, User{
			Id:       user.Id,
			Username: user.Username,
			Password: user.Password,
		})
	}
	return users
}
