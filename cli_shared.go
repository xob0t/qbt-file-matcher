package main

import (
	"fmt"
	"os"
	"slices"
)

const appVersion = "1.0.0"

func isCLICommand(arg string) bool {
	supportedCommands := []string{
		"match",
		"config",
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

	case "config":
		if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
			printConfigHelp()
			return
		}
		runConfigCommand()

	case "help", "--help", "-h":
		printCLIHelp()

	case "version", "--version", "-v":
		fmt.Printf("qbittorrent-file-matcher v%s\n", appVersion)

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
	fmt.Println("  config      Show or set configuration")
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
	fmt.Println("  --save-config            Save connection settings to config file")
	fmt.Println()
	fmt.Println("Config file:")
	fmt.Println("  Connection settings (URL, username, password) can be saved to a config file.")
	fmt.Println("  Use --save-config to save settings, or use 'config set' command.")
	fmt.Println("  Config file location:", GetConfigPath())
	fmt.Println()
	fmt.Println("Interactive mode:")
	fmt.Println("  When multiple files match the same size, you'll be prompted to select one.")
	fmt.Println("  Use --auto to skip prompts and auto-select the first match.")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  qbittorrent-file-matcher match --url http://localhost:8080 --hash abc123 --path /downloads")
}

func printConfigHelp() {
	fmt.Println("Usage: qbittorrent-file-matcher config [command]")
	fmt.Println()
	fmt.Println("Manage CLI configuration")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  show                     Show current configuration")
	fmt.Println("  set [flags]              Set configuration values")
	fmt.Println("  clear                    Remove configuration file")
	fmt.Println("  path                     Show configuration file path")
	fmt.Println()
	fmt.Println("Set flags:")
	fmt.Println("  --url <url>              qBittorrent WebUI URL")
	fmt.Println("  -u, --username <user>    qBittorrent username")
	fmt.Println("  -p, --password <pass>    qBittorrent password")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  qbittorrent-file-matcher config show")
	fmt.Println("  qbittorrent-file-matcher config set --url http://localhost:8080 -u admin -p secret")
	fmt.Println("  qbittorrent-file-matcher config clear")
}

func runConfigCommand() {
	if len(os.Args) < 3 {
		printConfigHelp()
		return
	}

	subcommand := os.Args[2]

	switch subcommand {
	case "show":
		config, err := LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		if config == nil {
			fmt.Println("No configuration file found")
			fmt.Printf("Config file path: %s\n", GetConfigPath())
			return
		}
		fmt.Println("Current configuration:")
		fmt.Printf("  URL:      %s\n", config.URL)
		fmt.Printf("  Username: %s\n", config.Username)
		if config.Password != "" {
			fmt.Printf("  Password: %s\n", "********")
		} else {
			fmt.Printf("  Password: (not set)\n")
		}
		fmt.Printf("\nConfig file: %s\n", GetConfigPath())

	case "set":
		config, _ := LoadConfig()
		if config == nil {
			config = &Config{}
		}

		// Parse flags
		args := os.Args[3:]
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--url":
				if i+1 < len(args) {
					config.URL = args[i+1]
					i++
				}
			case "--username", "-u":
				if i+1 < len(args) {
					config.Username = args[i+1]
					i++
				}
			case "--password", "-p":
				if i+1 < len(args) {
					config.Password = args[i+1]
					i++
				}
			}
		}

		if err := SaveConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Configuration saved to %s\n", GetConfigPath())

	case "clear":
		configPath := GetConfigPath()
		if err := os.Remove(configPath); err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No configuration file to remove")
				return
			}
			fmt.Fprintf(os.Stderr, "Error removing config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Configuration file removed")

	case "path":
		fmt.Println(GetConfigPath())

	default:
		fmt.Fprintf(os.Stderr, "Unknown config command: %s\n\n", subcommand)
		printConfigHelp()
		os.Exit(1)
	}
}
