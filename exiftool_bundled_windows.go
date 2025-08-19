//go:build bundled && windows

package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// DO NOT USE (-k), rename file to avoid enter prompt
//
//go:embed bin/windows/exiftool.exe
var exiftoolExe []byte

//go:embed bin/windows/exiftool_files/*
var exiftoolFiles embed.FS

func getExifTool() (string, error) {
	return extractWindowsExifTool()
}

func extractWindowsExifTool() (string, error) {
	tempDir := filepath.Join(os.TempDir(), "timekeeper-exiftool")
	exePath := filepath.Join(tempDir, "exiftool.exe")

	// Check if already extracted
	if _, err := os.Stat(exePath); err == nil {
		return exePath, nil
	}

	// Remove existing temp directory to ensure clean extraction
	// os.RemoveAll(tempDir)

	// Create temp directory
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	// Extract main executable
	if err := os.WriteFile(exePath, exiftoolExe, 0755); err != nil {
		return "", fmt.Errorf("failed to extract exiftool.exe: %v", err)
	}

	// Extract exiftool_files directory
	err := fs.WalkDir(exiftoolFiles, "bin/windows/exiftool_files", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == "bin/windows/exiftool_files" {
			return nil
		}

		// Get path relative to "bin/windows/exiftool_files" instead of "bin/windows"
		relPath, _ := filepath.Rel("bin/windows/exiftool_files", path)
		localPath := filepath.Join(tempDir, "exiftool_files", relPath)

		if d.IsDir() {
			// fmt.Printf("Creating directory: %s\n", localPath)
			return os.MkdirAll(localPath, 0755)
		}

		// Ensure parent directory exists
		parentDir := filepath.Dir(localPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory %s: %v", parentDir, err)
		}

		// Read and write file
		data, err := exiftoolFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %v", path, err)
		}

		// fmt.Printf("Extracting file: %s\n", localPath)
		return os.WriteFile(localPath, data, 0644)
	})

	if err != nil {
		return "", fmt.Errorf("failed to extract exiftool_files: %v", err)
	}

	return exePath, nil
}
