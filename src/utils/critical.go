package utils

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"sdr/labo1/src/utils/colors"
	"time"
)

var enableCriticDebug = false

func SetCriticDebug(enable bool) {
	enableCriticDebug = enable
}

// createCriticalSection access a critical section (for debug)
func CreateCriticalSection(name string) (start func(), end func()) {
	if !enableCriticDebug {
		return func() {}, func() {}
	}

	// Generate an identifier the critical section
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	id := hex.EncodeToString(b)

	start = func() {
		Log(true, fmt.Sprintf("CRITIC START [%s]", id), colors.BackgroundRed, fmt.Sprintf("🔒%s", name))
		time.Sleep(time.Second * 2)
	}
	end = func() {
		Log(true, fmt.Sprintf("CRITIC END   [%s]", id), colors.BackgroundRed, fmt.Sprintf("🔓%s", name))
	}
	return
}
