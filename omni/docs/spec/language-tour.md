# OmniLang Language Tour (Developer Preview)

OmniLang is a statically typed systems language targeting modern operating
systems. The initial developer preview focuses on a clear syntax, strong static
analysis and a modern toolchain experience.

## Hello, Omni

```
func main() {
    print("Hello, Omni!\n")
}
```

## Variables

```
let answer:int = 42
var counter:int = 0
```

## Control Flow

```
if answer > 0 {
    println("positive")
} else {
    println("non-positive")
}
```

Loops use either C-style or range syntax:

```
for i:int = 0; i < 10; i++ {
    println(i)
}

for item in items {
    println(item)
}
```

More sections will follow as the parser, type checker and backend mature.
