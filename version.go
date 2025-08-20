// version.go
package main

import (
	"fmt"
	"runtime"
)

// Version information - set via ldflags during build
var (
	version   = "dev"     // Set via -ldflags "-X main.version=v1.0.0"
	commit    = "unknown" // Set via -ldflags "-X main.commit=$(git rev-parse HEAD)"
	buildTime = "unknown" // Set via -ldflags "-X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
)

// Version holds version information
type Version struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// GetVersion returns version information
func GetVersion() Version {
	return Version{
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// shortCommit returns a shortened version of the commit hash
func (v Version) ShortCommit() string {
	if len(v.Commit) > 8 {
		return v.Commit[:8]
	}
	return v.Commit
}

// String returns a formatted version string
func (v Version) String() string {
	return fmt.Sprintf("%s (%s) built on %s with %s for %s",
		v.Version, v.ShortCommit(), v.BuildTime, v.GoVersion, v.Platform)
}

// main.go additions
func init() {
	// You can add version checking logic here if needed
}

// Add this to your main function or as a separate command
func printVersion() {
	v := GetVersion()
	fmt.Printf("timekeeper version %s\n", v.String())
}
