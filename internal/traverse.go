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

// patternMatcher pairs a glob matcher with its original pattern string
type patternMatcher struct {
	pattern string
	matcher glob.Glob
}

func (pm *patternMatcher) Match(value string) bool {
	return pm.matcher.Match(value)
}

func (pm *patternMatcher) String() string {
	return pm.pattern
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

	return filterValidPaths(paths, includeMatchers), nil
}

// shouldIncludePath determines if a path should be included based on the patterns
func shouldIncludePath(path string, isDir bool, includeMatchers, excludeMatchers []patternMatcher) bool {
	// Append trailing slash for directories to match directory-specific patterns
	if isDir {
		path += "/"
	}
	paths := []string{path}

	// If the path doesn't contain a directory separator, also try matching it
	// with an implicit ./ prefix to handle root-level files with **/ patterns
	if !strings.Contains(path, "/") {
		paths = append(paths, "./"+path)
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

	// Even if directory doesn't match directly, check if it could contain matching files
	return isDir && couldContainMatches(path, includeMatchers)
}

// couldContainMatches checks if a directory could potentially contain files that match the patterns
func couldContainMatches(dirPath string, includeMatchers []patternMatcher) bool {
	dirPath = strings.TrimSuffix(dirPath, "/")

	// If the directory path itself matches any pattern, it could contain matches
	for _, matcher := range includeMatchers {
		pattern := matcher.String()

		// If pattern is just checking file extension or name (no directories),
		// then any directory could contain matching files
		if !strings.Contains(pattern, "/") {
			return true
		}

		// If pattern uses **, any directory could contain matches
		if strings.Contains(pattern, "**") {
			return true
		}

		// For patterns with explicit directory structure, check if this directory
		// is a potential parent of matching files
		parts := strings.Split(pattern, "/")
		dirParts := strings.Split(dirPath, "/")

		// If directory path is shorter than pattern, it could contain matches
		if len(dirParts) < len(parts) {
			prefixMatches := true
			for i := range dirParts {
				if parts[i] != dirParts[i] && parts[i] != "*" {
					prefixMatches = false
					break
				}
			}
			if prefixMatches {
				return true
			}
		}
	}
	return false
}

// filterValidPaths removes directories that don't match patterns and don't contain matching files
func filterValidPaths(paths []PathInfo, includeMatchers []patternMatcher) []PathInfo {
	// First, identify all directories that contain matching files
	hasMatchingFile := make(map[string]bool)

	// Process in reverse order so we handle deeper paths first
	for i := len(paths) - 1; i >= 0; i-- {
		path := paths[i]

		if !path.IsDir {
			// If it's a file and it's in our paths slice, it must have matched
			dirPath := filepath.Dir(path.RelativePath)
			for dirPath != "." && dirPath != "/" {
				hasMatchingFile[dirPath] = true
				dirPath = filepath.Dir(dirPath)
			}
			hasMatchingFile["."] = true // Root directory
		}
	}

	// Filter the paths
	validPaths := make([]PathInfo, 0, len(paths))
	for _, path := range paths {
		if !path.IsDir {
			// Files in our slice have already been verified to match
			validPaths = append(validPaths, path)
			continue
		}

		// For directories, check if they match directly or contain matching files
		dirPath := path.RelativePath
		if dirPath == "." {
			dirPath = ""
		}

		matches := false
		// Check if directory matches patterns directly
		for _, matcher := range includeMatchers {
			if matcher.Match(dirPath + "/") {
				matches = true
				break
			}
		}

		// If directory doesn't match directly, check if it contains matching files
		if !matches && !hasMatchingFile[dirPath] {
			continue
		}

		validPaths = append(validPaths, path)
	}

	return validPaths
}

func createPatternMatchers(patterns []string) ([]patternMatcher, error) {
	matchers := make([]patternMatcher, len(patterns))
	for i, pattern := range patterns {
		g, err := glob.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern '%s': %w", pattern, err)
		}
		matchers[i] = patternMatcher{
			pattern: pattern,
			matcher: g,
		}
	}
	return matchers, nil
}
