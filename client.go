package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sdr/labo1/types"
	"strings"
)

// StringPrompt asks for a string value using the label
func StringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

func readConfig() {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	configuration := types.Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(configuration)
}

func main() {
	readConfig()
	name := StringPrompt("What is your name?")
	fmt.Printf("Hello, %s!\n", name)
}
