package player

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const ffplayURL = "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip"

// DownloadFFplay downloads and extracts ffplay.exe and required DLLs on Windows.
func DownloadFFplay(ctx context.Context) (string, error) {
	if runtime.GOOS != "windows" {
		return "", errors.New("ffplay auto-download is supported only on Windows")
	}

	dir, err := downloadDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	ffplayPath := filepath.Join(dir, "ffplay.exe")
	if isExecutable(ffplayPath) {
		return ffplayPath, nil
	}

	client := &http.Client{Timeout: 10 * time.Minute}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ffplayURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("failed to download ffplay")
	}

	zipFile, err := os.CreateTemp(dir, "ffplay-*.zip")
	if err != nil {
		return "", err
	}
	zipPath := zipFile.Name()
	if _, err := io.Copy(zipFile, resp.Body); err != nil {
		_ = zipFile.Close()
		_ = os.Remove(zipPath)
		return "", err
	}
	if err := zipFile.Close(); err != nil {
		_ = os.Remove(zipPath)
		return "", err
	}

	if err := extractFFplay(zipPath, dir); err != nil {
		_ = os.Remove(zipPath)
		return "", err
	}
	_ = os.Remove(zipPath)

	if !isExecutable(ffplayPath) {
		return "", errors.New("ffplay download did not produce ffplay.exe")
	}
	return ffplayPath, nil
}

func extractFFplay(zipPath string, destDir string) error {
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	extracted := false
	for _, file := range zipReader.File {
		name := strings.ReplaceAll(file.Name, "\\", "/")
		if !strings.Contains(name, "/bin/") {
			continue
		}
		base := filepath.Base(name)
		lower := strings.ToLower(base)
		if lower != "ffplay.exe" && !strings.HasSuffix(lower, ".dll") {
			continue
		}
		if err := extractZipFile(file, filepath.Join(destDir, base)); err != nil {
			return err
		}
		if lower == "ffplay.exe" {
			extracted = true
		}
	}

	if !extracted {
		return errors.New("ffplay.exe not found in archive")
	}
	return nil
}

func extractZipFile(file *zip.File, destPath string) error {
	if file.FileInfo().IsDir() {
		return nil
	}
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()

	output, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, reader)
	return err
}

func downloadDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "valvefm", "bin"), nil
}

func findDownloadedPlayer() (string, string) {
	dir, err := downloadDir()
	if err != nil {
		return "", ""
	}

	candidates := []struct {
		backend string
		name    string
	}{
		{backend: "ffplay", name: "ffplay.exe"},
		{backend: "ffplay", name: "ffplay"},
		{backend: "mpv", name: "mpv.exe"},
		{backend: "mpv", name: "mpv"},
	}
	for _, candidate := range candidates {
		path := filepath.Join(dir, candidate.name)
		if isExecutable(path) {
			return path, candidate.backend
		}
	}
	return "", ""
}
