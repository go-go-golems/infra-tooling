package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"

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
		if code, ok := parseWrappedExitCode(err.Error()); ok {
			os.Exit(code)
		}
		os.Exit(1)
	}
	if code := exitcode.Requested(); code != 0 {
		os.Exit(code)
	}
}

var wrappedExitCodeRE = regexp.MustCompile(`\(exit code ([0-9]+)\)`)

func parseWrappedExitCode(s string) (int, bool) {
	m := wrappedExitCodeRE.FindStringSubmatch(s)
	if m == nil {
		return 0, false
	}
	code, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, false
	}
	return code, true
}
