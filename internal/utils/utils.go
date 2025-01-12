package utils

import (
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
	mapPathToChildren := make(map[string][]traverse.PathInfo)
	for _, path := range paths {
		if path.Depth == 0 {
			continue
		}
		parent := filepath.Dir(path.Path)
		mapPathToChildren[parent] = append(mapPathToChildren[parent], path)
	}

	var sb strings.Builder
	var printTree func(path traverse.PathInfo, prefix string, isLast bool)
	printTree = func(path traverse.PathInfo, prefix string, isLast bool) {
		// Print current item
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		name := filepath.Base(path.Path)
		if path.IsDir {
			name += "/"
		}
		sb.WriteString(fmt.Sprintf("%s%s%s\n", prefix, connector, name))

		// Print children
		childPrefix := prefix + "│   "
		if isLast {
			childPrefix = prefix + "    "
		}

		pathChildren := mapPathToChildren[path.Path]
		for i, child := range pathChildren {
			printTree(child, childPrefix, i == len(pathChildren)-1)
		}
	}

	// Process root level items
	rootPaths := mapPathToChildren[filepath.Dir(paths[0].Path)]
	for i, path := range rootPaths {
		printTree(path, "", i == len(rootPaths)-1)
	}

	return sb.String()
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

// FormatTimestamp returns a formatted timestamp string
func FormatTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
