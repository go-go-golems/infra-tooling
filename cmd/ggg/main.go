package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/infra-tooling/internal/cli"
)

func main() {
	root, err := cli.NewRootCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
