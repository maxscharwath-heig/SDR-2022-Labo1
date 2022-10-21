// SDR - Labo 2
// Nicolas Crausaz & Maxime Scharwath

package config

// ClientConfiguration contains the information needed for the client to connects to a server
type ClientConfiguration struct {
	Servers       []string `json:"servers"`
	ShowInfosLogs bool     `json:"showInfosLogs"`
}
