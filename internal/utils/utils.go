package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Broderick-Westrope/amalgo/internal/traverse"
)

// IsBinaryFile determines if a file is binary by checking its contents
func IsBinaryFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}
	buf = buf[:n]

	// Check for null bytes
	if bytes.IndexByte(buf, 0) != -1 {
		return true, nil
	}

	// Look for non-text characters
	for _, b := range buf {
		if b < 32 && b != 9 && b != 10 && b != 13 { // Not tab, LF, or CR
			return true, nil
		}
	}

	return false, nil
}

// GenerateTree creates a textual representation of the directory structure
func GenerateTree(paths []traverse.PathInfo) string {
	var builder strings.Builder
	builder.WriteString("Directory Tree\n")

	for _, path := range paths {
		if path.Depth == 0 {
			continue
		}

		var indent string
		if path.Depth == 1 {
			indent = "├── "
		} else {
			indent = strings.Repeat("│   ", path.Depth-2) + "├── "
		}

		name := filepath.Base(path.Path)
		if path.IsDir {
			name += "/"
		}
		builder.WriteString(fmt.Sprintf("%s%s\n", indent, name))
	}

	return builder.String()
}

// WriteOutput writes content to a file or stdout
func WriteOutput(path string, content string) error {
	if path == "stdout" || path == "-" {
		_, err := fmt.Print(content)
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// AppendOutput appends content to a file or writes to stdout
func AppendOutput(path string, content string) error {
	if path == "stdout" || path == "-" {
		_, err := fmt.Print(content)
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}

// Clean removes empty lines and trims whitespace
func Clean(content string) string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

// FormatTimestamp returns a formatted timestamp string
func FormatTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// ConfigureLogging sets up logging with the specified verbosity
// func ConfigureLogging(verbose bool) {
// 	logLevel := log.LevelWarn
// 	if verbose {
// 		logLevel = log.LevelInfo
// 	}

// 	log.SetLevel(logLevel)
// 	log.SetFormatter(&log.TextFormatter{
// 		FullTimestamp:   true,
// 		TimestampFormat: "2006-01-02 15:04:05",
// 	})
// }
