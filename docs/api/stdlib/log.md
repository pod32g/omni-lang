## std.log

Structured logging for OmniLang programs, powered by [`simple-logger`](https://github.com/pod32g/simple-logger). The logger instance is initialised once per process and shared with the compiler, runner, and runtime.

### Module Overview

```omni
import std

func main():int {
    std.log.info("service starting")

    if std.log.set_level("debug") {
        std.log.debug("debug logging enabled")
    } else {
        std.log.warn("invalid log level requested")
    }

    std.log.error("fatal path encountered")
    return 0
}
```

### Functions

| Function | Description |
| -------- | ----------- |
| `std.log.debug(message:string)` | Emit a debug-level log line. Only visible when the active level is `DEBUG`. |
| `std.log.info(message:string)` | Emit an info-level log line. Visible for `INFO` and below. |
| `std.log.warn(message:string)` | Emit a warning. |
| `std.log.error(message:string)` | Emit an error. |
| `std.log.set_level(level:string):bool` | Attempt to change the active log level. Returns `true` on success. Accepted values: `debug`, `info`, `warn`, `error`. |

### Environment Configuration

Set environment variables before launching an Omni binary (`omnic`, `omnir`, `omnipkg`, or generated executables) to control logging behaviour:

- `LOG_LEVEL` - default `info`. Accepts `debug`, `info`, `warn`, `error`, `fatal`.
- `LOG_OUTPUT` - `stdout`, `stderr`, or a filesystem path. Defaults to `stderr`.
- `LOG_FORMAT` - `text` (default) or `json`.
- `LOG_COLORIZE` - `true`/`false`. Colourise text output when supported.
- `LOG_ENABLE_CALLER` - include caller information (`true`/`false`).
- `LOG_SYNC_WRITES` - force synchronous writes (`true`/`false`, default `true`).
- `LOG_TIME_FORMAT` - Go time layout (e.g. `2006-01-02T15:04:05Z07:00`).
- `LOG_INCLUDE_STACKTRACE` - append stacktraces on error logs (`true`/`false`).
- `LOG_ROTATE` - enable file rotation (`true`/`false`) with supporting keys:
  - `LOG_ROTATE_MAX_SIZE` (megabytes, default `100`)
  - `LOG_ROTATE_MAX_AGE` (days, default `30`)
  - `LOG_ROTATE_MAX_BACKUPS` (default `7`)
  - `LOG_ROTATE_COMPRESS` (`true`/`false`, default `true`)

When no configuration is supplied Omni defaults to:

- Level: `INFO`
- Output: `stderr`
- Format: human-readable text
- Synchronous, non-colourised writes

### Integration Notes

- The `-verbose` flag for Omni CLIs (`omnic`, `omnir`, `omnipkg`) temporarily escalates the level to `DEBUG`.
- Generated programs can freely change levels at runtime via `std.log.set_level`, affecting subsequent log statements in the same process.
- The C runtime exposes `omni_log_debug|info|warn|error|set_level`, so logging works consistently across the C backend, VM backend, and packaged binaries.

