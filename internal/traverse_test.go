package internal

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldIncludePath(t *testing.T) {
	tests := map[string]struct {
		path           string
		isDir          bool
		includePattern string
		excludePattern string
		wantInclude    bool
		wantReason     matchReason
	}{
		// Original test cases
		"root go file with no patterns": {
			path:           "main.go",
			isDir:          false,
			includePattern: "",
			excludePattern: "",
			wantInclude:    false,
			wantReason:     noMatch,
		},
		"nested go file with no patterns": {
			path:           "internal/parser/parser.go",
			isDir:          false,
			includePattern: "",
			excludePattern: "",
			wantInclude:    false,
			wantReason:     noMatch,
		},
		"root go file with simple pattern": {
			path:           "main.go",
			isDir:          false,
			includePattern: "*.go",
			excludePattern: "",
			wantInclude:    true,
			wantReason:     directMatch,
		},
		"nested go file with simple pattern": {
			path:           "internal/parser/parser.go",
			isDir:          false,
			includePattern: "*.go",
			excludePattern: "",
			wantInclude:    true,
			wantReason:     directMatch,
		},
		"root go file with double-star pattern": {
			path:           "main.go",
			isDir:          false,
			includePattern: "**/*.go",
			excludePattern: "",
			wantInclude:    true,
			wantReason:     directMatch,
		},
		"nested go file with double-star pattern": {
			path:           "internal/parser/parser.go",
			isDir:          false,
			includePattern: "**/*.go",
			excludePattern: "",
			wantInclude:    true,
			wantReason:     directMatch,
		},

		// New test cases for directory handling
		"directory with potential go files": {
			path:           "src",
			isDir:          true,
			includePattern: "**/*.go",
			excludePattern: "",
			wantInclude:    true,
			wantReason:     potentialParent,
		},
		"nested directory with potential go files": {
			path:           "src/utils",
			isDir:          true,
			includePattern: "**/*.go",
			excludePattern: "",
			wantInclude:    true,
			wantReason:     potentialParent,
		},
		"excluded directory": {
			path:           "vendor",
			isDir:          true,
			includePattern: "**/*.go",
			excludePattern: "vendor/**",
			wantInclude:    false,
			wantReason:     noMatch,
		},
		"nested excluded directory": {
			path:           "vendor/pkg",
			isDir:          true,
			includePattern: "**/*.go",
			excludePattern: "vendor/**",
			wantInclude:    false,
			wantReason:     noMatch,
		},
		"file in excluded directory": {
			path:           "vendor/pkg/main.go",
			isDir:          false,
			includePattern: "**/*.go",
			excludePattern: "vendor/**",
			wantInclude:    false,
			wantReason:     noMatch,
		},

		// Tests for specific directory patterns
		"directory matching specific pattern": {
			path:           "internal",
			isDir:          true,
			includePattern: "internal/*.go",
			excludePattern: "",
			wantInclude:    true,
			wantReason:     potentialParent,
		},
		"non-matching directory with specific pattern": {
			path:           "pkg",
			isDir:          true,
			includePattern: "internal/*.go",
			excludePattern: "",
			wantInclude:    false,
			wantReason:     noMatch,
		},

		// Tests for multiple patterns
		"file matching one of multiple patterns": {
			path:           "main.go",
			isDir:          false,
			includePattern: "*.txt,*.go",
			excludePattern: "",
			wantInclude:    true,
			wantReason:     directMatch,
		},
		"file not matching any of multiple patterns": {
			path:           "main.rs",
			isDir:          false,
			includePattern: "*.txt,*.go",
			excludePattern: "",
			wantInclude:    false,
			wantReason:     noMatch,
		},

		// Edge cases
		"root directory with double-star": {
			path:           ".",
			isDir:          true,
			includePattern: "**/*.go",
			excludePattern: "",
			wantInclude:    true,
			wantReason:     potentialParent,
		},
		"empty path": {
			path:           "",
			isDir:          true,
			includePattern: "**/*.go",
			excludePattern: "",
			wantInclude:    true,
			wantReason:     potentialParent,
		},
		"path with special characters": {
			path:           "test[1].go",
			isDir:          false,
			includePattern: "*.go",
			excludePattern: "",
			wantInclude:    true,
			wantReason:     directMatch,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var includeMatchers []patternMatcher
			var excludeMatchers []patternMatcher
			var err error

			if tt.includePattern != "" {
				includeMatchers, err = createPatternMatchers(strings.Split(tt.includePattern, ","))
				require.NoError(t, err)
			}
			if tt.excludePattern != "" {
				excludeMatchers, err = createPatternMatchers(strings.Split(tt.excludePattern, ","))
				require.NoError(t, err)
			}

			gotInclude, gotReason := shouldIncludePath(tt.path, tt.isDir, includeMatchers, excludeMatchers)
			assert.Equal(t, tt.wantInclude, gotInclude, "path: %s, include: %s, exclude: %s",
				tt.path, tt.includePattern, tt.excludePattern)
			assert.Equal(t, tt.wantReason, gotReason, "path: %s, include: %s, exclude: %s",
				tt.path, tt.includePattern, tt.excludePattern)
		})
	}
}

func TestCouldContainMatches(t *testing.T) {
	tests := map[string]struct {
		dirPath string
		pattern string
		want    bool
	}{
		"simple top-level pattern": {
			dirPath: "internal",
			pattern: "internal/*.go",
			want:    true,
		},
		"nested directory pattern": {
			dirPath: "src/internal",
			pattern: "src/internal/*.go",
			want:    true,
		},
		"non-matching directory": {
			dirPath: "src/other",
			pattern: "src/internal/*.go",
			want:    false,
		},
		"wildcard in directory": {
			dirPath: "src/v1/internal",
			pattern: "src/*/internal/*.go",
			want:    true,
		},
		"partial directory match": {
			dirPath: "src",
			pattern: "src/internal/*.go",
			want:    true,
		},
		"exact file pattern": {
			dirPath: "src",
			pattern: "src/main.go",
			want:    true,
		},
		"deep directory structure": {
			dirPath: "src/v1/internal/pkg",
			pattern: "src/**/internal/**/*.go",
			want:    true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			matchers, err := createPatternMatchers([]string{tt.pattern})
			require.NoError(t, err)

			got := couldContainMatches(tt.dirPath, matchers)
			assert.Equal(t, tt.want, got,
				"directory: %s, pattern: %s", tt.dirPath, tt.pattern)
		})
	}
}
