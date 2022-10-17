package main

import (
	"fmt"
	server "sdr/labo1/src"
	"sdr/labo1/src/config"
	"sdr/labo1/src/core"
	"sdr/labo1/src/utils"
)

func main() {
	utils.PrintServerWelcome()
	go server.Start(core.ReadConfig("server.json", &config.ServerConfiguration{}))
	var input string
	for {
		fmt.Scanln(&input)
		if input == "quit" {
			server.Stop()
			break
		}
	}
}
