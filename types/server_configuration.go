package types

import "fmt"

type ServerConfiguration struct {
	Type   string  `json:"type"`
	Host   string  `json:"host"`
	Port   int     `json:"port"`
	Users  []User  `json:"users"`
	Events []Event `json:events`
}

func (config ServerConfiguration) FullUrl() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
