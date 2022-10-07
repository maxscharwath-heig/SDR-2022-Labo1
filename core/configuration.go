package core

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadConfig[T any](path string, config T) T {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error reading file:", err.Error())
		return config
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error parsing file:", err.Error())
		return config
	}
	return config
}
