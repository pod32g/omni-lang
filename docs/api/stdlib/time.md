# std.time - Time And Duration Utilities

The `std.time` module provides UTC time conversion helpers, simple duration arithmetic, sleep functions, and local timezone metadata.

## Types

```omni
struct Time {
    year:int
    month:int
    day:int
    hour:int
    minute:int
    second:int
    nanosecond:int
}

struct Duration {
    seconds:int
    nanoseconds:int
}
```

`Time` conversion helpers use UTC and RFC3339-style strings such as `"2024-04-27T12:30:05Z"`.

## Current Time

| Function | Description |
|----------|-------------|
| `now(): Time` | Current UTC time as a `Time` struct. |
| `unix_timestamp(): int` | Current Unix timestamp in seconds. |
| `unix_nano(): int` | Current Unix timestamp in nanoseconds. On today's 32-bit `int` surface this can overflow; use only when the value is known to fit. |

## Time Creation And Conversion

| Function | Description |
|----------|-------------|
| `time_create(year, month, day, hour, minute, second): Time` | Builds a `Time` with `nanosecond = 0`. |
| `time_from_unix(timestamp:int): Time` | Converts a Unix timestamp in seconds to UTC fields. |
| `time_from_string(time_str:string): Time` | Parses RFC3339/RFC3339Nano text. Invalid input falls back to the Unix epoch. |
| `time_to_unix(t:Time): int` | Converts UTC fields to a Unix timestamp in seconds. |
| `time_to_unix_nano(t:Time): int` | Converts UTC fields to nanoseconds since the Unix epoch. |
| `time_to_string(t:Time): string` | Formats UTC fields as RFC3339/RFC3339Nano text. |
| `time_format(t, layout)` | Currently aliases `time_to_string`; custom layouts are pending. |
| `time_parse(time_str, layout)` | Currently aliases `time_from_string`; custom layouts are pending. |

```omni
import std

func main():int {
    let t = std.time.time_from_string("2024-04-27T12:30:05Z")
    let unix = std.time.time_to_unix(t)
    if unix != 1714221005 {
        return 1
    }
    std.io.println(std.time.time_to_string(t))
    return 0
}
```

## Comparison

| Function | Description |
|----------|-------------|
| `time_equal(a, b): bool` | True when every field matches. |
| `time_before(a, b): bool` | Lexicographic UTC field comparison. |
| `time_after(a, b): bool` | Equivalent to `time_before(b, a)`. |

## Durations

The runtime-backed duration surface is `duration_to_string`. The remaining duration creation, conversion, arithmetic, and comparison helpers are OmniLang source helpers; they work when the module body is loaded, but C-backend parity still needs either std.time body-loading or dedicated intrinsics.

| Function | Description |
|----------|-------------|
| `duration_create(seconds, nanoseconds): Duration` | Builds a duration. |
| `duration_from_seconds(seconds:float): Duration` | Converts fractional seconds. |
| `duration_from_milliseconds(milliseconds:int): Duration` | Converts milliseconds. |
| `duration_from_minutes(minutes:float): Duration` | Converts minutes. |
| `duration_from_hours(hours:float): Duration` | Converts hours. |
| `duration_from_days(days:float): Duration` | Converts days. |
| `duration_to_seconds(d): float` | Converts to fractional seconds. |
| `duration_to_milliseconds(d): int` | Converts to milliseconds. |
| `duration_to_minutes(d): float` | Converts to minutes. |
| `duration_to_hours(d): float` | Converts to hours. |
| `duration_to_days(d): float` | Converts to days. |
| `duration_to_string(d): string` | Formats as seconds, for example `"2.500000000s"`. |
| `duration_add(a, b): Duration` | Adds and normalizes nanosecond overflow. |
| `duration_sub(a, b): Duration` | Subtracts and normalizes nanosecond underflow. |
| `duration_mul(d, scalar): Duration` | Multiplies by a scalar. |
| `duration_div(d, scalar): Duration` | Divides by a scalar; divide-by-zero returns zero duration. |
| `duration_equal`, `duration_less`, `duration_greater` | Duration comparisons. |

## Sleep And Timezone

| Function | Description |
|----------|-------------|
| `sleep(duration:Duration)` | Sleeps for the duration. |
| `sleep_seconds(seconds:float)` | Sleeps for fractional seconds. |
| `sleep_milliseconds(milliseconds:int)` | Sleeps for milliseconds. |
| `time_zone_offset(): int` | Local timezone offset in seconds. |
| `time_zone_name(): string` | Local timezone name from the host runtime. |

## Backend Status

`std.time` has VM and C backend coverage for deterministic runtime-backed conversions (`time_from_unix`, `time_from_string`, `time_to_unix`, `time_to_string`, `time_to_unix_nano`), `duration_to_string`, timezone helpers, and sleep calls. `time_format` and `time_parse` are intentionally partial aliases until custom layout parsing is designed.
