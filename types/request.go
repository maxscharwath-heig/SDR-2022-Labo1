package types

import "encoding/json"

type Request[T any] struct {
	Credentials Credentials `json:"credentials"`
	Data        T           `json:"data"`
}

func (r Request[T]) ToJson() string {
	bytes, _ := json.Marshal(r)
	return string(bytes)
}

func (r Request[T]) FromJson(data string) Request[T] {
	json.Unmarshal([]byte(data), &r)
	return r
}
