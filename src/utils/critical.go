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

func CreateCriticalSection(name string, callback func()) {
	if !enableCriticDebug {
		callback()
		return
	}
	// Generate an identifier the critical section
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	id := hex.EncodeToString(b)

	Log(true, fmt.Sprintf("CRITIC START [%s]", id), colors.BackgroundRed, fmt.Sprintf("🔒%s", name))
	time.Sleep(time.Second * 2)
	callback()
	Log(true, fmt.Sprintf("CRITIC END   [%s]", id), colors.BackgroundRed, fmt.Sprintf("🔓%s", name))
}
