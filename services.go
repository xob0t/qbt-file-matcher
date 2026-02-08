package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/autobrr/go-qbittorrent"

	"qbittorrent-file-matcher/internal/matcher"
)

// QBitService handles qBittorrent operations
type QBitService struct {
	client *qbittorrent.Client
}

// ConnectionConfig represents connection settings
type ConnectionConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Connect connects to qBittorrent
func (s *QBitService) Connect(config ConnectionConfig) error {
	s.client = qbittorrent.NewClient(qbittorrent.Config{
		Host:     config.URL,
		Username: config.Username,
		Password: config.Password,
	})

	if err := s.client.Login(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	return nil
}

// Disconnect disconnects from qBittorrent
func (s *QBitService) Disconnect() error {
	s.client = nil
	return nil
}

// IsConnected returns whether we're connected to qBittorrent
func (s *QBitService) IsConnected() bool {
	return s.client != nil
}

// GetVersion returns the qBittorrent version
func (s *QBitService) GetVersion() (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("not connected")
	}
	return s.client.GetAppVersion()
}

// TorrentInfo represents torrent information for the frontend
type TorrentInfo struct {
	Hash        string  `json:"hash"`
	Name        string  `json:"name"`
	Size        int64   `json:"size"`
	Progress    float64 `json:"progress"`
	State       string  `json:"state"`
	SavePath    string  `json:"savePath"`
	ContentPath string  `json:"contentPath"`
}

// GetTorrents returns all torrents
func (s *QBitService) GetTorrents() ([]TorrentInfo, error) {
	if s.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	torrents, err := s.client.GetTorrents(qbittorrent.TorrentFilterOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]TorrentInfo, len(torrents))
	for i, t := range torrents {
		result[i] = TorrentInfo{
			Hash:        t.Hash,
			Name:        t.Name,
			Size:        t.Size,
			Progress:    t.Progress,
			State:       string(t.State),
			SavePath:    t.SavePath,
			ContentPath: t.ContentPath,
		}
	}

	return result, nil
}

// TorrentFileInfo represents a file in a torrent for the frontend
type TorrentFileInfo struct {
	Index    int     `json:"index"`
	Name     string  `json:"name"`
	Size     int64   `json:"size"`
	Progress float64 `json:"progress"`
}

// GetTorrentFiles returns files for a specific torrent
func (s *QBitService) GetTorrentFiles(hash string) ([]TorrentFileInfo, error) {
	if s.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	files, err := s.client.GetFilesInformation(hash)
	if err != nil {
		return nil, err
	}

	result := make([]TorrentFileInfo, len(*files))
	for i, f := range *files {
		result[i] = TorrentFileInfo{
			Index:    f.Index,
			Name:     f.Name,
			Size:     f.Size,
			Progress: float64(f.Progress),
		}
	}

	return result, nil
}

// RenameFile renames a file in qBittorrent
func (s *QBitService) RenameFile(hash string, oldPath string, newPath string) error {
	if s.client == nil {
		return fmt.Errorf("not connected")
	}
	return s.client.RenameFile(hash, oldPath, newPath)
}

// SetTorrentLocation sets the download location for a torrent
func (s *QBitService) SetTorrentLocation(hash string, location string) error {
	if s.client == nil {
		return fmt.Errorf("not connected")
	}
	return s.client.SetLocation([]string{hash}, location)
}

// MatcherService handles file matching operations
type MatcherService struct{}

// DiskFileInfo represents a file on disk for the frontend
type DiskFileInfo struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// ScanDirectory scans a directory and returns all files
func (s *MatcherService) ScanDirectory(path string) ([]DiskFileInfo, error) {
	// Expand ~ to home directory
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, path[1:])
	}

	files, err := matcher.ScanDirectory(path)
	if err != nil {
		return nil, err
	}

	result := make([]DiskFileInfo, len(files))
	for i, f := range files {
		result[i] = DiskFileInfo{
			Path: f.Path,
			Name: f.Name,
			Size: f.Size,
		}
	}

	return result, nil
}

// MatchRequest represents a request to find matches
type MatchRequest struct {
	TorrentFiles         []matcher.TorrentFileInfo `json:"torrentFiles"`
	DiskFiles            []matcher.DiskFile        `json:"diskFiles"`
	RequireSameExtension bool                      `json:"requireSameExtension"`
}

// MatchResponse represents the match results
type MatchResponse struct {
	Matches      []MatchInfo               `json:"matches"`
	Unmatched    []matcher.TorrentFileInfo `json:"unmatched"`
	TotalFiles   int                       `json:"totalFiles"`
	MatchedCount int                       `json:"matchedCount"`
}

// MatchInfo represents a single match for the frontend
type MatchInfo struct {
	TorrentFile matcher.TorrentFileInfo `json:"torrentFile"`
	DiskFiles   []matcher.DiskFile      `json:"diskFiles"`
	Selected    *matcher.DiskFile       `json:"selected"`
	AutoMatched bool                    `json:"autoMatched"`
}

// FindMatches finds matches between torrent files and disk files
func (s *MatcherService) FindMatches(req MatchRequest) MatchResponse {
	result := matcher.FindMatches(req.TorrentFiles, req.DiskFiles, req.RequireSameExtension)

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
	Matches            []MatchInfo `json:"matches"`
	TorrentContentPath string      `json:"torrentContentPath"`
}

// RenameOperation represents a single rename operation
type RenameOperation struct {
	OldPath     string                  `json:"oldPath"`
	NewPath     string                  `json:"newPath"`
	TorrentFile matcher.TorrentFileInfo `json:"torrentFile"`
	DiskFile    matcher.DiskFile        `json:"diskFile"`
}

// GenerateRenames generates rename operations based on matches
func (s *MatcherService) GenerateRenames(req RenameRequest) []RenameOperation {
	// Convert MatchInfo to matcher.Match
	matches := make([]matcher.Match, len(req.Matches))
	for i, m := range req.Matches {
		matches[i] = matcher.Match{
			TorrentFile: m.TorrentFile,
			DiskFiles:   m.DiskFiles,
			Selected:    m.Selected,
			AutoMatched: m.AutoMatched,
		}
	}

	renames := matcher.GenerateRenames(matches, req.TorrentContentPath)

	result := make([]RenameOperation, len(renames))
	for i, r := range renames {
		result[i] = RenameOperation{
			OldPath:     r.OldPath,
			NewPath:     r.NewPath,
			TorrentFile: r.TorrentFile,
			DiskFile:    r.DiskFile,
		}
	}

	return result
}

// DirectoryExists checks if a directory exists
func (s *MatcherService) DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// DialogService handles native dialogs
type DialogService struct {
	ctx context.Context
}

// SetContext sets the application context for dialogs
func (s *DialogService) SetContext(ctx context.Context) {
	s.ctx = ctx
}
