package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"qbittorrent-file-matcher/backend"
)

// CLI config for match command
type matchConfig struct {
	url           string
	username      string
	password      string
	hash          string
	path          string
	sameExtension bool
	skipUnmatched bool
	dryRun        bool
	autoSelect    bool // Auto-select first match without prompting
	recheck       bool // Trigger recheck after applying renames
}

func runMatchCommand() {
	config := matchConfig{
		sameExtension: true, // default
	}

	// Load config file first (command line args will override)
	savedConfig, err := backend.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config file: %v\n", err)
	} else if savedConfig != nil {
		config.url = savedConfig.URL
		config.username = savedConfig.Username
		config.password = savedConfig.Password
	}

	saveConfig := false

	// Parse flags (override config file values)
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--url":
			if i+1 < len(args) {
				config.url = args[i+1]
				i++
			}
		case "--username", "-u":
			if i+1 < len(args) {
				config.username = args[i+1]
				i++
			}
		case "--password", "-p":
			if i+1 < len(args) {
				config.password = args[i+1]
				i++
			}
		case "--hash":
			if i+1 < len(args) {
				config.hash = args[i+1]
				i++
			}
		case "--path":
			if i+1 < len(args) {
				config.path = args[i+1]
				i++
			}
		case "--same-ext":
			config.sameExtension = true
		case "--no-same-ext":
			config.sameExtension = false
		case "--skip-unmatched":
			config.skipUnmatched = true
		case "--dry-run":
			config.dryRun = true
		case "--auto", "-a":
			config.autoSelect = true
		case "--recheck", "-r":
			config.recheck = true
		case "--save-config":
			saveConfig = true
		}
	}

	// Save config if requested
	if saveConfig && config.url != "" {
		err := backend.SaveConfig(&backend.Config{
			URL:      config.url,
			Username: config.username,
			Password: config.password,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save config: %v\n", err)
		} else {
			fmt.Printf("Config saved to %s\n", backend.GetConfigPath())
		}
	}

	// Validate required flags
	if config.url == "" {
		fmt.Fprintln(os.Stderr, "Error: --url is required (or set in config file)")
		os.Exit(1)
	}
	if config.hash == "" {
		fmt.Fprintln(os.Stderr, "Error: --hash is required")
		os.Exit(1)
	}
	if config.path == "" {
		fmt.Fprintln(os.Stderr, "Error: --path is required")
		os.Exit(1)
	}

	// Validate path exists
	if _, err := os.Stat(config.path); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: path does not exist: %s\n", config.path)
		os.Exit(1)
	}

	// Run the match
	if execErr := executeMatch(config); execErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", execErr)
		os.Exit(1)
	}
}

func executeMatch(config matchConfig) error {
	// Connect to qBittorrent
	fmt.Printf("Connecting to qBittorrent at %s...\n", config.url)

	qbitService := &backend.QBitService{}
	err := qbitService.Connect(backend.ConnectionConfig{
		URL:      config.url,
		Username: config.username,
		Password: config.password,
	})
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	fmt.Println("Connected!")

	// Get torrent files
	fmt.Printf("Getting files for torrent %s...\n", config.hash)
	torrentFiles, err := qbitService.GetTorrentFiles(config.hash)
	if err != nil {
		return fmt.Errorf("failed to get torrent files: %w", err)
	}
	fmt.Printf("Found %d files in torrent\n", len(torrentFiles))

	// Scan directory
	fmt.Printf("Scanning directory %s...\n", config.path)
	diskFiles, err := backend.ScanDirectory(config.path)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}
	fmt.Printf("Found %d files on disk\n", len(diskFiles))

	// Convert to backend types
	torrentFileInfos := make([]backend.TorrentFileInfo, len(torrentFiles))
	for i, f := range torrentFiles {
		torrentFileInfos[i] = backend.TorrentFileInfo{
			Index: f.Index,
			Name:  f.Name,
			Size:  f.Size,
		}
	}

	// Find matches
	fmt.Println("Finding matches...")
	matchResult := backend.FindMatches(torrentFileInfos, diskFiles, config.sameExtension)

	// Handle interactive selection for files with multiple candidates
	if !config.autoSelect && !config.dryRun {
		matchResult = handleInteractiveSelection(matchResult)
	}

	fmt.Printf("Matched: %d, Unmatched: %d\n", matchResult.MatchedCount, len(matchResult.Unmatched))

	// Generate renames
	renames := backend.GenerateRenames(matchResult.Matches, config.path)

	renamesApplied := false
	if len(renames) == 0 {
		fmt.Println("No renames needed - all files already have correct paths")
	} else {
		fmt.Printf("\nRenames to apply (%d):\n", len(renames))
		for _, r := range renames {
			fmt.Printf("  %s\n    -> %s\n", r.OldPath, r.NewPath)
		}

		if config.dryRun {
			fmt.Println("\n[DRY RUN] No changes made")
		} else {
			fmt.Println("\nApplying renames...")
			successCount := 0
			errorCount := 0

			for _, r := range renames {
				err := qbitService.RenameFile(config.hash, r.OldPath, r.NewPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  Failed to rename %s: %v\n", r.OldPath, err)
					errorCount++
				} else {
					successCount++
				}
			}

			fmt.Printf("Renamed %d files successfully", successCount)
			if errorCount > 0 {
				fmt.Printf(", %d failed", errorCount)
			}
			fmt.Println()

			if successCount > 0 {
				renamesApplied = true
			}
		}
	}

	// Handle unmatched files
	if len(matchResult.Unmatched) > 0 && config.skipUnmatched {
		fmt.Printf("\nSkipping %d unmatched files (setting priority to 0)...\n", len(matchResult.Unmatched))

		if config.dryRun {
			fmt.Println("[DRY RUN] Would skip:")
			for _, f := range matchResult.Unmatched {
				fmt.Printf("  %s\n", f.Name)
			}
		} else {
			// Build comma-separated list of file indices
			indices := make([]string, len(matchResult.Unmatched))
			for i, f := range matchResult.Unmatched {
				indices[i] = fmt.Sprintf("%d", f.Index)
			}
			fileIDs := strings.Join(indices, ",")

			err := qbitService.SetFilePriority(config.hash, fileIDs, 0)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to set priority: %v\n", err)
			} else {
				fmt.Printf("Set priority to 0 for %d files\n", len(matchResult.Unmatched))
			}
		}
	}

	// Print unmatched files
	if len(matchResult.Unmatched) > 0 && !config.skipUnmatched {
		fmt.Printf("\nUnmatched files (%d):\n", len(matchResult.Unmatched))
		for _, f := range matchResult.Unmatched {
			fmt.Printf("  %s (%s)\n", f.Name, formatSize(f.Size))
		}
		fmt.Println("\nUse --skip-unmatched to set priority to 0 for these files")
	}

	// Trigger recheck if requested and changes were made
	if config.recheck && renamesApplied && !config.dryRun {
		fmt.Println("\nTriggering torrent recheck...")
		err := qbitService.RecheckTorrent(config.hash)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to trigger recheck: %v\n", err)
		} else {
			fmt.Println("Recheck started - qBittorrent will verify file integrity")
		}
	}

	return nil
}

// handleInteractiveSelection prompts user to select files when there are multiple candidates
func handleInteractiveSelection(matchResult backend.MatchResult) backend.MatchResult {
	reader := bufio.NewReader(os.Stdin)

	for i := range matchResult.Matches {
		match := &matchResult.Matches[i]

		// Skip if already matched or no candidates
		if match.Selected != nil || len(match.DiskFiles) <= 1 {
			continue
		}

		fmt.Printf("\nMultiple matches found for: %s (%s)\n", match.TorrentFile.Name, formatSize(match.TorrentFile.Size))
		fmt.Println("Select a file:")

		for j, df := range match.DiskFiles {
			fmt.Printf("  [%d] %s\n", j+1, df.Path)
		}
		fmt.Printf("  [0] Skip this file\n")
		fmt.Print("Enter choice: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 0 || choice > len(match.DiskFiles) {
			fmt.Println("Invalid choice, skipping...")
			continue
		}

		if choice == 0 {
			// User chose to skip
			continue
		}

		// Select the chosen file
		match.Selected = &match.DiskFiles[choice-1]
		matchResult.MatchedCount++
	}

	return matchResult
}

func formatSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	const k = 1024
	sizes := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	size := float64(bytes)
	for size >= k && i < len(sizes)-1 {
		size /= k
		i++
	}
	return fmt.Sprintf("%.1f %s", size, sizes[i])
}

// Ensure path uses forward slashes (for consistency)
func normalizePath(path string) string {
	return filepath.ToSlash(path)
}
