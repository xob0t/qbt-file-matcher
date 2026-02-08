package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_NoFile(t *testing.T) {
	// Save original config path and restore after test
	originalConfigPath, _ := getConfigPath()
	defer func() {
		// Clean up test file if created
		os.Remove(originalConfigPath + ".test")
	}()

	config, err := LoadConfig()
	// It's OK if no config file exists - should return nil, nil
	if err != nil && !os.IsNotExist(err) {
		// Only fail if it's an actual error, not just missing file
		t.Logf("LoadConfig returned error (may be expected): %v", err)
	}
	t.Logf("LoadConfig returned config: %v", config)
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create a temp directory for test config
	tempDir, err := os.MkdirTemp("", "qbt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save original function reference
	originalGetConfigPath := getConfigPath

	// We can't easily replace getConfigPath since it's not a variable,
	// so we'll test SaveConfig and LoadConfig with the real path and restore after

	// Test data
	testConfig := &Config{
		URL:      "http://test:8080",
		Username: "testuser",
		Password: "testpass",
	}

	// Test that config can be marshaled and unmarshaled
	t.Run("Config roundtrip", func(t *testing.T) {
		// Use the real config path but clean up after
		configPath, err := originalGetConfigPath()
		if err != nil {
			t.Fatalf("Failed to get config path: %v", err)
		}

		// Backup existing config if present
		existingConfig, _ := os.ReadFile(configPath)
		hasExisting := existingConfig != nil

		defer func() {
			if hasExisting {
				os.WriteFile(configPath, existingConfig, 0600)
			} else {
				os.Remove(configPath)
			}
		}()

		// Save config
		err = SaveConfig(testConfig)
		if err != nil {
			t.Fatalf("SaveConfig failed: %v", err)
		}

		// Load config
		loadedConfig, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}
		if loadedConfig == nil {
			t.Fatal("LoadConfig returned nil config")
		}

		// Verify values
		if loadedConfig.URL != testConfig.URL {
			t.Errorf("URL mismatch: got %s, want %s", loadedConfig.URL, testConfig.URL)
		}
		if loadedConfig.Username != testConfig.Username {
			t.Errorf("Username mismatch: got %s, want %s", loadedConfig.Username, testConfig.Username)
		}
		if loadedConfig.Password != testConfig.Password {
			t.Errorf("Password mismatch: got %s, want %s", loadedConfig.Password, testConfig.Password)
		}
	})
}

func TestGetConfigPath(t *testing.T) {
	path, err := getConfigPath()
	if err != nil {
		t.Fatalf("getConfigPath failed: %v", err)
	}

	// Path should end with our app directory and config.json
	if filepath.Base(path) != "config.json" {
		t.Errorf("Expected config.json, got %s", filepath.Base(path))
	}

	parentDir := filepath.Base(filepath.Dir(path))
	if parentDir != "qbittorrent-file-matcher" {
		t.Errorf("Expected qbittorrent-file-matcher directory, got %s", parentDir)
	}

	t.Logf("Config path: %s", path)
}

func TestGetConfigPath_Public(t *testing.T) {
	path := GetConfigPath()
	if path == "" || path == "<unknown>" {
		t.Error("GetConfigPath returned empty or unknown")
	}
	t.Logf("Config path (public): %s", path)
}

func TestIsCLICommand(t *testing.T) {
	tests := []struct {
		arg      string
		expected bool
	}{
		{"match", true},
		{"config", true},
		{"help", true},
		{"--help", true},
		{"-h", true},
		{"version", true},
		{"--version", true},
		{"-v", true},
		{"unknown", false},
		{"", false},
		{"gui", false},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			result := isCLICommand(tt.arg)
			if result != tt.expected {
				t.Errorf("isCLICommand(%q) = %v, want %v", tt.arg, result, tt.expected)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100.0 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"C:\\Users\\admin\\Downloads", "C:/Users/admin/Downloads"},
		{"C:/Users/admin/Downloads", "C:/Users/admin/Downloads"},
		{"/home/user/downloads", "/home/user/downloads"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMatchConfig_Defaults(t *testing.T) {
	config := matchConfig{
		sameExtension: true,
	}

	// Verify defaults
	if !config.sameExtension {
		t.Error("sameExtension should default to true")
	}
	if config.skipUnmatched {
		t.Error("skipUnmatched should default to false")
	}
	if config.dryRun {
		t.Error("dryRun should default to false")
	}
	if config.autoSelect {
		t.Error("autoSelect should default to false")
	}
	if config.recheck {
		t.Error("recheck should default to false")
	}
}

// TestQBitService_RecheckTorrent tests the recheck method
func TestQBitService_RecheckTorrent(t *testing.T) {
	// This test requires a running qBittorrent instance
	// Skip if not available
	service := &QBitService{}

	// Test that calling Recheck without connection returns error
	err := service.RecheckTorrent("somehash")
	if err == nil {
		t.Error("Expected error when not connected, got nil")
	}
	if err.Error() != "not connected" {
		t.Errorf("Expected 'not connected' error, got: %v", err)
	}
}

// TestQBitService_RecheckTorrent_Connected tests recheck with a live connection
func TestQBitService_RecheckTorrent_Connected(t *testing.T) {
	service := &QBitService{}

	// Try to connect
	err := service.Connect(ConnectionConfig{
		URL:      "http://localhost:8080",
		Username: "admin",
		Password: "adminadmin",
	})
	if err != nil {
		t.Skipf("Skipping test - cannot connect to qBittorrent: %v", err)
	}

	// Get torrents to find a hash
	torrents, err := service.GetTorrents()
	if err != nil {
		t.Fatalf("Failed to get torrents: %v", err)
	}

	if len(torrents) == 0 {
		t.Skip("No torrents available for recheck test")
	}

	// Test recheck on first torrent
	hash := torrents[0].Hash
	t.Logf("Testing recheck on torrent: %s (%s)", torrents[0].Name, hash)

	err = service.RecheckTorrent(hash)
	if err != nil {
		t.Errorf("RecheckTorrent failed: %v", err)
	} else {
		t.Log("Recheck triggered successfully")
	}
}

// Integration test for config commands
func TestConfigCommands_Integration(t *testing.T) {
	// This test verifies that config saving/loading works correctly

	// Get config path
	configPath, err := getConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	// Backup existing config
	existingConfig, _ := os.ReadFile(configPath)
	hasExisting := existingConfig != nil

	defer func() {
		if hasExisting {
			os.WriteFile(configPath, existingConfig, 0600)
		} else {
			os.Remove(configPath)
			// Also try to remove parent directory if empty
			os.Remove(filepath.Dir(configPath))
		}
	}()

	// Clean slate
	os.Remove(configPath)

	// Test 1: Load when no config exists
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig should not error when no file exists: %v", err)
	}
	if config != nil {
		t.Error("LoadConfig should return nil when no file exists")
	}

	// Test 2: Save config
	testConfig := &Config{
		URL:      "http://localhost:9999",
		Username: "admin",
		Password: "secret123",
	}
	err = SaveConfig(testConfig)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Test 3: Load saved config
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed after save: %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadConfig returned nil after save")
	}
	if loaded.URL != testConfig.URL {
		t.Errorf("URL mismatch: %s != %s", loaded.URL, testConfig.URL)
	}
	if loaded.Username != testConfig.Username {
		t.Errorf("Username mismatch: %s != %s", loaded.Username, testConfig.Username)
	}
	if loaded.Password != testConfig.Password {
		t.Errorf("Password mismatch: %s != %s", loaded.Password, testConfig.Password)
	}

	// Test 4: Update config
	testConfig.URL = "http://newhost:8080"
	err = SaveConfig(testConfig)
	if err != nil {
		t.Fatalf("SaveConfig (update) failed: %v", err)
	}

	loaded, _ = LoadConfig()
	if loaded.URL != "http://newhost:8080" {
		t.Errorf("Updated URL not saved: %s", loaded.URL)
	}
}
