package network

import (
	"encoding/json"
	"sdr/labo1/types"
)

type Request[T any] struct {
	Credentials types.Credentials `json:"credentials,omitempty"`
	Data        T                 `json:"data"`
}

func (r Request[T]) ToJson() string {
	bytes, _ := json.Marshal(r)
	return string(bytes)
}

func FromJson[T any](jsonString string) Request[T] {
	var request Request[T]
	_ = json.Unmarshal([]byte(jsonString), &request)
	return request
}
