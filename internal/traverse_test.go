package internal

import (
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
		want           bool
	}{
		"root go file with no patterns": {
			path:           "main.go",
			isDir:          false,
			includePattern: "",
			excludePattern: "",
			want:           false,
		},
		"nested go file with no patterns": {
			path:           "internal/parser/parser.go",
			isDir:          false,
			includePattern: "",
			excludePattern: "",
			want:           false,
		},

		"root go file with simple pattern": {
			path:           "main.go",
			isDir:          false,
			includePattern: "*.go",
			excludePattern: "",
			want:           true,
		},
		"nested go file with simple pattern": {
			path:           "internal/parser/parser.go",
			isDir:          false,
			includePattern: "*.go",
			excludePattern: "",
			want:           true,
		},

		"root go file with double-star pattern": {
			path:           "main.go",
			isDir:          false,
			includePattern: "**/*.go",
			excludePattern: "",
			want:           true,
		},
		"nested go file with double-star pattern": {
			path:           "internal/parser/parser.go",
			isDir:          false,
			includePattern: "**/*.go",
			excludePattern: "",
			want:           true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			includeMatchers, err := createPatternMatchers([]string{tt.includePattern})
			require.NoError(t, err)

			excludeMatchers, err := createPatternMatchers([]string{tt.excludePattern})
			require.NoError(t, err)

			got := shouldIncludePath(tt.path, tt.isDir, includeMatchers, excludeMatchers)
			assert.Equal(t, tt.want, got)
		})
	}
}
