package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Level represents a log level.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging to both stdout and a log file.
type Logger struct {
	mu          sync.Mutex
	level       Level
	file        *os.File
	filePath    string
	writers     []io.Writer
	secrets     map[string]bool // secret values to mask in output
	initialized bool
}

const (
	maxLogFiles = 10
	logDir      = ".templatr/logs"
)

// New creates a new logger. Call Init() to set up the log file.
func New() *Logger {
	return &Logger{
		level:   INFO,
		secrets: make(map[string]bool),
	}
}

// Init sets up the log file in ~/.templatr/logs/.
func (l *Logger) Init() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	dir := filepath.Join(home, logDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Rotate old log files
	l.rotateFiles(dir)

	// Create new log file
	timestamp := time.Now().Format("2006-01-02_150405")
	l.filePath = filepath.Join(dir, fmt.Sprintf("setup-%s.log", timestamp))

	f, err := os.Create(l.filePath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	l.file = f
	l.writers = []io.Writer{f}
	l.initialized = true

	// Write header
	l.writeToFile("=== templatr-setup log started at %s ===\n", time.Now().Format(time.RFC3339))

	return nil
}

// SetLevel sets the minimum log level for stdout output.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// AddSecret adds a value that should be masked in all log output.
func (l *Logger) AddSecret(secret string) {
	if secret == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.secrets[secret] = true
}

// Debug logs a debug message (file only, unless verbose).
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message.
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// FilePath returns the path to the current log file.
func (l *Logger) FilePath() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.filePath
}

// Close closes the log file.
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		l.writeToFile("=== templatr-setup log ended at %s ===\n", time.Now().Format(time.RFC3339))
		l.file.Close()
		l.file = nil
	}
}

// RecentLogFiles returns the paths of recent log files, newest first.
func RecentLogFiles(max int) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(home, logDir)
	return RecentLogFilesInDir(dir, max)
}

// RecentLogFilesInDir returns log files from a specific directory, newest first.
func RecentLogFilesInDir(dir string, max int) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "setup-") && strings.HasSuffix(e.Name(), ".log") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}

	// Sort descending by name (timestamp-based names sort chronologically)
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	if max > 0 && len(files) > max {
		files = files[:max]
	}

	return files, nil
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	msg = l.maskSecrets(msg)

	timestamp := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[%s] %s: %s", timestamp, level, msg)

	l.mu.Lock()
	defer l.mu.Unlock()

	// Always write to log file (all levels)
	if l.initialized {
		l.writeToFile("%s\n", line)
	}

	// Write to stdout only if level >= configured level
	if level >= l.level {
		switch level {
		case ERROR:
			fmt.Fprintf(os.Stderr, "  ERROR: %s\n", msg)
		case WARN:
			fmt.Printf("  WARN: %s\n", msg)
		default:
			fmt.Printf("  %s\n", msg)
		}
	}
}

func (l *Logger) writeToFile(format string, args ...interface{}) {
	if l.file != nil {
		fmt.Fprintf(l.file, format, args...)
	}
}

func (l *Logger) maskSecrets(msg string) string {
	for secret := range l.secrets {
		msg = strings.ReplaceAll(msg, secret, "****")
	}
	return msg
}

func (l *Logger) rotateFiles(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var logFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "setup-") && strings.HasSuffix(e.Name(), ".log") {
			logFiles = append(logFiles, filepath.Join(dir, e.Name()))
		}
	}

	// Keep only the most recent files
	if len(logFiles) >= maxLogFiles {
		sort.Strings(logFiles)
		// Remove oldest files to make room
		toRemove := len(logFiles) - maxLogFiles + 1
		for i := 0; i < toRemove; i++ {
			os.Remove(logFiles[i])
		}
	}
}
