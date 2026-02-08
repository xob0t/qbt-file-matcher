package backend

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_NoFile(t *testing.T) {
	config, err := LoadConfig()
	// It's OK if no config file exists - should return nil, nil
	if err != nil && !os.IsNotExist(err) {
		t.Logf("LoadConfig returned error (may be expected): %v", err)
	}
	t.Logf("LoadConfig returned config: %v", config)
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Test data
	testConfig := &Config{
		URL:      "http://test:8080",
		Username: "testuser",
		Password: "testpass",
	}

	t.Run("Config roundtrip", func(t *testing.T) {
		// Get config path
		configPath := GetConfigPath()

		// Backup existing config if present
		existingConfig, _ := os.ReadFile(configPath)
		hasExisting := existingConfig != nil

		defer func() {
			if hasExisting {
				_ = os.WriteFile(configPath, existingConfig, 0600)
			} else {
				os.Remove(configPath)
			}
		}()

		// Save config
		err := SaveConfig(testConfig)
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

func TestConfigCommands_Integration(t *testing.T) {
	// Get config path
	configPath := GetConfigPath()

	// Backup existing config
	existingConfig, _ := os.ReadFile(configPath)
	hasExisting := existingConfig != nil

	defer func() {
		if hasExisting {
			_ = os.WriteFile(configPath, existingConfig, 0600)
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
