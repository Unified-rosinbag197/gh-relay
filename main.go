package main

import (
	"fmt"
	"os"

	"github.ibm.com/soub4i/gh-relay/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
