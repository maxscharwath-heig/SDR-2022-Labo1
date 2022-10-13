package main

import (
	"fmt"
	"golang.org/x/term"
	"net"
	. "sdr/labo1/core"
	"sdr/labo1/network"
	"sdr/labo1/types"
	"strconv"
	"strings"
	"syscall"
)

func stringPrompt(label string) string {
	for {
		fmt.Print(label + " ")
		var input string
		_, err := fmt.Scanln(&input)
		if err == nil {
			return strings.TrimSpace(input)
		}
	}
}

func passPrompt(label string) string {
	fmt.Print(label + " ")
	pass, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println("****")
	return string(pass)
}

// intPrompt asks for an int value using the label
func intPrompt(label string) int {
	for {
		fmt.Print(label + " ")
		var input string
		_, err := fmt.Scanln(&input)
		if err == nil {
			if i, err := strconv.Atoi(input); err == nil {
				return i
			}
		}
	}
}

func printWelcome() {
	fmt.Println("\n   _____ ____  ____ \n  / ___// __ \\/ __ \\\n  \\__ \\/ / / / /_/ /\n ___/ / /_/ / _, _/ \n/____/_____/_/ |_|")
	fmt.Println("Welcome to the SDR-Labo1 client")
	fmt.Println("")
}

func printHelp() {
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

func authenticate() types.Credentials {
	return types.Credentials{
		Username: stringPrompt("Enter username:"),
		Password: passPrompt("Enter password:"),
	}
}

func parseArgs(cmdRaw string) (string, []string, map[string]bool) {
	parsed := strings.Split(cmdRaw, " ")
	cmd := parsed[0]
	var args []string
	flags := make(map[string]bool)
	for _, arg := range parsed[1:] {
		if strings.HasPrefix(arg, "--") {
			flags[strings.TrimPrefix(arg, "--")] = true
		} else if strings.HasPrefix(arg, "-") {
			flags[strings.TrimPrefix(arg, "-")] = true
		} else {
			args = append(args, arg)
		}
	}
	return cmd, args, flags
}

func clientProcess(configuration types.ClientConfiguration) {
	printWelcome()
	printHelp()

	conn, entryMessages := connect(configuration.Type, configuration.FullUrl())

	for {
		cmd, args, flags := parseArgs(stringPrompt("Enter command [press h for help]:"))

		switch cmd {
		case "h":
			printHelp()
		case "create":
			request := network.Request[types.Event]{
				Credentials: authenticate(),
				Data: types.Event{
					Name: stringPrompt("Enter event name:"),
				},
			}
			jobsMap := make(map[string]types.Job)
			for {
				job := types.Job{
					Name:     stringPrompt("Enter job name:"),
					Capacity: intPrompt("Enter job capacity:"),
				}
				jobsMap[job.Name] = job
				if stringPrompt("Add another job? [y/n]") == "n" {
					break
				}
			}
			jobs := make([]types.Job, len(jobsMap))
			for _, job := range jobsMap {
				jobs = append(jobs, job)
			}
			request.Data.Jobs = jobs
			network.SendRequest(conn, "create", request)

		case "close":
			request := network.Request[int]{
				Credentials: authenticate(),
				Data:        intPrompt("Enter event id:"),
			}
			network.SendRequest(conn, "close", request)
		case "register":
			request := network.Request[types.Registration]{
				Credentials: authenticate(),
				Data: types.Registration{
					EventId: intPrompt("Enter event id:"),
					JobId:   intPrompt("Enter job id:"),
				},
			}
			network.SendRequest(conn, "register", request)
		case "show":
			eventId := 0
			if len(args) > 0 {
				eventId, _ = strconv.Atoi(args[0])
			} else {
				eventId = -1
			}
			type ShowRequest struct {
				EventId int
				Resume  bool
			}
			request := network.Request[ShowRequest]{
				Data: ShowRequest{
					EventId: eventId,
					Resume:  flags["resume"],
				},
			}
			network.SendRequest(conn, "show", request)
			fmt.Println("Waiting for response...")
			data := <-entryMessages
			body := network.FromJson[any](data.Body)
			fmt.Println(body)

		case "quit":
			disconnect(conn)
			return
		default:
			fmt.Println("Invalid command, try again")
		}
		fmt.Println("")
	}
}

func connect(protocol string, address string) (*net.TCPConn, chan network.Message) {
	tcpAddr, _ := net.ResolveTCPAddr(protocol, address)
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)
	//create channel to receive messages
	entryMessages := make(chan network.Message)
	go network.HandleReceiveData(conn, entryMessages)
	return conn, entryMessages
}

func disconnect(conn net.Conn) {
	conn.Close()
}

func main() {
	config := ReadConfig("config/client.json", types.ClientConfiguration{})
	clientProcess(config)
}
