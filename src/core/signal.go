// SDR - Labo 1
// Nicolas Crausaz & Maxime Scharwath

package core

import (
	"os"
	"os/signal"
	"syscall"
)

func OnSigTerm(handler func()) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		handler()
		os.Exit(1)
	}()
}
