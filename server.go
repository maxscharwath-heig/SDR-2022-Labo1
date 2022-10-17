package main

import (
	"fmt"
	server "sdr/labo1/src"
	"sdr/labo1/src/config"
	"sdr/labo1/src/core"
)

func main() {
	go server.Start(core.ReadConfig("server.json", &config.ServerConfiguration{}))
	// close server when type "exit"
	var input string
	for {
		fmt.Scanln(&input)
		if input == "exit" {
			server.Stop()
			break
		}
	}
}
