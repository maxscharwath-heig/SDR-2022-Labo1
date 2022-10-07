package types

import "encoding/json"

type Event struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Organizer User   `json:"organizer"`
	Jobs      []Job  `json:"jobs"`
	Open      bool   `json:"open"`
}

func (e Event) ToJson() string {
	marshal, err := json.Marshal(e)
	if err != nil {
		return ""
	}
	return string(marshal)
}
