//go:build cli

package main

import (
	"os"
)

func main() {
	// CLI-only mode for Windows
	// This binary doesn't include Wails/WebView, so always run CLI
	if len(os.Args) < 2 {
		printCLIHelp()
		os.Exit(1)
	}

	runCLI()
}
