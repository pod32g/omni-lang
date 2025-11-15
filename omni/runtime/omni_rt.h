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
const char* omni_await_string(omni_promise_t* promise);
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
const char* omni_read_file(const char* path);
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

// Entry point
int32_t omni_main();

#endif // OMNI_RT_H
