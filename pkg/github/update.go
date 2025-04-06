package github

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/dipjyotimetia/jarvis/pkg/logger"
)

const (
	GithubOwner = "dipjyotimetia"
	GithubRepo  = "jarvis"
)

// GetReleaseDownloadURL constructs the GitHub release URL for the specified version
func GetReleaseDownloadURL(version string) (string, error) {
	// Remove v prefix if present
	version = strings.TrimPrefix(version, "v")

	// Create the download URL based on OS and architecture
	architecture := runtime.GOARCH
	osType := runtime.GOOS

	// Handle naming differences
	if architecture == "amd64" {
		architecture = "x86_64"
	}

	var extension string
	if osType == "windows" {
		extension = "zip"
	} else {
		extension = "tar.gz"
	}

	// Construct the URL for the release asset
	url := fmt.Sprintf(
		"https://github.com/%s/%s/releases/download/v%s/%s_%s_%s.%s",
		GithubOwner, GithubRepo, version,
		GithubRepo, strings.ToTitle(osType), architecture, extension,
	)

	return url, nil
}

// SelfUpdate updates the binary to the latest version
func SelfUpdate(currentVersion string) error {
	logger.Info("Checking for updates...")

	// Get the latest version from GitHub
	latestVersion, err := GetLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	// Parse the current version
	current, err := semver.NewVersion(currentVersion)
	if err != nil {
		return fmt.Errorf("failed to parse current version %s: %w", currentVersion, err)
	}

	// Compare versions
	if latestVersion.LessThanEqual(current) {
		logger.Info("You're already using the latest version of Jarvis!")
		return nil
	}

	logger.Info("%s", fmt.Sprintf("New version found: %s (current: %s)", latestVersion.String(), current.String()))
	logger.Info("Downloading the latest version...")

	// Get the download URL
	downloadURL, err := GetReleaseDownloadURL(latestVersion.String())
	if err != nil {
		return fmt.Errorf("failed to construct download URL: %w", err)
	}

	// Download to a temporary file
	tempFile, err := downloadToTempFile(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Extract the binary from the archive
	binPath, err := extractBinary(tempFile.Name())
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}
	defer os.Remove(binPath)

	// Get the current executable path
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Replace the current binary
	err = replaceBinary(binPath, executablePath)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	logger.Info("%s", fmt.Sprintf("Successfully updated to version %s", latestVersion.String()))
	return nil
}

// downloadToTempFile downloads a file from a URL to a temporary file
func downloadToTempFile(url string) (*os.File, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	tmpFile, err := os.CreateTemp("", "jarvis-update-*")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, err
	}

	tmpFile.Close()
	return tmpFile, nil
}

// extractBinary extracts the binary from the archive
func extractBinary(archivePath string) (string, error) {
	// Create a temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "jarvis-extract-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	// Check if it's a zip or tar.gz file based on the file extension
	if strings.HasSuffix(archivePath, ".zip") {
		err = unzip(archivePath, tempDir)
	} else if strings.HasSuffix(archivePath, ".tar.gz") {
		err = untar(archivePath, tempDir)
	} else {
		return "", fmt.Errorf("unsupported archive format: %s", archivePath)
	}

	if err != nil {
		return "", fmt.Errorf("extraction failed: %w", err)
	}

	// Find the binary in the extracted files
	binPath := ""
	binName := GithubRepo
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	// Look for the binary in the extracted files
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.Contains(info.Name(), binName) {
			binPath = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if binPath == "" {
		return "", fmt.Errorf("binary not found in the archive")
	}

	// Create a temporary file for the binary
	tmpBin, err := os.CreateTemp("", "jarvis-bin-*")
	if err != nil {
		return "", err
	}
	tmpBin.Close()

	// Copy the binary to the temporary location
	srcFile, err := os.Open(binPath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	destFile, err := os.Create(tmpBin.Name())
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return "", err
	}

	// Make the file executable
	err = os.Chmod(tmpBin.Name(), 0o755)
	if err != nil {
		return "", err
	}

	return tmpBin.Name(), nil
}

// unzip extracts a zip archive to the specified destination
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Store filename/path for returning and using later
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			// Create directory if it doesn't exist
			if err = os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		// Create containing directory if it doesn't exist
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		// Open file for writing
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// untar extracts a tar.gz archive to the specified destination
func untar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Get the target path
		target := filepath.Join(dest, header.Name)

		// Check for path traversal attacks
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", target)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory if it doesn't exist
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create containing directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}

			// Create the file
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()

			// Copy the contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}

	return nil
}

// replaceBinary replaces the current binary with the new one
func replaceBinary(newBinPath, currentBinPath string) error {
	// On Unix systems, we can directly rename/replace the executable
	// On Windows, we need to use a different approach due to file locking

	// For Unix-like systems (Linux/macOS)
	if runtime.GOOS != "windows" {
		return os.Rename(newBinPath, currentBinPath)
	}

	// For Windows, we need a more complex replacement strategy
	// Typically involves creating a .bat file that replaces the executable
	// after the current process exits

	// Generate a batch file to execute after this process exits
	batchPath := filepath.Join(os.TempDir(), "jarvis-update.bat")
	batchContent := fmt.Sprintf(`@echo off
:retry
timeout /t 1 /nobreak > NUL
copy /Y "%s" "%s"
if errorlevel 1 goto retry
del "%s"
del "%%~f0"
`, newBinPath, currentBinPath, newBinPath)

	err := os.WriteFile(batchPath, []byte(batchContent), 0o755)
	if err != nil {
		return err
	}

	// Execute the batch file
	cmd := exec.Command("cmd", "/C", "start", "/min", batchPath)
	err = cmd.Start()
	if err != nil {
		return err
	}

	return nil
}
