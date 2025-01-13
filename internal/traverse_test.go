package internal

import (
	"os"
	"path/filepath"
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
		directories    []string
		filterPatterns []string
		wantRelPaths   []string
		wantErr        bool
	}{
		"match go files in top directory": {
			directories:    []string{filepath.Join(tmpDir, "src")},
			filterPatterns: []string{"*.go"},
			wantRelPaths: []string{
				"main.go",
			},
		},
		"exclude directory": {
			directories:    []string{filepath.Join(tmpDir, "src")},
			filterPatterns: []string{"*.go", "**/*.go", "!internal/**"},
			wantRelPaths: []string{
				"main.go",
			},
		},
		"match specific directory": {
			directories:    []string{filepath.Join(tmpDir, "src", "internal")},
			filterPatterns: []string{"*"},
			wantRelPaths: []string{
				"util.go",
				"test.txt",
			},
		},
		"non-existent directory": {
			directories:    []string{filepath.Join(tmpDir, "nonexistent")},
			filterPatterns: []string{"**/*.go"},
			wantErr:        true,
		},
		"file as directory": {
			directories:    []string{filepath.Join(tmpDir, "src", "main.go")},
			filterPatterns: []string{"*.go"},
			wantRelPaths: []string{
				"main.go",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			paths, err := TraverseDirectories(tt.directories, tt.filterPatterns)

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
				expectedDepth := 1
				if p.RelativePath != "" {
					expectedDepth = len(filepath.SplitList(filepath.FromSlash(p.RelativePath)))
				}
				assert.Equal(t, expectedDepth, p.Depth, "Incorrect depth for path: %s", p.RelativePath)
			}
		})
	}
}
