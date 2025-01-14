package internal

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTraverseDirectories(t *testing.T) {
	// Create a temporary directory structure for testing.
	tmpDir := t.TempDir()

	// Create test directory structure. Bool indicates if it's a directory.
	testFiles := map[string]bool{
		"src":                   true,
		"src/main.go":           false,
		"src/README.md":         false,
		"src/internal":          true,
		"src/internal/util.go":  false,
		"src/internal/test.txt": false,
		"vendor":                true,
		"vendor/lib.go":         false,
	}

	// Create the test files and directories.
	for path, isDir := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		if isDir {
			require.NoError(t, os.MkdirAll(fullPath, 0755))
		} else {
			require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
			require.NoError(t, os.WriteFile(fullPath, []byte("test content"), 0644))
		}
	}

	tests := map[string]struct {
		directory      string
		filterPatterns []string
		wantRelPaths   []string
		wantErr        bool
	}{
		"match go files in top directory": {
			directory:      filepath.Join(tmpDir, "src"),
			filterPatterns: []string{"*.go"},
			wantRelPaths: []string{
				"src/main.go",
			},
		},
		"match all go files": {
			directory:      filepath.Join(tmpDir, "src"),
			filterPatterns: []string{"**/*.go"},
			wantRelPaths: []string{
				"src/main.go",
				"src/internal/util.go",
			},
		},
		"exclude directory": {
			directory:      filepath.Join(tmpDir, "src"),
			filterPatterns: []string{"*.go", "**/*.go", "!internal/**"},
			wantRelPaths: []string{
				"src/main.go",
			},
		},
		"match specific directory": {
			directory:      filepath.Join(tmpDir, "src", "internal"),
			filterPatterns: []string{"*"},
			wantRelPaths: []string{
				"internal/util.go",
				"internal/test.txt",
			},
		},
		"non-existent directory": {
			directory:      filepath.Join(tmpDir, "nonexistent"),
			filterPatterns: []string{"**/*.go"},
			wantErr:        true,
		},
		"file as directory": {
			directory:      filepath.Join(tmpDir, "src", "main.go"),
			filterPatterns: []string{"*.go"},
			wantRelPaths: []string{
				"src/main.go",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			paths, err := TraverseDirectory(tt.directory, tt.filterPatterns)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Convert PathInfo slice to relative paths for easier comparison.
			var gotPaths []string
			for _, p := range paths {
				if !p.IsDir { // Only include files in our comparison
					gotPaths = append(gotPaths, p.RelativePath)
				}
			}
			assert.ElementsMatch(t, tt.wantRelPaths, gotPaths)

			// Additional validation of PathInfo fields.
			for _, p := range paths {
				// Paths should be absolute.
				assert.True(t, filepath.IsAbs(p.Path), "Path should be absolute: %s", p.Path)

				// RelativePath should not be absolute.
				assert.False(t, filepath.IsAbs(p.RelativePath), "RelativePath should be relative: %s", p.RelativePath)

				// Depth should match the number of path separators plus one.
				expectedDepth := 0
				if p.RelativePath != "" {
					expectedDepth = len(strings.Split(p.RelativePath, "/"))
				}
				assert.Equal(t, expectedDepth, p.Depth, "Incorrect depth for path: %s", p.RelativePath)
			}
		})
	}
}

func TestProcessPaths(t *testing.T) {
	var nilPathInfoSlice []PathInfo = nil
	tests := map[string]struct {
		paths     *[]PathInfo
		wantPaths []PathInfo
		wantErr   error
	}{
		"single file adds parent dirs": {
			paths: &[]PathInfo{
				{
					Path:         "/Users/someuser/dev/program/internal/file1.go",
					RelativePath: "program/internal/file1.go",
					Depth:        3,
					IsDir:        false,
				},
			},
			wantPaths: []PathInfo{
				{
					Path:         "/Users/someuser/dev/program/internal/file1.go",
					RelativePath: "program/internal/file1.go",
					Depth:        3,
					IsDir:        false,
				},
				{
					Path:         "/Users/someuser/dev/program",
					RelativePath: "program",
					Depth:        1,
					IsDir:        true,
				},
				{
					Path:         "/Users/someuser/dev/program/internal",
					RelativePath: "program/internal",
					Depth:        2,
					IsDir:        true,
				},
			},
		},
		"multiple files same directory": {
			paths: &[]PathInfo{
				{
					Path:         "/Users/someuser/dev/program/internal/file1.go",
					RelativePath: "program/internal/file1.go",
					Depth:        3,
					IsDir:        false,
				},
				{
					Path:         "/Users/someuser/dev/program/internal/file2.go",
					RelativePath: "program/internal/file2.go",
					Depth:        3,
					IsDir:        false,
				},
			},
			wantPaths: []PathInfo{
				{
					Path:         "/Users/someuser/dev/program/internal/file1.go",
					RelativePath: "program/internal/file1.go",
					Depth:        3,
					IsDir:        false,
				},
				{
					Path:         "/Users/someuser/dev/program/internal/file2.go",
					RelativePath: "program/internal/file2.go",
					Depth:        3,
					IsDir:        false,
				},
				{
					Path:         "/Users/someuser/dev/program",
					RelativePath: "program",
					Depth:        1,
					IsDir:        true,
				},
				{
					Path:         "/Users/someuser/dev/program/internal",
					RelativePath: "program/internal",
					Depth:        2,
					IsDir:        true,
				},
			},
		},
		"different directory depths": {
			paths: &[]PathInfo{
				{
					Path:         "/Users/someuser/dev/program/file1.go",
					RelativePath: "program/file1.go",
					Depth:        2,
					IsDir:        false,
				},
				{
					Path:         "/Users/someuser/dev/program/internal/deep/file2.go",
					RelativePath: "program/internal/deep/file2.go",
					Depth:        4,
					IsDir:        false,
				},
			},
			wantPaths: []PathInfo{
				{
					Path:         "/Users/someuser/dev/program/file1.go",
					RelativePath: "program/file1.go",
					Depth:        2,
					IsDir:        false,
				},
				{
					Path:         "/Users/someuser/dev/program/internal/deep/file2.go",
					RelativePath: "program/internal/deep/file2.go",
					Depth:        4,
					IsDir:        false,
				},
				{
					Path:         "/Users/someuser/dev/program",
					RelativePath: "program",
					Depth:        1,
					IsDir:        true,
				},
				{
					Path:         "/Users/someuser/dev/program/internal",
					RelativePath: "program/internal",
					Depth:        2,
					IsDir:        true,
				},
				{
					Path:         "/Users/someuser/dev/program/internal/deep",
					RelativePath: "program/internal/deep",
					Depth:        3,
					IsDir:        true,
				},
			},
		},
		"directory included": {
			paths: &[]PathInfo{
				{
					Path:         "/Users/someuser/dev/program/internal",
					RelativePath: "program/internal",
					Depth:        2,
					IsDir:        true,
				},
				{
					Path:         "/Users/someuser/dev/program/internal/file1.go",
					RelativePath: "program/internal/file1.go",
					Depth:        3,
					IsDir:        false,
				},
			},
			wantPaths: []PathInfo{
				{
					Path:         "/Users/someuser/dev/program/internal",
					RelativePath: "program/internal",
					Depth:        2,
					IsDir:        true,
				},
				{
					Path:         "/Users/someuser/dev/program/internal/file1.go",
					RelativePath: "program/internal/file1.go",
					Depth:        3,
					IsDir:        false,
				},
				{
					Path:         "/Users/someuser/dev/program",
					RelativePath: "program",
					Depth:        1,
					IsDir:        true,
				},
			},
		},
		"empty slice": {
			paths:     &[]PathInfo{},
			wantPaths: []PathInfo{},
		},
		"nil slice": {
			paths:     &nilPathInfoSlice,
			wantPaths: []PathInfo{},
			wantErr:   errors.New("underlying paths slice cannot be nil"),
		},
		"nil pointer": {
			paths:     nil,
			wantPaths: []PathInfo{},
			wantErr:   errors.New("paths must be a pointer to a slice"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := processPaths(tt.paths)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.wantPaths, *tt.paths)
		})
	}
}
