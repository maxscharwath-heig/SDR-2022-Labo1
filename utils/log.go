package utils

import (
	"fmt"
	"time"
)

func LogInfo(prefix string, data ...any) {
	log(prefix, "\033[33m", "\033[0m", data)
}

func LogError(data ...any) {
	log("error", "\033[1;31m", "\033[0m", data)
}

func log(prefix string, color string, reset string, data ...any) {
	date := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(color, fmt.Sprintf("[%s] (%s):", date, prefix), reset, data)
}
