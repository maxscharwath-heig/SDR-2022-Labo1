package main

import (
	server "sdr/labo1/src"
	"sdr/labo1/src/config"
	"sdr/labo1/src/core"
)

func main() {
	server.Start(core.ReadConfig("server.json", &config.ServerConfiguration{}))
}
