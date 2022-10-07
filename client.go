package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	. "sdr/labo1/core"
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

func ClientProcess(configuration types.ClientConfiguration) {
	PrintWelcome()
	PrintHelp()

	connect(configuration.Type, configuration.FullUrl())

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

		case "quit":
			// quit server
		default:
			fmt.Println("Invalid command, try again")
		}
		fmt.Println("")
	}
}

func connect(network string, address string) {

	TestCmd := "blabla"

	tcpAddr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}

	_, err = conn.Write([]byte(TestCmd))
	if err != nil {
		println("Write to server failed:", err.Error())
		os.Exit(1)
	}

	println("write to server = ", TestCmd)

	reply := make([]byte, 1024)

	_, err = conn.Read(reply)
	if err != nil {
		println("Write to server failed:", err.Error())
		os.Exit(1)
	}

	println("reply from server=", string(reply))

	disconnect(conn)
}

func disconnect(conn net.Conn) {
	if conn != nil {
		err := conn.Close()
		if err != nil {
			return // error on close
		}
	}
}

func main() {
	config := ReadConfig("config/client.json", types.ClientConfiguration{})
	ClientProcess(config)
}
