# OmniLang Standard Library

This directory contains the standard library for OmniLang, providing essential functions and utilities for common programming tasks.

## Modules

### std.io
Input/Output functions for console operations.

**Functions:**
- `print(s:string)` - Print string without newline
- `println(s:string)` - Print string with newline
- `print_int(i:int)` - Print integer without newline
- `println_int(i:int)` - Print integer with newline
- `print_float(f:float)` - Print float without newline
- `println_float(f:float)` - Print float with newline
- `print_bool(b:bool)` - Print boolean without newline
- `println_bool(b:bool)` - Print boolean with newline

### std.math
Mathematical functions and utilities.

**Basic Arithmetic:**
- `abs(x:int):int` - Absolute value of integer
- `abs_float(x:float):float` - Absolute value of float
- `max(a:int, b:int):int` - Maximum of two integers
- `min(a:int, b:int):int` - Minimum of two integers
- `max_float(a:float, b:float):float` - Maximum of two floats
- `min_float(a:float, b:float):float` - Minimum of two floats

**Power and Root Functions:**
- `pow(x:float, y:float):float` - Power function
- `pow_int(x:int, y:int):int` - Integer power function
- `sqrt(x:float):float` - Square root
- `cbrt(x:float):float` - Cube root

**Rounding Functions:**
- `floor(x:float):float` - Floor function
- `ceil(x:float):float` - Ceiling function
- `round(x:float):float` - Round function
- `trunc(x:float):float` - Truncate function

**Number Theory:**
- `gcd(a:int, b:int):int` - Greatest common divisor
- `lcm(a:int, b:int):int` - Least common multiple
- `is_prime(n:int):bool` - Check if number is prime
- `factorial(n:int):int` - Factorial function
- `fibonacci(n:int):int` - Fibonacci number

**Trigonometric Functions:**
- `sin(x:float):float` - Sine function
- `cos(x:float):float` - Cosine function
- `tan(x:float):float` - Tangent function
- `asin(x:float):float` - Arcsine function
- `acos(x:float):float` - Arccosine function
- `atan(x:float):float` - Arctangent function
- `atan2(y:float, x:float):float` - Two-argument arctangent

**Logarithmic and Exponential:**
- `exp(x:float):float` - Exponential function
- `log(x:float):float` - Natural logarithm
- `log10(x:float):float` - Base-10 logarithm
- `log2(x:float):float` - Base-2 logarithm

**Hyperbolic Functions:**
- `sinh(x:float):float` - Hyperbolic sine
- `cosh(x:float):float` - Hyperbolic cosine
- `tanh(x:float):float` - Hyperbolic tangent

**Statistical Functions:**
- `mean(values:array<int>):float` - Arithmetic mean
- `median(values:array<int>):float` - Median value

**Utility Functions:**
- `clamp(value:int, min_val:int, max_val:int):int` - Clamp value
- `clamp_float(value:float, min_val:float, max_val:float):float` - Clamp float
- `lerp(a:float, b:float, t:float):float` - Linear interpolation
- `deg_to_rad(degrees:float):float` - Convert degrees to radians
- `rad_to_deg(radians:float):float` - Convert radians to degrees

### std.string
String manipulation functions.

**Functions:**
- `length(s:string):int` - String length
- `concat(a:string, b:string):string` - Concatenate strings
- `substring(s:string, start:int, end:int):string` - Extract substring
- `char_at(s:string, index:int):char` - Get character at index
- `starts_with(s:string, prefix:string):bool` - Check prefix
- `ends_with(s:string, suffix:string):bool` - Check suffix
- `contains(s:string, substr:string):bool` - Check substring
- `index_of(s:string, substr:string):int` - Find substring index
- `last_index_of(s:string, substr:string):int` - Find last substring index
- `trim(s:string):string` - Remove whitespace
- `to_upper(s:string):string` - Convert to uppercase
- `to_lower(s:string):string` - Convert to lowercase
- `equals(a:string, b:string):bool` - String equality
- `compare(a:string, b:string):int` - String comparison

### std.array
Array manipulation functions.

**Functions:**
- `length<T>(arr:array<T>):int` - Array length
- `get<T>(arr:array<T>, index:int):T` - Get element at index
- `set<T>(arr:array<T>, index:int, value:T)` - Set element at index
- `append<T>(arr:array<T>, value:T):array<T>` - Append element
- `prepend<T>(arr:array<T>, value:T):array<T>` - Prepend element
- `insert<T>(arr:array<T>, index:int, value:T):array<T>` - Insert element
- `remove<T>(arr:array<T>, index:int):array<T>` - Remove element
- `contains<T>(arr:array<T>, value:T):bool` - Check if contains value
- `index_of<T>(arr:array<T>, value:T):int` - Find value index
- `reverse<T>(arr:array<T>):array<T>` - Reverse array
- `slice<T>(arr:array<T>, start:int, end:int):array<T>` - Extract slice
- `concat<T>(a:array<T>, b:array<T>):array<T>` - Concatenate arrays
- `fill<T>(arr:array<T>, value:T)` - Fill array with value
- `copy<T>(src:array<T>, dest:array<T>, count:int)` - Copy elements

### std.os
Operating system interface functions.

**Functions:**
- `exit(code:int)` - Terminate program
- `getenv(name:string):string` - Get environment variable
- `setenv(name:string, value:string):bool` - Set environment variable
- `unsetenv(name:string):bool` - Remove environment variable
- `getpid():int` - Get process ID
- `getppid():int` - Get parent process ID
- `getcwd():string` - Get current working directory
- `chdir(path:string):bool` - Change directory
- `mkdir(path:string):bool` - Create directory
- `rmdir(path:string):bool` - Remove directory
- `exists(path:string):bool` - Check if path exists
- `is_file(path:string):bool` - Check if path is file
- `is_dir(path:string):bool` - Check if path is directory
- `remove(path:string):bool` - Remove file/directory
- `rename(old_path:string, new_path:string):bool` - Rename file/directory
- `copy(src:string, dest:string):bool` - Copy file
- `read_file(path:string):string` - Read file contents
- `write_file(path:string, contents:string):bool` - Write file contents
- `append_file(path:string, contents:string):bool` - Append to file

### std.collections
Collection data structures.

**Map Functions:**
- `size<K,V>(m:map<K,V>):int` - Map size
- `get<K,V>(m:map<K,V>, key:K):V` - Get value by key
- `set<K,V>(m:map<K,V>, key:K, value:V)` - Set value by key
- `has<K,V>(m:map<K,V>, key:K):bool` - Check if key exists
- `remove<K,V>(m:map<K,V>, key:K):bool` - Remove key-value pair
- `clear<K,V>(m:map<K,V>)` - Clear all entries
- `keys<K,V>(m:map<K,V>):array<K>` - Get all keys
- `values<K,V>(m:map<K,V>):array<V>` - Get all values
- `copy<K,V>(m:map<K,V>):map<K,V>` - Copy map
- `merge<K,V>(a:map<K,V>, b:map<K,V>):map<K,V>` - Merge maps

**Set Functions:**
- `set_create<T>():set<T>` - Create new set
- `set_add<T>(s:set<T>, element:T):bool` - Add element to set
- `set_remove<T>(s:set<T>, element:T):bool` - Remove element from set
- `set_contains<T>(s:set<T>, element:T):bool` - Check if set contains element
- `set_size<T>(s:set<T>):int` - Get set size
- `set_clear<T>(s:set<T>)` - Clear set
- `set_union<T>(a:set<T>, b:set<T>):set<T>` - Set union
- `set_intersection<T>(a:set<T>, b:set<T>):set<T>` - Set intersection
- `set_difference<T>(a:set<T>, b:set<T>):set<T>` - Set difference

**Queue Functions:**
- `queue_create<T>():queue<T>` - Create new queue
- `queue_enqueue<T>(q:queue<T>, element:T)` - Add element to queue
- `queue_dequeue<T>(q:queue<T>):T` - Remove element from queue
- `queue_peek<T>(q:queue<T>):T` - Peek at front element
- `queue_is_empty<T>(q:queue<T>):bool` - Check if queue is empty
- `queue_size<T>(q:queue<T>):int` - Get queue size
- `queue_clear<T>(q:queue<T>)` - Clear queue

**Stack Functions:**
- `stack_create<T>():stack<T>` - Create new stack
- `stack_push<T>(s:stack<T>, element:T)` - Push element onto stack
- `stack_pop<T>(s:stack<T>):T` - Pop element from stack
- `stack_peek<T>(s:stack<T>):T` - Peek at top element
- `stack_is_empty<T>(s:stack<T>):bool` - Check if stack is empty
- `stack_size<T>(s:stack<T>):int` - Get stack size
- `stack_clear<T>(s:stack<T>)` - Clear stack

**Priority Queue Functions:**
- `priority_queue_create<T>():priority_queue<T>` - Create new priority queue
- `priority_queue_insert<T>(pq:priority_queue<T>, element:T, priority:int)` - Insert with priority
- `priority_queue_extract_max<T>(pq:priority_queue<T>):T` - Extract highest priority element
- `priority_queue_peek<T>(pq:priority_queue<T>):T` - Peek at highest priority element
- `priority_queue_is_empty<T>(pq:priority_queue<T>):bool` - Check if priority queue is empty
- `priority_queue_size<T>(pq:priority_queue<T>):int` - Get priority queue size

**Linked List Functions:**
- `linked_list_create<T>():linked_list<T>` - Create new linked list
- `linked_list_append<T>(ll:linked_list<T>, element:T)` - Append element
- `linked_list_prepend<T>(ll:linked_list<T>, element:T)` - Prepend element
- `linked_list_insert<T>(ll:linked_list<T>, index:int, element:T):bool` - Insert at index
- `linked_list_remove<T>(ll:linked_list<T>, index:int):bool` - Remove at index
- `linked_list_get<T>(ll:linked_list<T>, index:int):T` - Get element at index
- `linked_list_set<T>(ll:linked_list<T>, index:int, element:T):bool` - Set element at index
- `linked_list_size<T>(ll:linked_list<T>):int` - Get linked list size
- `linked_list_is_empty<T>(ll:linked_list<T>):bool` - Check if linked list is empty
- `linked_list_clear<T>(ll:linked_list<T>)` - Clear linked list

**Tree Functions:**
- `binary_tree_create<T>():binary_tree<T>` - Create new binary tree
- `binary_tree_insert<T>(bt:binary_tree<T>, element:T)` - Insert element
- `binary_tree_search<T>(bt:binary_tree<T>, element:T):bool` - Search for element
- `binary_tree_remove<T>(bt:binary_tree<T>, element:T):bool` - Remove element
- `binary_tree_size<T>(bt:binary_tree<T>):int` - Get tree size
- `binary_tree_is_empty<T>(bt:binary_tree<T>):bool` - Check if tree is empty
- `binary_tree_clear<T>(bt:binary_tree<T>)` - Clear tree

### std.algorithms
Common algorithms for sorting, searching, and data manipulation.

**Sorting Algorithms:**
- `bubble_sort(arr:array<int>):array<int>` - Bubble sort
- `selection_sort(arr:array<int>):array<int>` - Selection sort
- `insertion_sort(arr:array<int>):array<int>` - Insertion sort

**Searching Algorithms:**
- `linear_search(arr:array<int>, target:int):int` - Linear search
- `binary_search(arr:array<int>, target:int):int` - Binary search

**Array Manipulation:**
- `reverse<T>(arr:array<T>):array<T>` - Reverse array
- `rotate<T>(arr:array<T>, k:int):array<T>` - Rotate array
- `shuffle<T>(arr:array<T>):array<T>` - Shuffle array

**Mathematical Algorithms:**
- `euclidean_distance(x1:float, y1:float, x2:float, y2:float):float` - Euclidean distance
- `manhattan_distance(x1:float, y1:float, x2:float, y2:float):float` - Manhattan distance

**String Algorithms:**
- `levenshtein_distance(s1:string, s2:string):int` - Levenshtein distance

**Utility Algorithms:**
- `find_max(arr:array<int>):int` - Find maximum element
- `find_min(arr:array<int>):int` - Find minimum element
- `count_occurrences<T>(arr:array<T>, value:T):int` - Count occurrences
- `unique<T>(arr:array<T>):array<T>` - Remove duplicates

**Graph Algorithms:**
- `is_connected(adjacency_list:array<array<int>>):bool` - Check graph connectivity

### std.time
Time and date utilities.

**Time Structure:**
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
```

**Current Time Functions:**
- `now():Time` - Get current time
- `unix_timestamp():int` - Get current Unix timestamp
- `unix_nano():int` - Get current Unix timestamp in nanoseconds

**Time Creation:**
- `time_create(year:int, month:int, day:int, hour:int, minute:int, second:int):Time` - Create time
- `time_from_unix(timestamp:int):Time` - Create time from Unix timestamp
- `time_from_string(time_str:string):Time` - Parse time from string

**Time Conversion:**
- `time_to_unix(t:Time):int` - Convert time to Unix timestamp
- `time_to_string(t:Time):string` - Convert time to string
- `time_to_unix_nano(t:Time):int` - Convert time to Unix nanoseconds

**Time Comparison:**
- `time_equal(t1:Time, t2:Time):bool` - Check if times are equal
- `time_before(t1:Time, t2:Time):bool` - Check if t1 is before t2
- `time_after(t1:Time, t2:Time):bool` - Check if t1 is after t2

**Time Arithmetic:**
- `time_add_duration(t:Time, duration:Duration):Time` - Add duration to time
- `time_sub_duration(t:Time, duration:Duration):Time` - Subtract duration from time
- `time_sub_time(t1:Time, t2:Time):Duration` - Calculate duration between times

**Duration Structure:**
```omni
struct Duration {
    seconds:int
    nanoseconds:int
}
```

**Duration Creation:**
- `duration_create(seconds:int, nanoseconds:int):Duration` - Create duration
- `duration_from_seconds(seconds:float):Duration` - Create from seconds
- `duration_from_milliseconds(milliseconds:int):Duration` - Create from milliseconds
- `duration_from_minutes(minutes:float):Duration` - Create from minutes
- `duration_from_hours(hours:float):Duration` - Create from hours
- `duration_from_days(days:float):Duration` - Create from days

**Duration Conversion:**
- `duration_to_seconds(d:Duration):float` - Convert to seconds
- `duration_to_milliseconds(d:Duration):int` - Convert to milliseconds
- `duration_to_minutes(d:Duration):float` - Convert to minutes
- `duration_to_hours(d:Duration):float` - Convert to hours
- `duration_to_days(d:Duration):float` - Convert to days
- `duration_to_string(d:Duration):string` - Convert to string

**Duration Arithmetic:**
- `duration_add(d1:Duration, d2:Duration):Duration` - Add durations
- `duration_sub(d1:Duration, d2:Duration):Duration` - Subtract durations
- `duration_mul(d:Duration, scalar:float):Duration` - Multiply duration by scalar
- `duration_div(d:Duration, scalar:float):Duration` - Divide duration by scalar

**Duration Comparison:**
- `duration_equal(d1:Duration, d2:Duration):bool` - Check if durations are equal
- `duration_less(d1:Duration, d2:Duration):bool` - Check if d1 is less than d2
- `duration_greater(d1:Duration, d2:Duration):bool` - Check if d1 is greater than d2

**Utility Functions:**
- `sleep(duration:Duration)` - Sleep for duration
- `sleep_seconds(seconds:float)` - Sleep for seconds
- `sleep_milliseconds(milliseconds:int)` - Sleep for milliseconds

**Time Zone Functions:**
- `time_zone_offset():int` - Get time zone offset
- `time_zone_name():string` - Get time zone name

**Formatting Functions:**
- `time_format(t:Time, layout:string):string` - Format time
- `time_parse(time_str:string, layout:string):Time` - Parse time

**Duration Constants:**
- `NANOSECOND`, `MICROSECOND`, `MILLISECOND`, `SECOND`, `MINUTE`, `HOUR`, `DAY`, `WEEK`, `MONTH`, `YEAR`

### std.network
Basic networking utilities.

**Network Structures:**
```omni
struct IPAddress {
    address:string
    is_ipv4:bool
    is_ipv6:bool
}

struct URL {
    scheme:string
    host:string
    port:int
    path:string
    query:string
    fragment:string
}

struct HTTPRequest {
    method:string
    url:string
    headers:map<string, string>
    body:string
}

struct HTTPResponse {
    status_code:int
    status_text:string
    headers:map<string, string>
    body:string
}
```

**IP Address Functions:**
- `ip_parse(ip_str:string):IPAddress` - Parse IP address
- `ip_is_valid(ip_str:string):bool` - Check if IP is valid
- `ip_is_private(ip:IPAddress):bool` - Check if IP is private
- `ip_is_loopback(ip:IPAddress):bool` - Check if IP is loopback
- `ip_to_string(ip:IPAddress):string` - Convert IP to string

**URL Functions:**
- `url_parse(url_str:string):URL` - Parse URL
- `url_to_string(url:URL):string` - Convert URL to string
- `url_is_valid(url_str:string):bool` - Check if URL is valid

**DNS Functions:**
- `dns_lookup(hostname:string):array<IPAddress>` - DNS lookup
- `dns_reverse_lookup(ip:IPAddress):string` - Reverse DNS lookup

**HTTP Client Functions:**
- `http_get(url:string):HTTPResponse` - HTTP GET request
- `http_post(url:string, body:string):HTTPResponse` - HTTP POST request
- `http_put(url:string, body:string):HTTPResponse` - HTTP PUT request
- `http_delete(url:string):HTTPResponse` - HTTP DELETE request
- `http_request(req:HTTPRequest):HTTPResponse` - Custom HTTP request

**HTTP Response Functions:**
- `http_response_is_success(resp:HTTPResponse):bool` - Check if response is successful
- `http_response_is_client_error(resp:HTTPResponse):bool` - Check if client error
- `http_response_is_server_error(resp:HTTPResponse):bool` - Check if server error
- `http_response_get_header(resp:HTTPResponse, name:string):string` - Get header
- `http_response_set_header(resp:HTTPResponse, name:string, value:string):HTTPResponse` - Set header

**HTTP Request Functions:**
- `http_request_create(method:string, url:string):HTTPRequest` - Create request
- `http_request_set_header(req:HTTPRequest, name:string, value:string):HTTPRequest` - Set header
- `http_request_set_body(req:HTTPRequest, body:string):HTTPRequest` - Set body
- `http_request_get_header(req:HTTPRequest, name:string):string` - Get header

**Socket Functions:**
- `socket_create():int` - Create socket
- `socket_connect(socket:int, address:string, port:int):bool` - Connect socket
- `socket_bind(socket:int, address:string, port:int):bool` - Bind socket
- `socket_listen(socket:int, backlog:int):bool` - Listen on socket
- `socket_accept(socket:int):int` - Accept connection
- `socket_send(socket:int, data:string):int` - Send data
- `socket_receive(socket:int, buffer_size:int):string` - Receive data
- `socket_close(socket:int):bool` - Close socket

**Utility Functions:**
- `network_is_connected():bool` - Check if network is connected
- `network_get_local_ip():IPAddress` - Get local IP address
- `network_ping(host:string):bool` - Ping host

**Constants:**
- HTTP status codes: `HTTP_OK`, `HTTP_CREATED`, `HTTP_BAD_REQUEST`, `HTTP_NOT_FOUND`, etc.
- HTTP methods: `HTTP_GET`, `HTTP_POST`, `HTTP_PUT`, `HTTP_DELETE`, etc.
- Common ports: `PORT_HTTP`, `PORT_HTTPS`, `PORT_SSH`, etc.

## Usage

Import the standard library in your OmniLang programs:

```omni
import std

func main():int {
    std.io.println("Hello, World!")
    let result:int = std.math.max(10, 20)
    std.io.println_int(result)
    return 0
}
```

Or import specific modules:

```omni
import std.io
import std.math

func main():int {
    io.println("Hello, World!")
    let result:int = math.max(10, 20)
    io.println_int(result)
    return 0
}
```

## Implementation Status

⚠️ **Note**: Most standard library functions are currently declared as intrinsic functions that need to be implemented in the runtime backends. The current implementation provides:

- ✅ Function declarations and type signatures
- ✅ Basic math functions (implemented in Go)
- ❌ String manipulation (needs runtime implementation)
- ❌ Array operations (needs runtime implementation)
- ❌ OS operations (needs runtime implementation)
- ❌ Collection operations (needs runtime implementation)

## Examples

See the `examples/` directory for demonstration programs using the standard library.

## Contributing

To add new standard library functions:

1. Add the function declaration to the appropriate module
2. Implement the function in the runtime backends (VM and Cranelift)
3. Add tests for the new functionality
4. Update this documentation
