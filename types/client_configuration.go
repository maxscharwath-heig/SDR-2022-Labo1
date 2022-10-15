package types

import "fmt"

type ClientConfiguration struct {
	Host          string `json:"srvHost"`
	Port          int    `json:"srvPort"`
	ShowInfosLogs bool   `json:"showInfosLogs"`
}

func (config ClientConfiguration) FullUrl() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
