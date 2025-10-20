# std.math - Mathematical Functions

The `std.math` module provides mathematical functions and operations.

## Functions

### abs(x: int): int

Returns the absolute value of an integer.

**Parameters:**
- `x` (int): The integer value

**Returns:**
- `int`: The absolute value of x

**Example:**
```omni
import std.math as math

func main():int {
    let result1:int = math.abs(-5)    // 5
    let result2:int = math.abs(10)    // 10
    let result3:int = math.abs(0)     // 0
    
    return result1 + result2 + result3
}
```

### min(a: int, b: int): int

Returns the minimum of two integers.

**Parameters:**
- `a` (int): First integer
- `b` (int): Second integer

**Returns:**
- `int`: The smaller of a and b

**Example:**
```omni
import std.math as math

func main():int {
    let result1:int = math.min(5, 10)     // 5
    let result2:int = math.min(-3, 2)     // -3
    let result3:int = math.min(7, 7)      // 7
    
    return result1 + result2 + result3
}
```

### max(a: int, b: int): int

Returns the maximum of two integers.

**Parameters:**
- `a` (int): First integer
- `b` (int): Second integer

**Returns:**
- `int`: The larger of a and b

**Example:**
```omni
import std.math as math

func main():int {
    let result1:int = math.max(5, 10)     // 10
    let result2:int = math.max(-3, 2)     // 2
    let result3:int = math.max(7, 7)      // 7
    
    return result1 + result2 + result3
}
```

### toString(value: int): string

Converts an integer to its string representation.

**Parameters:**
- `value` (int): The integer to convert

**Returns:**
- `string`: String representation of the integer

**Example:**
```omni
import std.math as math
import std.io as io

func main():int {
    let num:int = 42
    let str:string = math.toString(num)
    io.println("Number as string: " + str)  // "Number as string: 42"
    
    return 0
}
```

### sqrt(x: int): int

Returns the square root of an integer (integer result).

**Parameters:**
- `x` (int): The integer value

**Returns:**
- `int`: The integer square root of x

**Example:**
```omni
import std.math as math

func main():int {
    let result1:int = math.sqrt(16)    // 4
    let result2:int = math.sqrt(25)    // 5
    let result3:int = math.sqrt(9)     // 3
    
    return result1 + result2 + result3
}
```

### is_prime(n: int): bool

Checks if an integer is prime.

**Parameters:**
- `n` (int): The integer to check

**Returns:**
- `bool`: true if n is prime, false otherwise

**Example:**
```omni
import std.math as math

func main():int {
    let result1:bool = math.is_prime(17)    // true
    let result2:bool = math.is_prime(15)    // false
    let result3:bool = math.is_prime(2)     // true
    
    return 0
}
```

### max_float(a: float, b: float): float

Returns the maximum of two floats.

**Parameters:**
- `a` (float): First float
- `b` (float): Second float

**Returns:**
- `float`: The larger of a and b

**Example:**
```omni
import std.math as math

func main():int {
    let result:float = math.max_float(3.14, 2.71)  // 3.14
    return 0
}
```

### min_float(a: float, b: float): float

Returns the minimum of two floats.

**Parameters:**
- `a` (float): First float
- `b` (float): Second float

**Returns:**
- `float`: The smaller of a and b

**Example:**
```omni
import std.math as math

func main():int {
    let result:float = math.min_float(3.14, 2.71)  // 2.71
    return 0
}
```

### abs_float(x: float): float

Returns the absolute value of a float.

**Parameters:**
- `x` (float): The float value

**Returns:**
- `float`: The absolute value of x

**Example:**
```omni
import std.math as math

func main():int {
    let result:float = math.abs_float(-3.14)  // 3.14
    return 0
}
```

## Usage Examples

### Basic Math Operations

```omni
import std.math as math
import std.io as io

func main():int {
    let a:int = 15
    let b:int = 25
    
    io.println("=== Basic Math Operations ===")
    io.println("a = " + math.toString(a))
    io.println("b = " + math.toString(b))
    io.println("min(a, b) = " + math.toString(math.min(a, b)))
    io.println("max(a, b) = " + math.toString(math.max(a, b)))
    io.println("abs(-a) = " + math.toString(math.abs(-a)))
    
    return 0
}
```

### Prime Number Check

```omni
import std.math as math
import std.io as io

func main():int {
    let numbers:array<int> = [2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17]
    
    io.println("=== Prime Number Check ===")
    for num in numbers {
        if math.is_prime(num) {
            io.println(math.toString(num) + " is prime")
        } else {
            io.println(math.toString(num) + " is not prime")
        }
    }
    
    return 0
}
```

### Square Root Examples

```omni
import std.math as math
import std.io as io

func main():int {
    let numbers:array<int> = [1, 4, 9, 16, 25, 36, 49, 64, 81, 100]
    
    io.println("=== Square Roots ===")
    for num in numbers {
        let sqrt_val:int = math.sqrt(num)
        io.println("sqrt(" + math.toString(num) + ") = " + math.toString(sqrt_val))
    }
    
    return 0
}
```

### Float Operations

```omni
import std.math as math
import std.io as io

func main():int {
    let pi:float = 3.14159
    let e:float = 2.71828
    
    io.println("=== Float Operations ===")
    io.println("pi = " + math.toString(math.abs_float(pi)))
    io.println("e = " + math.toString(math.abs_float(e)))
    io.println("max(pi, e) = " + math.toString(math.max_float(pi, e)))
    io.println("min(pi, e) = " + math.toString(math.min_float(pi, e)))
    
    return 0
}
```

## Notes

- All integer functions work with `int` type
- Float functions work with `float` type
- The `toString` function is essential for converting numbers to strings for output
- Prime checking is efficient for small to medium numbers
- Square root returns integer results (truncated)
