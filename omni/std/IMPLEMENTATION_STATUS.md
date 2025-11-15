# Standard Library Implementation Status

This document tracks which standard library functions are actually implemented vs. which are stubs.

## Fully Implemented (Runtime + Backend Wiring)

### std.io
- [IMPLEMENTED] `print(value)` - Wired to `omni_print_string`
- [IMPLEMENTED] `println(value)` - Wired to `omni_println_string`
- [IMPLEMENTED] `read_line()` - Wired to `omni_read_line`

### std.string
- [IMPLEMENTED] `length(s)` - Wired to `omni_strlen`
- [IMPLEMENTED] `concat(a, b)` - Wired to `omni_strcat`
- [IMPLEMENTED] `substring(s, start, end)` - Wired to `omni_substring`
- [IMPLEMENTED] `char_at(s, index)` - Wired to `omni_char_at`
- [IMPLEMENTED] `starts_with(s, prefix)` - Wired to `omni_starts_with`
- [IMPLEMENTED] `ends_with(s, suffix)` - Wired to `omni_ends_with`
- [IMPLEMENTED] `contains(s, substr)` - Wired to `omni_contains`
- [IMPLEMENTED] `index_of(s, substr)` - Wired to `omni_index_of`
- [IMPLEMENTED] `last_index_of(s, substr)` - Wired to `omni_last_index_of`
- [IMPLEMENTED] `trim(s)` - Wired to `omni_trim`
- [IMPLEMENTED] `to_upper(s)` - Wired to `omni_to_upper`
- [IMPLEMENTED] `to_lower(s)` - Wired to `omni_to_lower`
- [IMPLEMENTED] `equals(a, b)` - Wired to `omni_string_equals`
- [IMPLEMENTED] `compare(a, b)` - Wired to `omni_string_compare`
- [IMPLEMENTED] `find_all(s, substr)` - Implemented in OmniLang
- [IMPLEMENTED] `replace(s, old, new)` - Implemented in OmniLang
- [IMPLEMENTED] `replace_all(s, old, new)` - Implemented in OmniLang
- [IMPLEMENTED] `replace_first(s, old, new)` - Implemented in OmniLang
- [IMPLEMENTED] `replace_last(s, old, new)` - Implemented in OmniLang
- [IMPLEMENTED] `split(s, delimiter)` - Implemented in OmniLang
- [IMPLEMENTED] `split_lines(s)` - Implemented in OmniLang
- [IMPLEMENTED] `split_words(s)` - Implemented in OmniLang
- [IMPLEMENTED] `join(strings, delimiter)` - Implemented in OmniLang
- [IMPLEMENTED] `join_lines(strings)` - Implemented in OmniLang
- [IMPLEMENTED] `pad_left(s, length, pad_char)` - Implemented in OmniLang
- [IMPLEMENTED] `pad_right(s, length, pad_char)` - Implemented in OmniLang
- [IMPLEMENTED] `pad_center(s, length, pad_char)` - Implemented in OmniLang
- [IMPLEMENTED] `is_blank(s)` - Implemented in OmniLang
- [IMPLEMENTED] `is_alpha(s)` - Implemented in OmniLang
- [IMPLEMENTED] `is_digit(s)` - Implemented in OmniLang
- [IMPLEMENTED] `is_alnum(s)` - Implemented in OmniLang
- [IMPLEMENTED] `is_ascii(s)` - Implemented in OmniLang
- [IMPLEMENTED] `is_upper(s)` - Implemented in OmniLang
- [IMPLEMENTED] `is_lower(s)` - Implemented in OmniLang
- [IMPLEMENTED] `format(template, args)` - Implemented in OmniLang
- [IMPLEMENTED] `format_int(value, width, pad_char)` - Implemented in OmniLang
- [IMPLEMENTED] `format_float(value, precision)` - Implemented in OmniLang
- [IMPLEMENTED] `repeat(s, count)` - Implemented in OmniLang
- [IMPLEMENTED] `truncate(s, max_length)` - Implemented in OmniLang
- [IMPLEMENTED] `truncate_with_ellipsis(s, max_length)` - Implemented in OmniLang
- [IMPLEMENTED] `interpolate(template, variables)` - Implemented in OmniLang
- [IMPLEMENTED] `template(template, values)` - Implemented in OmniLang
- [IMPLEMENTED] `matches(s, pattern)` - Wired to `omni_string_matches` (POSIX regex)
- [IMPLEMENTED] `find_match(s, pattern)` - Wired to `omni_string_find_match` (POSIX regex)
- [IMPLEMENTED] `find_all_matches(s, pattern)` - Wired to `omni_string_find_all_matches` (POSIX regex)
- [IMPLEMENTED] `replace_regex(s, pattern, replacement)` - Wired to `omni_string_replace_regex` (POSIX regex)
- [IMPLEMENTED] `encode_base64(s)` - Wired to `omni_encode_base64`
- [IMPLEMENTED] `decode_base64(s)` - Wired to `omni_decode_base64`
- [IMPLEMENTED] `encode_url(s)` - Wired to `omni_encode_url`
- [IMPLEMENTED] `decode_url(s)` - Wired to `omni_decode_url`
- [IMPLEMENTED] `escape_html(s)` - Wired to `omni_escape_html`
- [IMPLEMENTED] `unescape_html(s)` - Wired to `omni_unescape_html`
- [IMPLEMENTED] `escape_json(s)` - Wired to `omni_escape_json`
- [IMPLEMENTED] `escape_shell(s)` - Wired to `omni_escape_shell`

### std.math
- [IMPLEMENTED] `abs(x)` - Wired to `omni_abs` (also implemented in OmniLang)
- [IMPLEMENTED] `max(a, b)` - Wired to `omni_max` (also implemented in OmniLang)
- [IMPLEMENTED] `min(a, b)` - Wired to `omni_min` (also implemented in OmniLang)
- [IMPLEMENTED] `pow(x, y)` - Wired to `omni_pow`
- [IMPLEMENTED] `sqrt(x)` - Wired to `omni_sqrt`
- [IMPLEMENTED] `floor(x)` - Wired to `omni_floor`
- [IMPLEMENTED] `ceil(x)` - Wired to `omni_ceil`
- [IMPLEMENTED] `round(x)` - Wired to `omni_round`
- [IMPLEMENTED] `trunc(x)` - Wired to `omni_trunc`
- [IMPLEMENTED] `cbrt(x)` - Wired to `omni_cbrt`
- [IMPLEMENTED] `gcd(a, b)` - Wired to `omni_gcd`
- [IMPLEMENTED] `lcm(a, b)` - Wired to `omni_lcm`
- [IMPLEMENTED] `factorial(n)` - Wired to `omni_factorial`
- [IMPLEMENTED] `is_prime(n)` - Implemented in OmniLang
- [IMPLEMENTED] `fibonacci(n)` - Implemented in OmniLang
- [IMPLEMENTED] `sin(x)` - Wired to `omni_sin`
- [IMPLEMENTED] `cos(x)` - Wired to `omni_cos`
- [IMPLEMENTED] `tan(x)` - Wired to `omni_tan`
- [IMPLEMENTED] `asin(x)` - Wired to `omni_asin`
- [IMPLEMENTED] `acos(x)` - Wired to `omni_acos`
- [IMPLEMENTED] `atan(x)` - Wired to `omni_atan`
- [IMPLEMENTED] `atan2(y, x)` - Wired to `omni_atan2`
- [IMPLEMENTED] `exp(x)` - Wired to `omni_exp`
- [IMPLEMENTED] `log(x)` - Wired to `omni_log`
- [IMPLEMENTED] `log10(x)` - Wired to `omni_log10`
- [IMPLEMENTED] `log2(x)` - Wired to `omni_log2`
- [IMPLEMENTED] `sinh(x)` - Wired to `omni_sinh`
- [IMPLEMENTED] `cosh(x)` - Wired to `omni_cosh`
- [IMPLEMENTED] `tanh(x)` - Wired to `omni_tanh`
- [IMPLEMENTED] `mean(values)` - Implemented in OmniLang
- [IMPLEMENTED] `median(values)` - Implemented in OmniLang
- [IMPLEMENTED] `clamp(value, min, max)` - Implemented in OmniLang
- [IMPLEMENTED] `clamp_float(value, min, max)` - Implemented in OmniLang
- [IMPLEMENTED] `lerp(a, b, t)` - Implemented in OmniLang
- [IMPLEMENTED] `deg_to_rad(degrees)` - Implemented in OmniLang
- [IMPLEMENTED] `rad_to_deg(radians)` - Implemented in OmniLang

### std.file / file
- [IMPLEMENTED] `open(filename, mode)` - Wired to `omni_file_open`
- [IMPLEMENTED] `close(handle)` - Wired to `omni_file_close`
- [IMPLEMENTED] `read(handle, buffer, size)` - Wired to `omni_file_read`
- [IMPLEMENTED] `write(handle, buffer, size)` - Wired to `omni_file_write`
- [IMPLEMENTED] `seek(handle, offset, whence)` - Wired to `omni_file_seek`
- [IMPLEMENTED] `tell(handle)` - Wired to `omni_file_tell`
- [IMPLEMENTED] `exists(filename)` - Wired to `omni_file_exists`
- [IMPLEMENTED] `size(filename)` - Wired to `omni_file_size`

### std.os
- [IMPLEMENTED] `exit(code)` - Wired to `omni_exit`
- [IMPLEMENTED] `read_file(path)` - Wired to `omni_read_file`
- [IMPLEMENTED] `write_file(path, content)` - Wired to `omni_write_file`
- [IMPLEMENTED] `append_file(path, content)` - Wired to `omni_append_file`
- [IMPLEMENTED] `getenv(name)` - Wired to `omni_getenv`
- [IMPLEMENTED] `setenv(name, value)` - Wired to `omni_setenv`
- [IMPLEMENTED] `unsetenv(name)` - Wired to `omni_unsetenv`
- [IMPLEMENTED] `getcwd()` - Wired to `omni_getcwd`
- [IMPLEMENTED] `chdir(path)` - Wired to `omni_chdir`
- [IMPLEMENTED] `mkdir(path)` - Wired to `omni_mkdir`
- [IMPLEMENTED] `rmdir(path)` - Wired to `omni_rmdir`
- [IMPLEMENTED] `remove(path)` - Wired to `omni_remove`
- [IMPLEMENTED] `rename(old_path, new_path)` - Wired to `omni_rename`
- [IMPLEMENTED] `copy(src, dest)` - Wired to `omni_copy`
- [IMPLEMENTED] `exists(path)` - Wired to `omni_exists`
- [IMPLEMENTED] `is_file(path)` - Wired to `omni_is_file`
- [IMPLEMENTED] `is_dir(path)` - Wired to `omni_is_dir`
- [IMPLEMENTED] `args()` - Wired to `omni_args`
- [IMPLEMENTED] `args_count()` - Wired to `omni_args_count`
- [IMPLEMENTED] `has_flag(name)` - Wired to `omni_has_flag`
- [IMPLEMENTED] `get_flag(name, default_value)` - Wired to `omni_get_flag`
- [IMPLEMENTED] `positional_arg(index, default_value)` - Wired to `omni_positional_arg`
- [IMPLEMENTED] `getpid()` - Wired to `omni_getpid`
- [IMPLEMENTED] `getppid()` - Wired to `omni_getppid`

### std.log
- [IMPLEMENTED] `debug(message)` - Wired to `omni_log_debug`
- [IMPLEMENTED] `info(message)` - Wired to `omni_log_info`
- [IMPLEMENTED] `warn(message)` - Wired to `omni_log_warn`
- [IMPLEMENTED] `error(message)` - Wired to `omni_log_error`
- [IMPLEMENTED] `set_level(level)` - Wired to `omni_log_set_level`

### std.time
- [IMPLEMENTED] `now()` - Implemented in OmniLang (uses `unix_timestamp` and `time_from_unix`)
- [IMPLEMENTED] `unix_timestamp()` - Wired to `omni_unix_timestamp`
- [IMPLEMENTED] `unix_nano()` - Wired to `omni_unix_nano`
- [IMPLEMENTED] `sleep_seconds(seconds)` - Wired to `omni_sleep_seconds`
- [IMPLEMENTED] `sleep_milliseconds(milliseconds)` - Wired to `omni_sleep_milliseconds`
- [IMPLEMENTED] `time_zone_offset()` - Wired to `omni_time_zone_offset`
- [IMPLEMENTED] `time_zone_name()` - Wired to `omni_time_zone_name`
- [IMPLEMENTED] `time_from_unix(timestamp)` - Wired to `omni_time_from_unix`
- [IMPLEMENTED] `time_from_string(time_str)` - Wired to `omni_time_from_string`
- [IMPLEMENTED] `time_to_unix(t)` - Wired to `omni_time_to_unix`
- [IMPLEMENTED] `time_to_string(t)` - Wired to `omni_time_to_string`
- [IMPLEMENTED] `time_to_unix_nano(t)` - Wired to `omni_time_to_unix_nano`
- [IMPLEMENTED] `duration_to_string(d)` - Wired to `omni_duration_to_string`
- [PARTIAL] `time_format(t, layout)` - Basic RFC3339 formatting available via `time_to_string`, custom layouts pending
- [PARTIAL] `time_parse(time_str, layout)` - Basic RFC3339 parsing available via `time_from_string`, custom layouts pending

### std.collections
- [IMPLEMENTED] `keys(m)` - Wired to `omni_map_keys_string_int`
- [IMPLEMENTED] `values(m)` - Wired to `omni_map_values_string_int`
- [IMPLEMENTED] `copy(m)` - Wired to `omni_map_copy_string_int`
- [IMPLEMENTED] `merge(a, b)` - Wired to `omni_map_merge_string_int`
- [IMPLEMENTED] `set_create()` - Wired to `omni_set_create`
- [IMPLEMENTED] `set_add(s, element)` - Wired to `omni_set_add`
- [IMPLEMENTED] `set_remove(s, element)` - Wired to `omni_set_remove`
- [IMPLEMENTED] `set_contains(s, element)` - Wired to `omni_set_contains`
- [IMPLEMENTED] `set_size(s)` - Wired to `omni_set_size`
- [IMPLEMENTED] `set_clear(s)` - Wired to `omni_set_clear`
- [IMPLEMENTED] `set_union(a, b)` - Wired to `omni_set_union`
- [IMPLEMENTED] `set_intersection(a, b)` - Wired to `omni_set_intersection`
- [IMPLEMENTED] `set_difference(a, b)` - Wired to `omni_set_difference`
- [IMPLEMENTED] `queue_create()` - Wired to `omni_queue_create`
- [IMPLEMENTED] `queue_enqueue(q, element)` - Wired to `omni_queue_enqueue`
- [IMPLEMENTED] `queue_dequeue(q)` - Wired to `omni_queue_dequeue`
- [IMPLEMENTED] `queue_peek(q)` - Wired to `omni_queue_peek`
- [IMPLEMENTED] `queue_is_empty(q)` - Wired to `omni_queue_is_empty`
- [IMPLEMENTED] `queue_size(q)` - Wired to `omni_queue_size`
- [IMPLEMENTED] `queue_clear(q)` - Wired to `omni_queue_clear`
- [IMPLEMENTED] `stack_create()` - Wired to `omni_stack_create`
- [IMPLEMENTED] `stack_push(s, element)` - Wired to `omni_stack_push`
- [IMPLEMENTED] `stack_pop(s)` - Wired to `omni_stack_pop`
- [IMPLEMENTED] `stack_peek(s)` - Wired to `omni_stack_peek`
- [IMPLEMENTED] `stack_is_empty(s)` - Wired to `omni_stack_is_empty`
- [IMPLEMENTED] `stack_size(s)` - Wired to `omni_stack_size`
- [IMPLEMENTED] `stack_clear(s)` - Wired to `omni_stack_clear`
- [IMPLEMENTED] `priority_queue_create()` - Wired to `omni_priority_queue_create`
- [IMPLEMENTED] `priority_queue_insert(pq, element, priority)` - Wired to `omni_priority_queue_insert`
- [IMPLEMENTED] `priority_queue_extract_max(pq)` - Wired to `omni_priority_queue_extract_max`
- [IMPLEMENTED] `priority_queue_peek(pq)` - Wired to `omni_priority_queue_peek`
- [IMPLEMENTED] `priority_queue_is_empty(pq)` - Wired to `omni_priority_queue_is_empty`
- [IMPLEMENTED] `priority_queue_size(pq)` - Wired to `omni_priority_queue_size`
- [IMPLEMENTED] `linked_list_create()` - Wired to `omni_linked_list_create`
- [IMPLEMENTED] `linked_list_append(ll, element)` - Wired to `omni_linked_list_append`
- [IMPLEMENTED] `linked_list_prepend(ll, element)` - Wired to `omni_linked_list_prepend`
- [IMPLEMENTED] `linked_list_insert(ll, index, element)` - Wired to `omni_linked_list_insert`
- [IMPLEMENTED] `linked_list_remove(ll, index)` - Wired to `omni_linked_list_remove`
- [IMPLEMENTED] `linked_list_get(ll, index)` - Wired to `omni_linked_list_get`
- [IMPLEMENTED] `linked_list_set(ll, index, element)` - Wired to `omni_linked_list_set`
- [IMPLEMENTED] `linked_list_size(ll)` - Wired to `omni_linked_list_size`
- [IMPLEMENTED] `linked_list_is_empty(ll)` - Wired to `omni_linked_list_is_empty`
- [IMPLEMENTED] `linked_list_clear(ll)` - Wired to `omni_linked_list_clear`
- [IMPLEMENTED] `binary_tree_create()` - Wired to `omni_binary_tree_create`
- [IMPLEMENTED] `binary_tree_insert(bt, element)` - Wired to `omni_binary_tree_insert`
- [IMPLEMENTED] `binary_tree_search(bt, element)` - Wired to `omni_binary_tree_search`
- [IMPLEMENTED] `binary_tree_remove(bt, element)` - Wired to `omni_binary_tree_remove`
- [IMPLEMENTED] `binary_tree_size(bt)` - Wired to `omni_binary_tree_size`
- [IMPLEMENTED] `binary_tree_is_empty(bt)` - Wired to `omni_binary_tree_is_empty`
- [IMPLEMENTED] `binary_tree_clear(bt)` - Wired to `omni_binary_tree_clear`

### std.network
- [IMPLEMENTED] `ip_parse(ip_str)` - Wired to `omni_ip_parse`
- [IMPLEMENTED] `ip_is_valid(ip_str)` - Wired to `omni_ip_is_valid`
- [IMPLEMENTED] `ip_is_private(ip)` - Wired to `omni_ip_is_private`
- [IMPLEMENTED] `ip_is_loopback(ip)` - Wired to `omni_ip_is_loopback`
- [IMPLEMENTED] `ip_to_string(ip)` - Implemented in OmniLang
- [IMPLEMENTED] `url_parse(url_str)` - Wired to `omni_url_parse`
- [IMPLEMENTED] `url_to_string(url)` - Implemented in OmniLang (also wired to `omni_url_to_string`)
- [IMPLEMENTED] `url_is_valid(url_str)` - Wired to `omni_url_is_valid`
- [IMPLEMENTED] `socket_create()` - Wired to `omni_socket_create`
- [IMPLEMENTED] `socket_connect(socket, address, port)` - Wired to `omni_socket_connect`
- [IMPLEMENTED] `socket_bind(socket, address, port)` - Wired to `omni_socket_bind`
- [IMPLEMENTED] `socket_listen(socket, backlog)` - Wired to `omni_socket_listen`
- [IMPLEMENTED] `socket_accept(socket)` - Wired to `omni_socket_accept`
- [IMPLEMENTED] `socket_send(socket, data)` - Wired to `omni_socket_send`
- [IMPLEMENTED] `socket_receive(socket, buffer_size)` - Wired to `omni_socket_receive`
- [IMPLEMENTED] `socket_close(socket)` - Wired to `omni_socket_close`
- [PARTIAL] `dns_lookup(hostname)` - Stub implementation (returns empty array)
- [PARTIAL] `dns_reverse_lookup(ip)` - Stub implementation (returns empty string)
- [PARTIAL] `http_get(url)` - Stub implementation (returns default HTTPResponse)
- [PARTIAL] `http_post(url, body)` - Stub implementation (returns default HTTPResponse)
- [PARTIAL] `http_put(url, body)` - Stub implementation (returns default HTTPResponse)
- [PARTIAL] `http_delete(url)` - Stub implementation (returns default HTTPResponse)
- [PARTIAL] `http_request(req)` - Stub implementation (returns default HTTPResponse)
- [PARTIAL] `network_is_connected()` - Stub implementation (returns false)
- [PARTIAL] `network_get_local_ip()` - Stub implementation (returns localhost)
- [PARTIAL] `network_ping(host)` - Stub implementation (returns false)

### Type Conversions
- [IMPLEMENTED] `std.int_to_string(i)` - Wired to `omni_int_to_string`
- [IMPLEMENTED] `std.float_to_string(f)` - Wired to `omni_float_to_string`
- [IMPLEMENTED] `std.bool_to_string(b)` - Wired to `omni_bool_to_string`
- [IMPLEMENTED] `std.string_to_int(s)` - Wired to `omni_string_to_int`
- [IMPLEMENTED] `std.string_to_float(s)` - Wired to `omni_string_to_float`
- [IMPLEMENTED] `std.string_to_bool(s)` - Wired to `omni_string_to_bool`

### Testing
- [IMPLEMENTED] `std.test.start(name)` - Wired to `omni_test_start`
- [IMPLEMENTED] `std.test.end(name, passed)` - Wired to `omni_test_end`
- [IMPLEMENTED] `std.assert(condition, message)` - Wired to `omni_assert`

## Stubs (No Runtime Implementation)

### std.array
- [STUB] All generic array functions - Not implemented (arrays are fixed-size in C)
- [STUB] `append()`, `prepend()`, `insert()`, `remove()` - Not implemented
- [STUB] `slice()`, `concat()`, `fill()`, `copy()` - Not implemented
- [STUB] `reverse()`, `sort()` - Not implemented

### std.string
- [STUB] `trim_left()`, `trim_right()`, `trim_all()` - Not implemented
- [STUB] `to_title()`, `capitalize()`, `reverse()` - Not implemented
- [STUB] `equals_ignore_case()`, `compare_ignore_case()` - Not implemented
- [STUB] `count_occurrences()`, `count_lines()`, `count_words()` - Not implemented

### std.dev
- [STUB] All dev functions - Not implemented
- [STUB] `snapshot()`, `wait_for_change()`, `changed()`, `watch_loop()` - Not implemented

### std.algorithms
- [STUB] All algorithm functions - Not implemented
- [STUB] `bubble_sort()`, `selection_sort()`, `insertion_sort()` - Not implemented
- [STUB] `linear_search()`, `binary_search()` - Not implemented
- [STUB] `reverse()`, `rotate()`, `shuffle()` - Not implemented
- [STUB] `euclidean_distance()`, `manhattan_distance()`, `levenshtein_distance()` - Not implemented
- [STUB] `find_max()`, `find_min()`, `count_occurrences()`, `unique()` - Not implemented
- [STUB] `is_connected()` - Not implemented

## Warnings

When you call an unimplemented stdlib function, the compiler will now emit a warning:
```
WARNING: stdlib function 'std.string.find_all' is called but has no runtime implementation. 
It will return a default value or do nothing.
```

## Implementation Notes

- Functions marked as "intrinsic" in comments are supposed to be wired to runtime functions
- The backend now verifies that functions marked as runtime-provided actually have implementations
- Functions without implementations will use their stub bodies, which return default values
- This can lead to incorrect program behavior - always check warnings!
- Functions marked as [PARTIAL] have basic implementations but may lack full feature support
