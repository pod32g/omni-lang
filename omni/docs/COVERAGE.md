# Standard Library Coverage

This document describes how to use the coverage tracking system for the OmniLang standard library.

## Overview

The coverage system tracks which runtime-wired functions in the standard library are executed during tests. It measures both function coverage (percentage of functions called) and line coverage (percentage of lines executed).

## Usage

### Running Tests with Coverage

To run tests with coverage tracking enabled:

```bash
omnir --coverage --coverage-output coverage.json test.omni
```

Or for test harness mode:

```bash
omnir --coverage --coverage-output coverage.json --test test.omni
```

### Analyzing Coverage

Use the `omnicover` tool to analyze coverage data:

```bash
# Analyze coverage and print summary
omnicover analyze coverage.json std/

# Generate text report
omnicover report coverage.json std/ --format=text

# Generate HTML report
omnicover report coverage.json std/ --format=html coverage.html

# Check if coverage meets threshold (default 60%)
omnicover check coverage.json std/ --threshold=60
```

## Coverage Report Format

### Text Report

The text report shows:
- Overall function and line coverage percentages
- Per-file breakdown
- List of covered and uncovered functions

### HTML Report

The HTML report provides:
- Visual coverage bars
- Color-coded coverage levels (green ≥60%, yellow ≥30%, red <30%)
- Detailed function-level information
- Per-file statistics

## Coverage Data Format

Coverage data is stored in JSON format:

```json
{
  "entries": [
    {
      "function": "std.io.print",
      "file": "",
      "line": 0,
      "count": 5
    }
  ]
}
```

## Coverage Threshold

The default coverage threshold is 60%. This means:
- At least 60% of runtime-wired functions must be called
- At least 60% of lines in runtime-wired functions must be executed

## Runtime-Wired Functions

Only functions that are wired to the runtime (intrinsic functions) are tracked for coverage. These include:

- **std.io**: `print`, `println`, `read_line`
- **std.string**: `length`, `concat`, `substring`, `char_at`, `starts_with`, `ends_with`, `contains`, `index_of`, `last_index_of`, `trim`, `to_upper`, `to_lower`, `equals`, `compare`
- **std.math**: `abs`, `max`, `min`, `pow`, `sqrt`, `floor`, `ceil`, `round`
- **std.array**: `length`
- **std.file**: `open`, `close`, `read`, `write`, `seek`, `tell`, `exists`, `size`
- **std.os**: `exit`, `getenv`, `setenv`, `remove`
- And more...

## Integration with Tests

The Go test infrastructure (`omni/tests/std/std_test.go`) includes:

- `TestCoverageGeneration`: Verifies that coverage data is generated
- `TestCoverageThreshold`: Checks that coverage meets the 60% threshold

Run these tests with:

```bash
go test ./tests/std -v
```

## Limitations

- Coverage tracking only works with the VM backend (`--backend vm`)
- Only runtime-wired functions are tracked (functions implemented in C runtime)
- File paths and line numbers may not be available in VM mode
- Coverage data is accumulated during execution and exported at the end

