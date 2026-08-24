package main

import (
	"os"

	"github.com/ugurekmekci01/sshmark/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
