package backend

import (
	"log"
	"os"
	"path/filepath"
)

// DiskFile represents a file on the disk
type DiskFile struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// ScanDirectory scans a directory recursively and returns all files with their sizes
func ScanDirectory(root string) ([]DiskFile, error) {
	var files []DiskFile
	var skippedCount int

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log and skip files/directories we can't access
			skippedCount++
			log.Printf("Skipping inaccessible path: %s (%v)", path, err)
			return nil
		}

		if !info.IsDir() {
			files = append(files, DiskFile{
				Path: path,
				Name: info.Name(),
				Size: info.Size(),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if skippedCount > 0 {
		log.Printf("Skipped %d inaccessible files/directories during scan", skippedCount)
	}

	return files, nil
}

// GroupFilesBySize groups files by their size for efficient matching
func GroupFilesBySize(files []DiskFile) map[int64][]DiskFile {
	sizeMap := make(map[int64][]DiskFile)
	for _, f := range files {
		sizeMap[f.Size] = append(sizeMap[f.Size], f)
	}
	return sizeMap
}
