# OmniLang Language Tour (MVP)

## Hello World
```omni
import std.io

func main() {
    print("Hello, Omni!")
}
```

## Variables
```omni
let answer:int = 42
var counter:int = 0
counter = counter + 1
```

## Functions
```omni
func add(a:int, b:int):int { return a + b }
func sq(x:int):int => x * x
```

## Control Flow
```omni
if answer > 40 { print("big") } else { print("small") }

for i:int = 0; i < 3; i++ {
    print("i=" + i)
}

for item in [1,2,3] {
    print(item)
}
```

## Structs and Enums
```omni
struct Point { x:float y:float }
enum Color { RED GREEN BLUE }

let p:Point = Point{ x:1.0, y:2.0 }
```

## Containers
```omni
let xs:array<int> = [1,2,3]
let m:map<string,int> = { "a":1, "b":2 }
```
