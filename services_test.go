package main

import (
	"os"
	"testing"

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
