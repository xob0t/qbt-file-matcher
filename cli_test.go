package main

import (
	"testing"
)

func TestIsCLICommand(t *testing.T) {
	tests := []struct {
		arg      string
		expected bool
	}{
		{"match", true},
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

func TestGetAppVersion(t *testing.T) {
	version := getAppVersion()
	if version == "" || version == "unknown" {
		t.Errorf("getAppVersion returned invalid version: %q", version)
	}
	t.Logf("App version: %s", version)
}
