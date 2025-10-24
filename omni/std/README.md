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
Comprehensive string manipulation functions.

**Basic Operations:**
- `length(s:string):int` - String length
- `concat(a:string, b:string):string` - Concatenate strings
- `substring(s:string, start:int, end:int):string` - Extract substring
- `char_at(s:string, index:int):char` - Get character at index

**Searching and Matching:**
- `starts_with(s:string, prefix:string):bool` - Check prefix
- `ends_with(s:string, suffix:string):bool` - Check suffix
- `contains(s:string, substr:string):bool` - Check substring
- `index_of(s:string, substr:string):int` - Find substring index
- `last_index_of(s:string, substr:string):int` - Find last substring index
- `find_all(s:string, substr:string):array<int>` - Find all indices
- `matches(s:string, pattern:string):bool` - Regex pattern matching
- `find_match(s:string, pattern:string):string` - Find first regex match
- `find_all_matches(s:string, pattern:string):array<string>` - Find all regex matches

**String Transformation:**
- `trim(s:string):string` - Remove leading/trailing whitespace
- `trim_left(s:string):string` - Remove leading whitespace
- `trim_right(s:string):string` - Remove trailing whitespace
- `trim_all(s:string):string` - Remove all whitespace
- `to_upper(s:string):string` - Convert to uppercase
- `to_lower(s:string):string` - Convert to lowercase
- `to_title(s:string):string` - Convert to title case
- `capitalize(s:string):string` - Capitalize first character
- `reverse(s:string):string` - Reverse string

**String Comparison:**
- `equals(a:string, b:string):bool` - String equality
- `equals_ignore_case(a:string, b:string):bool` - Case-insensitive equality
- `compare(a:string, b:string):int` - Lexicographic comparison
- `compare_ignore_case(a:string, b:string):int` - Case-insensitive comparison

**Splitting and Joining:**
- `split(s:string, delimiter:string):array<string>` - Split by delimiter
- `split_lines(s:string):array<string>` - Split by newlines
- `split_words(s:string):array<string>` - Split by whitespace
- `join(strings:array<string>, delimiter:string):string` - Join with delimiter
- `join_lines(strings:array<string>):string` - Join with newlines

**String Replacement:**
- `replace(s:string, old:string, new:string):string` - Replace all occurrences
- `replace_first(s:string, old:string, new:string):string` - Replace first occurrence
- `replace_last(s:string, old:string, new:string):string` - Replace last occurrence
- `replace_regex(s:string, pattern:string, replacement:string):string` - Regex replacement

**String Padding and Alignment:**
- `pad_left(s:string, length:int, pad_char:char):string` - Left padding
- `pad_right(s:string, length:int, pad_char:char):string` - Right padding
- `pad_center(s:string, length:int, pad_char:char):string` - Center padding

**String Validation:**
- `is_empty(s:string):bool` - Check if empty
- `is_blank(s:string):bool` - Check if blank (empty or whitespace)
- `is_alpha(s:string):bool` - Check if alphabetic only
- `is_digit(s:string):bool` - Check if digits only
- `is_alnum(s:string):bool` - Check if alphanumeric only
- `is_ascii(s:string):bool` - Check if ASCII only
- `is_upper(s:string):bool` - Check if uppercase
- `is_lower(s:string):bool` - Check if lowercase

**String Formatting:**
- `format(template:string, args:array<string>):string` - Format with placeholders
- `format_int(value:int, width:int, pad_char:char):string` - Format integer
- `format_float(value:float, precision:int):string` - Format float

**String Encoding and Decoding:**
- `encode_base64(s:string):string` - Base64 encoding
- `decode_base64(s:string):string` - Base64 decoding
- `encode_url(s:string):string` - URL encoding
- `decode_url(s:string):string` - URL decoding

**String Escaping:**
- `escape_html(s:string):string` - Escape HTML characters
- `unescape_html(s:string):string` - Unescape HTML entities
- `escape_json(s:string):string` - Escape JSON characters
- `escape_shell(s:string):string` - Escape shell characters

**String Statistics:**
- `count_occurrences(s:string, substr:string):int` - Count substring occurrences
- `count_lines(s:string):int` - Count lines
- `count_words(s:string):int` - Count words
- `count_chars(s:string):int` - Count characters

**String Utilities:**
- `repeat(s:string, count:int):string` - Repeat string
- `truncate(s:string, max_length:int):string` - Truncate string
- `truncate_with_ellipsis(s:string, max_length:int):string` - Truncate with ellipsis
- `remove(s:string, substr:string):string` - Remove all occurrences
- `remove_first(s:string, substr:string):string` - Remove first occurrence
- `remove_last(s:string, substr:string):string` - Remove last occurrence

**String Interpolation:**
- `interpolate(template:string, variables:map<string, string>):string` - Variable interpolation
- `template(template:string, values:array<string>):string` - Template processing

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

**Map Functions (for map<string, int>):**
- `size(m:map<string, int>):int` - Map size
- `get(m:map<string, int>, key:string):int` - Get value by key
- `set(m:map<string, int>, key:string, value:int)` - Set value by key
- `has(m:map<string, int>, key:string):bool` - Check if key exists
- `remove(m:map<string, int>, key:string):bool` - Remove key-value pair
- `clear(m:map<string, int>)` - Clear all entries
- `keys(m:map<string, int>):array<string>` - Get all keys
- `values(m:map<string, int>):array<int>` - Get all values
- `copy(m:map<string, int>):map<string, int>` - Copy map
- `merge(a:map<string, int>, b:map<string, int>):map<string, int>` - Merge maps

**Set Functions (for set<int>):**
- `set_create():set<int>` - Create new set
- `set_add(s:set<int>, element:int):bool` - Add element to set
- `set_remove(s:set<int>, element:int):bool` - Remove element from set
- `set_contains(s:set<int>, element:int):bool` - Check if set contains element
- `set_size(s:set<int>):int` - Get set size
- `set_clear(s:set<int>)` - Clear set
- `set_union(a:set<int>, b:set<int>):set<int>` - Set union
- `set_intersection(a:set<int>, b:set<int>):set<int>` - Set intersection
- `set_difference(a:set<int>, b:set<int>):set<int>` - Set difference

**Queue Functions (for queue<int>):**
- `queue_create():queue<int>` - Create new queue
- `queue_enqueue(q:queue<int>, element:int)` - Add element to queue
- `queue_dequeue(q:queue<int>):int` - Remove element from queue
- `queue_peek(q:queue<int>):int` - Peek at front element
- `queue_is_empty(q:queue<int>):bool` - Check if queue is empty
- `queue_size(q:queue<int>):int` - Get queue size
- `queue_clear(q:queue<int>)` - Clear queue

**Stack Functions (for stack<int>):**
- `stack_create():stack<int>` - Create new stack
- `stack_push(s:stack<int>, element:int)` - Push element onto stack
- `stack_pop(s:stack<int>):int` - Pop element from stack
- `stack_peek(s:stack<int>):int` - Peek at top element
- `stack_is_empty(s:stack<int>):bool` - Check if stack is empty
- `stack_size(s:stack<int>):int` - Get stack size
- `stack_clear(s:stack<int>)` - Clear stack

**Priority Queue Functions (for priority_queue<int>):**
- `priority_queue_create():priority_queue<int>` - Create new priority queue
- `priority_queue_insert(pq:priority_queue<int>, element:int, priority:int)` - Insert with priority
- `priority_queue_extract_max(pq:priority_queue<int>):int` - Extract highest priority element
- `priority_queue_peek(pq:priority_queue<int>):int` - Peek at highest priority element
- `priority_queue_is_empty(pq:priority_queue<int>):bool` - Check if priority queue is empty
- `priority_queue_size(pq:priority_queue<int>):int` - Get priority queue size

**Linked List Functions (for linked_list<int>):**
- `linked_list_create():linked_list<int>` - Create new linked list
- `linked_list_append(ll:linked_list<int>, element:int)` - Append element
- `linked_list_prepend(ll:linked_list<int>, element:int)` - Prepend element
- `linked_list_insert(ll:linked_list<int>, index:int, element:int):bool` - Insert at index
- `linked_list_remove(ll:linked_list<int>, index:int):bool` - Remove at index
- `linked_list_get(ll:linked_list<int>, index:int):int` - Get element at index
- `linked_list_set(ll:linked_list<int>, index:int, element:int):bool` - Set element at index
- `linked_list_size(ll:linked_list<int>):int` - Get linked list size
- `linked_list_is_empty(ll:linked_list<int>):bool` - Check if linked list is empty
- `linked_list_clear(ll:linked_list<int>)` - Clear linked list

**Tree Functions (for binary_tree<int>):**
- `binary_tree_create():binary_tree<int>` - Create new binary tree
- `binary_tree_insert(bt:binary_tree<int>, element:int)` - Insert element
- `binary_tree_search(bt:binary_tree<int>, element:int):bool` - Search for element
- `binary_tree_remove(bt:binary_tree<int>, element:int):bool` - Remove element
- `binary_tree_size(bt:binary_tree<int>):int` - Get tree size
- `binary_tree_is_empty(bt:binary_tree<int>):bool` - Check if tree is empty
- `binary_tree_clear(bt:binary_tree<int>)` - Clear tree

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
