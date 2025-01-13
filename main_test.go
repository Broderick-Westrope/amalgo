package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestMain(m *testing.M) {
	color.NoColor = false
	os.Exit(testscript.RunMain(m,
		map[string]func() int{
			appName: run,
		},
	))
}

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
		Setup: func(env *testscript.Env) error {
			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"cmpfile": compareFiles,
			"showfile": func(ts *testscript.TestScript, neg bool, args []string) {
				if len(args) != 1 {
					ts.Fatalf("usage: showfile filename")
				}
				content, err := os.ReadFile(ts.MkAbs(args[0]))
				if err != nil {
					ts.Fatalf("reading %s: %v", args[0], err)
				}
				fmt.Fprintf(ts.Stdout(), "=== Content of %s ===\n", args[0])
				fmt.Fprintf(ts.Stdout(), "%s\n", content)
				fmt.Fprintf(ts.Stdout(), "=== End of %s ===\n", args[0])
			},
		},
	})
}

func compareFiles(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 2 {
		ts.Fatalf("usage: cmpfile actual expected")
	}
	temp := ts.MkAbs(args[0])
	actual, err := os.ReadFile(temp)
	if err != nil {
		ts.Fatalf("reading %q (actual): %v", temp, err)
	}
	temp = ts.MkAbs(args[1])
	expected, err := os.ReadFile(temp)
	if err != nil {
		ts.Fatalf("reading %q (expected): %v", temp, err)
	}

	// Split into lines
	actualLines := strings.Split(string(actual), "\n")
	expectedLines := strings.Split(string(expected), "\n")

	// Create debug versions with visible empty lines
	debugActual := make([]string, len(actualLines))
	debugExpected := make([]string, len(expectedLines))

	for i, line := range actualLines {
		if line == "" {
			debugActual[i] = "<empty>"
		} else {
			debugActual[i] = line
		}
	}

	for i, line := range expectedLines {
		if line == "" {
			debugExpected[i] = "<empty>"
		} else {
			debugExpected[i] = line
		}
	}

	matchFailed := false
	if len(actualLines) != len(expectedLines) {
		if !neg {
			matchFailed = true
		}
	}

	for i := 0; i < len(actualLines) && i < len(expectedLines); i++ {
		aLine := actualLines[i]
		eLine := expectedLines[i]
		// Skip timestamp line
		if strings.Contains(aLine, "Generated with Amalgo at:") || strings.Contains(aLine, "timestamp") {
			continue
		}
		if aLine != eLine {
			matchFailed = true
			break
		}
	}

	if matchFailed {
		diffStr := createDiff(
			fmt.Sprintf("Expected (%d lines)", len(expectedLines)),
			fmt.Sprintf("Actual (%d lines)", len(actualLines)),
			strings.Join(debugExpected, "\n"),
			strings.Join(debugActual, "\n"),
		)
		ts.Fatalf("Failed to match:\n%s", diffStr)
	} else if neg {
		ts.Fatalf("files match but should not")
	}
}

func createDiff(name1, name2, text1, text2 string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(text1, text2, false)
	return fmt.Sprintf(
		"%s\n%s\n%s",
		color.RedString("--- "+name1),
		color.GreenString("+++ "+name2),
		dmp.DiffPrettyText(diffs),
	)
}
