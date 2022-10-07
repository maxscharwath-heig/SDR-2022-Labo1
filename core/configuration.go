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
	}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error parsing file:", err.Error())
	}

	err = file.Close()
	if err != nil {
		fmt.Println("Error closing file:", err.Error())
	}
	return config
}
