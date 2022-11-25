// SDR - Labo 2
// Nicolas Crausaz & Maxime Scharwath

// Defines prompt utilities used in the application

package utils

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// StringPrompt get user input as a string
func StringPrompt(label string) string {
	for {
		fmt.Print(label + " ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err == nil {
			return strings.TrimSpace(input)
		}
	}
}

// PassPrompt get user password as a string with echo off
func PassPrompt(label string) string {
	fmt.Print(label + " ")
	pass, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println("****")
	return string(pass)
}

// IntPrompt get user input as a int
func IntPrompt(label string) int {
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

// ParseArgs parses the command line arguments
// and returns the command, the arguments and the flags
func ParseArgs(cmdRaw string) (string, []string, map[string]bool) {
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
