package internal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// IsBinaryFile determines if a file is binary by checking its contents
func IsBinaryFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	// Read first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("reading file: %w", err)
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
func GenerateTree(paths []PathInfo) string {
	if len(paths) == 0 {
		return "< no paths found >\n"
	}

	mapPathToChildren := make(map[string][]PathInfo)
	for _, path := range paths {
		if path.Depth == 0 {
			continue
		}
		parent := filepath.Dir(path.Path)
		mapPathToChildren[parent] = append(mapPathToChildren[parent], path)
	}

	var output string
	var printTree func(path PathInfo, prefix string, isLast bool)
	printTree = func(path PathInfo, prefix string, isLast bool) {
		// Print current item
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		name := filepath.Base(path.Path)
		if path.IsDir {
			name += "/"
		}
		output += fmt.Sprintf("%s%s%s\n", prefix, connector, name)

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

	shortestPath := paths[0]
	for _, path := range paths {
		if len(path.Path) < len(shortestPath.Path) {
			shortestPath = path
		}
	}

	// Find and process root level items.
	rootPaths := mapPathToChildren[filepath.Dir(shortestPath.Path)]
	for i, path := range rootPaths {
		printTree(path, "", i == len(rootPaths)-1)
	}
	return output
}

// WriteOutput writes content to a file or stdout
func WriteOutput(path string, content string) error {
	if path == "stdout" || path == "-" {
		_, err := fmt.Print(content)
		if err != nil {
			return fmt.Errorf("writing to stdout: %w", err)
		}
		return nil
	}

	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("creating directories along path %q: %w", dir, err)
	}

	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}
	return nil
}

// FormatTimestamp returns a formatted timestamp string
func FormatTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
