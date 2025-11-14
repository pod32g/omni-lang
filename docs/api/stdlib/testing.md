## std.testing

Lightweight helpers for organizing and reporting OmniLang test suites. These functions wrap the existing intrinsic hooks (`std.test.*`) and integrate cleanly with the new `omnir --test` mode.

### Structures

| Name | Description |
| ---- | ----------- |
| `Suite` | Opaque handle to a runtime-managed suite. The VM tracks counters and reporting for each handle. |

### Functions

| Function | Description |
| -------- | ----------- |
| `suite():Suite` | Create a brand new suite with zeroed counters. |
| `expect(state:Suite, name:string, condition:bool, message:string):Suite` | Record a boolean test outcome. The runtime prints test output and logs the failure message when `condition` is false. |
| `pass(state:Suite, name:string):Suite` | Convenience wrapper for a successful check. |
| `fail(state:Suite, name:string, message:string):Suite` | Convenience wrapper for a failed check. |
| `equal_int(state:Suite, name:string, expected:int, actual:int):Suite` | Assert that two integers match. |
| `equal_bool(state:Suite, name:string, expected:bool, actual:bool):Suite` | Assert that two booleans match. |
| `equal_string(state:Suite, name:string, expected:string, actual:string):Suite` | Assert that two strings match. |
| `equal_float(state:Suite, name:string, expected:float, actual:float):Suite` | Assert float equality with a default precision of 6 decimal places. |
| `equal_float_precision(state:Suite, name:string, expected:float, actual:float, precision:int):Suite` | Assert float equality with caller-provided precision. |
| `total(state:Suite):int` | Retrieve how many tests have been recorded. |
| `failures(state:Suite):int` | Retrieve how many tests have failed. |
| `summary(state:Suite):int` | Emit a log summary (and call `std.test.summary`) returning the number of failed tests. |
| `passed(state:Suite):bool` | Convenience helper that returns `true` when no tests have failed. |
| `exit(state:Suite)` | Print a summary and terminate the process with the number of failed tests as the exit code. Ideal for CLI integration. |

### Example

```omni
import std
import std.testing

func main():int {
    var suite = std.testing.suite()
    suite = std.testing.pass(suite, "initial setup")
    suite = std.testing.equal_int(suite, "addition", 4, 2 + 2)
    suite = std.testing.expect(suite, "boolean check", 1 < 2, "ordering should hold")

    let failed = std.testing.summary(suite)
    if failed == 0 {
        return 0
    }
    return failed
}
```

