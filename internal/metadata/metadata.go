package metadata

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var supportedExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".tiff": true, ".tif": true,
	".raw": true, ".cr2": true, ".nef": true, ".arw": true, ".dng": true,
	".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".wmv": true,
	".m4v": true, ".3gp": true, ".webm": true,
}

func IsMediaFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return supportedExtensions[ext]
}

func ExtractDateTime(exiftoolPath, filePath string) (time.Time, error) {
	// Try to extract various date fields in order of preference
	dateFields := []string{
		"DateTimeOriginal",
		"CreateDate",
		"DateTime",
		"FileModifyDate",
	}

	// log.Printf("Using ExifTool: %s", exiftoolPath)

	for _, field := range dateFields {
		// Use simple -FieldName syntax instead of -p
		cmd := exec.Command(exiftoolPath, "-s", "-s", "-s", "-"+field, filePath)
		output, err := cmd.Output()

		if err != nil {
			log.Printf("Command failed: %s - Error: %v", strings.Join(cmd.Args, " "), err)
			continue
		}

		dateStr := strings.TrimSpace(string(output))
		// log.Printf("Field %s: raw output='%s'", field, dateStr)

		if dateStr == "" {
			continue
		}

		// Parse the date string
		if dateTime, err := ParseExifDate(dateStr); err == nil {
			// log.Printf("Successfully parsed: %s -> %s", dateStr, dateTime.Format("2006-01-02 15:04:05"))
			return dateTime, nil
		} else {
			log.Printf("Failed to parse '%s': %v", dateStr, err)
		}
	}

	return time.Time{}, fmt.Errorf("no valid date found in EXIF data")
}

func ParseExifDate(dateStr string) (time.Time, error) {
	// ExifTool date formats to try
	formats := []string{
		"2006:01:02 15:04:05",
		"2006:01:02 15:04:05-07:00",
		"2006:01:02 15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func ProcessFileWithFallback(sourcePath, destBase string, dryRun bool) error {
	// Use file mod time as fallback
	// go does not have creation date because older unix filesystems had no support
	info, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	dateTime := info.ModTime()

	// Create destination directory structure
	destDir := filepath.Join(destBase,
		fmt.Sprintf("%d", dateTime.Year()),
		fmt.Sprintf("%s", dateTime.Month().String()))

	filename := filepath.Base(sourcePath)
	destPath := filepath.Join(destDir, filename)

	// Handle file name conflicts
	destPath = GetUniqueFilePath(destPath)

	if dryRun {
		log.Printf("[DRY RUN] [FALLBACK] Would move: %s -> %s", sourcePath, destPath)
		return nil
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", destDir, err)
	}

	// Move the file
	log.Printf("[FALLBACK] Moving: %s -> %s", sourcePath, destPath)
	return os.Rename(sourcePath, destPath)
}

func GetUniqueFilePath(originalPath string) string {
	if _, err := os.Stat(originalPath); os.IsNotExist(err) {
		return originalPath
	}

	dir := filepath.Dir(originalPath)
	base := filepath.Base(originalPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	counter := 1
	for {
		newPath := filepath.Join(dir, fmt.Sprintf("%s_%d%s", name, counter, ext))
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}
