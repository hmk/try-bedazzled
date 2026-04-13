package main

import (
	"fmt"
	"os"
)

// Version is set at build time via -ldflags
var Version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("try-bedazzled " + Version)
		os.Exit(0)
	}

	// TODO: Wire up Cobra commands
	fmt.Fprintln(os.Stderr, "try-bedazzled: not yet implemented")
	os.Exit(1)
}
