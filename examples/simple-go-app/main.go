package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "app",
		Short: "A simple example app",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello from godeps-guard example!")
		},
	}
	rootCmd.Execute()
}
