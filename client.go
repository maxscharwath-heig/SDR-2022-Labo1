package main

import (
	"fmt"
	"golang.org/x/term"
	"net"
	. "sdr/labo1/core"
	"sdr/labo1/dto"
	"sdr/labo1/network"
	"sdr/labo1/types"
	"sdr/labo1/utils"
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
	pass, _ := term.ReadPassword(syscall.Stdin)
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
	utils.PrintWelcome()
	utils.PrintHelp()

	conn, entryMessages := connect(configuration.Type, configuration.FullUrl())

	for {
		cmd, args, flags := parseArgs(stringPrompt("Enter command [press h for help]:"))

		switch cmd {
		case "h":
			utils.PrintHelp()
		case "create":
			request := network.Request[dto.EventCreate]{
				Credentials: authenticate(),
				Data: dto.EventCreate{
					Name: stringPrompt("Enter event name:"),
				},
			}
			jobsMap := make(map[string]dto.Job)
			for {
				job := dto.Job{
					Name:     stringPrompt("Enter job name:"),
					Capacity: intPrompt("Enter job capacity:"),
				}
				jobsMap[job.Name] = job
				if stringPrompt("Add another job? [y/n]") == "n" {
					break
				}
			}
			jobs := make([]dto.Job, len(jobsMap))
			for _, job := range jobsMap {
				jobs = append(jobs, job)
			}
			request.Data.Jobs = jobs
			network.SendRequest(conn, "create", request)

		case "close":
			request := network.Request[dto.EventClose]{
				Credentials: authenticate(),
				Data: dto.EventClose{
					EventId: intPrompt("Enter event id:"),
				},
			}
			network.SendRequest(conn, "close", request)
		case "register":
			request := network.Request[dto.EventRegister]{
				Credentials: authenticate(),
				Data: dto.EventRegister{
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

			request := network.Request[dto.EventShow]{
				Data: dto.EventShow{
					EventId: eventId,
					Resume:  flags["resume"],
				},
			}
			network.SendRequest(conn, "show", request)
			data := <-entryMessages
			body := network.RequestFromJson[[]types.Event](data.Body)

			if eventId > 0 {
				if flags["resume"] {
					displayEventFromIdResume(body.Data)
				} else {
					displayEventFromId(body.Data)
				}
			} else {
				displayEvents(body.Data)
			}

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

func displayEvents(events []types.Event) {
	headers := []string{"Number", "Name", "Organizer name", "open"}
	var printableEventRow []string
	for _, event := range events {
		printableEventRow = append(printableEventRow, event.ToRow())
	}

	utils.PrintTable(headers, printableEventRow)
}

func displayEventFromId(events []types.Event) {
	// TODO
}

func displayEventFromIdResume(events []types.Event) {
	// TODO
}

func main() {
	config := ReadConfig("config/client.json", types.ClientConfiguration{})
	clientProcess(config)
}
