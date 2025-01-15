// Package filter provides functionality for matching paths against exclusion patterns
// using a syntax similar to gitignore - patterns indicate what to include (match against)
// unless prefixed with '!'. This can be used for both file inclusion and exclusion.
package filter

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Filter wraps a list of filter patterns.
type Filter struct {
	patterns []*Pattern
}

// Pattern encapsulates a regexp pattern and whether it is negated.
type Pattern struct {
	Pattern *regexp.Regexp
	Negate  bool
	LineNo  int
	Line    string
}

// MergeWithoutPrecedence prepends the patterns of the given filter.
// The provided patterns will be used first, meaning they may be
// overruled by the existing patterns (on this Filter).
func (f *Filter) MergeWithoutPrecedence(other *Filter) {
	f.patterns = append(other.patterns, f.patterns...)
}

// MergeWithoutPrecedence appends the patterns of the given filter.
// The provided patterns will be used last, meaning they may
// overrule the existing patterns (on this Filter).
func (f *Filter) MergeWithPrecedence(other *Filter) {
	f.patterns = append(f.patterns, other.patterns...)
}

func (f *Filter) NegateAll() {
	for i := range f.patterns {
		f.patterns[i].Negate = !f.patterns[i].Negate

		pattern, found := strings.CutPrefix(f.patterns[i].Line, "!")
		if !found {
			pattern = "!" + f.patterns[i].Line
		}
		f.patterns[i].Line = pattern
	}
}

// MatchesPath returns true if the path matches the patterns.
func (f *Filter) MatchesPath(path string) bool {
	matches, _ := f.MatchesPathHow(path)
	return matches
}

// MatchesPathHow returns whether the path matches and which pattern matched it.
func (f *Filter) MatchesPathHow(path string) (bool, *Pattern) {
	// Normalize path separators.
	path = filepath.ToSlash(path)

	var matchingPattern *Pattern
	matchesPath := false

	for _, pattern := range f.patterns {
		if pattern.Pattern.MatchString(path) {
			if !pattern.Negate {
				matchesPath = true
				matchingPattern = pattern
			} else if matchesPath {
				// Path was previously matched but now negated.
				matchesPath = false
				matchingPattern = pattern
			}
		}
	}
	return matchesPath, matchingPattern
}

// CompileFilterPatterns accepts a variadic set of strings and returns a Filterer
// instance with the compiled patterns.
func CompileFilterPatterns(patterns ...string) *Filter {
	f := new(Filter)
	for i, pattern := range patterns {
		pattern = strings.TrimRight(pattern, "\r")
		pattern = strings.TrimSpace(pattern)
		compiledPattern, isNegated := getPatternFromLine(pattern)
		if compiledPattern != nil {
			fp := &Pattern{compiledPattern, isNegated, i + 1, pattern}
			f.patterns = append(f.patterns, fp)
		}
	}
	return f
}

// CompileFilterPatternFile reads patterns from a file and compiles them.
func CompileFilterPatternFile(path string) (*Filter, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	patterns := strings.Split(string(bs), "\n")
	return CompileFilterPatterns(patterns...), nil
}

// CompileExcludePatternFileAndLines compiles patterns from both a file and additional lines.
func CompileFilterPatternFileAndLines(path string, lines ...string) (*Filter, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	patterns := append(strings.Split(string(bs), "\n"), lines...)
	return CompileFilterPatterns(patterns...), nil
}

// getPatternFromLine converts a single pattern line into a regexp and bool indicating
// if it's a negated pattern. The rules follow .gitignore syntax.
func getPatternFromLine(line string) (*regexp.Regexp, bool) {
	// Strip comments.
	if strings.HasPrefix(line, "#") {
		return nil, false
	}

	// Skip empty lines.
	if line == "" {
		return nil, false
	}

	// Check for negation prefix. Several will negate the previous negation (ie. toggling).
	negatePattern := false
	for line[0] == '!' {
		negatePattern = !negatePattern
		line = line[1:]
	}

	// Create a copy of the line to modify. The original is maintained for later checks.
	expr := line

	// Ignore a prefix of escaped '#' or '!'.
	if regexp.MustCompile(`^(\#|\!)`).MatchString(expr) {
		expr = expr[1:]
	}

	// Escape dots.
	expr = regexp.MustCompile(`\.`).ReplaceAllString(expr, `\.`)

	// This 'magic star" is used temporarily when handling other single-star cases.
	magicStar := "#$~"

	// Handle '/**/' patterns.
	if strings.HasPrefix(expr, "/**/") {
		expr = expr[1:]
	}
	expr = regexp.MustCompile(`/\*\*/`).ReplaceAllString(expr, `(/|/.+/)`)
	expr = regexp.MustCompile(`\*\*/`).ReplaceAllString(expr, `(|.`+magicStar+`/)`)
	expr = regexp.MustCompile(`/\*\*`).ReplaceAllString(expr, `(|/.`+magicStar+`)`)

	// Handle wildcards.
	expr = regexp.MustCompile(`\\\*`).ReplaceAllString(expr, `\`+magicStar)
	expr = regexp.MustCompile(`\*`).ReplaceAllString(expr, `([^/]*)`) // '*' may be any number of characters other than '/'
	expr = strings.Replace(expr, "?", `[^/]`, -1)                     // '?' may be any single character other than '/'
	expr = strings.Replace(expr, magicStar, "*", -1)

	// Build final regex.
	if strings.HasSuffix(line, "/") {
		expr += "(|.*)$"
	} else {
		expr += "(|/.*)$"
	}

	// Only add directory prefix for patterns starting with /
	switch {
	case strings.HasPrefix(line, "/"):
		expr = "^(|/)" + expr[1:]

	case strings.HasPrefix(line, "**/"):
		// Pattern contains a slash but doesn't start with one
		expr = "^(|.*/)" + expr

	default:
		// Simple pattern like *.go - should only match in current directory
		expr = "^" + expr
	}

	pattern, _ := regexp.Compile(expr)
	return pattern, negatePattern
}
