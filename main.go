package main

import (
	"os"

	"github.com/forge/forge/src/forge/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
