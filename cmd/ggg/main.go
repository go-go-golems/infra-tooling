package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-go-golems/infra-tooling/internal/cli"
	"github.com/go-go-golems/infra-tooling/internal/exitcode"
)

func main() {
	root, err := cli.NewRootCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if err := root.Execute(); err != nil {
		var exitErr exitcode.Error
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}
