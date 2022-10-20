// SDR - Labo 1
// Nicolas Crausaz & Maxime Scharwath

package core

import (
	"os"
	"os/signal"
	"syscall"
)

// OnSigTerm handle end of execution with a custom handler
func OnSigTerm(handler func()) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		handler()
		os.Exit(1)
	}()
}
