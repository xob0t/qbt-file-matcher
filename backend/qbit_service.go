package backend

import (
	"fmt"

	"github.com/autobrr/go-qbittorrent"
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
type TorrentFile struct {
	Index    int     `json:"index"`
	Name     string  `json:"name"`
	Size     int64   `json:"size"`
	Progress float64 `json:"progress"`
}

// GetTorrentFiles returns files for a specific torrent
func (s *QBitService) GetTorrentFiles(hash string) ([]TorrentFile, error) {
	if s.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	files, err := s.client.GetFilesInformation(hash)
	if err != nil {
		return nil, err
	}

	// Handle nil response
	if files == nil {
		return []TorrentFile{}, nil
	}

	result := make([]TorrentFile, len(*files))
	for i, f := range *files {
		result[i] = TorrentFile{
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

// SetFilePriority sets the priority for files in a torrent
// IDs is a comma-separated list of file indices (e.g., "0,1,2")
// Priority: 0 = do not download, 1 = normal, 6 = high, 7 = maximum
func (s *QBitService) SetFilePriority(hash string, fileIDs string, priority int) error {
	if s.client == nil {
		return fmt.Errorf("not connected")
	}
	return s.client.SetFilePriority(hash, fileIDs, priority)
}

// RecheckTorrent triggers a hash recheck for the torrent
func (s *QBitService) RecheckTorrent(hash string) error {
	if s.client == nil {
		return fmt.Errorf("not connected")
	}
	return s.client.Recheck([]string{hash})
}
