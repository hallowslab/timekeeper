<#
.SYNOPSIS
    Build script for Timekeeper (PowerShell replacement for Makefile)
.DESCRIPTION
    Provides commands for building, testing, cleaning, and releasing the Go project.
    Usage:
        ./build.ps1 build
        ./build.ps1 build-all
        ./build.ps1 release-local
        ./build.ps1 test
        ./build.ps1 version
#>

param(
    [string]$Target = "help"
)

# --- Version info ---
$VERSION = (git describe --tags --always --dirty 2>$null)
if (-not $VERSION) { $VERSION = "dev" }
$COMMIT = (git rev-parse --short HEAD 2>$null)
if (-not $COMMIT) { $COMMIT = "unknown" }
$BUILD_TIME = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ")
$LDFLAGS = "-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.buildTime=$BUILD_TIME"

# --- Helper functions ---
function Run($cmd) {
    Write-Host ">>> $cmd" -ForegroundColor Cyan
    Invoke-Expression $cmd
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

switch ($Target) {
    "build" {
        $ext = if ($IsWindows) { ".exe" } else { "" }
        Run "go build -ldflags='$LDFLAGS' -o timekeeper$ext ."
    }
    "build-bundled" {
        if ($IsWindows) {
            Run "go build -tags bundled -ldflags='$LDFLAGS' -o timekeeper.exe ."
        } else {
            Write-Host "Bundled build only supported on Windows" -ForegroundColor Yellow
            & $PSCommandPath build
        }
    }
    "build-all" {
        & $PSCommandPath clean
        New-Item -ItemType Directory -Force -Path dist | Out-Null

        # Windows (bundled)
        Run "set GOOS=windows; set GOARCH=amd64; go build -tags bundled -ldflags='$LDFLAGS' -o dist/timekeeper-windows-amd64.exe ."
        Run "set GOOS=windows; set GOARCH=arm64; go build -tags bundled -ldflags='$LDFLAGS' -o dist/timekeeper-windows-arm64.exe ."

        # Linux
        Run "set GOOS=linux; set GOARCH=amd64; go build -ldflags='$LDFLAGS' -o dist/timekeeper-linux-amd64 ."
        Run "set GOOS=linux; set GOARCH=arm64; go build -ldflags='$LDFLAGS' -o dist/timekeeper-linux-arm64 ."

        # macOS
        Run "set GOOS=darwin; set GOARCH=amd64; go build -ldflags='$LDFLAGS' -o dist/timekeeper-macos-amd64 ."
        Run "set GOOS=darwin; set GOARCH=arm64; go build -ldflags='$LDFLAGS' -o dist/timekeeper-macos-arm64 ."
    }
    "release-local" {
        & $PSCommandPath build-all
        New-Item -ItemType Directory -Force -Path dist/releases | Out-Null

        # zip/tar releases
        Run "Compress-Archive -Path dist/timekeeper-windows-amd64.exe -DestinationPath dist/releases/timekeeper-$VERSION-windows-amd64.zip -Force"
        Run "Compress-Archive -Path dist/timekeeper-windows-arm64.exe -DestinationPath dist/releases/timekeeper-$VERSION-windows-arm64.zip -Force"
        Run "tar -czf dist/releases/timekeeper-$VERSION-linux-amd64.tar.gz -C dist timekeeper-linux-amd64"
        Run "tar -czf dist/releases/timekeeper-$VERSION-linux-arm64.tar.gz -C dist timekeeper-linux-arm64"
        Run "tar -czf dist/releases/timekeeper-$VERSION-macos-amd64.tar.gz -C dist timekeeper-macos-amd64"
        Run "tar -czf dist/releases/timekeeper-$VERSION-macos-arm64.tar.gz -C dist timekeeper-macos-arm64"

        Write-Host "Release files created in dist/releases/" -ForegroundColor Green
    }
    "test" {
        Run "go test -v ./..."
    }
    "test-coverage" {
        Run "go test -v -race -coverprofile=coverage.out ./..."
        Run "go tool cover -html=coverage.out -o coverage.html"
    }
    "clean" {
        Remove-Item -Recurse -Force dist, timekeeper, timekeeper.exe, coverage.out, coverage.html -ErrorAction SilentlyContinue
    }
    "version" {
        Write-Host "Version: $VERSION"
        Write-Host "Commit: $COMMIT"
        Write-Host "Build Time: $BUILD_TIME"
    }
    "dev" {
        Run "go build -o timekeeper ."
    }
    "install" {
        Run "go install -ldflags='$LDFLAGS' ."
    }
    "fmt" {
        Run "go fmt ./..."
    }
    "lint" {
        Run "golangci-lint run"
    }
    "deps-update" {
        Run "go get -u ./..."
        Run "go mod tidy"
    }
    "help" {
        Write-Host "Available targets:"
        Write-Host "  build         - Build for current platform"
        Write-Host "  build-bundled - Build with bundled dependencies (Windows only)"
        Write-Host "  build-all     - Build for all platforms"
        Write-Host "  test          - Run tests"
        Write-Host "  test-coverage - Run tests with coverage"
        Write-Host "  clean         - Clean build artifacts"
        Write-Host "  version       - Show version information"
        Write-Host "  release-local - Create local release packages"
        Write-Host "  dev           - Fast development build"
        Write-Host "  install       - Install to GOPATH/bin"
        Write-Host "  fmt           - Format code"
        Write-Host "  lint          - Lint code"
        Write-Host "  deps-update   - Update dependencies"
    }
    default {
        Write-Host "Unknown target: $Target" -ForegroundColor Red
        exit 1
    }
}
