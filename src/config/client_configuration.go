// SDR - Labo 1
// Nicolas Crausaz & Maxime Scharwath

package config

import "fmt"

// ClientConfiguration contains the information needed for the client to connects
// to a server
type ClientConfiguration struct {
	Host          string `json:"srvHost"`
	Port          int    `json:"srvPort"`
	ShowInfosLogs bool   `json:"showInfosLogs"`
}

// FullUrl gets the formatted connection URL
func (config ClientConfiguration) FullUrl() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
