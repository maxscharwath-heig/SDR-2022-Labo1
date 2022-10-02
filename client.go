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
// TODO: add boolean param to remove echo
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

func PrintWelcome() {
	fmt.Println("\n   _____ ____  ____ \n  / ___// __ \\/ __ \\\n  \\__ \\/ / / / /_/ /\n ___/ / /_/ / _, _/ \n/____/_____/_/ |_|")
	fmt.Println("Welcome to the SDR-Labo1 client")
	fmt.Println("")
}

func PrintHelp() {
	fmt.Println("Please type the wished command")
	fmt.Println("List of commands:")
	fmt.Println("- create")
	fmt.Println("- close")
	fmt.Println("- register")
	fmt.Println("- show")
	fmt.Println("- show [number]")
	fmt.Println("- show [number] --resume")
	fmt.Println("_________________________")
}

// AskCredentials gets the user input for its credential, validates them and returns them
func AskCredentials() (string, string) {
	username := StringPrompt("Enter your username:")
	password := StringPrompt("Enter your password:") // TODO: hide echo while typing password

	// TODO: validate inputs

	return username, password
}

func Authenticate(username string, password string) {
	// TODO: auth to the server
}

func ClientProcess() {
	PrintWelcome()
	PrintHelp()

	for {
		cmd := StringPrompt("Enter command [press h for help]:")

		switch cmd {
		case "h":
			PrintHelp()
		case "create":
			Authenticate(AskCredentials())

			// TODO
			name := StringPrompt("    Enter event name:")
			jobs := StringPrompt("    Enter jobs: [name, number of slots]")
			fmt.Println(name, jobs)

		case "close":
			Authenticate(AskCredentials())
			fmt.Println("")
		case "register":
			Authenticate(AskCredentials())
		case "show":
			fmt.Println("")
		default:
			fmt.Println("Invalid command, try again")
		}
		fmt.Println("")
	}
}

func main() {
	readConfig()
	ClientProcess()
}
