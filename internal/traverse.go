package internal

import (
	"errors"
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

// TraverseDirectory traverses the directory and collects path information using the filter package
func TraverseDirectory(dir string, filterPatterns, gitignorePaths []string) ([]PathInfo, error) {
	// Create the filter from filter patterns.
	f := filter.CompileFilterPatterns(filterPatterns...)
	// Create the gitignore filter from gitignore file paths.
	gi := new(filter.Filter)
	for _, giPath := range gitignorePaths {
		tempFilter, err := filter.CompileFilterPatternFile(giPath)
		if err != nil {
			return nil, fmt.Errorf("compiling patterns from gitignore file %q: %w", giPath, err)
		}
		gi.MergeWithPrecedence(tempFilter)
	}

	paths := make([]PathInfo, 0)
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

	// Base parent allows getting the relative path in relation to the parent.
	baseParent := filepath.Dir(basePath)

	err = filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("at path %q: %w", path, err)
		}

		// Skip directories as the filter system is built to process file paths.
		// Skip the base path as it's already processed.
		if d.IsDir() || path == basePath {
			return nil
		}

		relPath, err := filepath.Rel(basePath, path)
		if err != nil {
			return fmt.Errorf("getting relative path between %q and %q: %w", basePath, path, err)
		}
		relPath = filepath.ToSlash(relPath)

		if gi.MatchesPath(relPath) || !f.MatchesPath(relPath) {
			return nil
		}

		relPath, err = filepath.Rel(baseParent, path)
		if err != nil {
			return fmt.Errorf("getting relative path between %q and %q: %w", basePath, path, err)
		}
		relPath = filepath.ToSlash(relPath)

		paths = append(paths, PathInfo{
			Path:         path,
			RelativePath: relPath,
			Depth:        strings.Count(relPath, "/") + 1,
			IsDir:        false,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking directory %q: %w", basePath, err)
	}

	err = processPaths(&paths)
	if err != nil {
		return nil, err
	}
	return paths, nil
}

// ProcessPaths adds all parent directory paths to the given slice of PathInfo.
// The function modifies the input slice in place, adding parent directories in order
// from shallowest to deepest, followed by the original paths.
func processPaths(paths *[]PathInfo) error {
	if paths == nil {
		return errors.New("paths must be a pointer to a slice")
	} else if *paths == nil {
		return errors.New("underlying paths slice cannot be nil")
	}

	// Create a map to deduplicate paths.
	seen := make(map[string]struct{})
	for _, path := range *paths {
		seen[path.Path] = struct{}{}
	}

	result := make([]PathInfo, len(*paths))
	if copy(result, *paths) != len(*paths) {
		return errors.New("failed to copy paths to result slice")
	}

	for _, p := range *paths {
		// Split the relative path to process each component.
		components := strings.Split(p.RelativePath, "/")
		basePath := filepath.Dir(p.Path[:len(p.Path)-len(p.RelativePath)])

		// Process each level of the path.
		currentRel := ""
		currentAbs := basePath
		for i, comp := range components {
			if i == len(components)-1 && !p.IsDir {
				// Skip the last component if it's a file - we'll add it from the original slice.
				continue
			}

			if currentRel == "" {
				currentRel = comp
			} else {
				currentRel = filepath.Join(currentRel, comp)
			}
			currentAbs = filepath.Join(currentAbs, comp)

			// Only add if we haven't seen this path before.
			if _, exists := seen[currentAbs]; !exists {
				seen[currentAbs] = struct{}{}
				result = append(result, PathInfo{
					Path:         currentAbs,
					RelativePath: currentRel,
					Depth:        i + 1,
					IsDir:        true,
				})
			}
		}
	}

	// Update the input slice with the result.
	*paths = result
	return nil
}
