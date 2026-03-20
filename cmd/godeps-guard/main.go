package main

import (
	"os"

	"github.com/ashishsalunkhe/godeps-guard/internal/cli"
	"github.com/joho/godotenv"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	if err := cli.Execute(version); err != nil {
		os.Exit(1)
	}
}
