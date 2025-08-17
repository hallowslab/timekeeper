//go:build bundled
// +build bundled

package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed vendor/windows
var windowsExifTool embed.FS // Embed the entire exiftool directory structure

//go:embed vendor/linux/exiftool
var linuxExifTool []byte

//go:embed vendor/darwin/exiftool
var darwinExifTool []byte

func getExifTool() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return extractWindowsExifTool()
	case "linux":
		return extractLinuxExifTool()
	case "darwin":
		return extractDarwinExifTool()
	default:
		return "", fmt.Errorf("bundled ExifTool not available for %s", runtime.GOOS)
	}
}

func extractWindowsExifTool() (string, error) {
	// Create a temporary directory for ExifTool
	tempDir := filepath.Join(os.TempDir(), "timekeeper-exiftool")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", err
	}

	// Extract all Windows files from embed.FS
	err := fs.WalkDir(windowsExifTool, "assets/windows", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Convert embedded path to local path
		relPath, _ := filepath.Rel("assets/windows", path)
		localPath := filepath.Join(tempDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(localPath, 0755)
		}

		// Read file from embedded FS
		data, err := windowsExifTool.ReadFile(path)
		if err != nil {
			return err
		}

		// Write to temp directory
		return os.WriteFile(localPath, data, 0755)
	})

	if err != nil {
		return "", err
	}

	return filepath.Join(tempDir, "exiftool.exe"), nil
}

func extractLinuxExifTool() (string, error) {
	tempPath := filepath.Join(os.TempDir(), "timekeeper-exiftool")
	if err := os.WriteFile(tempPath, linuxExifTool, 0755); err != nil {
		return "", err
	}
	return tempPath, nil
}

func extractDarwinExifTool() (string, error) {
	tempPath := filepath.Join(os.TempDir(), "timekeeper-exiftool-mac")
	if err := os.WriteFile(tempPath, darwinExifTool, 0755); err != nil {
		return "", err
	}
	return tempPath, nil
}
