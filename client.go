package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	. "sdr/labo1/core"
	"sdr/labo1/types"
	"strconv"
	"strings"
)

// StringPrompt asks for a string value using the label
// TODO: add boolean param to remove echo
func StringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		println(label + " ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

// IntPrompt asks for an int value using the label
func IntPrompt(label string) int {
	var i int
	r := bufio.NewReader(os.Stdin)
	for {
		println(label + " ")
		_, err := fmt.Fscan(r, &i)
		if err == nil {
			break
		}
	}
	return i
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

func Authenticate() types.Credentials {
	return types.Credentials{
		Username: StringPrompt("Enter username:"),
		Password: StringPrompt("Enter password:"),
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

func ClientProcess(configuration types.ClientConfiguration) {
	PrintWelcome()
	PrintHelp()

	connect(configuration.Type, configuration.FullUrl())

	for {
		cmd, args, flags := parseArgs(StringPrompt("Enter command [press h for help]:"))

		switch cmd {
		case "h":
			PrintHelp()
		case "create":
			request := types.Request[types.Event]{
				Credentials: Authenticate(),
				Data: types.Event{
					Name: StringPrompt("Enter event name:"),
				},
			}
			jobsMap := make(map[string]types.Job)
			for {
				job := types.Job{
					Name:     StringPrompt("Enter job name:"),
					Capacity: IntPrompt("Enter job capacity:"),
				}
				jobsMap[job.Name] = job
				if StringPrompt("Add another job? [y/n]") == "n" {
					break
				}
			}
			jobs := make([]types.Job, len(jobsMap))
			for _, job := range jobsMap {
				jobs = append(jobs, job)
			}
			request.Data.Jobs = jobs
			println(ToJson(request))

		case "close":
			request := types.Request[int]{
				Credentials: Authenticate(),
				Data:        IntPrompt("Enter event id:"),
			}
			println(ToJson(request))
		case "register":
			request := types.Request[types.Registration]{
				Credentials: Authenticate(),
				Data: types.Registration{
					EventId: IntPrompt("Enter event id:"),
					JobId:   IntPrompt("Enter job id:"),
				},
			}
			println(ToJson(request))
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
			request := types.Request[ShowRequest]{
				Credentials: Authenticate(),
				Data: ShowRequest{
					EventId: eventId,
					Resume:  flags["resume"],
				},
			}
			println(ToJson(request))

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
	conn.Close()
}

func main() {
	config := ReadConfig("config/client.json", types.ClientConfiguration{})
	ClientProcess(config)
}
