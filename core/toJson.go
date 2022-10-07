package core

import "encoding/json"

func ToJson[T any](value T) string {
	marshal, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(marshal)
}
