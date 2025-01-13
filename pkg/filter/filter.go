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

// Filterer wraps a list of filter patterns.
type Filterer struct {
	patterns []*Pattern
}

// Pattern encapsulates a regexp pattern and whether it is negated.
type Pattern struct {
	Pattern *regexp.Regexp
	Negate  bool
	LineNo  int
	Line    string
}

// MatchesPath returns true if the path matches the patterns.
func (f *Filterer) MatchesPath(path string) bool {
	matches, _ := f.MatchesPathHow(path)
	return matches
}

// MatchesPathHow returns whether the path matches and which pattern matched it.
func (f *Filterer) MatchesPathHow(path string) (bool, *Pattern) {
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
func CompileFilterPatterns(patterns ...string) *Filterer {
	f := new(Filterer)
	for i, pattern := range patterns {
		compiledPattern, isNegated := getPatternFromLine(pattern)
		if compiledPattern != nil {
			fp := &Pattern{compiledPattern, isNegated, i + 1, pattern}
			f.patterns = append(f.patterns, fp)
		}
	}
	return f
}

// CompileFilterPatternFile reads patterns from a file and compiles them.
func CompileFilterPatternFile(path string) (*Filterer, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	patterns := strings.Split(string(bs), "\n")
	return CompileFilterPatterns(patterns...), nil
}

// CompileExcludePatternFileAndLines compiles patterns from both a file and additional lines.
func CompileFilterPatternFileAndLines(path string, lines ...string) (*Filterer, error) {
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
	// Trim OS-specific carriage returns.
	line = strings.TrimRight(line, "\r")

	// Strip comments.
	if strings.HasPrefix(line, "#") {
		return nil, false
	}

	// Trim whitespace.
	line = strings.Trim(line, " ")

	// Skip empty lines.
	if line == "" {
		return nil, false
	}

	// Check for exclusion prefix.
	negatePattern := false
	if line[0] == '!' {
		negatePattern = true
		line = line[1:]
	}

	// Ignore a prefix of escaped '#' or '!'.
	if regexp.MustCompile(`^(\#|\!)`).MatchString(line) {
		line = line[1:]
	}

	// Add leading '/' for patterns like 'foo/*.txt'.
	if regexp.MustCompile(`([^\/+])/.*\*\.`).MatchString(line) && line[0] != '/' {
		line = "/" + line
	}

	// Escape dots.
	line = regexp.MustCompile(`\.`).ReplaceAllString(line, `\.`)

	// This 'magic star" is used temporarily when handling other single-star cases.
	magicStar := "#$~"

	// Handle '/**/' patterns.
	if strings.HasPrefix(line, "/**/") {
		line = line[1:]
	}
	line = regexp.MustCompile(`/\*\*/`).ReplaceAllString(line, `(/|/.+/)`)
	line = regexp.MustCompile(`\*\*/`).ReplaceAllString(line, `(|.`+magicStar+`/)`)
	line = regexp.MustCompile(`/\*\*`).ReplaceAllString(line, `(|/.`+magicStar+`)`)

	// Handle wildcards.
	line = regexp.MustCompile(`\\\*`).ReplaceAllString(line, `\`+magicStar)
	line = regexp.MustCompile(`\*`).ReplaceAllString(line, `([^/]*)`)
	line = strings.Replace(line, "?", `\?`, -1)
	line = strings.Replace(line, magicStar, "*", -1)

	// Build final regex.
	var expr string
	if strings.HasSuffix(line, "/") {
		expr = line + "(|.*)$"
	} else {
		expr = line + "(|/.*)$"
	}
	if strings.HasPrefix(expr, "/") {
		expr = "^(|/)" + expr[1:]
	} else {
		expr = "^(|.*/)" + expr
	}

	pattern, _ := regexp.Compile(expr)
	return pattern, negatePattern
}
