package traverse

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

// Options configures the traversal behavior
type Options struct {
	IncludePatterns []string
	ExcludePatterns []string
	IncludeHidden   bool
	IgnoreDirs      []string
}

// GetPaths traverses directories and collects path information
func GetPaths(directories []string, opts Options) ([]PathInfo, error) {
	if len(directories) == 0 {
		directories = []string{"."}
	}

	includeMatchers, err := createPatternMatchers(opts.IncludePatterns)
	if err != nil {
		return nil, fmt.Errorf("creating include pattern matchers")
	}

	excludeMatchers, err := createPatternMatchers(opts.ExcludePatterns)
	if err != nil {
		return nil, fmt.Errorf("creating exclude pattern matchers")
	}

	var paths []PathInfo
	for _, dir := range directories {
		basePath, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}

		parentPath := filepath.Dir(basePath)
		baseInfo, err := os.Stat(basePath)
		if err != nil {
			return nil, err
		}

		if !baseInfo.IsDir() {
			continue
		}

		// Add base directory
		paths = append(paths, PathInfo{
			Path:         basePath,
			RelativePath: filepath.ToSlash(filepath.Join(filepath.Base(parentPath), filepath.Base(basePath))),
			Depth:        1,
			IsDir:        true,
		})

		// Walk the directory tree
		err = filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Skip the root directory as it's already added
			if path == basePath {
				return nil
			}

			// Check if path should be ignored
			for _, ignoreDir := range opts.IgnoreDirs {
				ignorePath, err := filepath.Abs(ignoreDir)
				if err != nil {
					continue
				}
				if strings.HasPrefix(path, ignorePath) {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}

			// Handle hidden files/directories
			if !opts.IncludeHidden && strings.HasPrefix(filepath.Base(path), ".") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			isDir := d.IsDir()
			if !isDir {
				matched := false
				_, name := filepath.Split(path)
				for _, matcher := range excludeMatchers {
					if matcher.Match(name) {
						matched = true
						break
					}
				}
				if matched {
					return nil
				}

				matched = false
				_, name = filepath.Split(path)
				for _, matcher := range includeMatchers {
					if matcher.Match(name) {
						matched = true
						break
					}
				}
				if !matched {
					return nil
				}
			}

			relPath, err := filepath.Rel(parentPath, path)
			if err != nil {
				return err
			}

			depth := strings.Count(relPath, string(os.PathSeparator)) + 1

			paths = append(paths, PathInfo{
				Path:         path,
				RelativePath: filepath.ToSlash(relPath),
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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func createPatternMatchers(patterns []string) ([]glob.Glob, error) {
	matchers := make([]glob.Glob, len(patterns))
	for i, pattern := range patterns {
		g, err := glob.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("compiling glob pattern '%s': %w", pattern, err)
		}
		matchers[i] = g
	}
	return matchers, nil
}
