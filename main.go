package main

import (
	"os"

	"github.com/BillyDyball/grep-news/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
