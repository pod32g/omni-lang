# Standard Library Implementation Status

This document tracks which standard library functions are actually implemented vs. which are stubs.

## Fully Implemented (Runtime + Backend Wiring)

### std.io
- [IMPLEMENTED] `print(value)` - Wired to `omni_print_string`
- [IMPLEMENTED] `println(value)` - Wired to `omni_println_string`
- [IMPLEMENTED] `eprint(value)` - Wired to `omni_eprint_string` (stderr)
- [IMPLEMENTED] `eprintln(value)` - Wired to `omni_eprintln_string` (stderr)
- [IMPLEMENTED] `flush()` - Wired to `omni_io_flush`
- [IMPLEMENTED] `is_terminal()` - Wired to `omni_io_is_terminal` (stdout TTY check)
- [IMPLEMENTED] `read_line()` - Wired to `omni_read_line`
- [IMPLEMENTED] `read_all()` - Wired to `omni_read_all`
- [IMPLEMENTED] `read_lines()` - Wired to `omni_io_read_lines` (split + drop trailing empty)
- [IMPLEMENTED] `read_int()` - Wired (read_line + omni_io_parse_int)
- [IMPLEMENTED] `read_float()` - Wired (read_line + omni_io_parse_float)
- [IMPLEMENTED] `prompt(message)` - Wired to `omni_io_prompt` (print + flush + read_line)
- [IMPLEMENTED] `sprint(value)` - Compile-time dispatched to omni_*_to_string
- [IMPLEMENTED] `sprintln(value)` - Wired to `omni_io_sprintln` (sprint + "\n")
- [IMPLEMENTED] `sprintf(format, args)` - Wired to `omni_io_sprintf` (`%s` substitution, `%%` escape)
- [IMPLEMENTED] `parse_int(s)` - Wired to `omni_io_parse_int` (returns 0 on failure)
- [IMPLEMENTED] `parse_float(s)` - Wired to `omni_io_parse_float` (returns 0.0 on failure)
- [IMPLEMENTED] `is_int(s)` - Wired to `omni_io_is_int` (predicate)
- [IMPLEMENTED] `is_float(s)` - Wired to `omni_io_is_float` (predicate)
- [IMPLEMENTED] `printf(format, args)` - Wired to `omni_io_printf` (sprintf + write to stdout)
- [IMPLEMENTED] `eprintf(format, args)` - Wired to `omni_io_eprintf` (sprintf + write to stderr)
- [IMPLEMENTED] `print_each(items)` - Wired to `omni_io_print_each` (one line per item)
- [IMPLEMENTED] `eprint_each(items)` - Wired to `omni_io_eprint_each` (stderr)
- [IMPLEMENTED] `eprompt(message)` - Wired to `omni_io_eprompt` (prompt to stderr)
- [IMPLEMENTED] `confirm(message)` - Wired to `omni_io_confirm` (y/n prompt)
- [IMPLEMENTED] `flush_stderr()` - Wired to `omni_io_flush_stderr`
- [IMPLEMENTED] `style(s, code)` - Wired to `omni_io_style` (generic ANSI SGR wrap)
- [IMPLEMENTED] `bold(s)`, `dim(s)`, `italic(s)`, `underline(s)` - ANSI text styles
- [IMPLEMENTED] `red(s)`, `green(s)`, `yellow(s)`, `blue(s)`, `magenta(s)`, `cyan(s)` - ANSI foreground colors
- [IMPLEMENTED] `fprint(handle, value)` - Wired to `omni_io_fprint` (write to file handle)
- [IMPLEMENTED] `fprintln(handle, value)` - Wired to `omni_io_fprintln`
- [IMPLEMENTED] `fprintf(handle, format, args)` - Wired to `omni_io_fprintf`

Surface mirrors Go's `fmt` + `bufio` + `io` as far as makes sense
without varargs, byte arrays, or Reader/Writer interfaces. ANSI
helpers always emit escape codes — gate on `is_terminal()` if you
want to skip when stdout isn't a TTY.

VM and C backend parity pinned by `tests/e2e/std_io_basic.omni`,
`std_io_read.omni`, `std_io_extras.omni`, `std_io_lines.omni`,
`std_io_format.omni`, `std_io_confirm.omni` (`TestStdIoBasic`,
`TestStdIoRead`, `TestStdIoExtras`, `TestStdIoReadLines`,
`TestStdIoFormat`, `TestStdIoConfirm`).

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
- [IMPLEMENTED] `find_all(s, substr)` - Wired to `omni_string_find_all`; returns array<int> with offsets via out-pointer length forwarding
- [IMPLEMENTED] `replace(s, old, new)` - Wired to `omni_string_replace_all` (alias)
- [IMPLEMENTED] `replace_all(s, old, new)` - Wired to `omni_string_replace_all`
- [IMPLEMENTED] `replace_first(s, old, new)` - Wired to `omni_string_replace_first`
- [IMPLEMENTED] `replace_last(s, old, new)` - Wired to `omni_string_replace_last`
- [IMPLEMENTED] `split(s, delimiter)` - Wired to `omni_string_split`; out-pointer length forwarded
- [IMPLEMENTED] `split_lines(s)` - Wired to `omni_string_split_lines`
- [IMPLEMENTED] `split_words(s)` - Wired to `omni_string_split_words` (whitespace runs)
- [IMPLEMENTED] `join(strings, delimiter)` - Wired to `omni_string_join`; receives array length companion
- [IMPLEMENTED] `join_lines(strings)` - OmniLang body delegates to `join` (still works via VM body-load)
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
- [IMPLEMENTED] `trim_left(s)` - Wired to `omni_trim_left` (C); OmniLang body for VM
- [IMPLEMENTED] `trim_right(s)` - Wired to `omni_trim_right` (C); OmniLang body for VM
- [IMPLEMENTED] `trim_all(s)` - Wired to `omni_trim_all` (C); OmniLang body for VM
- [IMPLEMENTED] `to_title(s)` - Wired to `omni_to_title` (C); OmniLang body for VM
- [IMPLEMENTED] `capitalize(s)` - Wired to `omni_capitalize` (C); OmniLang body for VM
- [IMPLEMENTED] `reverse(s)` - Wired to `omni_string_reverse` (C); OmniLang body for VM
- [IMPLEMENTED] `equals_ignore_case(a, b)` - Wired to `omni_string_equals_ignore_case` (C); OmniLang body for VM
- [IMPLEMENTED] `compare_ignore_case(a, b)` - Wired to `omni_string_compare_ignore_case` (C); OmniLang body for VM
- [IMPLEMENTED] `count_occurrences(s, substr)` - Wired to `omni_count_occurrences` (C); OmniLang body for VM
- [IMPLEMENTED] `count_lines(s)` - Wired to `omni_count_lines` (C); OmniLang body for VM
- [IMPLEMENTED] `count_words(s)` - Wired to `omni_count_words` (C); OmniLang body for VM
- [IMPLEMENTED] `is_empty(s)` - Wired to `omni_string_is_empty` (C); OmniLang body delegates to `std.string.length` for VM

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
- [IMPLEMENTED] `random_seed(seed)` - Wired to `omni_random_seed` (C); xorshift32 state, mirrored in the VM
- [IMPLEMENTED] `random_int(bound)` - Wired to `omni_random_int` (C); returns int in [0, bound)

### std.file / file
- [IMPLEMENTED] `open(filename, mode)` - Wired to `omni_file_open`
- [IMPLEMENTED] `close(handle)` - Wired to `omni_file_close`
- [IMPLEMENTED] `read(handle, buffer, size)` - Wired to `omni_file_read`; returns byte count, buffer mutation awaits a mutable byte-buffer ABI
- [IMPLEMENTED] `write(handle, buffer, size)` - Wired to `omni_file_write`
- [IMPLEMENTED] `seek(handle, offset, whence)` - Wired to `omni_file_seek`
- [IMPLEMENTED] `tell(handle)` - Wired to `omni_file_tell`
- [IMPLEMENTED] `exists(filename)` - Wired to `omni_file_exists`
- [IMPLEMENTED] `size(filename)` - Wired to `omni_file_size`
- [IMPLEMENTED] `read_all(handle)` - Wired to `omni_file_read_all_handle` (returns remaining content as string)
- [IMPLEMENTED] `read_line(handle)` - Wired to `omni_file_read_line_handle` (one line, newline stripped)
- [IMPLEMENTED] `write_string(handle, s)` - Wired to `omni_file_write_string` (returns bytes written)

### std.os
- [IMPLEMENTED] `exit(code)` - Wired to `omni_exit`
- [IMPLEMENTED] `read_file(path)` - Wired to `omni_read_file`
- [IMPLEMENTED] `write_file(path, content)` - Wired to `omni_write_file`
- [IMPLEMENTED] `append_file(path, content)` - Wired to `omni_append_file`
- [IMPLEMENTED] `read_file_lines(path)` - Wired to `omni_os_read_file_lines` (drops trailing empty)
- [IMPLEMENTED] `write_file_lines(path, lines)` - Wired to `omni_os_write_file_lines`
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

VM and C backend parity is pinned by `tests/e2e/std_os_ops.omni` /
`TestStdOsOps`, which exercises mkdir/rmdir, write/read/append, copy/
rename/remove, exists/is_file/is_dir, set/get/unsetenv, getcwd, and
getpid. `omni_getenv` and `omni_getcwd` return `""` rather than `NULL`
on missing-var / failed-syscall so the C backend can chain string ops
without crashing.

### std.log
- [IMPLEMENTED] `debug(message)` - Wired to `omni_log_debug`
- [IMPLEMENTED] `info(message)` - Wired to `omni_log_info`
- [IMPLEMENTED] `warn(message)` - Wired to `omni_log_warn`
- [IMPLEMENTED] `error(message)` - Wired to `omni_log_error`
- [IMPLEMENTED] `set_level(level)` - Wired to `omni_log_set_level`

### std.time
- [IMPLEMENTED] `now()` - VM intrinsic; C backend lowers through `omni_time_now_unix` + `omni_time_from_unix`
- [IMPLEMENTED] `unix_timestamp()` - Wired to `omni_time_now_unix`
- [IMPLEMENTED] `unix_nano()` - Wired to `omni_time_now_unix_nano`
- [IMPLEMENTED] `sleep_seconds(seconds)` - Wired to `omni_time_sleep_seconds`
- [IMPLEMENTED] `sleep_milliseconds(milliseconds)` - Wired to `omni_time_sleep_milliseconds`
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

Runtime-backed `std.time` functions above have VM and C backend smoke
coverage in `tests/e2e/std_time_ops.omni`. Pure Omni helper bodies such
as `time_create`, `time_equal`, and duration arithmetic are still a C
backend follow-up because `std.time` is not body-loaded there.

### std.collections
- [IMPLEMENTED] `size(m)` - Wired to `omni_map_size`
- [IMPLEMENTED] `get(m, key)` - Wired to `omni_map_get_string_int` (map<string, int>)
- [IMPLEMENTED] `set(m, key, value)` - Wired to `omni_map_put_string_int`
- [IMPLEMENTED] `has(m, key)` - Wired to `omni_map_has_string`
- [IMPLEMENTED] `remove(m, key)` - Wired to `omni_map_remove_string` (returns bool whether the key existed)
- [IMPLEMENTED] `clear(m)` - Wired to `omni_map_clear`
- [IMPLEMENTED] `keys(m)` - Wired to `omni_map_keys_string_int`
- [IMPLEMENTED] `values(m)` - Wired to `omni_map_values_string_int`
- [IMPLEMENTED] `copy(m)` - Wired to `omni_map_copy_string_int`
- [IMPLEMENTED] `merge(a, b)` - Wired to `omni_map_merge_string_int`
- [IMPLEMENTED] `queue_create / enqueue / dequeue / peek / is_empty / size / clear` - C backend now maps `queue<T>` → `omni_queue_t*`; VM uses a `*[]int` carrier
- [IMPLEMENTED] `stack_create / push / pop / peek / is_empty / size / clear` - C backend maps `stack<T>` → `omni_stack_t*`; VM uses a `*[]int` carrier
- [IMPLEMENTED] `set_create / add / remove / contains / size / clear / union / intersection / difference` - C backend maps `set<T>` → `omni_set_t*`; VM uses a `map[int]bool` carrier
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
- [IMPLEMENTED] `dns_lookup(hostname)` - Wired to `omni_dns_lookup` (uses getaddrinfo, supports IPv4 and IPv6)
- [IMPLEMENTED] `dns_reverse_lookup(ip)` - Wired to `omni_dns_reverse_lookup` (uses getnameinfo)
- [IMPLEMENTED] `http_get(url)` - Wired to `omni_http_get` (libcurl preferred, raw socket fallback)
- [IMPLEMENTED] `http_post(url, body)` - Wired to `omni_http_post` (libcurl preferred, raw socket fallback)
- [IMPLEMENTED] `http_put(url, body)` - Wired to `omni_http_put` (libcurl preferred, raw socket fallback)
- [IMPLEMENTED] `http_delete(url)` - Wired to `omni_http_delete` (libcurl preferred, raw socket fallback)
- [IMPLEMENTED] `http_request(req)` - Wired to `omni_http_request` (libcurl preferred, raw socket fallback)
- [IMPLEMENTED] `network_is_connected()` - Wired to `omni_network_is_connected` (platform-specific network interface checking)
- [IMPLEMENTED] `network_get_local_ip()` - Wired to `omni_network_get_local_ip` (returns first non-loopback IPv4 address)
- [IMPLEMENTED] `network_ping(host)` - Wired to `omni_network_ping` (ICMP on Windows, TCP fallback on POSIX)

Offline parts (`ip_is_valid`, `ip_parse`, `ip_is_loopback`, `ip_is_private`,
`ip_to_string`, `url_is_valid`) pinned by `tests/e2e/std_network_basic.omni` /
`TestStdNetworkBasic` on both backends. The audit caught `omni_ip_is_valid`
accepting out-of-range IPv4 segments and `omni_url_is_valid` returning true
for any string containing `://`; both runtime helpers now do real
validation. Network-touching functions (HTTP, DNS, sockets,
`network_ping`/`is_connected`/`get_local_ip`) are wired but not exercised
in CI. C-backend struct-field access for `omni_url_t*` and undeclared
`http_response_is_*` user helpers are pre-existing gaps tracked
separately.

### std.web
- [IMPLEMENTED] `server_create(port, options)` - Wired to `omni_server_create`
- [IMPLEMENTED] `server_listen(server)` - Wired to `omni_server_listen`
- [IMPLEMENTED] `server_listen_tls(server, cert_file, key_file)` - Wired to `omni_server_listen_tls` (currently falls back to `server_listen`)
- [IMPLEMENTED] `server_close(server)` - Wired to `omni_server_close`
- [IMPLEMENTED] `server_graceful_shutdown(server, timeout)` - Wired to `omni_server_graceful_shutdown`
- [IMPLEMENTED] `http_parse_request(raw_request)` - Wired to `omni_http_parse_request` (parses method, URL, headers, body)
- [IMPLEMENTED] `http_build_response(resp)` - Wired to `omni_http_build_response` (builds HTTP response string)
- [IMPLEMENTED] `http_parse_query(query_string, params)` - Wired to `omni_http_parse_query` (parses URL-encoded query strings)
- [IMPLEMENTED] `http_match_path(pattern, path, params)` - Wired to `omni_http_match_path` (matches URL patterns like `/user/:id`)
- [IMPLEMENTED] `json_parse(json_str)` - Wired to `omni_json_parse` (recursive descent parser)
- [IMPLEMENTED] `json_stringify(value, pretty)` - Wired to `omni_json_stringify` (supports int, string, float, bool, map, array)
- [IMPLEMENTED] `http_parse_form_urlencoded(body, params)` - Wired to `omni_http_parse_form_urlencoded`
- [IMPLEMENTED] `http_parse_multipart(body, boundary, fields, files)` - Wired to `omni_http_parse_multipart` (extracts form fields and file data)
- [IMPLEMENTED] `file_upload_save(data, size, filename, upload_dir)` - Wired to `omni_file_upload_save`
- [IMPLEMENTED] `file_upload_validate(filename, size, allowed_types, max_size)` - Wired to `omni_file_upload_validate`
- [IMPLEMENTED] `file_read_binary(path, size)` - Wired to `omni_file_read_binary`
- [IMPLEMENTED] `file_get_mime_type(filename)` - Wired to `omni_file_get_mime_type` (based on file extension)
- [IMPLEMENTED] `file_get_size(path)` - Wired to `omni_file_get_size`
- [PARTIAL] `http_compress_gzip(data, len, compressed_len)` - Wired to `omni_http_compress_gzip` (currently returns uncompressed data)
- [PARTIAL] `http_decompress_gzip(compressed, len, decompressed_len)` - Wired to `omni_http_decompress_gzip` (currently returns uncompressed data)
- [IMPLEMENTED] `validate_string(value, pattern, min_len, max_len)` - Wired to `omni_validate_string` (regex pattern and length validation)
- [IMPLEMENTED] `validate_int(value, min, max)` - Wired to `omni_validate_int` (integer conversion and range check)
- [IMPLEMENTED] `validate_email(email)` - Wired to `omni_validate_email` (basic email format validation)
- [IMPLEMENTED] `validate_url(url)` - Wired to `omni_validate_url` (basic URL scheme validation)
- [IMPLEMENTED] `sanitize_html(html)` - Wired to `omni_sanitize_html` (escapes HTML special characters)
- [IMPLEMENTED] `sanitize_sql(sql)` - Wired to `omni_sanitize_sql` (escapes SQL special characters)
- [PARTIAL] `websocket_handshake(request_headers)` - Wired to `omni_websocket_handshake` (simplified handshake)
- [PARTIAL] `websocket_frame_create(data, len, opcode, mask)` - Wired to `omni_websocket_frame_create` (basic frame creation)
- [PARTIAL] `websocket_frame_parse(frame, len)` - Wired to `omni_websocket_frame_parse` (basic frame parsing)
- [PARTIAL] `server_connection_pool_create(max_connections)` - Wired to `omni_server_connection_pool_create` (basic connection pool)
- [PARTIAL] `server_connection_pool_acquire(pool)` - Wired to `omni_server_connection_pool_acquire`
- [PARTIAL] `server_connection_pool_release(pool, connection)` - Wired to `omni_server_connection_pool_release`
- [PARTIAL] `server_thread_pool_create(num_threads)` - Wired to `omni_server_thread_pool_create` (tasks executed directly for now)
- [PARTIAL] `server_thread_pool_submit(pool, task)` - Wired to `omni_server_thread_pool_submit`
- [PARTIAL] `server_set_timeout(socket, timeout_seconds)` - Wired to `omni_server_set_timeout` (placeholder)
- [IMPLEMENTED] `server_set_max_request_size(max_size)` - Wired to `omni_server_set_max_request_size` (sets global limit)
- [IMPLEMENTED] `server_set_max_headers_size(max_size)` - Wired to `omni_server_set_max_headers_size` (sets global limit)
- [PARTIAL] `session_create(session_id)` - Wired to `omni_session_create` (in-memory stub)
- [PARTIAL] `session_get(session, key)` - Wired to `omni_session_get`
- [PARTIAL] `session_set(session, key, value)` - Wired to `omni_session_set`
- [PARTIAL] `session_destroy(session)` - Wired to `omni_session_destroy`
- [PARTIAL] `session_store_create(store_type)` - Wired to `omni_session_store_create` (in-memory stub)
- [PARTIAL] `session_store_save(store, session)` - Wired to `omni_session_store_save`
- [PARTIAL] `session_store_load(store, session_id)` - Wired to `omni_session_store_load`
- [PARTIAL] `auth_hash_password(password)` - Wired to `omni_auth_hash_password` (simple hashing)
- [PARTIAL] `auth_verify_password(password, hash)` - Wired to `omni_auth_verify_password`
- [PARTIAL] `auth_generate_token(user_id)` - Wired to `omni_auth_generate_token` (simple string manipulation)
- [PARTIAL] `auth_verify_token(token)` - Wired to `omni_auth_verify_token`
- [PARTIAL] `auth_check_permission(user_id, permission)` - Wired to `omni_auth_check_permission` (always returns true)
- [PARTIAL] `rate_limit_create(max_requests, window_seconds)` - Wired to `omni_rate_limit_create` (check always returns true)
- [PARTIAL] `rate_limit_check(limiter, identifier)` - Wired to `omni_rate_limit_check`
- [PARTIAL] `rate_limit_reset(limiter, identifier)` - Wired to `omni_rate_limit_reset`

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
- [IMPLEMENTED] (int arrays) `contains(arr, value)` - Wired to `omni_array_int_contains`
- [IMPLEMENTED] (int arrays) `index_of(arr, value)` - Wired to `omni_array_int_index_of`
- [IMPLEMENTED] (int arrays) `append(arr, value)` - Wired to `omni_array_int_append`; output length = input + 1
- [IMPLEMENTED] (int arrays) `prepend(arr, value)` - Wired to `omni_array_int_prepend`; output length = input + 1
- [IMPLEMENTED] (int arrays) `insert(arr, index, value)` - Wired to `omni_array_int_insert`; output length = input + 1
- [IMPLEMENTED] (int arrays) `remove(arr, index)` - Wired to `omni_array_int_remove`; output length = input - 1
- [IMPLEMENTED] (int arrays) `concat(a, b)` - Wired to `omni_array_int_concat`; output length = a + b
- [IMPLEMENTED] (int arrays) `slice(arr, start, end)` - Wired to `omni_array_int_slice`; output length = end - start
- [IMPLEMENTED] (string arrays) `contains`, `index_of`, `append`, `prepend`, `insert`, `remove`, `concat`, `slice` - Wired to `omni_array_str_*`; same shape as the int variants but element compares use strcmp and the pointer table is freshly allocated (payload strings are aliased)
- [STUB] (other element types) above ops still pass through to the input array unchanged; specialize per element type when needed
- [STUB] `fill()`, `copy()` - Not implemented (in-place mutation through parameter; needs a different ABI)
- [STUB] `reverse()` - Use `std.algorithms.reverse(arr)` instead (same shape, real implementation)
- [STUB] `length()`, `get()`, `set()` - Use the built-in `len(arr)` / `arr[i]` / `arr[i] = v` instead

### std.string
- (all previously-stubbed string helpers are now implemented; see the
  "Implemented" section above)

### std.dev
- [IMPLEMENTED] `snapshot(path)` - Implemented in OmniLang using `std.os.exists` and `std.file.size`
- [IMPLEMENTED] `changed(current, baseline)` - Implemented in OmniLang
- [IMPLEMENTED] `wait_for_change(path, poll_milliseconds)` - Implemented in OmniLang polling with `std.time.sleep_milliseconds`
- [IMPLEMENTED] `watch_loop(path, poll_milliseconds, iterations)` - Implemented in OmniLang

### std.algorithms
- [IMPLEMENTED] `euclidean_distance(x1, y1, x2, y2)` - Wired to `omni_euclidean_distance`
- [IMPLEMENTED] `manhattan_distance(x1, y1, x2, y2)` - Wired to `omni_manhattan_distance`
- [IMPLEMENTED] `levenshtein_distance(s1, s2)` - Wired to `omni_levenshtein_distance` (two-row DP, O(n) memory)
- [IMPLEMENTED] `bubble_sort(arr)`, `selection_sort(arr)`, `insertion_sort(arr)` - Wired to `omni_*_sort`; return a freshly allocated sorted copy
- [IMPLEMENTED] `linear_search(arr, target)` - Wired to `omni_linear_search`
- [IMPLEMENTED] `binary_search(arr, target)` - Wired to `omni_binary_search` (assumes `arr` is sorted ascending)
- [IMPLEMENTED] `find_max(arr)`, `find_min(arr)` - Wired to `omni_array_find_max` / `_min`
- [IMPLEMENTED] `count_occurrences(arr, value)` - Wired to `omni_array_count_occurrences`
- [IMPLEMENTED] `reverse(arr)` - Wired to `omni_array_reverse`; returns a freshly allocated reversed copy
- [IMPLEMENTED] `rotate(arr, k)` - Wired to `omni_array_rotate`; freshly allocated, normalizes k mod n
- [IMPLEMENTED] `shuffle(arr)` - Wired to `omni_array_<int|str>_shuffle`; Fisher–Yates on the shared xorshift32 PRNG. Both int and string element variants
- [IMPLEMENTED] `unique(arr)` - Wired to `omni_array_<int|str>_unique`; runtime writes the result length to a stack-allocated companion the codegen registers as the array's runtime length. Both element variants
- [STUB] `is_connected()` - Not implemented (needs a graph representation)

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
