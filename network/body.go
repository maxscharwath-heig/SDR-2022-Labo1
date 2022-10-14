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

func RequestFromJson[T any](jsonString string) Request[T] {
	var request Request[T]
	_ = json.Unmarshal([]byte(jsonString), &request)
	return request
}

type Response[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

func (r Response[T]) ToJson() string {
	bytes, _ := json.Marshal(r)
	return string(bytes)
}

func ResponseFromJson[T any](jsonString string) Response[T] {
	var response Response[T]
	_ = json.Unmarshal([]byte(jsonString), &response)
	return response
}
