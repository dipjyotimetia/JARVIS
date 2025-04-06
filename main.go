package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/dipjyotimetia/jarvis/cmd"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %v\n", r)
			fmt.Fprintf(os.Stderr, "%s\n", debug.Stack())
			os.Exit(1)
		}
	}()

	cmd.Execute()
}
