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

// Entry point
int32_t omni_main();

#endif // OMNI_RT_H
