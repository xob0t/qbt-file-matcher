package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"qbittorrent-file-matcher/internal/matcher"
)

func loadEnv(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		t.Skip("Skipping integration test: .env file not found")
	}
}

func getQBTConfig(t *testing.T) ConnectionConfig {
	loadEnv(t)

	url := os.Getenv("QBT_URL")
	username := os.Getenv("QBT_USERNAME")
	password := os.Getenv("QBT_PASSWORD")

	if url == "" || username == "" {
		t.Skip("Skipping integration test: QBT_URL or QBT_USERNAME not set")
	}

	return ConnectionConfig{
		URL:      url,
		Username: username,
		Password: password,
	}
}

func TestQBitService_Connect(t *testing.T) {
	config := getQBTConfig(t)

	service := &QBitService{}
	err := service.Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if !service.IsConnected() {
		t.Error("Expected IsConnected to return true after successful connect")
	}

	// Test GetVersion
	version, err := service.GetVersion()
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}
	if version == "" {
		t.Error("Expected non-empty version string")
	}
	t.Logf("Connected to qBittorrent %s", version)
}

func TestQBitService_GetTorrents(t *testing.T) {
	config := getQBTConfig(t)

	service := &QBitService{}
	err := service.Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	torrents, err := service.GetTorrents()
	if err != nil {
		t.Fatalf("Failed to get torrents: %v", err)
	}

	t.Logf("Found %d torrents", len(torrents))

	for i, torrent := range torrents {
		if i >= 3 { // Only log first 3
			break
		}
		t.Logf("  - %s (%.1f%%)", torrent.Name, torrent.Progress*100)
	}
}

func TestQBitService_GetTorrentFiles(t *testing.T) {
	config := getQBTConfig(t)

	service := &QBitService{}
	err := service.Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	torrents, err := service.GetTorrents()
	if err != nil {
		t.Fatalf("Failed to get torrents: %v", err)
	}

	if len(torrents) == 0 {
		t.Skip("No torrents available for testing")
	}

	// Get files for the first torrent
	files, err := service.GetTorrentFiles(torrents[0].Hash)
	if err != nil {
		t.Fatalf("Failed to get torrent files: %v", err)
	}

	t.Logf("Torrent '%s' has %d files", torrents[0].Name, len(files))

	for i, file := range files {
		if i >= 5 { // Only log first 5
			break
		}
		t.Logf("  - %s (%d bytes)", file.Name, file.Size)
	}
}

func TestQBitService_NotConnected(t *testing.T) {
	service := &QBitService{}

	if service.IsConnected() {
		t.Error("Expected IsConnected to return false before connect")
	}

	_, err := service.GetVersion()
	if err == nil {
		t.Error("Expected error when calling GetVersion without connection")
	}

	_, err = service.GetTorrents()
	if err == nil {
		t.Error("Expected error when calling GetTorrents without connection")
	}
}

func TestMatcherService_DirectoryExists(t *testing.T) {
	service := &MatcherService{}

	// Test existing directory
	if !service.DirectoryExists(".") {
		t.Error("Expected current directory to exist")
	}

	// Test non-existing directory
	if service.DirectoryExists("/nonexistent/path/12345") {
		t.Error("Expected non-existent path to return false")
	}
}

func TestMatcherService_ScanDirectory(t *testing.T) {
	service := &MatcherService{}

	// Scan current directory
	files, err := service.ScanDirectory(".")
	if err != nil {
		t.Fatalf("Failed to scan directory: %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected at least some files in current directory")
	}

	t.Logf("Found %d files in current directory", len(files))
}

func TestMatcherService_FindMatches(t *testing.T) {
	service := &MatcherService{}

	req := MatchRequest{
		TorrentFiles: []matcher.TorrentFileInfo{
			{Index: 0, Name: "test.txt", Size: 100},
			{Index: 1, Name: "other.txt", Size: 200},
		},
		DiskFiles: []matcher.DiskFile{
			{Path: "/path/test.txt", Name: "test.txt", Size: 100},
		},
		RequireSameExtension: false,
	}

	result := service.FindMatches(req)

	if result.TotalFiles != 2 {
		t.Errorf("Expected TotalFiles=2, got %d", result.TotalFiles)
	}
	if result.MatchedCount != 1 {
		t.Errorf("Expected MatchedCount=1, got %d", result.MatchedCount)
	}
	if len(result.Matches) != 1 {
		t.Errorf("Expected 1 match, got %d", len(result.Matches))
	}
	if len(result.Unmatched) != 1 {
		t.Errorf("Expected 1 unmatched, got %d", len(result.Unmatched))
	}
}

// TestSpecificTorrent_Debug tests the specific torrent to debug the rename issue
func TestSpecificTorrent_Debug(t *testing.T) {
	config := getQBTConfig(t)

	service := &QBitService{}
	err := service.Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Get the specific torrent
	hash := "cd510ce6662fb4fdcfa2b200118c379fff4c8622"
	files, err := service.GetTorrentFiles(hash)
	if err != nil {
		t.Fatalf("Failed to get torrent files: %v", err)
	}

	// Get torrent info
	torrents, err := service.GetTorrents()
	if err != nil {
		t.Fatalf("Failed to get torrents: %v", err)
	}

	var torrent *TorrentInfo
	for i := range torrents {
		if torrents[i].Hash == hash {
			torrent = &torrents[i]
			break
		}
	}

	if torrent == nil {
		t.Fatalf("Torrent %s not found", hash)
	}

	t.Logf("Torrent: %s", torrent.Name)
	t.Logf("SavePath: %s", torrent.SavePath)
	t.Logf("ContentPath: %s", torrent.ContentPath)
	t.Logf("")
	t.Logf("Files in torrent:")
	for _, f := range files {
		t.Logf("  [%d] %s (%d bytes)", f.Index, f.Name, f.Size)
	}
}

// TestSpecificTorrent_SimulateMatch simulates the matching and rename process
func TestSpecificTorrent_SimulateMatch(t *testing.T) {
	matcherService := &MatcherService{}

	// Simulate the scenario:
	// - Torrent file: "Le Bureau des Légendes S03E01.mkv" (no directory)
	// - Disk file: in subdirectory "Le Bureau des Légendes (The Bureau) - season 33"
	// - Search path: "C:\Users\admin\Downloads"

	searchPath := `C:\Users\admin\Downloads`

	// Scan the directory
	diskFiles, err := matcherService.ScanDirectory(searchPath)
	if err != nil {
		t.Fatalf("Failed to scan: %v", err)
	}

	// Find the S03E01 file
	var targetDiskFile *DiskFileInfo
	for i := range diskFiles {
		if diskFiles[i].Name == "Le Bureau des Légendes S03E01.mkv" {
			targetDiskFile = &diskFiles[i]
			break
		}
	}

	if targetDiskFile == nil {
		t.Skip("Target disk file not found")
	}

	t.Logf("Found disk file: %s", targetDiskFile.Path)

	// Create the match
	torrentFiles := []matcher.TorrentFileInfo{
		{Index: 0, Name: "Le Bureau des Légendes S03E01.mkv", Size: targetDiskFile.Size},
	}
	matchDiskFiles := []matcher.DiskFile{
		{Path: targetDiskFile.Path, Name: targetDiskFile.Name, Size: targetDiskFile.Size},
	}

	// Find matches
	matchResult := matcherService.FindMatches(MatchRequest{
		TorrentFiles:         torrentFiles,
		DiskFiles:            matchDiskFiles,
		RequireSameExtension: true,
	})

	t.Logf("Match result: %d matches, %d unmatched", matchResult.MatchedCount, len(matchResult.Unmatched))

	if len(matchResult.Matches) == 0 {
		t.Fatal("No matches found")
	}

	// Generate renames
	renames := matcherService.GenerateRenames(RenameRequest{
		Matches:    matchResult.Matches,
		SearchPath: searchPath,
	})

	t.Logf("")
	t.Logf("Generated renames:")
	for _, r := range renames {
		t.Logf("  Old: %s", r.OldPath)
		t.Logf("  New: %s", r.NewPath)
		t.Logf("")
	}

	// Verify the rename is correct
	if len(renames) == 0 {
		t.Error("Expected at least one rename operation")
	} else {
		// The new path should include the subdirectory
		expectedNewPath := "Le Bureau des Légendes (The Bureau) - season 33/Le Bureau des Légendes S03E01.mkv"
		if renames[0].NewPath != expectedNewPath {
			t.Errorf("Expected new path '%s', got '%s'", expectedNewPath, renames[0].NewPath)
		}
	}
}

// TestSpecificTorrent_ActualRename actually performs the rename to test it
func TestSpecificTorrent_ActualRename(t *testing.T) {
	config := getQBTConfig(t)

	qbitService := &QBitService{}
	err := qbitService.Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	matcherService := &MatcherService{}

	hash := "cd510ce6662fb4fdcfa2b200118c379fff4c8622"
	searchPath := `C:\Users\admin\Downloads`

	// Get current torrent files
	torrentFiles, err := qbitService.GetTorrentFiles(hash)
	if err != nil {
		t.Fatalf("Failed to get torrent files: %v", err)
	}

	t.Logf("Current torrent files:")
	for _, f := range torrentFiles {
		t.Logf("  [%d] %s", f.Index, f.Name)
	}

	// Scan disk
	diskFiles, err := matcherService.ScanDirectory(searchPath)
	if err != nil {
		t.Fatalf("Failed to scan: %v", err)
	}

	// Find a specific file to test with
	var targetTorrentFile *TorrentFileInfo
	var targetDiskFile *DiskFileInfo

	for i := range torrentFiles {
		if torrentFiles[i].Name == "Le Bureau des Légendes S03E01.mkv" {
			targetTorrentFile = &torrentFiles[i]
			break
		}
	}

	for i := range diskFiles {
		if diskFiles[i].Name == "Le Bureau des Légendes S03E01.mkv" {
			targetDiskFile = &diskFiles[i]
			break
		}
	}

	if targetTorrentFile == nil || targetDiskFile == nil {
		t.Skip("Target files not found")
	}

	t.Logf("")
	t.Logf("Torrent file: %s", targetTorrentFile.Name)
	t.Logf("Disk file: %s", targetDiskFile.Path)

	// Compute the new path
	relPath, err := filepath.Rel(searchPath, targetDiskFile.Path)
	if err != nil {
		t.Fatalf("Failed to compute relative path: %v", err)
	}
	newPath := filepath.ToSlash(relPath)

	t.Logf("")
	t.Logf("Rename operation:")
	t.Logf("  Old: %s", targetTorrentFile.Name)
	t.Logf("  New: %s", newPath)

	// Actually perform the rename
	t.Logf("")
	t.Logf("Attempting rename...")

	err = qbitService.RenameFile(hash, targetTorrentFile.Name, newPath)
	if err != nil {
		t.Fatalf("Failed to rename: %v", err)
	}
	t.Logf("Rename returned success!")

	// Wait a bit for qBittorrent to process
	time.Sleep(500 * time.Millisecond)

	// Verify by getting the files again
	torrentFilesAfter, err := qbitService.GetTorrentFiles(hash)
	if err != nil {
		t.Fatalf("Failed to get torrent files after rename: %v", err)
	}

	t.Logf("")
	t.Logf("Torrent files after rename:")
	for _, f := range torrentFilesAfter {
		t.Logf("  [%d] %s", f.Index, f.Name)
	}

	// Check if the rename actually happened
	foundNewPath := false
	for _, f := range torrentFilesAfter {
		if f.Name == newPath {
			foundNewPath = true
			break
		}
	}

	if !foundNewPath {
		t.Errorf("Rename did not take effect! Expected to find file with path '%s'", newPath)
	}
}
