## std.test

Thin intrinsic wrappers around the runtime test harness. These are typically used indirectly via `std.testing`, but remain available for low-level control or integration with legacy suites.

### Functions

| Function | Description |
| -------- | ----------- |
| `start(name:string)` | Begin a named test case. |
| `end(name:string, passed:bool)` | Mark the active test as passed/failed. |
| `summary()` | Emit a harness-managed summary of accumulated results. |

