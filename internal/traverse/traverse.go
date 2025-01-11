package traverse

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
	Extensions    []string
	IncludeAll    bool
	IncludeHidden bool
	IgnoreDirs    []string
}

// GetPaths traverses directories and collects path information
func GetPaths(directories []string, opts Options) ([]PathInfo, error) {
	if len(directories) == 0 {
		directories = []string{"."}
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
				if strings.HasPrefix(path, ignoreDir) {
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
			if !isDir && !opts.IncludeAll {
				ext := strings.ToLower(filepath.Ext(path))
				if !contains(opts.Extensions, ext) {
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
