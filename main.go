package main

import (
	"aks-coach/internal/cli"
	"os"
)

// entry point
func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
