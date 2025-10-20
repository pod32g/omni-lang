#include "omni_rt.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

// Basic I/O functions
void omni_print_int(int32_t value) {
    printf("%d", value);
}

void omni_print_string(const char* str) {
    printf("%s", str);
}

void omni_println_int(int32_t value) {
    printf("%d\n", value);
}

void omni_println_string(const char* str) {
    printf("%s\n", str);
}

void omni_print_float(double value) {
    printf("%f", value);
}

void omni_println_float(double value) {
    printf("%f\n", value);
}

void omni_print_bool(int32_t value) {
    printf("%s", value ? "true" : "false");
}

void omni_println_bool(int32_t value) {
    printf("%s\n", value ? "true" : "false");
}

// Memory management
void* omni_alloc(size_t size) {
    return malloc(size);
}

void omni_free(void* ptr) {
    free(ptr);
}

// String operations
char* omni_strcat(const char* str1, const char* str2) {
    size_t len1 = strlen(str1);
    size_t len2 = strlen(str2);
    char* result = malloc(len1 + len2 + 1);
    if (result) {
        strcpy(result, str1);
        strcat(result, str2);
    }
    return result;
}

int32_t omni_strlen(const char* str) {
    return (int32_t)strlen(str);
}

// Math operations
int32_t omni_add(int32_t a, int32_t b) {
    return a + b;
}

int32_t omni_sub(int32_t a, int32_t b) {
    return a - b;
}

int32_t omni_mul(int32_t a, int32_t b) {
    return a * b;
}

int32_t omni_div(int32_t a, int32_t b) {
    return b != 0 ? a / b : 0;
}

int32_t omni_abs(int32_t x) {
    return x < 0 ? -x : x;
}

int32_t omni_max(int32_t a, int32_t b) {
    return a > b ? a : b;
}

int32_t omni_min(int32_t a, int32_t b) {
    return a < b ? a : b;
}

char* omni_int_to_string(int32_t value) {
    char* str = malloc(32); // Enough for any int32_t
    if (str) {
        snprintf(str, 32, "%d", value);
    }
    return str;
}

// Entry point - this will be implemented by the generated code
// The generated code will provide the omni_main function
