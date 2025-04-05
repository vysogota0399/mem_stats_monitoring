# Static Analysis Client

This is a static analysis tool that combines multiple Go analyzers to perform comprehensive code analysis. The tool helps identify potential issues, bugs, and code quality problems in Go codebases.

## Features

The static analysis client includes the following analyzers:

### Built-in Go Analyzers
- `printf`: Checks for incorrect usage of Printf-style formatting
- `shadow`: Detects shadowed variables
- `structtag`: Validates struct field tags
- `errcheck`: Ensures error return values are checked
- `osexit`: Detects usage of `os.Exit()` calls that may cause unexpected program termination

### Staticcheck Analyzers
The following Staticcheck analyzers are enabled:

#### Style Checks (ST)
- `ST1023`: Unnecessary type conversion
- `ST1018`: Unnecessary use of fmt.Sprintf
- `ST1016`: Unnecessary use of fmt.Sprint
- `ST1013`: Unnecessary use of fmt.Sprintln
- `ST1012`: Unnecessary use of fmt.Sprintf with no arguments
- `ST1011`: Unnecessary use of fmt.Sprintf with no arguments
- `ST1006`: Poorly chosen name for variable of type error
- `ST1005`: Poorly chosen name for variable of type error

#### Simplification Checks (S)
- `S1036`: Unnecessary guard around call to delete
- `S1030`: Unnecessary use of fmt.Sprint
- `S1028`: Unnecessary use of fmt.Sprintf
- `S1016`: Use a constant instead of repeating the conversion
- `S1010`: Omit second value in type assertion
- `S1009`: Omit redundant nil check on ranged loop
- `S1001`: Use a simple channel send/receive instead of select with a single case

#### Quick Fix Checks (QF)
- `QF1010`: Convert slice expression to use pointer to array
- `QF1003`: Expand call to fmt.Sprintf

#### All SA (Static Analysis) Checks
All analyzers starting with "SA" are enabled by default.

## Usage

```bash
# Basic usage
staticclient ./...

# Analyze specific package
staticclient ./path/to/package

# Analyze multiple packages
staticclient ./pkg1/... ./pkg2/...
```

## Output

The tool will output diagnostic messages for any issues found in the code. Each diagnostic includes:
- File location
- Line number
- Issue description
- Suggested fixes (where applicable)

## Integration

This tool can be integrated into CI/CD pipelines or used as part of your development workflow to catch issues early.

## Requirements

- Go 1.16 or later
- Access to the required analyzer packages

## Dependencies

- github.com/kisielk/errcheck
- github.com/vysogota0399/mem_stats_monitoring/internal/multicheker/osexit
- golang.org/x/tools/go/analysis
- golang.org/x/tools/go/analysis/passes/printf
- golang.org/x/tools/go/analysis/passes/shadow
- golang.org/x/tools/go/analysis/passes/structtag
- honnef.co/go/tools/staticcheck 