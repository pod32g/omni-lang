#ifndef OMNI_RT_H
#define OMNI_RT_H

#include <stdint.h>
#include <stdio.h>

// OmniLang Runtime Library
// This provides the runtime support for OmniLang programs

// Basic I/O functions
void omni_print_string(const char* str);
void omni_println_string(const char* str);
char* omni_read_line(void);

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
char* omni_to_upper(const char* str);
char* omni_to_lower(const char* str);
int32_t omni_string_equals(const char* a, const char* b);
int32_t omni_string_compare(const char* a, const char* b);

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

// Array operations
int32_t omni_array_length(int32_t* arr);
// Array get/set operations with bounds checking
// length parameter must be passed by the backend for bounds checking
int32_t omni_array_get_int(int32_t* arr, int32_t index, int32_t length);
void omni_array_set_int(int32_t* arr, int32_t index, int32_t value, int32_t length);

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
typedef struct omni_struct omni_struct_t;
omni_struct_t* omni_struct_create();
void omni_struct_destroy(omni_struct_t* struct_ptr);
void omni_struct_set_string_field(omni_struct_t* struct_ptr, const char* field_name, const char* value);
void omni_struct_set_int_field(omni_struct_t* struct_ptr, const char* field_name, int32_t value);
void omni_struct_set_float_field(omni_struct_t* struct_ptr, const char* field_name, double value);
void omni_struct_set_bool_field(omni_struct_t* struct_ptr, const char* field_name, int32_t value);
const char* omni_struct_get_string_field(omni_struct_t* struct_ptr, const char* field_name);
int32_t omni_struct_get_int_field(omni_struct_t* struct_ptr, const char* field_name);
double omni_struct_get_float_field(omni_struct_t* struct_ptr, const char* field_name);
int32_t omni_struct_get_bool_field(omni_struct_t* struct_ptr, const char* field_name);

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
int32_t omni_file_read(intptr_t file_handle, char* buffer, int32_t size);
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

// Coverage tracking functions
void omni_coverage_init(void);
void omni_coverage_record(const char* function_name, const char* file_path, int32_t line_number);
char* omni_coverage_export(void);  // Returns JSON string, caller must free
void omni_coverage_reset(void);
int32_t omni_coverage_is_enabled(void);
void omni_coverage_set_enabled(int32_t enabled);

#endif // OMNI_RT_H
