package internal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Broderick-Westrope/amalgo/pkg/filter"
)

// PathInfo represents information about a file or directory
type PathInfo struct {
	Path         string
	RelativePath string
	Depth        int
	IsDir        bool
}

// TraverseDirectories traverses directories and collects path information using the filter package
func TraverseDirectories(directories []string, filterPatterns []string) ([]PathInfo, error) {
	// Create the filterer from patterns
	f := filter.CompileFilterPatterns(filterPatterns...)

	var paths []PathInfo
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
		if f.MatchesPath(baseRelPath) {
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

			// Check if path should be included based on patterns
			if !f.MatchesPath(relPath) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			depth := strings.Count(relPath, "/") + 1
			paths = append(paths, PathInfo{
				Path:         path,
				RelativePath: relPath,
				Depth:        depth,
				IsDir:        d.IsDir(),
			})
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walking directory %q: %w", basePath, err)
		}
	}
	return paths, nil
}
