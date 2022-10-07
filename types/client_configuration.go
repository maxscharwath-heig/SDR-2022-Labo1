package types

import "fmt"

type ClientConfiguration struct {
	Type string `json:"type"`
	Host string `json:"srvHost"`
	Port int    `json:"srvPort"`
}

func (config ClientConfiguration) FullUrl() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
