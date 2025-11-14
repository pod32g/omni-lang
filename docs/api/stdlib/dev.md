## std.dev

Utilities aimed at developer workflows such as file watching and lightweight change detection. These helpers are intentionally simple—ideal for scripting rebuild loops inside OmniLang without shelling out to external tooling.

### Structures

| Name | Description |
| ---- | ----------- |
| `WatchSnapshot` | Captures whether a path exists and its current size in bytes. |

### Functions

| Function | Description |
| -------- | ----------- |
| `snapshot(path:string):WatchSnapshot` | Produce a snapshot for the current path state. Missing paths report `exists = false` and `size = -1`. |
| `wait_for_change(path:string, poll_milliseconds:int):WatchSnapshot` | Block until the path’s snapshot changes (creation, deletion, or size change) while polling at the requested interval. |
| `changed(current:WatchSnapshot, baseline:WatchSnapshot):bool` | Convenience helper to compare two snapshots. |
| `watch_loop(path:string, poll_milliseconds:int, iterations:int):WatchSnapshot` | Repeatedly wait for changes `iterations` times (defaulting to 1) and return the latest snapshot. |

### Example

```omni
import std
import std.dev

func main():int {
    std.io.println("Waiting for changes to src/main.omni…")
    let updated = std.dev.wait_for_change("src/main.omni", 250)
    if updated.exists {
        std.io.println("File updated (size=" + std.int_to_string(updated.size) + ")")
    } else {
        std.io.println("File removed")
    }
    return 0
}
```

