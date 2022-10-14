package utils

import "encoding/json"

func FromJson[T any](data string) T {
	var result T
	_ = json.Unmarshal([]byte(data), &result)
	return result
}
