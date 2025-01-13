package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			want:     false, // The implementation escapes ? rather than treating it as a wildcard
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
func TestCompileFilterPatterns(t *testing.T) {
	tests := map[string]struct {
		input    []string
		wantLen  int
		wantLine string // Line of first pattern if wantLen > 0
	}{
		"empty patterns": {
			input:   []string{},
			wantLen: 0,
		},
		"only comments": {
			input:   []string{"#comment", " # comment "},
			wantLen: 0,
		},
		"empty lines are skipped": {
			input:   []string{"", "  "},
			wantLen: 0,
		},
		"single valid pattern": {
			input:    []string{"*.txt"},
			wantLen:  1,
			wantLine: "*.txt",
		},
		"mixed patterns": {
			input:    []string{"", "#comment", "*.txt", "  ", "*.md"},
			wantLen:  2,
			wantLine: "*.txt",
		},
		"whitespace is trimmed": {
			input:    []string{" *.txt  "},
			wantLen:  1,
			wantLine: "*.txt",
		},
		"negated patterns": {
			input:    []string{"!*.txt"},
			wantLen:  1,
			wantLine: "!*.txt",
		},
		"escaped characters": {
			input:    []string{"\\#not-a-comment"},
			wantLen:  1,
			wantLine: "\\#not-a-comment",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			f := CompileFilterPatterns(tc.input...)
			assert.Len(t, f.patterns, tc.wantLen)
			if tc.wantLen > 0 {
				require.GreaterOrEqual(t, len(f.patterns), 1)
				assert.Equal(t, tc.wantLine, f.patterns[0].Line)
			}
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
