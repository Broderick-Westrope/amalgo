# amalgo

[![Version](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go)](https://go.dev/doc/install)
[![Reference](https://img.shields.io/badge/Go-Reference-00ADD8?style=flat&logo=go)](https://pkg.go.dev/github.com/Broderick-Westrope/amalgo)
[![License](https://img.shields.io/badge/license-GNU%20GPLv3-blue.svg)](https://github.com/Broderick-Westrope/amalgo/blob/main/LICENSE)

Amalgo is a command-line tool that creates consolidated snapshots (ie. an amalgamation) of source code for analysis, documentation, and sharing with [LLMs](https://en.wikipedia.org/wiki/Large_language_model). It helps developers gather and organize their codebase into a single, well-structured document.

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Output Format](#output-format)
- [Example Use Cases](#example-use-cases)
- [Contributing](#contributing)
- [License](#license)

## Features

- 📁 **Directory Tree Generation**: Creates a visual representation of your project structure.
- 📝 **Code Content Dumping**: Consolidates all source files into a single document.
- 🎯 **Flexible Filtering**: Include/exclude files using the gitignore pattern syntax.
- 🔍 **Language-Specific Outlines**: Generates structural outlines for supported programming languages.
- 🎨 **Syntax Support**: The language outlines feature currently supports Go, with extensibility for other languages. All other features are language agnostic.
- 🚫 **Binary File Handling**: Option to skip or include binary files.

## Installation

```bash
# Using Homebrew
brew install Broderick-Westrope/tap/amalgo

# Using Go install
go install github.com/Broderick-Westrope/amalgo@latest
```

You can also find various builds on [the releases page](https://github.com/Broderick-Westrope/amalgo/releases).

## Usage

Use the help flag to get more information on usage:

```bash
amalgo --help
```

Example commands:

```bash
# Analyze current directory, excluding hidden files and directories by default
amalgo

# Analyze a specific directory
amalgo internal/

# Output to a specific file
amalgo -o output.txt

# Print output to stdout
amalgo --stdout

# Include only specific file types (eg. Go files without any Go tests or hidden files/directories)
amalgo -f '**/*.go,!**/*_test.go,!.*'

# Exclude certain directories (eg. include everything except the .git directory)
amalgo -f '*,!.git/'

# Include hidden files and directories
amalgo -f '*'

# Generate only the language-specific outline
amalgo --no-tree --no-dump --outline
```

### Positional Arguments

- `dir`
  - **Description:** Directory to analyze. If a file is provided it's parent directory will be used.
  - **Optional:** `true`
  - **Default:** `.` (current directory)

### Flags

Each flag has a corresponding environment variable which can be used to set the value. Flags override environment variables.

- `-o, --output`
  - **Description:** Specifies the destination path for the output file. The file extension will automatically adjust based on the selected format (see `--format`).
  - **Default:** `amalgo.txt`
  - **Environment Variable:** `$AMALGO_OUTPUT`

- `stdout`
  - **Description:** Redirects all output to standard output (terminal) instead of writing to a file. Useful for piping output to other commands.
  - **Default:** `false`
  - **Environment Variable:** `$AMALGO_STDOUT`

- `-f, --filter`
  - **Description:** Controls which files are processed using glob patterns similar to gitignore. Include patterns are processed first, then exclude patterns (prefixed with `!`). Hidden files and directories are excluded by default.
  - **Default:** `*,!.*`
  - **Environment Variable:** `$AMALGO_FILTER`
  - **Examples:**
    - `*.go,*.{js,ts}` - Include only Go, JavaScript, and TypeScript files.
    - `*,!*.md` - Include everything except Markdown files.

- `--no-tree`
  - **Description:** Skips the inclusion of the file tree in the output.
  - **Default:** `false`
  - **Environment Variable:** `$AMALGO_NO_TREE`

- `--no-dump`
  - **Description:** Skips the inclusion of file contents in the output.
  - **Default:** `false`
  - **Environment Variable:** `$AMALGO_NO_DUMP`

- `--outline`
  - **Description:** Includes in the output a language-aware outline of code files, showing functions, classes, and other significant elements. Only available for specific file extensions: `.go`.
  - **Default:** `false`
  - **Environment Variable:** `$AMALGO_OUTLINE`

- `--no-color`
  - **Description:** Disables ANSI color codes in the output.
  - **Default:** `false`
  - **Environment Variable:** `$AMALGO_NO_COLOR`

- `--include-binary`
  - **Description:** Processes binary files instead of skipping them. Use with caution as this may produce large or unreadable output.
  - **Default:** `false`
  - **Environment Variable:** `$AMALGO_INCLUDE_BINARY`

- `--format`
  - **Description:** Selects an alternative output format. This affects both the structure and the file extension of the output. Options: `default`, `json`.
  - **Default:** `"default"`
  - **Environment Variable:** `$AMALGO_FORMAT`

- `-v, --version`
  - **Description:** Displays the current version of the tool and exits immediately.
  - **Default:** `false`
  - **Environment Variable:** `$AMALGO_VERSION`

## Output Format

Examples of each output format can be found in [examples/formats/](https://github.com/Broderick-Westrope/amalgo/tree/main/examples/formats).

## Example Use Cases

### 1. Project Onboarding & Understanding
Get up to speed quickly with new codebases:
- Generate comprehensive snapshots for LLM-powered project exploration
- Understand architectural patterns and design decisions
- Identify key components and their relationships
- Perfect for new team members and project handovers

### 2. Project Documentation Generation
Quickly create and maintain project documentation:
- Generate READMEs that accurately reflect the current codebase
- Create architecture diagrams and explanations backed by actual code
- Build API documentation with real usage examples
- Keep documentation synchronized with code as projects evolve

### 3. Smart Code Reviews & Pull Requests
Generate context-rich snapshots of changes that help LLMs provide deeper insights:
- Create comprehensive PR descriptions that understand implementation context
- Generate targeted review checklists based on affected code patterns
- Identify potential impacts on dependent modules and services
- Go beyond simple diffs to understand architectural implications

### 2. Security Audit Assistant
Leverage full codebase context for better security analysis:
- Generate complete snapshots including configs, dependencies, and source code
- Enable security-focused LLMs to identify complex vulnerability patterns
- Catch security issues that emerge from component interactions
- Perfect for pre-release audits and open-source project maintenance

### 4. Architectural Decision Analysis
Make informed architectural decisions with full context:
- Compare different approaches using comprehensive snapshots
- Generate impact analysis reports for proposed changes
- Document architectural decisions with complete implementation context
- Track the evolution of architectural choices over time

### 5. Enhanced Bug Resolution
Provide LLMs with all the context needed for better bug fixing:
- Include related tests, configs, and error contexts in snapshots
- Enable root cause analysis with full system context
- Get fix suggestions that consider potential side effects
- Generate targeted test cases for better coverage

## Contributing

Contributions are welcome! Feel free to open issues and submit pull requests.

I encourage you to create an issue and spark a discussion there before beginning work on a large change. This way we can be clear on the goals and acceptance criteria before investing time on it.

### What could be improved?

Here are some suggestions:

- **Language parsers:** If you would like to add first-class support for a new language that would be great! The `parser` package contains a `Parser` interface that defines what your new parser should include.
- **Ideas:** The thing I'm most interested in hearing is ideas for unique use cases. If the use case requires some modifications that's fine. Similarly, if you think something can be done better I'd love to hear it. This is a relatively small CLI utility, so a bit of growth/change is acceptable for a cool enough use case ;)

## License

This project is licensed under the GNU GPL v3 License - see the [LICENSE](./LICENSE) file for details.
