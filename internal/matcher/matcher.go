package matcher

import (
	"path"
	"path/filepath"
	"strings"
)

// TorrentFileInfo represents a file from a torrent that needs matching
type TorrentFileInfo struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	Size  int64  `json:"size"`
}

// Match represents a potential match between a torrent file and a disk file
type Match struct {
	TorrentFile TorrentFileInfo `json:"torrentFile"`
	DiskFiles   []DiskFile      `json:"diskFiles"`
	Selected    *DiskFile       `json:"selected,omitempty"`
	AutoMatched bool            `json:"autoMatched"`
}

// MatchResult represents the result of matching a torrent with disk files
type MatchResult struct {
	Matches      []Match           `json:"matches"`
	Unmatched    []TorrentFileInfo `json:"unmatched"`
	TotalFiles   int               `json:"totalFiles"`
	MatchedCount int               `json:"matchedCount"`
}

// FindMatches finds potential matches between torrent files and disk files
func FindMatches(torrentFiles []TorrentFileInfo, diskFiles []DiskFile, requireSameExtension bool) MatchResult {
	result := MatchResult{
		Matches:    []Match{},
		Unmatched:  []TorrentFileInfo{},
		TotalFiles: len(torrentFiles),
	}

	// Group disk files by size for O(1) lookup
	sizeMap := GroupFilesBySize(diskFiles)

	for _, tf := range torrentFiles {
		candidates := sizeMap[tf.Size]

		if len(candidates) == 0 {
			result.Unmatched = append(result.Unmatched, tf)
			continue
		}

		// Filter by extension if required
		if requireSameExtension {
			tfExt := strings.ToLower(filepath.Ext(tf.Name))
			filtered := []DiskFile{}
			for _, c := range candidates {
				if strings.ToLower(filepath.Ext(c.Name)) == tfExt {
					filtered = append(filtered, c)
				}
			}
			candidates = filtered
		}

		if len(candidates) == 0 {
			result.Unmatched = append(result.Unmatched, tf)
			continue
		}

		match := Match{
			TorrentFile: tf,
			DiskFiles:   candidates,
			AutoMatched: false,
		}

		// Auto-match if there's exactly one candidate
		if len(candidates) == 1 {
			match.Selected = &candidates[0]
			match.AutoMatched = true
			result.MatchedCount++
		} else {
			// Try to find exact name match
			for i, c := range candidates {
				if strings.EqualFold(filepath.Base(c.Path), tf.Name) {
					match.Selected = &candidates[i]
					match.AutoMatched = true
					result.MatchedCount++
					break
				}
			}
		}

		result.Matches = append(result.Matches, match)
	}

	return result
}

// GenerateRenames generates the rename operations needed to match files
// Note: qBittorrent API uses forward slashes for paths on all platforms
func GenerateRenames(matches []Match, torrentContentPath string) []RenameOperation {
	var renames []RenameOperation

	for _, m := range matches {
		if m.Selected == nil {
			continue
		}

		// The new path should be relative to the torrent content path
		// and match the selected disk file's name
		oldPath := m.TorrentFile.Name

		// Keep the directory structure from the torrent, just change the filename
		// Use path (not filepath) to ensure forward slashes for qBittorrent API
		dir := path.Dir(oldPath)
		newName := filepath.Base(m.Selected.Path)

		var newPath string
		if dir == "." {
			newPath = newName
		} else {
			newPath = path.Join(dir, newName)
		}

		// Only add rename if paths are different
		if oldPath != newPath {
			renames = append(renames, RenameOperation{
				OldPath:     oldPath,
				NewPath:     newPath,
				TorrentFile: m.TorrentFile,
				DiskFile:    *m.Selected,
			})
		}
	}

	return renames
}

// RenameOperation represents a single rename operation
type RenameOperation struct {
	OldPath     string          `json:"oldPath"`
	NewPath     string          `json:"newPath"`
	TorrentFile TorrentFileInfo `json:"torrentFile"`
	DiskFile    DiskFile        `json:"diskFile"`
}
