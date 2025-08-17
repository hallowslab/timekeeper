package metadata

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// ExtractDate extracts date from various file types
func ExtractDate(filePath string) (*time.Time, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".jpg", ".jpeg", ".tiff":
		return extractImageDate(filePath)
	case ".mp4", ".m4v":
		return extractMP4Date(filePath)
	case ".mov":
		return extractMOVDate(filePath)
	case ".mkv", ".avi":
		return extractVideoDate(filePath)
	default:
		return extractFileSystemDate(filePath)
	}
}

func extractImageDate(filePath string) (*time.Time, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v", err)
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return extractFileSystemDate(filePath)
	}

	dateTime, err := x.DateTime()
	if err != nil {
		return extractFileSystemDate(filePath)
	}

	return &dateTime, nil
}

// TODO: Implement these
func extractMP4Date(filePath string) (*time.Time, error) {
	log.Println("MP4 date extraction not implemented yet, using filesystem date")
	return extractFileSystemDate(filePath)
}

func extractMOVDate(filePath string) (*time.Time, error) {
	log.Println("MOV date extraction not implemented yet, using filesystem date")
	return extractFileSystemDate(filePath)
}

func extractVideoDate(filePath string) (*time.Time, error) {
	log.Println("MOV date extraction not implemented yet, using filesystem date")
	return extractFileSystemDate(filePath)
}

func extractFileSystemDate(filePath string) (*time.Time, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	modTime := fileInfo.ModTime()
	if modTime.IsZero() {
		return nil, fmt.Errorf("no valid date found")
	}

	return &modTime, nil
}
