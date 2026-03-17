package util

import (
	"fmt"
	"os"
)

var (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	Bold   = "\033[1m"
)

func init() {
	// Disable colors if NO_COLOR is set or if stdout is not a TTY
	// (Simple check: if we aren't in a terminal, we might want to disable)
	if os.Getenv("NO_COLOR") != "" {
		DisableColors()
	}
}

func DisableColors() {
	Reset = ""
	Red = ""
	Green = ""
	Yellow = ""
	Blue = ""
	Purple = ""
	Cyan = ""
	Gray = ""
	Bold = ""
}

func Colorize(s string, color string) string {
	return color + s + Reset
}

func Success(s string) string {
	return Colorize(s, Green)
}

func Error(s string) string {
	return Colorize(s, Red)
}

func Warning(s string) string {
	return Colorize(s, Yellow)
}

func Info(s string) string {
	return Colorize(s, Cyan)
}

func RiskScore(score int) string {
	s := fmt.Sprintf("%d/10", score)
	if score >= 7 {
		return Colorize(s, Red)
	}
	if score >= 4 {
		return Colorize(s, Yellow)
	}
	return Colorize(s, Green)
}
