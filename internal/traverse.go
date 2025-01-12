package internal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

// PathInfo represents information about a file or directory
type PathInfo struct {
	Path         string
	RelativePath string
	Depth        int
	IsDir        bool
}

// TraverseDirectories traverses directories and collects path information
func TraverseDirectories(directories []string, includePatterns []string, excludePatterns []string) ([]PathInfo, error) {
	includeMatchers, err := createPatternMatchers(includePatterns)
	if err != nil {
		return nil, fmt.Errorf("creating include pattern matchers: %w", err)
	}

	excludeMatchers, err := createPatternMatchers(excludePatterns)
	if err != nil {
		return nil, fmt.Errorf("creating exclude pattern matchers: %w", err)
	}

	var paths []PathInfo
	for _, dir := range directories {
		basePath, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}

		baseInfo, err := os.Stat(basePath)
		if err != nil {
			return nil, err
		}

		if !baseInfo.IsDir() {
			basePath = filepath.Dir(basePath)
			baseInfo, err = os.Stat(basePath)
			if err != nil {
				return nil, err
			}
			if !baseInfo.IsDir() {
				return nil, fmt.Errorf("expected base path %q to be a directory", basePath)
			}
		}

		// Add base directory if it matches patterns
		baseRelPath := filepath.Base(basePath)
		if shouldIncludePath(baseRelPath, true, includeMatchers, excludeMatchers) {
			paths = append(paths, PathInfo{
				Path:         basePath,
				RelativePath: baseRelPath,
				Depth:        1,
				IsDir:        true,
			})
		}

		// Walk the directory tree
		err = filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Skip the root directory as it's already processed
			if path == basePath {
				return nil
			}

			relPath, err := filepath.Rel(basePath, path)
			if err != nil {
				return err
			}

			// Convert to forward slashes for consistent pattern matching
			relPath = filepath.ToSlash(relPath)
			isDir := d.IsDir()

			// Check if path should be included based on patterns
			if !shouldIncludePath(relPath, isDir, includeMatchers, excludeMatchers) {
				if isDir {
					return filepath.SkipDir
				}
				return nil
			}

			depth := strings.Count(relPath, "/") + 1

			paths = append(paths, PathInfo{
				Path:         path,
				RelativePath: relPath,
				Depth:        depth,
				IsDir:        isDir,
			})

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return paths, nil
}

// shouldIncludePath determines if a path should be included based on the patterns
func shouldIncludePath(path string, isDir bool, includeMatchers, excludeMatchers []glob.Glob) bool {
	paths := []string{path}

	// If the path doesn't contain a directory separator, also try matching it
	// with an implicit ./ prefix to handle root-level files with **/ patterns
	if !strings.Contains(path, "/") {
		paths = append(paths, "./"+path)
	}

	// Append trailing slash for directories to match directory-specific patterns
	if isDir {
		path += "/"
	}

	for _, matcher := range excludeMatchers {
		for _, p := range paths {
			if matcher.Match(p) {
				return false
			}
		}
	}

	for _, matcher := range includeMatchers {
		for _, p := range paths {
			if matcher.Match(p) {
				return true
			}
		}
	}
	return false
}

func createPatternMatchers(patterns []string) ([]glob.Glob, error) {
	matchers := make([]glob.Glob, len(patterns))
	for i, pattern := range patterns {
		g, err := glob.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern '%s': %w", pattern, err)
		}
		matchers[i] = g
	}
	return matchers, nil
}
