#ifndef OMNI_RT_H
#define OMNI_RT_H

#include <stdint.h>
#include <stdio.h>

// OmniLang Runtime Library
// This provides the runtime support for OmniLang programs

// Basic I/O functions
void omni_print_int(int32_t value);
void omni_print_string(const char* str);
void omni_println_int(int32_t value);
void omni_println_string(const char* str);
void omni_print_float(double value);
void omni_println_float(double value);
void omni_print_bool(int32_t value);
void omni_println_bool(int32_t value);

// Memory management
void* omni_alloc(size_t size);
void omni_free(void* ptr);
void* omni_malloc(size_t size);
void* omni_realloc(void* ptr, size_t new_size);

// String operations
char* omni_strcat(const char* str1, const char* str2);
int32_t omni_strlen(const char* str);

// Array operations
int32_t omni_len(void* array, size_t element_size);

// Math operations
int32_t omni_add(int32_t a, int32_t b);
int32_t omni_sub(int32_t a, int32_t b);
int32_t omni_mul(int32_t a, int32_t b);
int32_t omni_div(int32_t a, int32_t b);
int32_t omni_abs(int32_t x);
int32_t omni_max(int32_t a, int32_t b);
int32_t omni_min(int32_t a, int32_t b);
char* omni_int_to_string(int32_t value);

// File I/O operations
int32_t omni_file_open(const char* filename, const char* mode);
int32_t omni_file_close(int32_t file_handle);
int32_t omni_file_read(int32_t file_handle, char* buffer, int32_t size);
int32_t omni_file_write(int32_t file_handle, const char* buffer, int32_t size);
int32_t omni_file_seek(int32_t file_handle, int32_t offset, int32_t whence);
int32_t omni_file_tell(int32_t file_handle);
int32_t omni_file_exists(const char* filename);
int32_t omni_file_size(const char* filename);

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

// System operations
void omni_exit(int32_t code);

// Entry point
int32_t omni_main();

#endif // OMNI_RT_H
