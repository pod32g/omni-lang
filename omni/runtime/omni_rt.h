#ifndef OMNI_RT_H
#define OMNI_RT_H

#include <stdint.h>
#include <stdio.h>
#include <strings.h>  // For strcasecmp on POSIX systems

// Optional libcurl support for HTTP client
#ifdef HAVE_LIBCURL
#include <curl/curl.h>
#endif

// strcasecmp declaration for Windows compatibility
#ifdef _WIN32
#ifndef strcasecmp
#define strcasecmp _stricmp
#endif
#endif

// OmniLang Runtime Library
// This provides the runtime support for OmniLang programs

// Basic I/O functions
void omni_print_string(const char* str);
void omni_println_string(const char* str);
void omni_eprint_string(const char* str);
void omni_eprintln_string(const char* str);
void omni_io_flush(void);
int32_t omni_io_is_terminal(void);
char* omni_read_line(void);
char* omni_read_all(void);
char* omni_io_sprintf(const char* format, const char** args, int32_t args_len);
int32_t omni_io_parse_int(const char* str);
double omni_io_parse_float(const char* str);
int32_t omni_io_is_int(const char* str);
int32_t omni_io_is_float(const char* str);
char* omni_io_prompt(const char* message);
const char** omni_io_read_lines(int32_t* out_len);
char* omni_io_sprintln(const char* str);
void omni_io_printf(const char* format, const char** args, int32_t args_len);
void omni_io_eprintf(const char* format, const char** args, int32_t args_len);
void omni_io_print_each(const char** items, int32_t n);
void omni_io_eprint_each(const char** items, int32_t n);
char* omni_io_eprompt(const char* message);
int32_t omni_io_confirm(const char* message);
void omni_io_flush_stderr(void);
char* omni_io_style(const char* s, const char* code);
char* omni_io_bold(const char* s);
char* omni_io_dim(const char* s);
char* omni_io_italic(const char* s);
char* omni_io_underline(const char* s);
char* omni_io_red(const char* s);
char* omni_io_green(const char* s);
char* omni_io_yellow(const char* s);
char* omni_io_blue(const char* s);
char* omni_io_magenta(const char* s);
char* omni_io_cyan(const char* s);

// Logging functions
void omni_log_debug(const char* message);
void omni_log_info(const char* message);
void omni_log_warn(const char* message);
void omni_log_error(const char* message);
int32_t omni_log_set_level(const char* level);

// Memory management
void* omni_alloc(size_t size);
void omni_free(void* ptr);
void* omni_malloc(size_t size);
void* omni_realloc(void* ptr, size_t new_size);

// String operations
char* omni_strcat(const char* str1, const char* str2);
int32_t omni_strlen(const char* str);
char* omni_substring(const char* str, int32_t start, int32_t end);
char omni_char_at(const char* str, int32_t index);
int32_t omni_starts_with(const char* str, const char* prefix);
int32_t omni_ends_with(const char* str, const char* suffix);
int32_t omni_contains(const char* str, const char* substr);
int32_t omni_index_of(const char* str, const char* substr);
int32_t omni_last_index_of(const char* str, const char* substr);
char* omni_trim(const char* str);
char* omni_trim_left(const char* str);
char* omni_trim_right(const char* str);
char* omni_trim_all(const char* str);
char* omni_to_upper(const char* str);
char* omni_to_lower(const char* str);
char* omni_to_title(const char* str);
char* omni_capitalize(const char* str);
char* omni_string_reverse(const char* str);
int32_t omni_string_equals(const char* a, const char* b);
int32_t omni_string_compare(const char* a, const char* b);
int32_t omni_string_equals_ignore_case(const char* a, const char* b);
int32_t omni_string_compare_ignore_case(const char* a, const char* b);
int32_t omni_count_occurrences(const char* str, const char* substr);
int32_t omni_count_lines(const char* str);
int32_t omni_count_words(const char* str);
int32_t omni_string_is_empty(const char* str);

// String split / join. split returns a freshly allocated `const char**`
// and writes the count to `*out_len`; each element is a freshly
// allocated copy the caller owns (the existing string-array helpers
// only manage the pointer table, so deep ownership lives here).
const char** omni_string_split(const char* s, const char* delim, int32_t* out_len);
const char** omni_string_split_lines(const char* s, int32_t* out_len);
const char** omni_string_split_words(const char* s, int32_t* out_len);
char* omni_string_join(const char** parts, int32_t n, const char* sep);
char* omni_string_replace_all(const char* s, const char* old, const char* repl);
char* omni_string_replace_first(const char* s, const char* old, const char* repl);
char* omni_string_replace_last(const char* s, const char* old, const char* repl);
// find_all returns a freshly allocated int array with the byte offset
// of every non-overlapping occurrence; *out_len gets the count.
int32_t* omni_string_find_all(const char* s, const char* sub, int32_t* out_len);

// Random number generation. The runtime keeps a single global xorshift
// state seeded once on first use; std.math.random_seed lets a caller
// pin it for deterministic tests.
void omni_random_seed(int32_t seed);
int32_t omni_random_int(int32_t bound);

// Algorithms — distance metrics
double omni_euclidean_distance(double x1, double y1, double x2, double y2);
double omni_manhattan_distance(double x1, double y1, double x2, double y2);
int32_t omni_levenshtein_distance(const char* s1, const char* s2);

// Algorithms — array operations. The C backend pairs every array
// argument with its length, so each of these takes (T* arr, int32_t n)
// even though the OmniLang signature is just (arr: array<int>).
// Sorts return a freshly heap-allocated copy of `arr` sorted in place;
// callers receive the new pointer (also of length `n`).
int32_t* omni_bubble_sort(int32_t* arr, int32_t n);
int32_t* omni_selection_sort(int32_t* arr, int32_t n);
int32_t* omni_insertion_sort(int32_t* arr, int32_t n);
int32_t omni_linear_search(int32_t* arr, int32_t n, int32_t target);
int32_t omni_binary_search(int32_t* arr, int32_t n, int32_t target);
int32_t omni_array_find_max(int32_t* arr, int32_t n);
int32_t omni_array_find_min(int32_t* arr, int32_t n);
int32_t omni_array_count_occurrences(int32_t* arr, int32_t n, int32_t value);
int32_t* omni_array_reverse(int32_t* arr, int32_t n);
int32_t* omni_array_rotate(int32_t* arr, int32_t n, int32_t k);

// std.array — int32_t-specialized implementations of the generic
// list operations. The C backend recognizes std.array.<op> on
// `array<int>` values and routes to these. Other element types still
// fall through to the placeholder branch.
int32_t omni_array_int_contains(int32_t* arr, int32_t n, int32_t value);
int32_t omni_array_int_index_of(int32_t* arr, int32_t n, int32_t value);
int32_t* omni_array_int_append(int32_t* arr, int32_t n, int32_t value);
int32_t* omni_array_int_prepend(int32_t* arr, int32_t n, int32_t value);
int32_t* omni_array_int_insert(int32_t* arr, int32_t n, int32_t index, int32_t value);
int32_t* omni_array_int_remove(int32_t* arr, int32_t n, int32_t index);
int32_t* omni_array_int_concat(int32_t* a, int32_t alen, int32_t* b, int32_t blen);
int32_t* omni_array_int_slice(int32_t* arr, int32_t n, int32_t start, int32_t end);
int32_t* omni_array_int_shuffle(int32_t* arr, int32_t n);
// omni_array_int_unique writes the result length to `*out_len` and
// returns the freshly allocated deduplicated array. Order of first
// occurrence is preserved.
int32_t* omni_array_int_unique(int32_t* arr, int32_t n, int32_t* out_len);

// std.array — string-specialized siblings. Element compares use
// strcmp; output arrays alias the input strings (no deep copy of the
// payloads, only of the pointer table).
int32_t omni_array_str_contains(const char** arr, int32_t n, const char* value);
int32_t omni_array_str_index_of(const char** arr, int32_t n, const char* value);
const char** omni_array_str_append(const char** arr, int32_t n, const char* value);
const char** omni_array_str_prepend(const char** arr, int32_t n, const char* value);
const char** omni_array_str_insert(const char** arr, int32_t n, int32_t index, const char* value);
const char** omni_array_str_remove(const char** arr, int32_t n, int32_t index);
const char** omni_array_str_concat(const char** a, int32_t alen, const char** b, int32_t blen);
const char** omni_array_str_slice(const char** arr, int32_t n, int32_t start, int32_t end);
const char** omni_array_str_shuffle(const char** arr, int32_t n);
const char** omni_array_str_unique(const char** arr, int32_t n, int32_t* out_len);

// Promise/Async support (simplified synchronous implementation)
typedef struct {
    void* value;
    int32_t type;  // 0=int, 1=string, 2=float, 3=bool
    int32_t done;
} omni_promise_t;

// Create a resolved promise (synchronous implementation)
omni_promise_t* omni_promise_create_int(int32_t value);
omni_promise_t* omni_promise_create_string(const char* value);
omni_promise_t* omni_promise_create_float(double value);
omni_promise_t* omni_promise_create_bool(int32_t value);

// Await a promise (synchronous - just extracts the value)
int32_t omni_await_int(omni_promise_t* promise);
// Returns a newly allocated copy of the string - caller must free it
char* omni_await_string(omni_promise_t* promise);
double omni_await_float(omni_promise_t* promise);
int32_t omni_await_bool(omni_promise_t* promise);

// Free a promise
void omni_promise_free(omni_promise_t* promise);

// Array operations
// omni_len returns the length of an array. The length must be passed explicitly
// by the backend since C arrays don't carry length metadata.
int32_t omni_len(void* array, size_t element_size, int32_t array_length);

// Math operations
int32_t omni_add(int32_t a, int32_t b);
int32_t omni_sub(int32_t a, int32_t b);
int32_t omni_mul(int32_t a, int32_t b);
int32_t omni_div(int32_t a, int32_t b);
int32_t omni_abs(int32_t x);
int32_t omni_max(int32_t a, int32_t b);
int32_t omni_min(int32_t a, int32_t b);
char* omni_int_to_string(int32_t value);
char* omni_float_to_string(double value);
char* omni_bool_to_string(int32_t value);
int32_t omni_string_to_int(const char* str);
double omni_string_to_float(const char* str);
int32_t omni_string_to_bool(const char* str);
int32_t omni_char_code(int32_t c);
int32_t omni_char_from_code(int32_t code);
char* omni_char_to_string(int32_t c);

// Array operations
int32_t omni_array_length(int32_t* arr);
// Array get/set operations with bounds checking
// length parameter must be passed by the backend for bounds checking
int32_t omni_array_get_int(int32_t* arr, int32_t index, int32_t length);
void omni_array_set_int(int32_t* arr, int32_t index, int32_t value, int32_t length);

// Generic dynamic array for JSON parsing and multipart file uploads
typedef struct omni_array {
    void** items;
    int32_t count;
    int32_t capacity;
} omni_array_t;

omni_array_t* omni_array_create();
void omni_array_destroy(omni_array_t* arr);
void omni_array_append(omni_array_t* arr, void* item);
void* omni_array_get(omni_array_t* arr, int32_t index);
int32_t omni_array_size(omni_array_t* arr);

// Slice support — heap-allocated arrays with a hidden length/capacity header
// just before the data pointer. Used for OmniLang `[]T` arrays so they can
// be appended to and sliced at runtime.
void* omni_slice_make(int64_t len, int64_t cap, int64_t elem_size);
int64_t omni_slice_len_real(void* slice);
int64_t omni_slice_cap(void* slice);
void* omni_slice_append(void* slice, const void* elem);
void* omni_slice_subslice(void* slice, int64_t lo, int64_t hi);

// Channel + spawn support — pthread-backed bounded ring buffers and
// detached-thread spawning. The C backend lowers `make(chan T, n)` to
// omni_chan_make, `c <- v` to omni_chan_send, and `<-c` to omni_chan_recv.
typedef struct omni_chan omni_chan_t;
omni_chan_t* omni_chan_make(int64_t cap, int64_t elem_size);
void omni_chan_send(omni_chan_t* ch, const void* elem);
void omni_chan_recv(omni_chan_t* ch, void* out);
void omni_chan_close(omni_chan_t* ch);
void omni_chan_recv_ok(omni_chan_t* ch, void* out, int32_t* ok);
void omni_chan_destroy(omni_chan_t* ch);
// Non-blocking channel ops used by `select`. Return 1 on success, 0 if
// not ready (empty recv / full send), -1 if closed.
int32_t omni_chan_try_send(omni_chan_t* ch, const void* elem);
int32_t omni_chan_try_recv(omni_chan_t* ch, void* out);
int32_t omni_chan_try_recv_ok(omni_chan_t* ch, void* out, int32_t* ok);
int omni_spawn(void* (*thunk)(void*), void* ctx);

// `select` dispatch. The C codegen builds an array of case descriptors
// and calls omni_select, which returns the chosen index. Blocking and
// default semantics match Go (see the implementation alongside the
// definition).
#define OMNI_SELECT_KIND_SEND    0
#define OMNI_SELECT_KIND_RECV    1
#define OMNI_SELECT_KIND_RECV_OK 2
#define OMNI_SELECT_KIND_DEFAULT 3
typedef struct {
    int32_t      kind;
    omni_chan_t* ch;
    const void*  send_value;
    void*        recv_dest;
    int32_t*     recv_ok;
} omni_select_case_t;
int32_t omni_select(int32_t n, omni_select_case_t* cases);

// Type constants for any type support
#define OMNI_TYPE_INT 1
#define OMNI_TYPE_STRING 2
#define OMNI_TYPE_FLOAT 3
#define OMNI_TYPE_BOOL 4
#define OMNI_TYPE_MAP 5
#define OMNI_TYPE_ARRAY 6
#define OMNI_TYPE_STRUCT 7
#define OMNI_TYPE_ANY 8

// Map operations
typedef struct omni_map omni_map_t;
omni_map_t* omni_map_create();
void omni_map_destroy(omni_map_t* map);

// Map put operations for all type combinations
void omni_map_put_string_int(omni_map_t* map, const char* key, int32_t value);
void omni_map_put_string_string(omni_map_t* map, const char* key, const char* value);
void omni_map_put_string_float(omni_map_t* map, const char* key, double value);
void omni_map_put_string_bool(omni_map_t* map, const char* key, int32_t value);
void omni_map_put_int_int(omni_map_t* map, int32_t key, int32_t value);
void omni_map_put_int_string(omni_map_t* map, int32_t key, const char* value);
void omni_map_put_int_float(omni_map_t* map, int32_t key, double value);
void omni_map_put_int_bool(omni_map_t* map, int32_t key, int32_t value);

// Map put operations for any type support
void omni_map_put_string_any(omni_map_t* map, const char* key, void* value, int32_t value_type);
void omni_map_put_int_any(omni_map_t* map, int32_t key, void* value, int32_t value_type);
void omni_map_put_any_string(omni_map_t* map, void* key, int32_t key_type, const char* value);
void omni_map_put_any_int(omni_map_t* map, void* key, int32_t key_type, int32_t value);
void omni_map_put_any_float(omni_map_t* map, void* key, int32_t key_type, double value);
void omni_map_put_any_bool(omni_map_t* map, void* key, int32_t key_type, int32_t value);
void omni_map_put_any_any(omni_map_t* map, void* key, int32_t key_type, void* value, int32_t value_type);

// Map get operations for all type combinations
int32_t omni_map_get_string_int(omni_map_t* map, const char* key);
const char* omni_map_get_string_string(omni_map_t* map, const char* key);
double omni_map_get_string_float(omni_map_t* map, const char* key);
int32_t omni_map_get_string_bool(omni_map_t* map, const char* key);
int32_t omni_map_get_int_int(omni_map_t* map, int32_t key);
const char* omni_map_get_int_string(omni_map_t* map, int32_t key);
double omni_map_get_int_float(omni_map_t* map, int32_t key);
int32_t omni_map_get_int_bool(omni_map_t* map, int32_t key);

int32_t omni_map_contains_string(omni_map_t* map, const char* key);
int32_t omni_map_contains_int(omni_map_t* map, int32_t key);
int32_t omni_map_size(omni_map_t* map);
int32_t omni_map_has_string(omni_map_t* map, const char* key);
int32_t omni_map_has_int(omni_map_t* map, int32_t key);
int32_t omni_map_remove_string(omni_map_t* map, const char* key);
void omni_map_clear(omni_map_t* map);
void omni_map_delete_string(omni_map_t* map, const char* key);
void omni_map_delete_int(omni_map_t* map, int32_t key);

// Map utility functions
// Note: These return arrays which need to be allocated. For simplicity, we use fixed-size buffers.
// In a production implementation, these would need dynamic array allocation.
int32_t omni_map_keys_string_int(omni_map_t* map, char** keys_buffer, int32_t buffer_size);
int32_t omni_map_values_string_int(omni_map_t* map, int32_t* values_buffer, int32_t buffer_size);
omni_map_t* omni_map_copy_string_int(omni_map_t* map);
omni_map_t* omni_map_merge_string_int(omni_map_t* a, omni_map_t* b);

// Collection data structures (using maps as base implementation)
// Sets, queues, stacks, priority queues, linked lists, and binary trees
// For now, we'll use simplified implementations that can be extended later
typedef struct omni_set omni_set_t;
typedef struct omni_queue omni_queue_t;
typedef struct omni_stack omni_stack_t;
typedef struct omni_priority_queue omni_priority_queue_t;
typedef struct omni_linked_list omni_linked_list_t;
typedef struct omni_binary_tree omni_binary_tree_t;

// Set operations (using map internally for O(1) lookups)
omni_set_t* omni_set_create();
void omni_set_destroy(omni_set_t* set);
int32_t omni_set_add(omni_set_t* set, int32_t element);
int32_t omni_set_remove(omni_set_t* set, int32_t element);
int32_t omni_set_contains(omni_set_t* set, int32_t element);
int32_t omni_set_size(omni_set_t* set);
void omni_set_clear(omni_set_t* set);
omni_set_t* omni_set_union(omni_set_t* a, omni_set_t* b);
omni_set_t* omni_set_intersection(omni_set_t* a, omni_set_t* b);
omni_set_t* omni_set_difference(omni_set_t* a, omni_set_t* b);

// Queue operations (FIFO)
omni_queue_t* omni_queue_create();
void omni_queue_destroy(omni_queue_t* queue);
void omni_queue_enqueue(omni_queue_t* queue, int32_t element);
int32_t omni_queue_dequeue(omni_queue_t* queue);
int32_t omni_queue_peek(omni_queue_t* queue);
int32_t omni_queue_is_empty(omni_queue_t* queue);
int32_t omni_queue_size(omni_queue_t* queue);
void omni_queue_clear(omni_queue_t* queue);

// Stack operations (LIFO)
omni_stack_t* omni_stack_create();
void omni_stack_destroy(omni_stack_t* stack);
void omni_stack_push(omni_stack_t* stack, int32_t element);
int32_t omni_stack_pop(omni_stack_t* stack);
int32_t omni_stack_peek(omni_stack_t* stack);
int32_t omni_stack_is_empty(omni_stack_t* stack);
int32_t omni_stack_size(omni_stack_t* stack);
void omni_stack_clear(omni_stack_t* stack);

// Priority queue operations (max-heap)
omni_priority_queue_t* omni_priority_queue_create();
void omni_priority_queue_destroy(omni_priority_queue_t* pq);
void omni_priority_queue_insert(omni_priority_queue_t* pq, int32_t element, int32_t priority);
int32_t omni_priority_queue_extract_max(omni_priority_queue_t* pq);
int32_t omni_priority_queue_peek(omni_priority_queue_t* pq);
int32_t omni_priority_queue_is_empty(omni_priority_queue_t* pq);
int32_t omni_priority_queue_size(omni_priority_queue_t* pq);

// Linked list operations
omni_linked_list_t* omni_linked_list_create();
void omni_linked_list_destroy(omni_linked_list_t* ll);
void omni_linked_list_append(omni_linked_list_t* ll, int32_t element);
void omni_linked_list_prepend(omni_linked_list_t* ll, int32_t element);
int32_t omni_linked_list_insert(omni_linked_list_t* ll, int32_t index, int32_t element);
int32_t omni_linked_list_remove(omni_linked_list_t* ll, int32_t index);
int32_t omni_linked_list_get(omni_linked_list_t* ll, int32_t index);
int32_t omni_linked_list_set(omni_linked_list_t* ll, int32_t index, int32_t element);
int32_t omni_linked_list_size(omni_linked_list_t* ll);
int32_t omni_linked_list_is_empty(omni_linked_list_t* ll);
void omni_linked_list_clear(omni_linked_list_t* ll);

// Binary tree operations (BST)
omni_binary_tree_t* omni_binary_tree_create();
void omni_binary_tree_destroy(omni_binary_tree_t* bt);
void omni_binary_tree_insert(omni_binary_tree_t* bt, int32_t element);
int32_t omni_binary_tree_search(omni_binary_tree_t* bt, int32_t element);
int32_t omni_binary_tree_remove(omni_binary_tree_t* bt, int32_t element);
int32_t omni_binary_tree_size(omni_binary_tree_t* bt);
int32_t omni_binary_tree_is_empty(omni_binary_tree_t* bt);
void omni_binary_tree_clear(omni_binary_tree_t* bt);

// Network structures and functions
typedef struct omni_ip_address {
    char address[64];
    int32_t is_ipv4;
    int32_t is_ipv6;
} omni_ip_address_t;

typedef struct omni_url {
    char scheme[32];
    char host[256];
    int32_t port;
    char path[512];
    char query[512];
    char fragment[256];
} omni_url_t;

typedef struct omni_http_request {
    char method[16];
    char url[512];
    omni_map_t* headers;
    char* body;
} omni_http_request_t;

typedef struct omni_http_response {
    int32_t status_code;
    char status_text[64];
    omni_map_t* headers;
    char* body;
} omni_http_response_t;

// IP address functions
omni_ip_address_t* omni_ip_parse(const char* ip_str);
int32_t omni_ip_is_valid(const char* ip_str);
int32_t omni_ip_is_private(omni_ip_address_t* ip);
int32_t omni_ip_is_loopback(omni_ip_address_t* ip);
char* omni_ip_to_string(omni_ip_address_t* ip);

// URL functions
omni_url_t* omni_url_parse(const char* url_str);
char* omni_url_to_string(omni_url_t* url);
int32_t omni_url_is_valid(const char* url_str);

// DNS functions
omni_ip_address_t** omni_dns_lookup(const char* hostname, int32_t* count);
char* omni_dns_reverse_lookup(omni_ip_address_t* ip);

// HTTP client functions
omni_http_response_t* omni_http_get(const char* url);
omni_http_response_t* omni_http_post(const char* url, const char* body);
omni_http_response_t* omni_http_put(const char* url, const char* body);
omni_http_response_t* omni_http_delete(const char* url);
omni_http_response_t* omni_http_request(omni_http_request_t* req);
void omni_http_response_destroy(omni_http_response_t* resp);
int32_t omni_http_response_is_success(omni_http_response_t* resp);
int32_t omni_http_response_is_client_error(omni_http_response_t* resp);
int32_t omni_http_response_is_server_error(omni_http_response_t* resp);
char* omni_http_response_get_header(omni_http_response_t* resp, const char* name);
omni_http_request_t* omni_http_request_create(const char* method, const char* url);
void omni_http_request_set_header(omni_http_request_t* req, const char* name, const char* value);
void omni_http_request_set_body(omni_http_request_t* req, const char* body);
char* omni_http_request_get_header(omni_http_request_t* req, const char* name);
void omni_http_request_destroy(omni_http_request_t* req);

// Forward declarations for web framework (needed before function declarations)
// Note: omni_struct_t is defined later, so we use a forward declaration here
struct omni_struct;
struct omni_server;
typedef struct omni_struct omni_struct_t;
typedef struct omni_server omni_server_t;

// HTTP server functions
omni_http_request_t* omni_http_parse_request(const char* raw_request);
char* omni_http_build_response(omni_http_response_t* resp);
void omni_http_parse_query(const char* query_string, omni_map_t* params);
int32_t omni_http_match_path(const char* pattern, const char* path, omni_map_t* params);

// Context functions (for std.web framework)
const char* omni_context_param(omni_struct_t* ctx, const char* name);
const char* omni_context_query(omni_struct_t* ctx, const char* name);
omni_map_t* omni_context_query_all(omni_struct_t* ctx);
const char* omni_context_header(omni_struct_t* ctx, const char* name);
void omni_context_set_header(omni_struct_t* ctx, const char* name, const char* value);
void omni_context_status(omni_struct_t* ctx, int32_t code);
omni_struct_t* omni_context_html(omni_struct_t* ctx, const char* html);
omni_struct_t* omni_context_redirect(omni_struct_t* ctx, const char* url, int32_t code);
omni_struct_t* omni_context_cookie(omni_struct_t* ctx, const char* name, const char* value, omni_map_t* options);
const char* omni_context_get_cookie(omni_struct_t* ctx, const char* name);
const char* omni_context_body(omni_struct_t* ctx);
omni_struct_t* omni_context_set_state(omni_struct_t* ctx, const char* key, void* value, int32_t value_type);
void* omni_context_get_state(omni_struct_t* ctx, const char* key, int32_t* value_type);
void* omni_context_body_json(omni_struct_t* ctx);
omni_struct_t* omni_context_text(omni_struct_t* ctx, const char* text);
omni_struct_t* omni_context_json(omni_struct_t* ctx, void* data);
omni_struct_t* omni_context_file(omni_struct_t* ctx, const char* path);
omni_map_t* omni_context_body_form(omni_struct_t* ctx);
omni_array_t* omni_context_files(omni_struct_t* ctx);

// WebSocket functions (stubs)
void omni_server_websocket(omni_server_t* server, const char* pattern, void* handler);
int32_t omni_websocket_send(void* conn, const char* data, int32_t len);
int32_t omni_websocket_receive(void* conn, char* buf, int32_t max_len);
void omni_websocket_close(void* conn);

// Server routing functions (for std.web framework)
void omni_server_get(omni_server_t* server, const char* pattern, void* handler);
void omni_server_post(omni_server_t* server, const char* pattern, void* handler);
void omni_server_put(omni_server_t* server, const char* pattern, void* handler);
void omni_server_delete(omni_server_t* server, const char* pattern, void* handler);
void omni_server_patch(omni_server_t* server, const char* pattern, void* handler);
void omni_server_all(omni_server_t* server, const char* pattern, void* handler);
void omni_server_route(omni_server_t* server, const char* method, const char* pattern, void* handler);
omni_struct_t* omni_server_group(omni_server_t* server, const char* prefix);
void omni_server_use(omni_server_t* server, void* middleware);
void omni_server_use_before(omni_server_t* server, void* middleware);
void omni_server_use_after(omni_server_t* server, void* middleware);
void omni_group_get(omni_struct_t* group, const char* pattern, void* handler);
void omni_group_post(omni_struct_t* group, const char* pattern, void* handler);
void omni_group_use(omni_struct_t* group, void* middleware);

// Middleware functions
omni_struct_t* omni_middleware_logger(omni_struct_t* ctx);
omni_struct_t* omni_middleware_cors(omni_struct_t* ctx, omni_map_t* options);
omni_struct_t* omni_middleware_json_parser(omni_struct_t* ctx);
omni_struct_t* omni_middleware_form_parser(omni_struct_t* ctx);
omni_struct_t* omni_middleware_multipart_parser_impl(omni_struct_t* ctx);
void* omni_middleware_multipart_parser(omni_struct_t* ctx, int32_t max_size);
omni_struct_t* omni_middleware_static_impl(omni_struct_t* ctx);
void* omni_middleware_static(omni_server_t* server, const char* path, const char* dir, omni_map_t* options);

// Template functions
char* omni_template_render(const char* template, omni_map_t* data);
char* omni_template_load(const char* path);
void omni_template_cache_enable(int32_t enable);

// Validation functions
omni_map_t* omni_validate_request(omni_struct_t* ctx, omni_map_t* rules);

// Test client functions
void* omni_test_client_create(omni_server_t* server);
void* omni_test_client_get(void* client, const char* path, omni_map_t* headers);
void* omni_test_client_post(void* client, const char* path, const char* body, omni_map_t* headers);
int32_t omni_test_response_status(void* resp);
const char* omni_test_response_body(void* resp);
omni_map_t* omni_test_response_headers(void* resp);
void* omni_test_response_json(void* resp);

// Memory management functions
void omni_panic(const char* message);
void* omni_malloc(size_t size);
void omni_free(void* ptr);
void* omni_realloc(void* ptr, size_t new_size);

// JSON functions
void* omni_json_parse(const char* json_str);
char* omni_json_stringify(void* value, int32_t value_type, int32_t pretty);

// Form data and file upload functions
void omni_http_parse_form_urlencoded(const char* body, omni_map_t* params);
void omni_http_parse_multipart(const char* body, const char* boundary, omni_map_t* fields, omni_array_t* files);
char* omni_file_upload_save(const char* data, int32_t size, const char* filename, const char* upload_dir);
int32_t omni_file_upload_validate(const char* filename, int32_t size, const char* allowed_types, int32_t max_size);

// Static file serving functions
char* omni_file_read_binary(const char* path, int32_t* size);
const char* omni_file_get_mime_type(const char* filename);
int32_t omni_file_get_size(const char* path);

// Response compression functions
char* omni_http_compress_gzip(const char* data, int32_t len, int32_t* compressed_len);
char* omni_http_decompress_gzip(const char* compressed, int32_t len, int32_t* decompressed_len);

// Validation and sanitization functions
int32_t omni_validate_string(const char* value, const char* pattern, int32_t min_len, int32_t max_len);
int32_t omni_validate_int(const char* value, int32_t min, int32_t max);
int32_t omni_validate_email(const char* email);
int32_t omni_validate_url(const char* url);
char* omni_sanitize_html(const char* html);
char* omni_sanitize_sql(const char* sql);

// WebSocket functions
char* omni_websocket_handshake(const char* request_headers);
char* omni_websocket_frame_create(const char* data, int32_t len, int32_t opcode, int32_t mask);
void* omni_websocket_frame_parse(const char* frame, int32_t len);

// Server concurrency and connection management
typedef struct omni_connection_pool omni_connection_pool_t;
typedef struct omni_thread_pool omni_thread_pool_t;
omni_connection_pool_t* omni_server_connection_pool_create(int32_t max_connections);
int32_t omni_server_connection_pool_acquire(omni_connection_pool_t* pool);
void omni_server_connection_pool_release(omni_connection_pool_t* pool, int32_t socket);
omni_thread_pool_t* omni_server_thread_pool_create(int32_t num_threads);
void omni_server_thread_pool_submit(omni_thread_pool_t* pool, void (*task)(void*), void* arg);

// Server timeouts and limits
void omni_server_set_timeout(int32_t socket, int32_t timeout_seconds);
void omni_server_set_max_request_size(int32_t max_size);
void omni_server_set_max_headers_size(int32_t max_size);

// Server lifecycle
// Note: omni_server_t is forward declared above
omni_server_t* omni_server_create(int32_t port, omni_map_t* options);
int32_t omni_server_listen(omni_server_t* server);
int32_t omni_server_listen_tls(omni_server_t* server, const char* cert_file, const char* key_file);
void omni_server_close(omni_server_t* server);
void omni_server_graceful_shutdown(omni_server_t* server, int32_t timeout_seconds);

// Session management
typedef struct omni_session omni_session_t;
typedef struct omni_session_store omni_session_store_t;
omni_session_t* omni_session_create(const char* session_id, int32_t timeout_seconds);
const char* omni_session_get(omni_session_t* session, const char* key);
void omni_session_set(omni_session_t* session, const char* key, const char* value);
void omni_session_destroy(omni_session_t* session);
omni_session_store_t* omni_session_store_create(const char* storage_type);
void omni_session_store_save(omni_session_store_t* store, omni_session_t* session);
omni_session_t* omni_session_store_load(omni_session_store_t* store, const char* session_id);

// Authentication/Authorization
char* omni_auth_hash_password(const char* password, const char* salt);
int32_t omni_auth_verify_password(const char* password, const char* hash);
char* omni_auth_generate_token(const char* user_id, const char* secret, int32_t expires_in);
const char* omni_auth_verify_token(const char* token, const char* secret);
int32_t omni_auth_check_permission(const char* user_id, const char* resource, const char* action);

// Rate limiting
typedef struct omni_rate_limiter omni_rate_limiter_t;
omni_rate_limiter_t* omni_rate_limit_create(int32_t max_requests, int32_t window_seconds);
int32_t omni_rate_limit_check(omni_rate_limiter_t* limiter, const char* key);
void omni_rate_limit_reset(omni_rate_limiter_t* limiter, const char* key);

// Socket functions
int32_t omni_socket_create();
int32_t omni_socket_connect(int32_t socket, const char* address, int32_t port);
int32_t omni_socket_bind(int32_t socket, const char* address, int32_t port);
int32_t omni_socket_listen(int32_t socket, int32_t backlog);
int32_t omni_socket_accept(int32_t socket);
int32_t omni_socket_send(int32_t socket, const char* data);
int32_t omni_socket_receive(int32_t socket, char* buffer, int32_t buffer_size);
int32_t omni_socket_close(int32_t socket);

// Network utility functions
int32_t omni_network_is_connected();
omni_ip_address_t* omni_network_get_local_ip();
int32_t omni_network_ping(const char* host);

// Struct operations
// Note: omni_struct_t is forward declared above for web framework functions
omni_struct_t* omni_struct_create();
void omni_struct_destroy(omni_struct_t* struct_ptr);
void omni_struct_set_type_name(omni_struct_t* struct_ptr, const char* type_name);
const char* omni_struct_get_type_name(omni_struct_t* struct_ptr);
void omni_struct_set_string_field(omni_struct_t* struct_ptr, const char* field_name, const char* value);
void omni_struct_set_int_field(omni_struct_t* struct_ptr, const char* field_name, int32_t value);
void omni_struct_set_float_field(omni_struct_t* struct_ptr, const char* field_name, double value);
void omni_struct_set_bool_field(omni_struct_t* struct_ptr, const char* field_name, int32_t value);
void omni_struct_set_array_field(omni_struct_t* struct_ptr, const char* field_name, void* array_value, int32_t element_type, int32_t array_length);
void omni_struct_set_map_field(omni_struct_t* struct_ptr, const char* field_name, omni_map_t* map_value);
void omni_struct_set_struct_field(omni_struct_t* struct_ptr, const char* field_name, omni_struct_t* struct_value);
void omni_struct_set_null_field(omni_struct_t* struct_ptr, const char* field_name);
const char* omni_struct_get_string_field(omni_struct_t* struct_ptr, const char* field_name);
int32_t omni_struct_get_int_field(omni_struct_t* struct_ptr, const char* field_name);
double omni_struct_get_float_field(omni_struct_t* struct_ptr, const char* field_name);
int32_t omni_struct_get_bool_field(omni_struct_t* struct_ptr, const char* field_name);
omni_struct_t* omni_struct_get_struct_field(omni_struct_t* struct_ptr, const char* field_name);

double omni_pow(double x, double y);
double omni_sqrt(double x);
double omni_floor(double x);
double omni_ceil(double x);
double omni_round(double x);
int32_t omni_gcd(int32_t a, int32_t b);
int32_t omni_lcm(int32_t a, int32_t b);
int32_t omni_factorial(int32_t n);

// Trigonometric functions
double omni_sin(double x);
double omni_cos(double x);
double omni_tan(double x);
double omni_asin(double x);
double omni_acos(double x);
double omni_atan(double x);
double omni_atan2(double y, double x);

// Logarithmic and exponential functions
double omni_exp(double x);
double omni_log(double x);
double omni_log10(double x);
double omni_log2(double x);

// Hyperbolic functions
double omni_sinh(double x);
double omni_cosh(double x);
double omni_tanh(double x);

// Additional math functions
double omni_cbrt(double x);
double omni_trunc(double x);

// File I/O operations
// Use intptr_t for file handles to safely store FILE* pointers on 64-bit platforms
intptr_t omni_file_open(const char* filename, const char* mode);
int32_t omni_file_close(intptr_t file_handle);
int32_t omni_file_read(intptr_t file_handle, const char* buffer, int32_t size);
int32_t omni_file_write(intptr_t file_handle, const char* buffer, int32_t size);
int32_t omni_file_seek(intptr_t file_handle, int32_t offset, int32_t whence);
int32_t omni_file_tell(intptr_t file_handle);
int32_t omni_file_exists(const char* filename);
int32_t omni_file_size(const char* filename);

// File I/O convenience functions (for async operations)
// Returns a newly allocated string - caller must free it using free()
char* omni_read_file(const char* path);
int32_t omni_write_file(const char* path, const char* content);
int32_t omni_append_file(const char* path, const char* content);

// Testing framework
void omni_test_start(const char* test_name);
void omni_test_end(const char* test_name, int32_t passed);
void omni_assert(int32_t condition, const char* message);
void omni_assert_eq_int(int32_t expected, int32_t actual, const char* message);
void omni_assert_eq_string(const char* expected, const char* actual, const char* message);
void omni_assert_eq_float(double expected, double actual, const char* message);
void omni_assert_true(int32_t condition, const char* message);
void omni_assert_false(int32_t condition, const char* message);
int32_t omni_test_summary();
void omni_test_reset(); // Reset test counters (useful when running multiple test files)

// System operations
void omni_exit(int32_t code);

// Environment variable operations
const char* omni_getenv(const char* name);
int32_t omni_setenv(const char* name, const char* value);
int32_t omni_unsetenv(const char* name);

// Directory operations
const char* omni_getcwd(void);
int32_t omni_chdir(const char* path);

// File system operations
int32_t omni_mkdir(const char* path);
int32_t omni_rmdir(const char* path);
int32_t omni_remove(const char* path);
int32_t omni_rename(const char* old_path, const char* new_path);
int32_t omni_copy(const char* src_path, const char* dst_path);
int32_t omni_exists(const char* path);
int32_t omni_is_file(const char* path);
int32_t omni_is_dir(const char* path);

// String validation functions
int32_t omni_string_is_alpha(const char* str);
int32_t omni_string_is_digit(const char* str);
int32_t omni_string_is_alnum(const char* str);
int32_t omni_string_is_ascii(const char* str);
int32_t omni_string_is_upper(const char* str);
int32_t omni_string_is_lower(const char* str);

// String encoding/escaping functions
char* omni_encode_base64(const char* str);
char* omni_decode_base64(const char* str);
char* omni_encode_url(const char* str);
char* omni_decode_url(const char* str);
char* omni_escape_html(const char* str);
char* omni_unescape_html(const char* str);
char* omni_escape_json(const char* str);
char* omni_escape_shell(const char* str);

// Regex functions (using POSIX regex)
int32_t omni_string_matches(const char* str, const char* pattern);
char* omni_string_find_match(const char* str, const char* pattern);
char* omni_string_find_all_matches(const char* str, const char* pattern, int32_t* count);
char* omni_string_replace_regex(const char* str, const char* pattern, const char* replacement);

// Time functions
int64_t omni_time_now_unix(void);
int64_t omni_time_now_unix_nano(void);
void omni_time_sleep_seconds(double seconds);
void omni_time_sleep_milliseconds(int32_t milliseconds);
int32_t omni_time_zone_offset(void);
const char* omni_time_zone_name(void);
void omni_time_from_unix(int64_t timestamp, int32_t* year, int32_t* month, int32_t* day, int32_t* hour, int32_t* minute, int32_t* second, int32_t* nanosecond);
void omni_time_from_string(const char* time_str, int32_t* year, int32_t* month, int32_t* day, int32_t* hour, int32_t* minute, int32_t* second, int32_t* nanosecond);
int64_t omni_time_to_unix(int32_t year, int32_t month, int32_t day, int32_t hour, int32_t minute, int32_t second, int32_t nanosecond);
char* omni_time_to_string(int32_t year, int32_t month, int32_t day, int32_t hour, int32_t minute, int32_t second, int32_t nanosecond);
int64_t omni_time_to_unix_nano(int32_t year, int32_t month, int32_t day, int32_t hour, int32_t minute, int32_t second, int32_t nanosecond);
char* omni_duration_to_string(int32_t seconds, int32_t nanoseconds);

// Command-line argument functions
void omni_args_init(int argc, char** argv);
int32_t omni_args_count(void);
const char* omni_args_get(int32_t index);
int32_t omni_args_has_flag(const char* name);
const char* omni_args_get_flag(const char* name, const char* default_value);
const char* omni_args_positional(int32_t index, const char* default_value);

// Process ID functions
int32_t omni_getpid(void);
int32_t omni_getppid(void);

// Entry point
int32_t omni_main();

// Deferred-call support (Go-style `defer`). Each function that uses defer
// declares a local omni_defer_frame_t, pushes per-site thunks for each
// deferred call, and calls omni_defer_run_all before returning.
typedef void (*omni_defer_thunk_t)(void*);
typedef struct omni_defer_node {
    omni_defer_thunk_t thunk;
    void* ctx;
    struct omni_defer_node* next;
} omni_defer_node_t;
typedef struct {
    omni_defer_node_t* head;
} omni_defer_frame_t;
void omni_defer_push(omni_defer_frame_t* frame, omni_defer_thunk_t thunk, void* ctx);
void omni_defer_run_all(omni_defer_frame_t* frame);

// Coverage tracking functions
void omni_coverage_init(void);
void omni_coverage_record(const char* function_name, const char* file_path, int32_t line_number);
char* omni_coverage_export(void);  // Returns JSON string, caller must free
void omni_coverage_reset(void);
int32_t omni_coverage_is_enabled(void);
void omni_coverage_set_enabled(int32_t enabled);

#endif // OMNI_RT_H
