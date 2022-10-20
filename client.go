package main

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"net"
	"os"
	"sdr/labo1/src/config"
	"sdr/labo1/src/core"
	"sdr/labo1/src/dto"
	"sdr/labo1/src/network"
	"sdr/labo1/src/types"
	"sdr/labo1/src/utils"
	"sdr/labo1/src/utils/colors"
	"strconv"
	"strings"
	"syscall"
	"time"
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

// authenticate prompts the user for his credentials
func authenticate() types.Credentials {
	return types.Credentials{
		Username: stringPrompt("Enter username:"),
		Password: passPrompt("Enter password:"),
	}
}

// parseArgs parses the command line arguments
// and returns the command, the arguments and the flags
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

// clientProcess is the main function of the client
func clientProcess(configuration config.ClientConfiguration) {
	utils.PrintClientWelcome()
	conn := connect("tcp", configuration.FullUrl())
	protocol := network.CreateClientProtocol(conn, authenticate)
	core.OnSigTerm(func() {
		disconnect(conn)
	})
	protocol.OnClose(func() {
		fmt.Println()
		PrintError("Connection closed by server")
		os.Exit(1)
	})
	utils.PrintHelp()
	for {
		cmd, args, flags := parseArgs(stringPrompt("Enter command [press h for help]:"))

		switch cmd {
		case "h":
			utils.PrintHelp()
		case "create":
			json, err := protocol.SendRequest("create", func(auth network.AuthId) any {
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
			})
			if err != nil {
				PrintError(err.Error())
			} else {
				event, responseError := network.ParseResponse[*dto.Event](json)
				if responseError != nil {
					PrintError(responseError.Error())
				} else {
					PrintSuccess(fmt.Sprintf("Event created: %s#%d", event.Name, event.Id))
					displayEventFromId(event)
				}
			}
		case "close":
			json, err := protocol.SendRequest("close", func(auth network.AuthId) any {
				return dto.EventClose{
					EventId: intPrompt("Enter event id:"),
				}
			})
			if err != nil {
				PrintError(err.Error())
			} else {
				event, responseError := network.ParseResponse[*dto.Event](json)
				if responseError != nil {
					PrintError(responseError.Error())
				} else {
					PrintSuccess(fmt.Sprintf("Event closed: %s#%d", event.Name, event.Id))
				}
			}
		case "register":
			json, err := protocol.SendRequest("register", func(auth network.AuthId) any {
				return dto.EventRegister{
					EventId: intPrompt("Enter event id:"),
					JobId:   intPrompt("Enter job id:"),
				}
			})
			if err != nil {
				fmt.Println(colors.Red + err.Error() + colors.Reset)
			} else {
				event, responseError := network.ParseResponse[*dto.Event](json)
				if responseError != nil {
					PrintError(responseError.Error())
				} else {
					PrintSuccess(fmt.Sprintf("Registered to event: %s#%d", event.Name, event.Id))
				}
			}
		case "show":
			eventId := -1
			if len(args) > 0 {
				eventId, _ = strconv.Atoi(args[0])
			}
			json, err := protocol.SendRequest("show", func(auth network.AuthId) any {
				return dto.EventShow{
					EventId: eventId,
					Resume:  flags["resume"],
				}
			})
			if err != nil {
				fmt.Println(colors.Red + err.Error() + colors.Reset)
			} else {
				if eventId != -1 {
					event, responseError := network.ParseResponse[*dto.Event](json)

					if responseError != nil {
						PrintError(responseError.Error())
						break
					}

					if flags["resume"] {
						displayEventFromIdResume(event)
					} else {
						displayEventFromId(event)
					}
				} else {
					events, responseError := network.ParseResponse[[]dto.Event](json)
					if responseError != nil {
						PrintError(responseError.Error())
						break
					}
					displayEvents(events)
				}
			}
		case "quit":
			disconnect(conn)
			return
		default:
			PrintError(fmt.Sprintf("Unknown command \"%s\"", cmd))
		}
	}
}

func connect(protocol string, address string) *net.TCPConn {
	fmt.Print(colors.Yellow + "Connecting")
	// print dots while connecting
	isConnecting := make(chan bool)
	go func() {
		for {
			select {
			case <-isConnecting:
				return
			default:
				fmt.Print(".")
				time.Sleep(250 * time.Millisecond)
			}
		}
	}()
	tcpAddr, _ := net.ResolveTCPAddr(protocol, address)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	isConnecting <- true
	fmt.Print(colors.Reset)
	if err != nil {
		PrintError("Connection failed")
		os.Exit(1)
	}
	PrintSuccess("Connection established")
	return conn
}

func disconnect(conn net.Conn) {
	fmt.Print(colors.Yellow+"Disconnecting", colors.Reset)
	conn.Close()
}

func displayEvents(events []dto.Event) {
	headers := []string{"Number", "Name", "Organizer name", "open"}
	var printableEventRows []string
	for _, event := range events {
		printableEventRows = append(printableEventRows, event.ToRow())
	}

	utils.PrintTable(headers, printableEventRows)
}

func displayEventFromId(event *dto.Event) {
	if event == nil {
		return
	}

	fmt.Printf("Event #%d: %s \n", event.Id, event.Name)
	fmt.Println("List of jobs:")

	headers := []string{"Number", "Name", "Max capacity"}
	var printableJobsRow []string
	for _, job := range event.Jobs {
		printableJobsRow = append(printableJobsRow, job.ToRow())
	}

	utils.PrintTable(headers, printableJobsRow)
}

func displayEventFromIdResume(event *dto.Event) {
	if event == nil {
		return
	}
	fmt.Printf("Event #%d: %s \n", event.Id, event.Name)
	fmt.Println("Current board of registrations")

	headers := []string{" "}
	var rows []string

	type jobData struct {
		Index int
		Job   types.Job
	}
	var jobs = make(map[int]jobData)
	for index, job := range event.Jobs {
		headers = append(headers, fmt.Sprintf("%s#%d (%d/%d)", job.Name, job.Id, job.Count, job.Capacity))
		jobs[job.Id] = jobData{Index: index, Job: job}
	}
	for _, participant := range event.Participants {
		if job, ok := jobs[participant.JobId]; ok {
			participation := make([]bool, len(event.Jobs))
			participation[job.Index] = true
			rows = append(rows, formattedJobRow(participant.User.Username, participation))
		}
	}

	utils.PrintTable(headers, rows)
}

func formattedJobRow(username string, row []bool) string {
	values := []string{username}
	for _, value := range row {
		if value {
			values = append(values, "x")
		} else {
			values = append(values, " ")
		}
	}
	return strings.Join(values, "\t")
}

func PrintSuccess(message string) {
	fmt.Println("✅ " + colors.Green + message + colors.Reset)
}

func PrintError(message string) {
	fmt.Println("❌ " + colors.Red + message + colors.Reset)
}

func main() {
	clientConfiguration := core.ReadConfig("client.json", config.ClientConfiguration{})
	utils.SetEnabled(clientConfiguration.ShowInfosLogs)
	clientProcess(clientConfiguration)
}
