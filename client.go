package main

import (
	"fmt"
	"math/rand"
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
	"time"
)

// authenticate prompts the user for his credentials
func authenticate() types.Credentials {
	return types.Credentials{
		Username: utils.StringPrompt("Enter username:"),
		Password: utils.PassPrompt("Enter password:"),
	}
}

// clientProcess is the main function of the client
func clientProcess(configuration config.ClientConfiguration) {
	utils.PrintClientWelcome()

	server := utils.StringPrompt("Enter the server address (default: random):")
	if server == "" {
		server = configuration.Servers[rand.Intn(len(configuration.Servers))]
	}

	conn := connect("tcp", server)
	protocol := network.CreateClientProtocol(conn, authenticate)
	core.OnSigTerm(func() {
		disconnect(conn)
	})
	protocol.OnClose(func() {
		fmt.Println()
		utils.PrintError("Connection closed by server")
		os.Exit(1)
	})
	utils.PrintHelp()
	for {
		cmd, args, flags := utils.ParseArgs(utils.StringPrompt("Enter command [press h for help]:"))

		switch cmd {
		case "h":
			utils.PrintHelp()
		case "create":
			json, err := protocol.SendRequest("create", func(auth network.AuthId) any {
				event := dto.EventCreate{
					Name: utils.StringPrompt("Enter event name:"),
				}
				jobsMap := make(map[string]dto.Job)
				for {
					job := dto.Job{
						Name:     utils.StringPrompt("Enter job name:"),
						Capacity: utils.IntPrompt("Enter job capacity:"),
					}
					jobsMap[job.Name] = job
					if utils.StringPrompt("Add another job? [y/n]") == "n" {
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
				utils.PrintError(err.Error())
			} else {
				event, responseError := network.ParseResponse[*dto.Event](json)
				if responseError != nil {
					utils.PrintError(responseError.Error())
				} else {
					utils.PrintSuccess(fmt.Sprintf("Event created: %s#%d", event.Name, event.Id))
					displayEventFromId(event)
				}
			}
		case "close":
			json, err := protocol.SendRequest("close", func(auth network.AuthId) any {
				return dto.EventClose{
					EventId: utils.IntPrompt("Enter event id:"),
				}
			})
			if err != nil {
				utils.PrintError(err.Error())
			} else {
				event, responseError := network.ParseResponse[*dto.Event](json)
				if responseError != nil {
					utils.PrintError(responseError.Error())
				} else {
					utils.PrintSuccess(fmt.Sprintf("Event closed: %s#%d", event.Name, event.Id))
				}
			}
		case "register":
			json, err := protocol.SendRequest("register", func(auth network.AuthId) any {
				return dto.EventRegister{
					EventId: utils.IntPrompt("Enter event id:"),
					JobId:   utils.IntPrompt("Enter job id:"),
				}
			})
			if err != nil {
				fmt.Println(colors.Red + err.Error() + colors.Reset)
			} else {
				event, responseError := network.ParseResponse[*dto.Event](json)
				if responseError != nil {
					utils.PrintError(responseError.Error())
				} else {
					utils.PrintSuccess(fmt.Sprintf("Registered to event: %s#%d", event.Name, event.Id))
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
						utils.PrintError(responseError.Error())
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
						utils.PrintError(responseError.Error())
						break
					}
					displayEvents(events)
				}
			}
		case "quit":
			disconnect(conn)
			return
		default:
			utils.PrintError(fmt.Sprintf("Unknown command \"%s\"", cmd))
		}
	}
}

// connect makes client connects to server
func connect(protocol string, address string) *net.TCPConn {
	fmt.Print(colors.Yellow + fmt.Sprintf("Connecting to %s://%s", protocol, address) + colors.Reset)
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
		utils.PrintError("Connection failed")
		os.Exit(1)
	}
	utils.PrintSuccess("Connection established")
	return conn
}

// disconnect quit the client's connection
func disconnect(conn net.Conn) {
	fmt.Print(colors.Yellow+"Disconnecting", colors.Reset)
	conn.Close()
}

// Display events as table format
func displayEvents(events []dto.Event) {
	headers := []string{"Number", "Name", "Organizer name", "open"}
	var printableEventRows []string
	for _, event := range events {
		printableEventRows = append(printableEventRows, event.ToRow())
	}

	utils.PrintTable(headers, printableEventRows)
}

// Display an event as table format
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

// Display an event's resume as table format
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

func main() {
	clientConfiguration := core.ReadConfig("client.json", config.ClientConfiguration{})
	utils.SetEnabled(clientConfiguration.ShowInfosLogs)
	clientProcess(clientConfiguration)
}
