// SDR - Labo 1
// Nicolas Crausaz & Maxime Scharwath

// Defines logging utilities used in the application

package utils

import (
	"fmt"
	"sdr/labo1/src/utils/colors"
	"time"
)

var enabled = true

// SetEnabled enable trace login globally
func SetEnabled(enable bool) {
	enabled = enable
}

func LogInfo(prefix string, data ...any) {
	Log(false, fmt.Sprintf("ℹ️ INFO (%s)", prefix), colors.Blue, data...)
}

func LogWarning(prefix string, data ...any) {
	Log(false, fmt.Sprintf("⚠️ WARNING (%s)", prefix), colors.Yellow, data...)
}

func LogSuccess(prefix string, data ...any) {
	Log(false, fmt.Sprintf("✅ SUCCESS (%s)", prefix), colors.Green, data...)
}

func LogError(prefix string, data ...any) {
	Log(false, fmt.Sprintf("❌ ERROR (%s)", prefix), colors.Red, data...)
}

func Log(force bool, prefix string, color string, data ...any) {
	if !enabled && !force {
		return
	}
	date := time.Now().Format("2006-01-02 15:04:05")
	var result []any
	result = append(result, color+fmt.Sprintf("[%s] %s:", date, prefix)+colors.Reset)
	result = append(result, data...)
	fmt.Println(result...)
}
