package main

import (
	"embed"
	"fmt"
	"os"
	"slices"

	"qbittorrent-file-matcher/backend"
)

//go:embed build/windows/info.json
var versionInfo embed.FS

// getAppVersion returns version from embedded info.json
func getAppVersion() string {
	return backend.GetVersion(versionInfo)
}

func isCLICommand(arg string) bool {
	supportedCommands := []string{
		"match",
		"help", "--help", "-h",
		"version", "--version", "-v",
	}

	return slices.Contains(supportedCommands, arg)
}

func runCLI() {
	if len(os.Args) < 2 {
		printCLIHelp()
		os.Exit(1)
		return
	}

	command := os.Args[1]

	switch command {
	case "match":
		if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
			printMatchHelp()
			return
		}
		runMatchCommand()

	case "help", "--help", "-h":
		printCLIHelp()

	case "version", "--version", "-v":
		fmt.Printf("qbittorrent-file-matcher v%s\n", getAppVersion())

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n\n", command)
		printCLIHelp()
		os.Exit(1)
	}
}

func printCLIHelp() {
	fmt.Println("qbittorrent-file-matcher - Match torrent files with files on disk")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  qbittorrent-file-matcher              Launch GUI application")
	fmt.Println("  qbittorrent-file-matcher <command>    Run CLI command")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  match       Match and rename torrent files")
	fmt.Println("  help        Show this help message")
	fmt.Println("  version     Show version information")
	fmt.Println()
	fmt.Println("Run 'qbittorrent-file-matcher <command> --help' for more information on a command")
}

func printMatchHelp() {
	fmt.Println("Usage: qbittorrent-file-matcher match [flags]")
	fmt.Println()
	fmt.Println("Match torrent files with files on disk and rename them in qBittorrent")
	fmt.Println()
	fmt.Println("Required flags:")
	fmt.Println("  --url <url>              qBittorrent WebUI URL (e.g., http://localhost:8080)")
	fmt.Println("  --hash <hash>            Torrent hash to match")
	fmt.Println("  --path <path>            Directory path to scan for files")
	fmt.Println()
	fmt.Println("Optional flags:")
	fmt.Println("  -u, --username <user>    qBittorrent username")
	fmt.Println("  -p, --password <pass>    qBittorrent password")
	fmt.Println("  --same-ext               Only match files with same extension (default: true)")
	fmt.Println("  --no-same-ext            Allow matching files with different extensions")
	fmt.Println("  --skip-unmatched         Set priority to 0 for unmatched files")
	fmt.Println("  -r, --recheck            Trigger torrent recheck after applying renames")
	fmt.Println("  --dry-run                Show what would be done without making changes")
	fmt.Println("  -a, --auto               Auto-select first match (no interactive prompts)")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  QBT_URL                  Default qBittorrent WebUI URL")
	fmt.Println("  QBT_USERNAME             Default username")
	fmt.Println("  QBT_PASSWORD             Default password (more secure than command line)")
	fmt.Println()
	fmt.Println("Interactive mode:")
	fmt.Println("  When multiple files match the same size, you'll be prompted to select one.")
	fmt.Println("  Use --auto to skip prompts and auto-select the first match.")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  qbittorrent-file-matcher match --url http://localhost:8080 --hash abc123 --path /downloads")
}
