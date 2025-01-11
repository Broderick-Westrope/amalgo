# Amalgo

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Amalgo is a command-line tool that creates consolidated snapshots of source code for analysis, documentation, and sharing with [LLMs](https://en.wikipedia.org/wiki/Large_language_model). It helps developers gather and organize their codebase into a single, well-structured document.

## Features

- 📁 **Directory Tree Generation**: Creates a visual representation of your project structure
- 📝 **Code Content Dumping**: Consolidates all source files into a single document
- 🔍 **Language-Specific Outlines**: Generates structural outlines for supported programming languages
- 🎨 **Syntax Support**: Currently supports Go, with extensibility for other languages
- ⚡ **Flexible Filtering**: Include/exclude files using glob patterns
- 🚫 **Binary File Handling**: Option to skip or include binary files
- 🎯 **Selective Processing**: Ignore specific directories and hidden files

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

## License

This project is licensed under the GNU GPL v3 License - see the [LICENSE](./LICENSE) file for details.