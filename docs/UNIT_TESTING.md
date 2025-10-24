# Unit Testing in OmniLang

This document describes how to write and run unit tests in OmniLang using the language's built-in testing capabilities.

## Overview

OmniLang provides a comprehensive unit testing framework that allows developers to test individual functions, modules, and complete programs. While the language has built-in test framework functions (like `test.start()`, `test.end()`, `assert.eq()`, etc.), they are currently implemented as intrinsic functions in the VM and not available as standard library functions.

## Writing Unit Tests

### Basic Test Structure

Unit tests in OmniLang follow a simple pattern:

```omni
import std

func test_function_name():int {
    std.io.println("Testing function_name...")
    
    // Test setup
    let input:int = 42
    let expected:int = 84
    
    // Test execution
    let result:int = double(input)
    
    // Manual assertion
    if result != expected {
        std.io.println("FAIL: Expected " + std.int_to_string(expected) + 
                      ", got " + std.int_to_string(result))
        return 1
    }
    
    std.io.println("PASS: Function test")
    return 0
}

func main():int {
    var failed_tests:int = 0
    
    failed_tests = failed_tests + test_function_name()
    
    if failed_tests == 0 {
        std.io.println("=== All tests passed! ===")
    } else {
        std.io.println("=== " + std.int_to_string(failed_tests) + " test(s) failed ===")
    }
    
    return failed_tests
}
```

### Test Organization

1. **Separate Test Functions**: Each test should be in its own function
2. **Clear Naming**: Use descriptive names like `test_basic_math()`, `test_string_operations()`
3. **Manual Assertions**: Use `if` statements to check expected vs actual results
4. **Return Codes**: Return 0 for pass, 1 for fail
5. **Test Runner**: Use a main function to run all tests and report results

## Running Unit Tests

### Using the VM Runner

```bash
cd /path/to/omni-lang/omni
./bin/omnir examples/your_test_file.omni
```

### Expected Output

```
=== Running Unit Tests ===
Testing basic math operations...
PASS: Basic addition test
PASS: Basic multiplication test
Testing string operations...
PASS: String concatenation test
=== All tests passed! ===
```

## Test Categories

### Mathematical Functions

Test arithmetic operations, standard library math functions, and edge cases:

```omni
func test_basic_arithmetic():int {
    std.io.println("Testing basic arithmetic...")
    
    let a:int = 10
    let b:int = 5
    let sum:int = a + b
    
    if sum != 15 {
        std.io.println("FAIL: Addition test failed")
        return 1
    }
    
    std.io.println("PASS: Basic arithmetic test")
    return 0
}
```

### String Operations

Test string manipulation, concatenation, and conversion functions:

```omni
func test_string_operations():int {
    std.io.println("Testing string operations...")
    
    let greeting:string = "Hello"
    let name:string = "World"
    let full_greeting:string = greeting + " " + name
    
    if full_greeting != "Hello World" {
        std.io.println("FAIL: String concatenation failed")
        return 1
    }
    
    std.io.println("PASS: String operations test")
    return 0
}
```

### Array Operations

Test array creation, access, and manipulation:

```omni
func test_array_operations():int {
    std.io.println("Testing array operations...")
    
    let numbers:array<int> = [1, 2, 3, 4, 5]
    let length:int = std.array.length(numbers)
    
    if length != 5 {
        std.io.println("FAIL: Array length test failed")
        return 1
    }
    
    let first_element:int = std.array.get(numbers, 0)
    if first_element != 1 {
        std.io.println("FAIL: Array access test failed")
        return 1
    }
    
    std.io.println("PASS: Array operations test")
    return 0
}
```

### Control Flow

Test conditional statements, loops, and logical operators:

```omni
func test_control_flow():int {
    std.io.println("Testing control flow...")
    
    let x:int = 10
    let y:int = 5
    
    if x > y {
        std.io.println("PASS: Conditional test")
    } else {
        std.io.println("FAIL: Conditional test")
        return 1
    }
    
    return 0
}
```

### Function Testing

Test function definitions, parameters, and return values:

```omni
func add(a:int, b:int):int {
    return a + b
}

func test_function_calls():int {
    std.io.println("Testing function calls...")
    
    let result:int = add(5, 3)
    
    if result != 8 {
        std.io.println("FAIL: Function call test failed")
        return 1
    }
    
    std.io.println("PASS: Function call test")
    return 0
}
```

## Best Practices

### 1. Test Isolation
- Each test should be independent
- Don't rely on state from other tests
- Use local variables for test data

### 2. Clear Test Names
- Use descriptive function names
- Include what is being tested
- Example: `test_string_concatenation()`, `test_array_length()`

### 3. Assertion Messages
- Provide clear failure messages
- Include expected vs actual values
- Use descriptive error messages

### 4. Test Coverage
- Test normal cases
- Test edge cases (empty strings, zero values, etc.)
- Test error conditions

### 5. Test Organization
- Group related tests together
- Use consistent naming conventions
- Keep tests simple and focused

## Example Test Files

The following test files are available in the `examples/` directory:

- `basic_unit_test.omni` - Simple working example
- `unit_tests_math.omni` - Mathematical function tests
- `unit_tests_string.omni` - String operation tests
- `unit_tests_array.omni` - Array operation tests
- `unit_tests_control_flow.omni` - Control flow tests
- `unit_tests_functions.omni` - Function tests
- `unit_tests_types.omni` - Type system tests

## Integration with CI/CD

Unit tests can be integrated into CI/CD pipelines by:

1. **Exit Codes**: Tests return 0 for success, non-zero for failure
2. **Output Parsing**: Parse test output for pass/fail counts
3. **Test Discovery**: Automatically find and run test files

Example CI script:
```bash
#!/bin/bash
cd omni-lang/omni
for test_file in examples/unit_tests_*.omni; do
    echo "Running $test_file..."
    ./bin/omnir "$test_file"
    if [ $? -ne 0 ]; then
        echo "Test $test_file failed"
        exit 1
    fi
done
echo "All tests passed!"
```

## Future Enhancements

The OmniLang test framework could be enhanced with:

1. **Standard Library Integration**: Make test framework functions available in std library
2. **Test Discovery**: Automatic test discovery and execution
3. **Assertion Library**: More assertion functions (assert.greater, assert.less, etc.)
4. **Test Reporting**: Detailed test reports with timing and coverage
5. **Mocking**: Support for mocking external dependencies
6. **Parameterized Tests**: Support for running tests with different parameters

## Conclusion

OmniLang's unit testing framework provides a solid foundation for testing code quality and correctness. While the built-in test framework functions are not yet available as standard library functions, manual assertions provide a reliable and effective way to write comprehensive unit tests.

The framework supports testing all major language features including mathematical operations, string manipulation, array operations, control flow, functions, and type system features. With proper organization and best practices, developers can create maintainable and effective test suites.
