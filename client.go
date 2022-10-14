package main

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"net"
	"os"
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
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
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
	conn := connect(configuration.Type, configuration.FullUrl())
	protocol := network.ClientProtocol{Conn: conn, AuthFunc: authenticate}

	utils.PrintWelcome()
	utils.PrintHelp()
	for {
		cmd, args, flags := parseArgs(stringPrompt("Enter command [press h for help]:"))

		switch cmd {
		case "h":
			utils.PrintHelp()
		case "create":
			fmt.Println(protocol.SendRequest("create", func(auth any) any {
				event := dto.EventCreate{
					Name: stringPrompt("Enter event name:"),
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
				var jobs []dto.Job
				for _, job := range jobsMap {
					jobs = append(jobs, job)
				}
				event.Jobs = jobs
				return event
			}))
		case "close":
			fmt.Println(protocol.SendRequest("close", func(auth any) any {
				return dto.EventClose{
					EventId: intPrompt("Enter event id:"),
				}
			}))
		case "register":
			fmt.Println(protocol.SendRequest("register", func(auth any) any {
				return dto.EventRegister{
					EventId: intPrompt("Enter event id:"),
					JobId:   intPrompt("Enter job id:"),
				}
			}))
		case "show":
			eventId := 0
			if len(args) > 0 {
				eventId, _ = strconv.Atoi(args[0])
			} else {
				eventId = -1
			}
			response, _ := protocol.SendRequest("show", func(auth any) any {
				return dto.EventShow{
					EventId: eventId,
					Resume:  flags["resume"],
				}
			})
			if eventId != -1 {
				if flags["resume"] {
					displayEventFromIdResume(utils.FromJson[types.Event](response))
				} else {
					displayEventFromId(utils.FromJson[types.Event](response))
				}
			} else {
				displayEvents(utils.FromJson[[]types.Event](response))
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

func connect(protocol string, address string) *net.TCPConn {
	tcpAddr, _ := net.ResolveTCPAddr(protocol, address)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("Error connecting to server")
		os.Exit(1)
	}
	return conn
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

func displayEventFromId(event types.Event) {
	// TODO
	fmt.Println(event.ToRow())
}

func displayEventFromIdResume(event types.Event) {
	// TODO
	fmt.Println(event.ToRow())
	fmt.Println("RESUME")
}

func main() {
	config := ReadConfig("config/client.json", types.ClientConfiguration{})
	clientProcess(config)
}
