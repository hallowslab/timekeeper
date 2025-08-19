package main

import (
	"flag"
	"fmt"
	"github.com/hallowslab/timekeeper/internal/metadata"
	"log"
	"os"
	"path/filepath"
)

// Statistics for tracking progress
type Stats struct {
	Total         int
	Processed     int
	ExifCount     int
	FallbackCount int
	Errors        int
}

func (s *Stats) Print() {
	fmt.Printf("\n=== SUMMARY ===\n")
	fmt.Printf("Total files: %d\n", s.Total)
	fmt.Printf("Successfully processed: %d\n", s.Processed)
	if s.Processed > 0 {
		fmt.Printf("  - Using EXIF data: %d (%.1f%%)\n", s.ExifCount, float64(s.ExifCount)/float64(s.Processed)*100)
		fmt.Printf("  - Using fallback (ModTime): %d (%.1f%%)\n", s.FallbackCount, float64(s.FallbackCount)/float64(s.Processed)*100)
	}
	fmt.Printf("Errors: %d\n", s.Errors)
}

func main() {
	var source string
	var destination string
	var dryRun bool

	flag.StringVar(&source, "s", "", "Source file or directory")
	flag.StringVar(&destination, "d", "", "Destination directory")
	flag.BoolVar(&dryRun, "dry-run", false, "Show what would be done")
	flag.Parse()

	if source == "" || destination == "" {
		log.Fatal("Source (-s) and destination (-d) cannot be empty")
	}

	stats := &Stats{}
	if err := processPath(source, destination, dryRun, stats); err != nil {
		log.Fatal(err)
	}

	stats.Print()
}

func processPath(sourcePath, destBase string, dryRun bool, stats *Stats) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("cannot access source path: %v", err)
	}

	if info.IsDir() {
		// First pass: count total files
		fmt.Println("Scanning directory for media files...")
		if err := countMediaFiles(sourcePath, stats); err != nil {
			return err
		}
		fmt.Printf("Found %d media files to process\n\n", stats.Total)

		return processDirectory(sourcePath, destBase, dryRun, stats)
	}

	stats.Total = 1
	return processFile(sourcePath, destBase, dryRun, stats)
}

func countMediaFiles(sourceDir string, stats *Stats) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		if !info.IsDir() && metadata.IsMediaFile(path) {
			stats.Total++
		}

		return nil
	})
}

func processDirectory(sourceDir, destBase string, dryRun bool, stats *Stats) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing %s: %v", path, err)
			stats.Errors++
			return nil // Continue processing other files
		}

		if info.IsDir() {
			return nil
		}

		if metadata.IsMediaFile(path) {
			if err := processFile(path, destBase, dryRun, stats); err != nil {
				log.Printf("Error processing %s: %v", path, err)
				stats.Errors++
			}
		}

		return nil
	})
}

func processFile(sourcePath, destBase string, dryRun bool, stats *Stats) error {
	stats.Processed++
	filename := filepath.Base(sourcePath)
	prefix := ""
	if dryRun {
		prefix = "[DRY RUN]"
	}
	// Show progress
	// fmt.Printf("[%d/%d] Processing: %s", stats.Processed, stats.Total, filename)

	// Get ExifTool path (build-tag dependent)
	exiftoolPath, err := getExifTool()
	if err != nil {
		log.Printf(" (using fallback - ExifTool not available)\n")
		stats.FallbackCount++
		return metadata.ProcessFileWithFallback(sourcePath, destBase, dryRun)
	}

	// Extract metadata using ExifTool
	dateTime, err := metadata.ExtractDateTime(exiftoolPath, sourcePath)
	if err != nil {
		log.Printf(" (using fallback - EXIF extraction failed)\n")
		stats.FallbackCount++
		return metadata.ProcessFileWithFallback(sourcePath, destBase, dryRun)
	}

	// Success with EXIF
	// fmt.Printf(" (EXIF: %s)\n", dateTime.Format("2006-01-02"))
	stats.ExifCount++

	// Create destination directory structure
	destDir := filepath.Join(destBase,
		fmt.Sprintf("%d", dateTime.Year()),
		fmt.Sprintf("%s", dateTime.Month().String()))

	filename = filepath.Base(sourcePath)
	destPath := filepath.Join(destDir, filename)

	log.Printf("[%d/%d] %s Processed: %s -> %s", stats.Processed, stats.Total, prefix, filename, destPath)

	// Handle file name conflicts
	destPath = metadata.GetUniqueFilePath(destPath)

	if dryRun {
		return nil
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", destDir, err)
	}

	// Move the file
	log.Printf("  Moving: %s -> %s", sourcePath, destPath)
	return os.Rename(sourcePath, destPath)
}
