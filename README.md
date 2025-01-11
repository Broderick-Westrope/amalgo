# Amalgo

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Amalgo is a command-line tool that creates consolidated snapshots of source code for analysis, documentation, and sharing with [LLMs](https://en.wikipedia.org/wiki/Large_language_model). It helps developers gather and organize their codebase into a single, well-structured document.

- [Features](#features)
- [Example Use Cases](#example-use-cases)
- [Installation](#installation)
- [Usage](#usage)
- [Output Format](#output-format)
- [Contributing](#contributing)
- [License](#license)

## Features

- 📁 **Directory Tree Generation**: Creates a visual representation of your project structure
- 📝 **Code Content Dumping**: Consolidates all source files into a single document
- 🔍 **Language-Specific Outlines**: Generates structural outlines for supported programming languages
- 🎨 **Syntax Support**: Currently supports Go, with extensibility for other languages
- ⚡ **Flexible Filtering**: Include/exclude files using glob patterns
- 🚫 **Binary File Handling**: Option to skip or include binary files
- 🎯 **Selective Processing**: Ignore specific directories and hidden files

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

## Installation

```bash
go install github.com/Broderick-Westrope/amalgo@latest
```

## Usage

Basic usage:

```bash
amalgo [flags] [directories...]
```

Example commands:

```bash
# Analyze current directory
amalgo

# Analyze specific directories
amalgo ./src ./lib

# Output to a specific file
amalgo -o output.txt ./src

# Print output to stdout
amalgo --stdout ./src

# Include only specific file types
amalgo -f "*.go" -f "*.js" ./src

# Ignore certain directories
amalgo -i node_modules -i .git ./src

# Generate only the outline
amalgo --no-tree --no-dump --outline ./src

# Include hidden files
amalgo --include-hidden ./src
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-o, --output` | Output file path | `amalgo.txt` |
| `--stdout` | Write output to stdout instead of file | `false` |
| `-i, --ignore-dirs` | Directories to ignore | `[]` |
| `-f, --filter` | File patterns to include (e.g., '*.go', '*.{js,ts}') | `["*"]` |
| `--include-hidden` | Include hidden files and directories | `false` |
| `--no-tree` | Skip directory tree generation | `false` |
| `--no-dump` | Skip file content dumping | `false` |
| `--outline` | Generate language-specific outlines | `false` |
| `--no-color` | Don't use color in the terminal output | `false` |
| `--include-binary` | Include binary files | `false` |
| `-v, --version` | Print version information and quit | |

## Output Format

The generated output file includes:

1. **Header**: Timestamp and generation information
2. **Directory Tree**: Visual representation of the project structure (unless `--no-tree` is specified)
3. **Language-Specific Outlines**: Structural analysis of supported source files (if `--outline` is specified)
4. **File Contents**: Complete source code of all included files (unless `--no-dump` is specified)

Example output:

```
## Generated with Amalgo at: 2025-01-11 22:49:36

Directory Tree
└── project/
    ├── main.go
    ├── internal/
    │   ├── cli/
    │   │   └── cli.go
    │   └── utils/
    │       └── utils.go

## Language-Specific Outlines

### File: main.go
FUNCTION: main()
  Documentation:
    Entry point for the application

## File Contents

--- File: main.go
package main
...
```

## Contributing

Contributions are welcome! Feel free to open issues and submit pull requests.

I encourage you to create an issue and spark a discussion there before beginning work on a large change. This way we can be clear on the goals and acceptance criteria before investing time on it.

## License

This project is licensed under the GNU GPL v3 License - see the [LICENSE](./LICENSE) file for details.