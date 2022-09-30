package types

type Configuration struct {
	Host  string `json:"host"`
	Port  int    `json:"port"`
	Users []User `json:"users"`
}
