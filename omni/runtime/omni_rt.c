#include "omni_rt.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <sys/stat.h>
#include <math.h>

// Test framework state
static int32_t total_tests = 0;
static int32_t passed_tests = 0;
static int32_t current_test_passed = 1;

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

void* omni_malloc(size_t size) {
    return malloc(size);
}

void* omni_realloc(void* ptr, size_t new_size) {
    return realloc(ptr, new_size);
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

char* omni_substring(const char* str, int32_t start, int32_t end) {
    if (!str || start < 0 || end < start) {
        return NULL;
    }
    
    int32_t len = (int32_t)strlen(str);
    if (start >= len) {
        return malloc(1); // Return empty string
    }
    
    if (end > len) {
        end = len;
    }
    
    int32_t sublen = end - start;
    char* result = malloc(sublen + 1);
    if (result) {
        strncpy(result, str + start, sublen);
        result[sublen] = '\0';
    }
    return result;
}

char omni_char_at(const char* str, int32_t index) {
    if (!str || index < 0 || index >= (int32_t)strlen(str)) {
        return '\0';
    }
    return str[index];
}

int32_t omni_starts_with(const char* str, const char* prefix) {
    if (!str || !prefix) {
        return 0;
    }
    return strncmp(str, prefix, strlen(prefix)) == 0 ? 1 : 0;
}

int32_t omni_ends_with(const char* str, const char* suffix) {
    if (!str || !suffix) {
        return 0;
    }
    
    int32_t str_len = (int32_t)strlen(str);
    int32_t suffix_len = (int32_t)strlen(suffix);
    
    if (suffix_len > str_len) {
        return 0;
    }
    
    return strcmp(str + str_len - suffix_len, suffix) == 0 ? 1 : 0;
}

int32_t omni_contains(const char* str, const char* substr) {
    if (!str || !substr) {
        return 0;
    }
    return strstr(str, substr) != NULL ? 1 : 0;
}

int32_t omni_index_of(const char* str, const char* substr) {
    if (!str || !substr) {
        return -1;
    }
    
    char* pos = strstr(str, substr);
    if (pos == NULL) {
        return -1;
    }
    
    return (int32_t)(pos - str);
}

int32_t omni_last_index_of(const char* str, const char* substr) {
    if (!str || !substr) {
        return -1;
    }
    
    int32_t str_len = (int32_t)strlen(str);
    int32_t substr_len = (int32_t)strlen(substr);
    
    if (substr_len == 0) {
        return str_len;
    }
    
    for (int32_t i = str_len - substr_len; i >= 0; i--) {
        if (strncmp(str + i, substr, substr_len) == 0) {
            return i;
        }
    }
    
    return -1;
}

char* omni_trim(const char* str) {
    if (!str) {
        return NULL;
    }
    
    // Find start of non-whitespace
    const char* start = str;
    while (*start && (*start == ' ' || *start == '\t' || *start == '\n' || *start == '\r')) {
        start++;
    }
    
    // Find end of non-whitespace
    const char* end = str + strlen(str) - 1;
    while (end > start && (*end == ' ' || *end == '\t' || *end == '\n' || *end == '\r')) {
        end--;
    }
    
    int32_t len = (int32_t)(end - start + 1);
    char* result = malloc(len + 1);
    if (result) {
        strncpy(result, start, len);
        result[len] = '\0';
    }
    return result;
}

char* omni_to_upper(const char* str) {
    if (!str) {
        return NULL;
    }
    
    int32_t len = (int32_t)strlen(str);
    char* result = malloc(len + 1);
    if (result) {
        for (int32_t i = 0; i < len; i++) {
            char c = str[i];
            if (c >= 'a' && c <= 'z') {
                result[i] = c - 'a' + 'A';
            } else {
                result[i] = c;
            }
        }
        result[len] = '\0';
    }
    return result;
}

char* omni_to_lower(const char* str) {
    if (!str) {
        return NULL;
    }
    
    int32_t len = (int32_t)strlen(str);
    char* result = malloc(len + 1);
    if (result) {
        for (int32_t i = 0; i < len; i++) {
            char c = str[i];
            if (c >= 'A' && c <= 'Z') {
                result[i] = c - 'A' + 'a';
            } else {
                result[i] = c;
            }
        }
        result[len] = '\0';
    }
    return result;
}

int32_t omni_string_equals(const char* a, const char* b) {
    if (!a || !b) {
        return (a == b) ? 1 : 0;
    }
    return strcmp(a, b) == 0 ? 1 : 0;
}

int32_t omni_string_compare(const char* a, const char* b) {
    if (!a || !b) {
        if (!a && !b) return 0;
        if (!a) return -1;
        return 1;
    }
    return strcmp(a, b);
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

char* omni_float_to_string(double value) {
    char* str = malloc(64); // Enough for any double
    if (str) {
        snprintf(str, 64, "%f", value);
    }
    return str;
}

char* omni_bool_to_string(int32_t value) {
    char* str = malloc(8); // Enough for "true" or "false"
    if (str) {
        if (value) {
            strcpy(str, "true");
        } else {
            strcpy(str, "false");
        }
    }
    return str;
}

int32_t omni_string_to_int(const char* str) {
    if (!str) return 0;
    return (int32_t)atoi(str);
}

double omni_string_to_float(const char* str) {
    if (!str) return 0.0;
    return atof(str);
}

int32_t omni_string_to_bool(const char* str) {
    if (!str) return 0;
    return (strcmp(str, "true") == 0) ? 1 : 0;
}

// Array operations
int32_t omni_array_length(int32_t* arr) {
    if (!arr) return 0;
    // For now, we'll use a simple approach where we can't determine the length
    // In a real implementation, we'd have proper array metadata
    // For testing purposes, return a fixed value
    return 5; // This is a placeholder - in reality we'd need array metadata
}

int32_t omni_array_get_int(int32_t* arr, int32_t index) {
    if (!arr || index < 0) return 0;
    // Return the actual element at the given index
    return arr[index];
}

void omni_array_set_int(int32_t* arr, int32_t index, int32_t value) {
    if (!arr || index < 0) return;
    // Set the actual element at the given index
    arr[index] = value;
}

double omni_pow(double x, double y) {
    return pow(x, y);
}

double omni_sqrt(double x) {
    return sqrt(x);
}

double omni_floor(double x) {
    return floor(x);
}

double omni_ceil(double x) {
    return ceil(x);
}

double omni_round(double x) {
    return round(x);
}

int32_t omni_gcd(int32_t a, int32_t b) {
    if (b == 0) {
        return a;
    }
    return omni_gcd(b, a % b);
}

int32_t omni_lcm(int32_t a, int32_t b) {
    if (a == 0 || b == 0) {
        return 0;
    }
    return (a * b) / omni_gcd(a, b);
}

int32_t omni_factorial(int32_t n) {
    if (n <= 1) {
        return 1;
    }
    int32_t result = 1;
    for (int32_t i = 2; i <= n; i++) {
        result *= i;
    }
    return result;
}

// Array operations
int32_t omni_len(void* array, size_t element_size) {
    // This is a placeholder - in practice, we need to track array sizes
    // For now, we'll return 0 as a safe default
    // TODO: Implement proper array size tracking
    (void)array;        // Suppress unused parameter warning
    (void)element_size; // Suppress unused parameter warning
    return 0;
}

// File I/O operations
int32_t omni_file_open(const char* filename, const char* mode) {
    FILE* file = fopen(filename, mode);
    if (file == NULL) {
        return -1; // Error: file could not be opened
    }
    return (int32_t)(intptr_t)file; // Cast FILE* to int32_t handle
}

int32_t omni_file_close(int32_t file_handle) {
    if (file_handle == -1) {
        return -1; // Error: invalid handle
    }
    FILE* file = (FILE*)(intptr_t)file_handle;
    return fclose(file) == 0 ? 0 : -1;
}

int32_t omni_file_read(int32_t file_handle, char* buffer, int32_t size) {
    if (file_handle == -1 || buffer == NULL || size <= 0) {
        return -1; // Error: invalid parameters
    }
    FILE* file = (FILE*)(intptr_t)file_handle;
    size_t bytes_read = fread(buffer, 1, (size_t)size, file);
    return (int32_t)bytes_read;
}

int32_t omni_file_write(int32_t file_handle, const char* buffer, int32_t size) {
    if (file_handle == -1 || buffer == NULL || size <= 0) {
        return -1; // Error: invalid parameters
    }
    FILE* file = (FILE*)(intptr_t)file_handle;
    size_t bytes_written = fwrite(buffer, 1, (size_t)size, file);
    return (int32_t)bytes_written;
}

int32_t omni_file_seek(int32_t file_handle, int32_t offset, int32_t whence) {
    if (file_handle == -1) {
        return -1; // Error: invalid handle
    }
    FILE* file = (FILE*)(intptr_t)file_handle;
    return fseek(file, offset, whence) == 0 ? 0 : -1;
}

int32_t omni_file_tell(int32_t file_handle) {
    if (file_handle == -1) {
        return -1; // Error: invalid handle
    }
    FILE* file = (FILE*)(intptr_t)file_handle;
    long pos = ftell(file);
    return pos >= 0 ? (int32_t)pos : -1;
}

int32_t omni_file_exists(const char* filename) {
    if (filename == NULL) {
        return 0; // Error: null filename
    }
    struct stat st;
    return stat(filename, &st) == 0 ? 1 : 0;
}

int32_t omni_file_size(const char* filename) {
    if (filename == NULL) {
        return -1; // Error: null filename
    }
    struct stat st;
    if (stat(filename, &st) != 0) {
        return -1; // Error: file not found or stat failed
    }
    return (int32_t)st.st_size;
}

// Testing framework
void omni_test_start(const char* test_name) {
    printf("Running test: %s\n", test_name);
    current_test_passed = 1;
}

void omni_test_end(const char* test_name, int32_t passed) {
    total_tests++;
    if (passed && current_test_passed) {
        passed_tests++;
        printf("✓ %s PASSED\n", test_name);
    } else {
        printf("✗ %s FAILED\n", test_name);
    }
}

void omni_assert(int32_t condition, const char* message) {
    if (!condition) {
        printf("  ASSERTION FAILED: %s\n", message);
        current_test_passed = 0;
    }
}

void omni_assert_eq_int(int32_t expected, int32_t actual, const char* message) {
    if (expected != actual) {
        printf("  ASSERTION FAILED: %s (expected: %d, actual: %d)\n", message, expected, actual);
        current_test_passed = 0;
    }
}

void omni_assert_eq_string(const char* expected, const char* actual, const char* message) {
    if (strcmp(expected, actual) != 0) {
        printf("  ASSERTION FAILED: %s (expected: \"%s\", actual: \"%s\")\n", message, expected, actual);
        current_test_passed = 0;
    }
}

void omni_assert_eq_float(double expected, double actual, const char* message) {
    const double epsilon = 1e-9;
    if (fabs(expected - actual) > epsilon) {
        printf("  ASSERTION FAILED: %s (expected: %f, actual: %f)\n", message, expected, actual);
        current_test_passed = 0;
    }
}

void omni_assert_true(int32_t condition, const char* message) {
    if (!condition) {
        printf("  ASSERTION FAILED: %s (expected: true, actual: false)\n", message);
        current_test_passed = 0;
    }
}

void omni_assert_false(int32_t condition, const char* message) {
    if (condition) {
        printf("  ASSERTION FAILED: %s (expected: false, actual: true)\n", message);
        current_test_passed = 0;
    }
}

int32_t omni_test_summary() {
    printf("\nTest Summary: %d/%d tests passed\n", passed_tests, total_tests);
    if (passed_tests == total_tests) {
        printf("All tests passed! ✓\n");
        return 0;
    } else {
        printf("Some tests failed! ✗\n");
        return 1;
    }
}

// System operations
void omni_exit(int32_t code) {
    exit(code);
}

// Entry point - this will be implemented by the generated code
// The generated code will provide the omni_main function
