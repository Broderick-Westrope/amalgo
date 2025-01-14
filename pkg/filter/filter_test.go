package filter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchesPath(t *testing.T) {
	tests := map[string]struct {
		patterns []string
		path     string
		want     bool
	}{
		"double asterisk at start": {
			patterns: []string{"**/test.txt"},
			path:     "deep/nested/test.txt",
			want:     true,
		},
		"double asterisk in middle": {
			patterns: []string{"src/**/test.txt"},
			path:     "src/deeply/nested/test.txt",
			want:     true,
		},
		"double asterisk at end": {
			patterns: []string{"src/**"},
			path:     "src/any/number/of/subdirs",
			want:     true,
		},
		"leading slash": {
			patterns: []string{"/root.txt"},
			path:     "root.txt",
			want:     true,
		},
		"single character wildcard": {
			patterns: []string{"test?.txt"},
			path:     "test1.txt",
			want:     true,
		},
		"actual dots in filenames": {
			patterns: []string{"*.txt"},
			path:     "file.txt",
			want:     true,
		},
		"multiple extensions separate patterns": {
			patterns: []string{"*.txt", "*.md", "*.json"},
			path:     "file.json",
			want:     true,
		},
		"complex nesting with negation": {
			patterns: []string{
				"**/*.go",
				"!vendor/**",
				"vendor/allowed/*.go",
			},
			path: "vendor/forbidden/file.go",
			want: false,
		},
		"multiple patterns with precedence": {
			patterns: []string{
				"*.txt",
				"!important/*.txt",
				"important/keepthis.txt",
			},
			path: "important/keepthis.txt",
			want: true,
		},
		"spaces in pattern": {
			patterns: []string{" *.txt ", "  ", " # comment "},
			path:     "file.txt",
			want:     true,
		},
		"carriage return handling": {
			patterns: []string{"*.txt\r", "*.md\r"},
			path:     "file.txt",
			want:     true,
		},
		"simple match": {
			patterns: []string{"*.txt"},
			path:     "file.txt",
			want:     true,
		},
		"no match": {
			patterns: []string{"*.txt"},
			path:     "file.go",
			want:     false,
		},
		"negated pattern": {
			patterns: []string{"*.txt", "!test.txt"},
			path:     "test.txt",
			want:     false,
		},
		"directory match": {
			patterns: []string{"src/**/*.go"},
			path:     "src/pkg/file.go",
			want:     true,
		},
		"escaped characters": {
			patterns: []string{"\\#file.txt"},
			path:     "#file.txt",
			want:     true,
		},
		"multiple patterns with override": {
			patterns: []string{"*.txt", "!test.txt", "test.txt"},
			path:     "test.txt",
			want:     true,
		},
		"directory trailing slash": {
			patterns: []string{"logs/"},
			path:     "logs/debug.log",
			want:     true,
		},
		"comment and empty lines": {
			patterns: []string{"", "# comment", "*.txt"},
			path:     "file.txt",
			want:     true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			f := CompileFilterPatterns(tc.patterns...)
			got := f.MatchesPath(tc.path)
			assert.Equal(t, got, tc.want, "path: %q, patterns: %v", tc.path, tc.patterns)
		})
	}
}

func TestMatchesPathHow(t *testing.T) {
	tests := map[string]struct {
		patterns        []string
		path            string
		wantMatch       bool
		wantMatchLine   string
		wantMatchNegate bool
	}{
		"matches first pattern": {
			patterns:        []string{"*.txt", "*.go"},
			path:            "file.txt",
			wantMatch:       true,
			wantMatchLine:   "*.txt",
			wantMatchNegate: false,
		},
		"matches negated pattern": {
			patterns:        []string{"*.txt", "!test.txt"},
			path:            "test.txt",
			wantMatch:       false,
			wantMatchLine:   "!test.txt",
			wantMatchNegate: true,
		},
		"no match returns nil pattern": {
			patterns:        []string{"*.txt"},
			path:            "file.go",
			wantMatch:       false,
			wantMatchLine:   "",
			wantMatchNegate: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			f := CompileFilterPatterns(tc.patterns...)
			gotMatch, gotPattern := f.MatchesPathHow(tc.path)

			assert.Equal(t, tc.wantMatch, gotMatch)

			if tc.wantMatchLine == "" {
				assert.Nil(t, gotPattern)
				return
			}

			assert.NotNil(t, gotPattern)
			assert.Equal(t, tc.wantMatchLine, gotPattern.Line)
			assert.Equal(t, tc.wantMatchNegate, gotPattern.Negate)
		})
	}
}

func TestCompileAndMatchPatterns(t *testing.T) {
	tests := map[string]struct {
		patterns    []string
		pathsToWant map[string]bool
	}{
		"match all files": {
			patterns: []string{"*"},
			pathsToWant: map[string]bool{
				"main.go":              true,
				"internal/util.go":     true,
				"src/main.go":          true,
				"src/internal/util.go": true,
				"README.md":            true,
				"internal/test.txt":    true,
			},
		},
		"match all files except top-level hidden files and directories": {
			patterns: []string{"*", "!.*"},
			pathsToWant: map[string]bool{
				"main.go":              true,
				"internal/util.go":     true,
				"src/main.go":          true,
				"src/internal/util.go": true,
				"src/.env":             true,
				"README.md":            true,
				"internal/test.txt":    true,

				".git/config.txt": false,
				".env":            false,
			},
		},
		"match all files except recursive hidden files and directories": {
			patterns: []string{"*", "!**/.*"},
			pathsToWant: map[string]bool{
				"main.go":              true,
				"internal/util.go":     true,
				"src/main.go":          true,
				"src/internal/util.go": true,
				"README.md":            true,
				"internal/test.txt":    true,

				"src/.env":        false,
				".git/config.txt": false,
				".env":            false,
			},
		},
		"match top-level go files": {
			patterns: []string{"*.go"},
			pathsToWant: map[string]bool{
				"main.go": true,

				"internal/util.go":     false,
				"src/main.go":          false,
				"src/internal/util.go": false,
				"README.md":            false,
				"internal/test.txt":    false,
			},
		},
		"match go files recursively": {
			patterns: []string{"**/*.go"},
			pathsToWant: map[string]bool{
				"main.go":              true,
				"internal/util.go":     true,
				"src/main.go":          true,
				"src/internal/util.go": true,

				"README.md":         false,
				"internal/test.txt": false,
			},
		},
		"match go files recursively with negation": {
			patterns: []string{"**/*.go", "!internal/**"},
			pathsToWant: map[string]bool{
				"main.go":              true,
				"src/main.go":          true,
				"src/internal/util.go": true,

				"internal/util.go":  false,
				"README.md":         false,
				"internal/test.txt": false,
			},
		},
		"match with escaped special characters": {
			patterns: []string{`\!important.txt`, `\#comment.txt`},
			pathsToWant: map[string]bool{
				"!important.txt": true,
				"#comment.txt":   true,

				"important.txt":       false,
				"comment.txt":         false,
				"test/!important.txt": false,
			},
		},
		"match directories with trailing slash": {
			patterns: []string{"docs/", "!docs/internal/"},
			pathsToWant: map[string]bool{
				"docs/readme.md":   true,
				"docs/api/spec.md": true,

				"docs/internal/dev.md":  false,
				"docs/internal/arch.md": false,
				"other/docs/readme.md":  false,
			},
		},
		"match with directory depth constraints": {
			patterns: []string{
				"/*/*.go",    // Matches files exactly one directory deep
				"!/**/test/", // Excludes any test directories at any depth
			},
			pathsToWant: map[string]bool{
				"cmd/main.go":        true,
				"internal/config.go": true,

				"main.go":              false,
				"pkg/sub/util.go":      false,
				"cmd/test/testutil.go": false,
			},
		},
		"recursively match subdirectory and extension": {
			patterns: []string{"src/**/*.go"},
			pathsToWant: map[string]bool{
				"src/pkg/file.go": true,

				"cmd/main.go":          false,
				"internal/config.go":   false,
				"main.go":              false,
				"pkg/sub/util.go":      false,
				"cmd/test/testutil.go": false,
			},
		},
		"complex nested directory matching": {
			patterns: []string{
				"src/*/test/**/*.go",
				"!src/*/test/vendor/**",
				"!src/temp/*/",
			},
			pathsToWant: map[string]bool{
				"src/project/test/unit/main_test.go":      true,
				"src/lib/test/integration/helper_test.go": true,

				"src/project/test/vendor/mock/mock.go": false,
				"src/temp/cache/data.txt":              false,
				"src/project/prod/main.go":             false,
			},
		},
		"match with multiple pattern ordering": {
			patterns: []string{
				"*.txt",
				"!important.txt",
				"!!important.txt",
				"!test/important.txt",
			},
			pathsToWant: map[string]bool{
				"readme.txt":         true,
				"important.txt":      true,
				"docs/notes.txt":     false,
				"test/important.txt": false,
			},
		},
		"match with question mark wildcards": {
			patterns: []string{"test?.txt", "lib/????.go"},
			pathsToWant: map[string]bool{
				"test1.txt":   true,
				"testa.txt":   true,
				"lib/util.go": true,
				"lib/main.go": true,

				"test.txt":     false,
				"test12.txt":   false,
				"lib/utils.go": false,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			f := CompileFilterPatterns(tt.patterns...)

			for path, want := range tt.pathsToWant {
				got := f.MatchesPath(path)
				assert.Equal(t, want, got, "Patterns: %q; Path: %q", strings.Join(tt.patterns, ", "), path)
			}
		})
	}
}
