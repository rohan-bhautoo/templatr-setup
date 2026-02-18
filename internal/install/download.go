package install

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ProgressFunc is called during downloads with bytes downloaded and total bytes.
type ProgressFunc func(downloaded, total int64)

// DownloadFile downloads a file from the given URL to destPath.
func DownloadFile(url, destPath string, progress ProgressFunc) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d for %s", resp.StatusCode, url)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", destPath, err)
	}
	defer out.Close()

	reader := io.Reader(resp.Body)
	if progress != nil {
		reader = &progressReader{
			reader:   resp.Body,
			total:    resp.ContentLength,
			progress: progress,
		}
	}

	if _, err := io.Copy(out, reader); err != nil {
		return fmt.Errorf("failed to write %s: %w", destPath, err)
	}

	return nil
}

type progressReader struct {
	reader     io.Reader
	progress   ProgressFunc
	total      int64
	downloaded int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.downloaded += int64(n)
	pr.progress(pr.downloaded, pr.total)
	return n, err
}

// VerifyChecksum checks that a file's SHA256 hash matches the expected value.
func VerifyChecksum(filePath, expectedHash string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open %s for checksum: %w", filePath, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("failed to compute checksum for %s: %w", filePath, err)
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(actual, expectedHash) {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", filepath.Base(filePath), expectedHash, actual)
	}
	return nil
}

// FetchChecksumFromURL downloads a SHASUMS256.txt-style file and returns the hash for the given filename.
func FetchChecksumFromURL(url, filename string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch checksums from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksums URL returned HTTP %d: %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read checksums: %w", err)
	}

	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "hash  filename" or "hash *filename"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := strings.TrimLeft(parts[1], "*")
			if name == filename {
				return parts[0], nil
			}
		}
	}

	return "", fmt.Errorf("checksum not found for %s in %s", filename, url)
}

// ExtractTarGz extracts a .tar.gz archive to destDir.
func ExtractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to open gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	cleanDest := filepath.Clean(destDir)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read error: %w", err)
		}

		target := filepath.Join(destDir, hdr.Name)

		// Prevent path traversal attacks
		if !strings.HasPrefix(filepath.Clean(target), cleanDest) {
			return fmt.Errorf("invalid path in archive: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)|0o200)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			// Remove existing symlink if present
			os.Remove(target)
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return err
			}
		}
	}
	return nil
}

// ExtractZip extracts a .zip archive to destDir.
func ExtractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	cleanDest := filepath.Clean(destDir)

	for _, f := range r.File {
		target := filepath.Join(destDir, f.Name)

		if !strings.HasPrefix(filepath.Clean(target), cleanDest) {
			return fmt.Errorf("invalid path in archive: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0o755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode()|0o200)
		if err != nil {
			rc.Close()
			return err
		}

		_, copyErr := io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

// ExtractArchive detects format from filename and extracts accordingly.
func ExtractArchive(archivePath, destDir string) error {
	lower := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return ExtractTarGz(archivePath, destDir)
	case strings.HasSuffix(lower, ".zip"):
		return ExtractZip(archivePath, destDir)
	default:
		return fmt.Errorf("unsupported archive format: %s", filepath.Base(archivePath))
	}
}

// ExtractAndFlatten extracts an archive and moves the contents of the single
// top-level directory to targetDir. Many runtime archives (Node, Go, etc.)
// contain a single top-level directory that we want to strip.
func ExtractAndFlatten(archivePath, targetDir string) error {
	// Extract to a temp directory next to the target
	tmpDir, err := os.MkdirTemp(filepath.Dir(targetDir), "extract-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := ExtractArchive(archivePath, tmpDir); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Find the single top-level entry
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return err
	}

	var sourceDir string
	if len(entries) == 1 && entries[0].IsDir() {
		// Single top-level directory — flatten it
		sourceDir = filepath.Join(tmpDir, entries[0].Name())
	} else {
		// No single top-level dir — use extracted contents directly
		sourceDir = tmpDir
	}

	// Ensure target parent exists
	if err := os.MkdirAll(filepath.Dir(targetDir), 0o755); err != nil {
		return err
	}

	// Remove target if it exists
	os.RemoveAll(targetDir)

	// Move source to target
	if err := os.Rename(sourceDir, targetDir); err != nil {
		// Rename can fail across filesystems — fall back to copy
		return copyDir(sourceDir, targetDir)
	}

	return nil
}

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else if entry.Type()&os.ModeSymlink != 0 {
			link, err := os.Readlink(srcPath)
			if err != nil {
				return err
			}
			os.Symlink(link, dstPath)
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// RuntimesDir returns the base directory for installed runtimes (~/.templatr/runtimes/).
func RuntimesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".templatr", "runtimes"), nil
}

// PlatformExt returns the typical archive extension for the current OS.
func PlatformExt() string {
	if runtime.GOOS == "windows" {
		return "zip"
	}
	return "tar.gz"
}

// PlatformArch returns the architecture name commonly used in download URLs.
func PlatformArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "arm64":
		return "arm64"
	default:
		return runtime.GOARCH
	}
}

// FetchJSON is a helper that fetches a URL and returns the response body as bytes.
func FetchJSON(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}
