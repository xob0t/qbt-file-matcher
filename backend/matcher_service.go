package backend

import (
	"os"
	"path/filepath"
)

// MatcherService handles file matching operations
type MatcherService struct{}

// DiskFileInfo represents a file on disk for the frontend
type DiskFileInfo struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// ScanDir scans a directory and returns all files
func (s *MatcherService) ScanDir(path string) ([]DiskFileInfo, error) {
	// Expand ~ to home directory
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, path[1:])
	}

	files, err := ScanDirectory(path)
	if err != nil {
		return nil, err
	}

	result := make([]DiskFileInfo, len(files))
	for i, f := range files {
		result[i] = DiskFileInfo(f)
	}

	return result, nil
}

// MatchRequest represents a request to find matches
type MatchRequest struct {
	TorrentFiles         []TorrentFileInfo `json:"torrentFiles"`
	DiskFiles            []DiskFile        `json:"diskFiles"`
	RequireSameExtension bool              `json:"requireSameExtension"`
}

// MatchResponse represents the match results
type MatchResponse struct {
	Matches      []MatchInfo       `json:"matches"`
	Unmatched    []TorrentFileInfo `json:"unmatched"`
	TotalFiles   int               `json:"totalFiles"`
	MatchedCount int               `json:"matchedCount"`
}

// MatchInfo represents a single match for the frontend
type MatchInfo struct {
	TorrentFile TorrentFileInfo `json:"torrentFile"`
	DiskFiles   []DiskFile      `json:"diskFiles"`
	Selected    *DiskFile       `json:"selected"`
	AutoMatched bool            `json:"autoMatched"`
}

// FindMatches finds matches between torrent files and disk files
func (s *MatcherService) FindMatches(req MatchRequest) MatchResponse {
	result := FindMatches(req.TorrentFiles, req.DiskFiles, req.RequireSameExtension)

	matches := make([]MatchInfo, len(result.Matches))
	for i, m := range result.Matches {
		matches[i] = MatchInfo{
			TorrentFile: m.TorrentFile,
			DiskFiles:   m.DiskFiles,
			Selected:    m.Selected,
			AutoMatched: m.AutoMatched,
		}
	}

	return MatchResponse{
		Matches:      matches,
		Unmatched:    result.Unmatched,
		TotalFiles:   result.TotalFiles,
		MatchedCount: result.MatchedCount,
	}
}

// RenameRequest represents a rename operation request
type RenameRequest struct {
	Matches    []MatchInfo `json:"matches"`
	SearchPath string      `json:"searchPath"`
}

// RenameOp represents a single rename operation
type RenameOp struct {
	OldPath     string          `json:"oldPath"`
	NewPath     string          `json:"newPath"`
	TorrentFile TorrentFileInfo `json:"torrentFile"`
	DiskFile    DiskFile        `json:"diskFile"`
}

// GenRenames generates rename operations based on matches
func (s *MatcherService) GenRenames(req RenameRequest) []RenameOp {
	// Convert MatchInfo to Match
	matches := make([]Match, len(req.Matches))
	for i, m := range req.Matches {
		matches[i] = Match{
			TorrentFile: m.TorrentFile,
			DiskFiles:   m.DiskFiles,
			Selected:    m.Selected,
			AutoMatched: m.AutoMatched,
		}
	}

	renames := GenerateRenames(matches, req.SearchPath)

	result := make([]RenameOp, len(renames))
	for i, r := range renames {
		result[i] = RenameOp(r)
	}

	return result
}

// DirExists checks if a directory exists
func (s *MatcherService) DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
