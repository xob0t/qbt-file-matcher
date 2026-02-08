package matcher

import (
	"testing"
)

func TestFindMatches_SingleExactMatch(t *testing.T) {
	torrentFiles := []TorrentFileInfo{
		{Index: 0, Name: "movie.mkv", Size: 1000000},
	}
	diskFiles := []DiskFile{
		{Path: "/data/movie.mkv", Name: "movie.mkv", Size: 1000000},
	}

	result := FindMatches(torrentFiles, diskFiles, false)

	if result.MatchedCount != 1 {
		t.Errorf("Expected 1 match, got %d", result.MatchedCount)
	}
	if len(result.Matches) != 1 {
		t.Errorf("Expected 1 match entry, got %d", len(result.Matches))
	}
	if result.Matches[0].Selected == nil {
		t.Error("Expected auto-match to select the file")
	}
	if !result.Matches[0].AutoMatched {
		t.Error("Expected AutoMatched to be true")
	}
}

func TestFindMatches_MultipleMatchCandidates(t *testing.T) {
	torrentFiles := []TorrentFileInfo{
		{Index: 0, Name: "video.mp4", Size: 500000},
	}
	diskFiles := []DiskFile{
		{Path: "/data/video1.mp4", Name: "video1.mp4", Size: 500000},
		{Path: "/data/video2.mp4", Name: "video2.mp4", Size: 500000},
	}

	result := FindMatches(torrentFiles, diskFiles, false)

	if result.MatchedCount != 0 {
		t.Errorf("Expected 0 auto-matches (multiple candidates), got %d", result.MatchedCount)
	}
	if len(result.Matches) != 1 {
		t.Errorf("Expected 1 match entry, got %d", len(result.Matches))
	}
	if len(result.Matches[0].DiskFiles) != 2 {
		t.Errorf("Expected 2 candidates, got %d", len(result.Matches[0].DiskFiles))
	}
	if result.Matches[0].Selected != nil {
		t.Error("Expected no auto-selection with multiple candidates")
	}
}

func TestFindMatches_NoMatch(t *testing.T) {
	torrentFiles := []TorrentFileInfo{
		{Index: 0, Name: "movie.mkv", Size: 1000000},
	}
	diskFiles := []DiskFile{
		{Path: "/data/other.mkv", Name: "other.mkv", Size: 2000000},
	}

	result := FindMatches(torrentFiles, diskFiles, false)

	if result.MatchedCount != 0 {
		t.Errorf("Expected 0 matches, got %d", result.MatchedCount)
	}
	if len(result.Unmatched) != 1 {
		t.Errorf("Expected 1 unmatched, got %d", len(result.Unmatched))
	}
}

func TestFindMatches_RequireSameExtension(t *testing.T) {
	torrentFiles := []TorrentFileInfo{
		{Index: 0, Name: "video.mkv", Size: 1000000},
	}
	diskFiles := []DiskFile{
		{Path: "/data/video.mp4", Name: "video.mp4", Size: 1000000},
		{Path: "/data/video.mkv", Name: "video.mkv", Size: 1000000},
	}

	// Without extension requirement - should find both
	result := FindMatches(torrentFiles, diskFiles, false)
	if len(result.Matches[0].DiskFiles) != 2 {
		t.Errorf("Expected 2 candidates without ext filter, got %d", len(result.Matches[0].DiskFiles))
	}

	// With extension requirement - should find only .mkv
	result = FindMatches(torrentFiles, diskFiles, true)
	if len(result.Matches[0].DiskFiles) != 1 {
		t.Errorf("Expected 1 candidate with ext filter, got %d", len(result.Matches[0].DiskFiles))
	}
	if result.Matches[0].DiskFiles[0].Name != "video.mkv" {
		t.Errorf("Expected video.mkv, got %s", result.Matches[0].DiskFiles[0].Name)
	}
}

func TestFindMatches_ExactNameMatch(t *testing.T) {
	torrentFiles := []TorrentFileInfo{
		{Index: 0, Name: "movie.mkv", Size: 1000000},
	}
	diskFiles := []DiskFile{
		{Path: "/data/other.mkv", Name: "other.mkv", Size: 1000000},
		{Path: "/data/movie.mkv", Name: "movie.mkv", Size: 1000000},
	}

	result := FindMatches(torrentFiles, diskFiles, false)

	// Should auto-match to exact name even with multiple size matches
	if result.MatchedCount != 1 {
		t.Errorf("Expected 1 auto-match via exact name, got %d", result.MatchedCount)
	}
	if result.Matches[0].Selected == nil {
		t.Error("Expected auto-selection via exact name match")
	}
	if result.Matches[0].Selected.Name != "movie.mkv" {
		t.Errorf("Expected movie.mkv, got %s", result.Matches[0].Selected.Name)
	}
}

func TestGenerateRenames(t *testing.T) {
	matches := []Match{
		{
			TorrentFile: TorrentFileInfo{Index: 0, Name: "old_name.mkv", Size: 1000000},
			Selected:    &DiskFile{Path: "/data/new_name.mkv", Name: "new_name.mkv", Size: 1000000},
		},
	}

	renames := GenerateRenames(matches, "/downloads")

	if len(renames) != 1 {
		t.Errorf("Expected 1 rename, got %d", len(renames))
	}
	if renames[0].OldPath != "old_name.mkv" {
		t.Errorf("Expected old path 'old_name.mkv', got '%s'", renames[0].OldPath)
	}
	if renames[0].NewPath != "new_name.mkv" {
		t.Errorf("Expected new path 'new_name.mkv', got '%s'", renames[0].NewPath)
	}
}

func TestGenerateRenames_NoChange(t *testing.T) {
	matches := []Match{
		{
			TorrentFile: TorrentFileInfo{Index: 0, Name: "same.mkv", Size: 1000000},
			Selected:    &DiskFile{Path: "/data/same.mkv", Name: "same.mkv", Size: 1000000},
		},
	}

	renames := GenerateRenames(matches, "/downloads")

	if len(renames) != 0 {
		t.Errorf("Expected 0 renames (same name), got %d", len(renames))
	}
}

func TestGenerateRenames_WithSubdirectory(t *testing.T) {
	matches := []Match{
		{
			TorrentFile: TorrentFileInfo{Index: 0, Name: "Season 1/episode01.mkv", Size: 1000000},
			Selected:    &DiskFile{Path: "/data/S01E01.mkv", Name: "S01E01.mkv", Size: 1000000},
		},
	}

	renames := GenerateRenames(matches, "/downloads")

	if len(renames) != 1 {
		t.Errorf("Expected 1 rename, got %d", len(renames))
	}
	if renames[0].OldPath != "Season 1/episode01.mkv" {
		t.Errorf("Expected old path 'Season 1/episode01.mkv', got '%s'", renames[0].OldPath)
	}
	// New path should keep directory structure
	if renames[0].NewPath != "Season 1/S01E01.mkv" {
		t.Errorf("Expected new path 'Season 1/S01E01.mkv', got '%s'", renames[0].NewPath)
	}
}

func TestGroupFilesBySize(t *testing.T) {
	files := []DiskFile{
		{Path: "/a.mkv", Name: "a.mkv", Size: 100},
		{Path: "/b.mkv", Name: "b.mkv", Size: 200},
		{Path: "/c.mkv", Name: "c.mkv", Size: 100},
	}

	grouped := GroupFilesBySize(files)

	if len(grouped[100]) != 2 {
		t.Errorf("Expected 2 files with size 100, got %d", len(grouped[100]))
	}
	if len(grouped[200]) != 1 {
		t.Errorf("Expected 1 file with size 200, got %d", len(grouped[200]))
	}
}
