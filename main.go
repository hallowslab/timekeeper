package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hallowslab/timekeeper/internal/metadata"
)

func main() {
	var source string
	var destination string
	var dryRun bool

	flag.StringVar(&source, "s", "", "Source file or directory")
	flag.StringVar(&destination, "d", "", "Destination directory")
	flag.BoolVar(&dryRun, "dry-run", false, "Show what would be done")
	flag.Parse()

	if source == "" || destination == "" {
		log.Fatal("Source or destination cannot be empty")
	}

	// This function exists in either exiftool_bundled.go OR exiftool_system.go
	// depending on build tags
	if err := processFile(source, destination, dryRun); err != nil {
		log.Fatal(err)
	}
}

func processFile(sourcePath, destBase string, dryRun bool) error {
	// Use getExifTool() - this function will be different depending on build
	exiftoolPath, err := getExifTool()
	if err != nil {
		fmt.Printf("ExifTool error: %v\n", err)
		// Fallback to filesystem dates
		return nil
	}

	// Rest of your existing logic...
	destDir := filepath.Join(destBase,
		fmt.Sprintf("%d", dateTime.Year()),
		fmt.Sprintf("%02d-%s", dateTime.Month(), dateTime.Month().String()))

	filename := filepath.Base(sourcePath)
	destPath := filepath.Join(destDir, filename)

	if dryRun {
		log.Printf("[DRY RUN] Would move: %s -> %s\n", sourcePath, destPath)
		return nil
	}

	if err := os.MkdirAll(destDir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", destDir, err)
	}

	log.Printf("Moving: %s -> %s\n", sourcePath, destPath)
	return nil
}
