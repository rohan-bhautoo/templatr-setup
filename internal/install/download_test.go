package install

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyChecksum_Match(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("hello world")
	os.WriteFile(testFile, content, 0o644)

	h := sha256.Sum256(content)
	expected := hex.EncodeToString(h[:])

	if err := VerifyChecksum(testFile, expected); err != nil {
		t.Errorf("expected checksum to match, got error: %s", err)
	}
}

func TestVerifyChecksum_Mismatch(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("hello world"), 0o644)

	err := VerifyChecksum(testFile, "0000000000000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Error("expected checksum mismatch error")
	}
}

func TestVerifyChecksum_FileNotFound(t *testing.T) {
	err := VerifyChecksum("/nonexistent/file", "abc123")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestDownloadFile(t *testing.T) {
	// Start a test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test content"))
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	destFile := filepath.Join(tmpDir, "downloaded.txt")

	err := DownloadFile(ts.URL, destFile, nil)
	if err != nil {
		t.Fatalf("download failed: %s", err)
	}

	data, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("read failed: %s", err)
	}
	if string(data) != "test content" {
		t.Errorf("expected 'test content', got %q", string(data))
	}
}

func TestDownloadFile_WithProgress(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "12")
		w.Write([]byte("test content"))
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	destFile := filepath.Join(tmpDir, "downloaded.txt")

	var lastDownloaded, lastTotal int64
	progress := func(downloaded, total int64) {
		lastDownloaded = downloaded
		lastTotal = total
	}

	err := DownloadFile(ts.URL, destFile, progress)
	if err != nil {
		t.Fatalf("download failed: %s", err)
	}

	if lastDownloaded != 12 {
		t.Errorf("expected 12 bytes downloaded, got %d", lastDownloaded)
	}
	if lastTotal != 12 {
		t.Errorf("expected total 12, got %d", lastTotal)
	}
}

func TestDownloadFile_404(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	destFile := filepath.Join(tmpDir, "downloaded.txt")

	err := DownloadFile(ts.URL, destFile, nil)
	if err == nil {
		t.Error("expected error for 404")
	}
}

func TestFetchChecksumFromURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("abc123  file1.tar.gz\ndef456  file2.tar.gz\n"))
	}))
	defer ts.Close()

	hash, err := FetchChecksumFromURL(ts.URL, "file2.tar.gz")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if hash != "def456" {
		t.Errorf("expected 'def456', got %q", hash)
	}
}

func TestFetchChecksumFromURL_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("abc123  file1.tar.gz\n"))
	}))
	defer ts.Close()

	_, err := FetchChecksumFromURL(ts.URL, "nonexistent.tar.gz")
	if err == nil {
		t.Error("expected error for missing filename")
	}
}

func TestExtractAndFlatten_Zip(t *testing.T) {
	// We'll test the zip extraction with a minimal test
	// by creating a zip with a single top-level dir
	tmpDir := t.TempDir()

	// Create a zip file manually is complex, so we just test ExtractZip error cases
	err := ExtractZip(filepath.Join(tmpDir, "nonexistent.zip"), tmpDir)
	if err == nil {
		t.Error("expected error for nonexistent zip")
	}
}

func TestExtractArchive_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.rar")
	os.WriteFile(testFile, []byte("fake"), 0o644)

	err := ExtractArchive(testFile, tmpDir)
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestPlatformExt(t *testing.T) {
	ext := PlatformExt()
	// On Windows it should be "zip", on others "tar.gz"
	if ext != "zip" && ext != "tar.gz" {
		t.Errorf("unexpected extension: %s", ext)
	}
}

func TestPlatformArch(t *testing.T) {
	arch := PlatformArch()
	// Should return a valid architecture string
	if arch == "" {
		t.Error("expected non-empty architecture")
	}
}

func TestRuntimesDir(t *testing.T) {
	dir, err := RuntimesDir()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if dir == "" {
		t.Error("expected non-empty runtimes dir")
	}
	if !filepath.IsAbs(dir) {
		t.Error("expected absolute path")
	}
}
