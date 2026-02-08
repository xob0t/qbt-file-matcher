package backend

import (
	"testing"
)

func TestFindMatches_SingleExactMatch(t *testing.T) {
	torrentFiles := []TorrentFileInfo{
		{Index: 0, Name: "video.mkv", Size: 1000},
	}
	diskFiles := []DiskFile{
		{Path: "/path/video.mkv", Name: "video.mkv", Size: 1000},
	}

	result := FindMatches(torrentFiles, diskFiles, true)

	if result.MatchedCount != 1 {
		t.Errorf("Expected 1 match, got %d", result.MatchedCount)
	}
	if len(result.Unmatched) != 0 {
		t.Errorf("Expected 0 unmatched, got %d", len(result.Unmatched))
	}
	if result.Matches[0].Selected == nil {
		t.Error("Expected auto-selected match")
	}
}

func TestFindMatches_MultipleMatchCandidates(t *testing.T) {
	torrentFiles := []TorrentFileInfo{
		{Index: 0, Name: "video.mkv", Size: 1000},
	}
	diskFiles := []DiskFile{
		{Path: "/path1/video.mkv", Name: "video.mkv", Size: 1000},
		{Path: "/path2/copy.mkv", Name: "copy.mkv", Size: 1000},
	}

	result := FindMatches(torrentFiles, diskFiles, true)

	if len(result.Matches) != 1 {
		t.Fatalf("Expected 1 match entry, got %d", len(result.Matches))
	}
	if len(result.Matches[0].DiskFiles) != 2 {
		t.Errorf("Expected 2 candidates, got %d", len(result.Matches[0].DiskFiles))
	}
	// Should auto-match the one with exact name
	if result.Matches[0].Selected == nil {
		t.Error("Expected auto-selected match based on name")
	}
}

func TestFindMatches_NoMatch(t *testing.T) {
	torrentFiles := []TorrentFileInfo{
		{Index: 0, Name: "video.mkv", Size: 1000},
	}
	diskFiles := []DiskFile{
		{Path: "/path/other.mkv", Name: "other.mkv", Size: 2000},
	}

	result := FindMatches(torrentFiles, diskFiles, true)

	if result.MatchedCount != 0 {
		t.Errorf("Expected 0 matches, got %d", result.MatchedCount)
	}
	if len(result.Unmatched) != 1 {
		t.Errorf("Expected 1 unmatched, got %d", len(result.Unmatched))
	}
}

func TestFindMatches_RequireSameExtension(t *testing.T) {
	torrentFiles := []TorrentFileInfo{
		{Index: 0, Name: "video.mkv", Size: 1000},
	}
	diskFiles := []DiskFile{
		{Path: "/path/video.avi", Name: "video.avi", Size: 1000},
	}

	// With same extension required - no match
	result := FindMatches(torrentFiles, diskFiles, true)
	if result.MatchedCount != 0 {
		t.Errorf("Expected 0 matches with same extension required, got %d", result.MatchedCount)
	}

	// Without same extension required - match
	result = FindMatches(torrentFiles, diskFiles, false)
	if result.MatchedCount != 1 {
		t.Errorf("Expected 1 match without extension requirement, got %d", result.MatchedCount)
	}
}

func TestFindMatches_ExactNameMatch(t *testing.T) {
	torrentFiles := []TorrentFileInfo{
		{Index: 0, Name: "video.mkv", Size: 1000},
	}
	diskFiles := []DiskFile{
		{Path: "/path1/other.mkv", Name: "other.mkv", Size: 1000},
		{Path: "/path2/video.mkv", Name: "video.mkv", Size: 1000},
	}

	result := FindMatches(torrentFiles, diskFiles, true)

	if result.Matches[0].Selected == nil {
		t.Fatal("Expected auto-selected match")
	}
	if result.Matches[0].Selected.Name != "video.mkv" {
		t.Errorf("Expected 'video.mkv' to be selected, got '%s'", result.Matches[0].Selected.Name)
	}
}

func TestGenerateRenames(t *testing.T) {
	matches := []Match{
		{
			TorrentFile: TorrentFileInfo{Index: 0, Name: "video.mkv", Size: 1000},
			Selected:    &DiskFile{Path: "/downloads/subdir/video.mkv", Name: "video.mkv", Size: 1000},
		},
	}

	renames := GenerateRenames(matches, "/downloads")

	if len(renames) != 1 {
		t.Fatalf("Expected 1 rename, got %d", len(renames))
	}
	if renames[0].OldPath != "video.mkv" {
		t.Errorf("Expected old path 'video.mkv', got '%s'", renames[0].OldPath)
	}
	if renames[0].NewPath != "subdir/video.mkv" {
		t.Errorf("Expected new path 'subdir/video.mkv', got '%s'", renames[0].NewPath)
	}
}

func TestGenerateRenames_NoChange(t *testing.T) {
	matches := []Match{
		{
			TorrentFile: TorrentFileInfo{Index: 0, Name: "video.mkv", Size: 1000},
			Selected:    &DiskFile{Path: "/downloads/video.mkv", Name: "video.mkv", Size: 1000},
		},
	}

	renames := GenerateRenames(matches, "/downloads")

	if len(renames) != 0 {
		t.Errorf("Expected 0 renames when paths match, got %d", len(renames))
	}
}

func TestGenerateRenames_WithSubdirectory(t *testing.T) {
	matches := []Match{
		{
			TorrentFile: TorrentFileInfo{Index: 0, Name: "Series/S01E01.mkv", Size: 1000},
			Selected:    &DiskFile{Path: "/downloads/Series/S01E01.mkv", Name: "S01E01.mkv", Size: 1000},
		},
	}

	renames := GenerateRenames(matches, "/downloads")

	if len(renames) != 0 {
		t.Errorf("Expected 0 renames when paths already match, got %d", len(renames))
	}
}

func TestGenerateRenames_DiskFileInSubdir(t *testing.T) {
	matches := []Match{
		{
			TorrentFile: TorrentFileInfo{Index: 0, Name: "S01E01.mkv", Size: 1000},
			Selected:    &DiskFile{Path: "/downloads/Series/S01E01.mkv", Name: "S01E01.mkv", Size: 1000},
		},
	}

	renames := GenerateRenames(matches, "/downloads")

	if len(renames) != 1 {
		t.Fatalf("Expected 1 rename, got %d", len(renames))
	}
	if renames[0].NewPath != "Series/S01E01.mkv" {
		t.Errorf("Expected new path 'Series/S01E01.mkv', got '%s'", renames[0].NewPath)
	}
}

func TestGroupFilesBySize(t *testing.T) {
	files := []DiskFile{
		{Path: "/a.txt", Name: "a.txt", Size: 100},
		{Path: "/b.txt", Name: "b.txt", Size: 100},
		{Path: "/c.txt", Name: "c.txt", Size: 200},
	}

	groups := GroupFilesBySize(files)

	if len(groups[100]) != 2 {
		t.Errorf("Expected 2 files of size 100, got %d", len(groups[100]))
	}
	if len(groups[200]) != 1 {
		t.Errorf("Expected 1 file of size 200, got %d", len(groups[200]))
	}
}

func TestMatcherService_DirExists(t *testing.T) {
	service := &MatcherService{}

	// Test existing directory
	if !service.DirExists(".") {
		t.Error("Expected current directory to exist")
	}

	// Test non-existing directory
	if service.DirExists("/nonexistent/path/12345") {
		t.Error("Expected non-existent path to return false")
	}
}

func TestMatcherService_ScanDir(t *testing.T) {
	service := &MatcherService{}

	// Scan current directory
	files, err := service.ScanDir(".")
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
		TorrentFiles: []TorrentFileInfo{
			{Index: 0, Name: "test.txt", Size: 100},
			{Index: 1, Name: "other.txt", Size: 200},
		},
		DiskFiles: []DiskFile{
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
