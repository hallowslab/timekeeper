//go:build !bundled
// +build !bundled

package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

// This function only exists in the system build
func getExifTool() (string, error) {
	if path, err := exec.LookPath("exiftool"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("exiftool not found. Install it with:\n%s", getInstallInstructions())
}

func getInstallInstructions() string {
	switch runtime.GOOS {
	case "windows":
		return "winget install ExifTool"
	case "linux":
		return "sudo apt install libimage-exiftool-perl"
	case "darwin":
		return "brew install exiftool"
	default:
		return "https://exiftool.org/"
	}
}
