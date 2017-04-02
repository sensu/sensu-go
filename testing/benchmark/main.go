package main

import (
	"log"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "bench",
		Short: "bench",
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
