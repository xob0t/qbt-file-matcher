package backend

import (
	"embed"
	"encoding/json"
)

type versionInfoJSON struct {
	Fixed struct {
		FileVersion string `json:"file_version"`
	} `json:"fixed"`
}

// GetVersion returns version from embedded info.json
func GetVersion(content embed.FS) string {
	data, err := content.ReadFile("build/windows/info.json")
	if err != nil {
		return "unknown"
	}
	var info versionInfoJSON
	if err := json.Unmarshal(data, &info); err != nil {
		return "unknown"
	}
	return info.Fixed.FileVersion
}
