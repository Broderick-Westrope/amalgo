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

	// matchReason tracks why this path was included
	matchReason matchReason
}

type matchReason int

const (
	noMatch         matchReason = iota
	directMatch                 // Path matches pattern directly
	potentialParent             // Directory that might contain matches
)

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

	var confirmedPaths []PathInfo
	var potentialPaths []PathInfo
	directoryHasMatches := make(map[string]struct{})

	for _, dir := range directories {
		basePath, err := filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("getting base path for directory %q: %w", dir, err)
		}

		baseInfo, err := os.Stat(basePath)
		if err != nil {
			return nil, fmt.Errorf("describing base path for directory %q: %w", dir, err)
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
		if include, reason := shouldIncludePath(baseRelPath, true, includeMatchers, excludeMatchers); include {
			confirmedPaths = append(confirmedPaths, PathInfo{
				Path:         basePath,
				RelativePath: baseRelPath,
				Depth:        1,
				IsDir:        true,
				matchReason:  reason,
			})
		}

		// Walk the directory tree
		err = filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("at path %q: %w", path, err)
			}

			// Skip the root directory as it's already processed
			if path == basePath {
				return nil
			}

			relPath, err := filepath.Rel(basePath, path)
			if err != nil {
				return fmt.Errorf("getting relative path between %q and %q: %w", basePath, path, err)
			}

			// Convert to forward slashes for consistent pattern matching
			relPath = filepath.ToSlash(relPath)
			isDir := d.IsDir()

			// Check if path should be included based on patterns
			include, reason := shouldIncludePath(relPath, isDir, includeMatchers, excludeMatchers)
			if !include {
				if isDir {
					return filepath.SkipDir
				}
				return nil
			}

			depth := strings.Count(relPath, "/") + 1
			info := PathInfo{
				Path:         path,
				RelativePath: relPath,
				Depth:        depth,
				IsDir:        isDir,
				matchReason:  reason,
			}

			if !isDir || reason == directMatch {
				confirmedPaths = append(confirmedPaths, info)

				dir := filepath.Dir(info.RelativePath)
				for dir != "." && dir != "/" {
					directoryHasMatches[dir] = struct{}{}
					dir = filepath.Dir(dir)
				}
				directoryHasMatches["."] = struct{}{}
			} else {
				potentialPaths = append(potentialPaths, info)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walking directory %q: %w", basePath, err)
		}
	}

	if len(confirmedPaths) > 0 {
		// Add directories that contain matches
		for _, p := range potentialPaths {
			if p.IsDir && p.matchReason == potentialParent {
				if _, found := directoryHasMatches[p.RelativePath]; found {
					confirmedPaths = append(confirmedPaths, p)
				}
			}
		}
	}
	return confirmedPaths, nil
}

// shouldIncludePath determines if a path should be included based on the patterns
func shouldIncludePath(path string, isDir bool, includeMatchers, excludeMatchers []patternMatcher) (bool, matchReason) {
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
				return false, noMatch
			}
		}
	}

	for _, matcher := range includeMatchers {
		for _, p := range paths {
			if matcher.Match(p) {
				return true, directMatch
			}
		}
	}

	// Even if directory doesn't match directly, check if it could contain matching files
	isMatch := isDir && couldContainMatches(path, includeMatchers)
	if isMatch {
		return isMatch, potentialParent
	}
	return false, noMatch
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

func createPatternMatchers(patterns []string) ([]patternMatcher, error) {
	matchers := make([]patternMatcher, len(patterns))
	for i, pattern := range patterns {
		g, err := glob.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("compiling pattern %q: %w", pattern, err)
		}
		matchers[i] = patternMatcher{
			pattern: pattern,
			matcher: g,
		}
	}
	return matchers, nil
}
