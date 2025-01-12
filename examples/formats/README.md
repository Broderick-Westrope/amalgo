# Format Examples

These files are examples of each available output format. Each file name is a valid value for the `--format` flag and the file extension is the extension used for the default file of that type. Normally, the generated files will be named "amalgo". These were generated with the following command:

```sh
amalgo -f '*,!.*,!examples/,!*amalgo*,!*.md,!LICENSE' --format <format>
```

Each generated output file includes:

1. **Header**: Timestamp and generation information
2. **Directory Tree**: Visual representation of the project structure (unless `--no-tree` is specified)
3. **Language-Specific Outlines**: Structural analysis of supported source files (if `--outline` is specified)
4. **File Contents**: Complete source code of all included files (unless `--no-dump` is specified)
