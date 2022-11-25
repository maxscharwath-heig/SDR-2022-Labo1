// SDR - Labo 2
// Nicolas Crausaz & Maxime Scharwath

package main

import (
	"flag"
	"fmt"
	server "sdr/labo1/src"
	"sdr/labo1/src/config"
	"sdr/labo1/src/core"
	"sdr/labo1/src/utils"
)

func main() {
	utils.PrintServerWelcome()

	config := core.ReadConfig("server.json", &config.ServerConfiguration{})
	flagId := flag.Int("id", 0, "# of the server")
	flag.Parse()
	config.Id = *flagId
	if config.Id < 0 || config.Id >= len(config.Servers) {
		panic("Invalid server number")
	}

	go server.Start(config)

	/*core.OnSigTerm(func() {
		fmt.Println("Stopping server...")
		server.Stop()
	})*/

	var input string
	for {
		fmt.Scanln(&input)
		if input == "quit" {
			server.Stop()
			break
		}
	}
}
