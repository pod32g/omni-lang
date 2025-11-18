// Enable POSIX and GNU extensions for functions like localtime_r, setenv, strdup, etc.
#ifndef _POSIX_C_SOURCE
#define _POSIX_C_SOURCE 200809L
#endif
#ifndef _GNU_SOURCE
#define _GNU_SOURCE
#endif

#include "omni_rt.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <sys/stat.h>
#include <math.h>
#include <time.h>
#include <ctype.h>
#include <stdint.h>
#include <errno.h>
#include <limits.h>
#include <locale.h>
#include <regex.h>
#ifdef _WIN32
#include <windows.h>
#include <direct.h>
#include <io.h>
#include <process.h>
#else
#include <pthread.h>
#include <unistd.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <time.h>
#endif

// Test framework state
static int32_t total_tests = 0;
static int32_t passed_tests = 0;
static int32_t current_test_passed = 1;

// Logging state
enum {
    OMNI_LOG_LEVEL_DEBUG = 0,
    OMNI_LOG_LEVEL_INFO = 1,
    OMNI_LOG_LEVEL_WARN = 2,
    OMNI_LOG_LEVEL_ERROR = 3,
};

static int32_t omni_current_log_level = OMNI_LOG_LEVEL_INFO;

// Thread-safety for logging
#ifdef _WIN32
static CRITICAL_SECTION omni_log_mutex;
static int omni_log_mutex_initialized = 0;
#else
static pthread_mutex_t omni_log_mutex = PTHREAD_MUTEX_INITIALIZER;
#endif

static int omni_equals_ignore_case(const char* a, const char* b) {
    if (!a || !b) {
        return 0;
    }
    while (*a && *b) {
        unsigned char ca = (unsigned char)(*a);
        unsigned char cb = (unsigned char)(*b);
        ca = (unsigned char)tolower(ca);
        cb = (unsigned char)tolower(cb);
        if (ca != cb) {
            return 0;
        }
        a++;
        b++;
    }
    return *a == '\0' && *b == '\0';
}

static void omni_log_write(int32_t level, const char* level_name, const char* message) {
    if (level < omni_current_log_level) {
        return;
    }
    
    // Thread-safe logging: acquire mutex
#ifdef _WIN32
    if (!omni_log_mutex_initialized) {
        InitializeCriticalSection(&omni_log_mutex);
        omni_log_mutex_initialized = 1;
    }
    EnterCriticalSection(&omni_log_mutex);
#else
    pthread_mutex_lock(&omni_log_mutex);
#endif
    
    time_t now = time(NULL);
    char timebuf[32];
#if defined(_WIN32)
    struct tm tm_info;
    if (localtime_s(&tm_info, &now) != 0) {
        strncpy(timebuf, "0000-00-00 00:00:00", sizeof(timebuf));
        timebuf[sizeof(timebuf) - 1] = '\0';
    } else {
        if (strftime(timebuf, sizeof(timebuf), "%Y-%m-%d %H:%M:%S", &tm_info) == 0) {
            strncpy(timebuf, "0000-00-00 00:00:00", sizeof(timebuf));
            timebuf[sizeof(timebuf) - 1] = '\0';
        }
    }
#else
    struct tm tm_info;
    if (localtime_r(&now, &tm_info) == NULL) {
        strncpy(timebuf, "0000-00-00 00:00:00", sizeof(timebuf));
        timebuf[sizeof(timebuf) - 1] = '\0';
    } else {
        if (strftime(timebuf, sizeof(timebuf), "%Y-%m-%d %H:%M:%S", &tm_info) == 0) {
            strncpy(timebuf, "0000-00-00 00:00:00", sizeof(timebuf));
            timebuf[sizeof(timebuf) - 1] = '\0';
        }
    }
#endif
    fprintf(stderr, "%s - [%s] %s\n", timebuf, level_name, message ? message : "");
    fflush(stderr);
    
    // Release mutex
#ifdef _WIN32
    LeaveCriticalSection(&omni_log_mutex);
#else
    pthread_mutex_unlock(&omni_log_mutex);
#endif
}

void omni_log_debug(const char* message) {
    omni_log_write(OMNI_LOG_LEVEL_DEBUG, "DEBUG", message);
}

void omni_log_info(const char* message) {
    omni_log_write(OMNI_LOG_LEVEL_INFO, "INFO", message);
}

void omni_log_warn(const char* message) {
    omni_log_write(OMNI_LOG_LEVEL_WARN, "WARN", message);
}

void omni_log_error(const char* message) {
    omni_log_write(OMNI_LOG_LEVEL_ERROR, "ERROR", message);
}

int32_t omni_log_set_level(const char* level) {
    if (!level) {
        return 0;
    }
    if (omni_equals_ignore_case(level, "DEBUG")) {
        omni_current_log_level = OMNI_LOG_LEVEL_DEBUG;
        return 1;
    }
    if (omni_equals_ignore_case(level, "INFO")) {
        omni_current_log_level = OMNI_LOG_LEVEL_INFO;
        return 1;
    }
    if (omni_equals_ignore_case(level, "WARN") || omni_equals_ignore_case(level, "WARNING")) {
        omni_current_log_level = OMNI_LOG_LEVEL_WARN;
        return 1;
    }
    if (omni_equals_ignore_case(level, "ERROR") || omni_equals_ignore_case(level, "ERR")) {
        omni_current_log_level = OMNI_LOG_LEVEL_ERROR;
        return 1;
    }
    return 0;
}

void omni_print_string(const char* str) {
    printf("%s", str);
}

void omni_println_string(const char* str) {
    printf("%s\n", str);
}

// NOTE: Returns a newly allocated string - caller must free it using free()
// This function allocates memory that must be freed by the caller to avoid leaks.
char* omni_read_line(void) {
    size_t capacity = 128;
    size_t length = 0;
    char* buffer = malloc(capacity);
    if (!buffer) {
        return NULL;
    }

    int c;
    while ((c = fgetc(stdin)) != EOF) {
        if (c == '\r') {
            int next = fgetc(stdin);
            if (next != '\n' && next != EOF) {
                ungetc(next, stdin);
            }
            break;
        }
        if (c == '\n') {
            break;
        }
        if (length + 1 >= capacity) {
            capacity *= 2;
            char* new_buffer = realloc(buffer, capacity);
            if (!new_buffer) {
                free(buffer);
                return NULL;
            }
            buffer = new_buffer;
        }
        buffer[length++] = (char)c;
    }

    if (c == EOF && length == 0) {
        buffer[0] = '\0';
        return buffer;
    }

    buffer[length] = '\0';
    return buffer;
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
// NOTE: Returns a newly allocated string - caller must free it using free()
// This function allocates memory that must be freed by the caller to avoid leaks.
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

// Helper function to find the start of the next UTF-8 rune
static const char* next_utf8_rune(const char* str) {
    if (!str || *str == '\0') return str;
    // Skip continuation bytes (0x80-0xBF)
    while ((*str & 0xC0) == 0x80) {
        str++;
    }
    return str;
}

// Helper function to find the start of the previous UTF-8 rune
static const char* __attribute__((unused)) prev_utf8_rune(const char* str, const char* start) {
    if (!str || str <= start) return start;
    str--;
    // Skip continuation bytes backwards
    while (str > start && (*str & 0xC0) == 0x80) {
        str--;
    }
    return str;
}

// NOTE: Returns a newly allocated string - caller must free it using free()
// This function allocates memory that must be freed by the caller to avoid leaks.
// WARNING: This function treats start/end as byte indices, not rune indices.
// For proper UTF-8 support, indices should be rune-based, but that requires
// counting runes which is O(n). This implementation at least ensures we don't
// split in the middle of a multi-byte character.
char* omni_substring(const char* str, int32_t start, int32_t end) {
    if (!str || start < 0 || end < start) {
        char* empty = (char*)malloc(1);
        if (empty) empty[0] = '\0';
        return empty;
    }
    
    int32_t len = (int32_t)strlen(str);
    if (start >= len) {
        char* empty = (char*)malloc(1);
        if (empty) empty[0] = '\0';
        return empty;
    }
    
    if (end > len) {
        end = len;
    }
    
    // Ensure we start at a valid UTF-8 boundary
    const char* start_ptr = str + start;
    if (start > 0) {
        start_ptr = next_utf8_rune(start_ptr);
    }
    
    // Ensure we end at a valid UTF-8 boundary
    const char* end_ptr = str + end;
    if (end < len) {
        end_ptr = next_utf8_rune(end_ptr);
    }
    
    int32_t sublen = (int32_t)(end_ptr - start_ptr);
    char* result = (char*)malloc(sublen + 1);
    if (result) {
        strncpy(result, start_ptr, sublen);
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

// NOTE: Returns a newly allocated string - caller must free it using free()
// This function allocates memory that must be freed by the caller to avoid leaks.
char* omni_trim(const char* str) {
    if (!str) {
        return NULL;
    }
    
    // Handle empty string
    size_t str_len = strlen(str);
    if (str_len == 0) {
        char* result = malloc(1);
        if (result) {
            result[0] = '\0';
        }
        return result;
    }
    
    // Find start of non-whitespace
    const char* start = str;
    while (*start && (*start == ' ' || *start == '\t' || *start == '\n' || *start == '\r')) {
        start++;
    }
    
    // If entire string is whitespace, return empty string
    if (*start == '\0') {
        char* result = malloc(1);
        if (result) {
            result[0] = '\0';
        }
        return result;
    }
    
    // Find end of non-whitespace (safe: we know start is valid)
    const char* end = str + str_len - 1;
    while (end >= start && (*end == ' ' || *end == '\t' || *end == '\n' || *end == '\r')) {
        end--;
    }
    
    // Calculate length (end >= start is guaranteed at this point)
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
    errno = 0;
    char* endptr;
    long result = strtol(str, &endptr, 10);
    // Check for conversion errors
    if (errno == ERANGE || result < INT32_MIN || result > INT32_MAX) {
        fprintf(stderr, "WARNING: String to int conversion overflow: %s\n", str);
        return 0;
    }
    if (endptr == str || *endptr != '\0') {
        // No digits found or invalid characters
        fprintf(stderr, "WARNING: Invalid integer string: %s\n", str);
        return 0;
    }
    return (int32_t)result;
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
// NOTE: Array length cannot be determined from a pointer alone in C.
// The backend should pass array length as a parameter or track it separately.
// For now, return 0 as a safe default (backend should use compile-time known lengths).
int32_t omni_array_length(int32_t* arr) {
    if (!arr) return 0;
    // Cannot determine length from pointer - backend must pass length explicitly
    // Returning 0 is safer than a hardcoded value
    return 0;
}

// Array get operation with bounds checking
int32_t omni_array_get_int(int32_t* arr, int32_t index, int32_t length) {
    if (!arr) {
        fprintf(stderr, "ERROR: Array access on NULL pointer\n");
        abort();
    }
    if (index < 0 || index >= length) {
        fprintf(stderr, "ERROR: Array index out of bounds: index=%d, length=%d\n", index, length);
        abort();
    }
    return arr[index];
}

// Array set operation with bounds checking
void omni_array_set_int(int32_t* arr, int32_t index, int32_t value, int32_t length) {
    if (!arr) {
        fprintf(stderr, "ERROR: Array access on NULL pointer\n");
        abort();
    }
    if (index < 0 || index >= length) {
        fprintf(stderr, "ERROR: Array index out of bounds: index=%d, length=%d\n", index, length);
        abort();
    }
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

// Trigonometric functions
double omni_sin(double x) {
    return sin(x);
}

double omni_cos(double x) {
    return cos(x);
}

double omni_tan(double x) {
    return tan(x);
}

double omni_asin(double x) {
    return asin(x);
}

double omni_acos(double x) {
    return acos(x);
}

double omni_atan(double x) {
    return atan(x);
}

double omni_atan2(double y, double x) {
    return atan2(y, x);
}

// Logarithmic and exponential functions
double omni_exp(double x) {
    return exp(x);
}

double omni_log(double x) {
    return log(x);
}

double omni_log10(double x) {
    return log10(x);
}

double omni_log2(double x) {
    return log2(x);
}

// Hyperbolic functions
double omni_sinh(double x) {
    return sinh(x);
}

double omni_cosh(double x) {
    return cosh(x);
}

double omni_tanh(double x) {
    return tanh(x);
}

// Additional math functions
double omni_cbrt(double x) {
    if (x == 0.0) return 0.0;
    if (x < 0.0) return -pow(-x, 1.0/3.0);
    return pow(x, 1.0/3.0);
}

double omni_trunc(double x) {
    return trunc(x);
}

// Array operations
// omni_len returns the length of an array. The length is passed explicitly by the backend.
int32_t omni_len(void* array, size_t element_size, int32_t array_length) {
    (void)array;        // Suppress unused parameter warning
    (void)element_size; // Suppress unused parameter warning
    // Return the length passed by the backend
    return array_length;
}

// File I/O operations
intptr_t omni_file_open(const char* filename, const char* mode) {
    FILE* file = fopen(filename, mode);
    if (file == NULL) {
        return (intptr_t)-1; // Error: file could not be opened
    }
    return (intptr_t)file; // Cast FILE* to intptr_t handle (safe on 64-bit)
}

int32_t omni_file_close(intptr_t file_handle) {
    if (file_handle == (intptr_t)-1) {
        return -1; // Error: invalid handle
    }
    FILE* file = (FILE*)(intptr_t)file_handle;
    return fclose(file) == 0 ? 0 : -1;
}

int32_t omni_file_read(intptr_t file_handle, char* buffer, int32_t size) {
    if (file_handle == (intptr_t)-1 || buffer == NULL || size <= 0) {
        return -1; // Error: invalid parameters
    }
    FILE* file = (FILE*)file_handle;
    size_t bytes_read = fread(buffer, 1, (size_t)size, file);
    return (int32_t)bytes_read;
}

int32_t omni_file_write(intptr_t file_handle, const char* buffer, int32_t size) {
    if (file_handle == (intptr_t)-1 || buffer == NULL || size <= 0) {
        return -1; // Error: invalid parameters
    }
    FILE* file = (FILE*)file_handle;
    size_t bytes_written = fwrite(buffer, 1, (size_t)size, file);
    return (int32_t)bytes_written;
}

int32_t omni_file_seek(intptr_t file_handle, int32_t offset, int32_t whence) {
    if (file_handle == (intptr_t)-1) {
        return -1; // Error: invalid handle
    }
    FILE* file = (FILE*)file_handle;
    return fseek(file, offset, whence) == 0 ? 0 : -1;
}

int32_t omni_file_tell(intptr_t file_handle) {
    if (file_handle == (intptr_t)-1) {
        return -1; // Error: invalid handle
    }
    FILE* file = (FILE*)file_handle;
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

void omni_test_reset() {
    // Reset test counters (useful when running multiple test files)
    total_tests = 0;
    passed_tests = 0;
    current_test_passed = 1;
}

// System operations
void omni_exit(int32_t code) {
    exit(code);
}

// Environment variable operations
const char* omni_getenv(const char* name) {
    if (!name) return NULL;
    char* value = getenv(name);
    if (!value) return NULL;
    // Return a pointer to the environment variable value
    // Note: This is valid until the environment is modified
    return value;
}

int32_t omni_setenv(const char* name, const char* value) {
    if (!name) return 0;
#ifdef _WIN32
    // Windows uses _putenv_s
    if (value) {
        char* env_str = (char*)malloc(strlen(name) + strlen(value) + 2);
        if (!env_str) return 0;
        sprintf(env_str, "%s=%s", name, value);
        int result = _putenv(env_str);
        free(env_str);
        return (result == 0) ? 1 : 0;
    } else {
        // Unset by setting to empty
        char* env_str = (char*)malloc(strlen(name) + 2);
        if (!env_str) return 0;
        sprintf(env_str, "%s=", name);
        int result = _putenv(env_str);
        free(env_str);
        return (result == 0) ? 1 : 0;
    }
#else
    // POSIX uses setenv
    if (value) {
        return (setenv(name, value, 1) == 0) ? 1 : 0;
    } else {
        return (unsetenv(name) == 0) ? 1 : 0;
    }
#endif
}

int32_t omni_unsetenv(const char* name) {
    if (!name) return 0;
#ifdef _WIN32
    // Windows: set to empty string
    char* env_str = (char*)malloc(strlen(name) + 2);
    if (!env_str) return 0;
    sprintf(env_str, "%s=", name);
    int result = _putenv(env_str);
    free(env_str);
    return (result == 0) ? 1 : 0;
#else
    return (unsetenv(name) == 0) ? 1 : 0;
#endif
}

// Directory operations
const char* omni_getcwd(void) {
    static char cwd[PATH_MAX];
#ifdef _WIN32
    if (_getcwd(cwd, sizeof(cwd)) != NULL) {
        return cwd;
    }
#else
    if (getcwd(cwd, sizeof(cwd)) != NULL) {
        return cwd;
    }
#endif
    return NULL;
}

int32_t omni_chdir(const char* path) {
    if (!path) return 0;
#ifdef _WIN32
    return (_chdir(path) == 0) ? 1 : 0;
#else
    return (chdir(path) == 0) ? 1 : 0;
#endif
}

// File system operations
int32_t omni_mkdir(const char* path) {
    if (!path) return 0;
#ifdef _WIN32
    return (_mkdir(path) == 0) ? 1 : 0;
#else
    return (mkdir(path, 0755) == 0) ? 1 : 0;
#endif
}

int32_t omni_rmdir(const char* path) {
    if (!path) return 0;
#ifdef _WIN32
    return (_rmdir(path) == 0) ? 1 : 0;
#else
    return (rmdir(path) == 0) ? 1 : 0;
#endif
}

int32_t omni_remove(const char* path) {
    if (!path) return 0;
    return (remove(path) == 0) ? 1 : 0;
}

int32_t omni_rename(const char* old_path, const char* new_path) {
    if (!old_path || !new_path) return 0;
    return (rename(old_path, new_path) == 0) ? 1 : 0;
}

int32_t omni_copy(const char* src_path, const char* dst_path) {
    if (!src_path || !dst_path) return 0;
    
    FILE* src = fopen(src_path, "rb");
    if (!src) return 0;
    
    FILE* dst = fopen(dst_path, "wb");
    if (!dst) {
        fclose(src);
        return 0;
    }
    
    char buffer[4096];
    size_t bytes;
    int32_t success = 1;
    
    while ((bytes = fread(buffer, 1, sizeof(buffer), src)) > 0) {
        if (fwrite(buffer, 1, bytes, dst) != bytes) {
            success = 0;
            break;
        }
    }
    
    fclose(src);
    fclose(dst);
    
    return success;
}

int32_t omni_exists(const char* path) {
    if (!path) return 0;
    struct stat st;
    return (stat(path, &st) == 0) ? 1 : 0;
}

int32_t omni_is_file(const char* path) {
    if (!path) return 0;
    struct stat st;
    if (stat(path, &st) != 0) return 0;
    return S_ISREG(st.st_mode) ? 1 : 0;
}

int32_t omni_is_dir(const char* path) {
    if (!path) return 0;
    struct stat st;
    if (stat(path, &st) != 0) return 0;
    return S_ISDIR(st.st_mode) ? 1 : 0;
}

// String validation functions
int32_t omni_string_is_alpha(const char* str) {
    if (!str) return 0;
    for (const char* p = str; *p; p++) {
        if (!isalpha((unsigned char)*p)) {
            return 0;
        }
    }
    return (*str != '\0') ? 1 : 0;
}

int32_t omni_string_is_digit(const char* str) {
    if (!str) return 0;
    for (const char* p = str; *p; p++) {
        if (!isdigit((unsigned char)*p)) {
            return 0;
        }
    }
    return (*str != '\0') ? 1 : 0;
}

int32_t omni_string_is_alnum(const char* str) {
    if (!str) return 0;
    for (const char* p = str; *p; p++) {
        if (!isalnum((unsigned char)*p)) {
            return 0;
        }
    }
    return (*str != '\0') ? 1 : 0;
}

int32_t omni_string_is_ascii(const char* str) {
    if (!str) return 1;
    for (const char* p = str; *p; p++) {
        if ((unsigned char)*p > 127) {
            return 0;
        }
    }
    return 1;
}

int32_t omni_string_is_upper(const char* str) {
    if (!str || *str == '\0') return 0;
    for (const char* p = str; *p; p++) {
        if (islower((unsigned char)*p)) {
            return 0;
        }
    }
    return 1;
}

int32_t omni_string_is_lower(const char* str) {
    if (!str || *str == '\0') return 0;
    for (const char* p = str; *p; p++) {
        if (isupper((unsigned char)*p)) {
            return 0;
        }
    }
    return 1;
}

// String encoding/escaping functions (base64 implementations moved below)

char* omni_encode_url(const char* str) {
    if (!str) return NULL;
    // URL encoding: %XX for non-ASCII and special chars
    size_t len = strlen(str);
    char* result = (char*)malloc(len * 3 + 1);
    if (!result) return NULL;
    
    size_t j = 0;
    for (size_t i = 0; i < len; i++) {
        unsigned char c = (unsigned char)str[i];
        if ((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || 
            (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '~') {
            result[j++] = c;
        } else {
            sprintf(result + j, "%%%02X", c);
            j += 3;
        }
    }
    result[j] = '\0';
    return result;
}

char* omni_decode_url(const char* str) {
    if (!str) return NULL;
    size_t len = strlen(str);
    char* result = (char*)malloc(len + 1);
    if (!result) return NULL;
    
    size_t j = 0;
    for (size_t i = 0; i < len; i++) {
        if (str[i] == '%' && i + 2 < len) {
            char hex[3] = {str[i+1], str[i+2], '\0'};
            int value;
            if (sscanf(hex, "%x", &value) == 1) {
                result[j++] = (char)value;
                i += 2;
            } else {
                result[j++] = str[i];
            }
        } else if (str[i] == '+') {
            result[j++] = ' ';
        } else {
            result[j++] = str[i];
        }
    }
    result[j] = '\0';
    return result;
}

char* omni_escape_html(const char* str) {
    if (!str) return NULL;
    size_t len = strlen(str);
    char* result = (char*)malloc(len * 6 + 1); // Worst case: all chars need escaping
    if (!result) return NULL;
    
    size_t j = 0;
    for (size_t i = 0; i < len; i++) {
        switch (str[i]) {
            case '&': memcpy(result + j, "&amp;", 5); j += 5; break;
            case '<': memcpy(result + j, "&lt;", 4); j += 4; break;
            case '>': memcpy(result + j, "&gt;", 4); j += 4; break;
            case '"': memcpy(result + j, "&quot;", 6); j += 6; break;
            case '\'': memcpy(result + j, "&#39;", 5); j += 5; break;
            default: result[j++] = str[i]; break;
        }
    }
    result[j] = '\0';
    return result;
}

char* omni_unescape_html(const char* str) {
    if (!str) return NULL;
    size_t len = strlen(str);
    char* result = (char*)malloc(len + 1);
    if (!result) return NULL;
    
    size_t j = 0;
    for (size_t i = 0; i < len; i++) {
        if (str[i] == '&' && i + 3 < len) {
            if (strncmp(str + i, "&lt;", 4) == 0) {
                result[j++] = '<';
                i += 3;
            } else if (strncmp(str + i, "&gt;", 4) == 0) {
                result[j++] = '>';
                i += 3;
            } else if (strncmp(str + i, "&amp;", 5) == 0) {
                result[j++] = '&';
                i += 4;
            } else if (strncmp(str + i, "&quot;", 6) == 0) {
                result[j++] = '"';
                i += 5;
            } else if (strncmp(str + i, "&#39;", 5) == 0) {
                result[j++] = '\'';
                i += 4;
            } else {
                result[j++] = str[i];
            }
        } else {
            result[j++] = str[i];
        }
    }
    result[j] = '\0';
    return result;
}

char* omni_escape_json(const char* str) {
    if (!str) return NULL;
    size_t len = strlen(str);
    char* result = (char*)malloc(len * 2 + 1);
    if (!result) return NULL;
    
    size_t j = 0;
    for (size_t i = 0; i < len; i++) {
        switch (str[i]) {
            case '"': result[j++] = '\\'; result[j++] = '"'; break;
            case '\\': result[j++] = '\\'; result[j++] = '\\'; break;
            case '\b': result[j++] = '\\'; result[j++] = 'b'; break;
            case '\f': result[j++] = '\\'; result[j++] = 'f'; break;
            case '\n': result[j++] = '\\'; result[j++] = 'n'; break;
            case '\r': result[j++] = '\\'; result[j++] = 'r'; break;
            case '\t': result[j++] = '\\'; result[j++] = 't'; break;
            default: result[j++] = str[i]; break;
        }
    }
    result[j] = '\0';
    return result;
}

char* omni_escape_shell(const char* str) {
    if (!str) return NULL;
    size_t len = strlen(str);
    char* result = (char*)malloc(len * 2 + 3); // Worst case: all chars escaped + quotes
    if (!result) return NULL;
    
    result[0] = '\'';
    size_t j = 1;
    for (size_t i = 0; i < len; i++) {
        if (str[i] == '\'') {
            result[j++] = '\'';
            result[j++] = '\\';
            result[j++] = '\'';
            result[j++] = '\'';
        } else {
            result[j++] = str[i];
        }
    }
    result[j++] = '\'';
    result[j] = '\0';
    return result;
}

// Base64 encoding/decoding (full implementation)
static const char base64_chars[] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

char* omni_encode_base64(const char* str) {
    if (!str) return NULL;
    size_t len = strlen(str);
    if (len == 0) {
        char* result = (char*)malloc(1);
        if (result) result[0] = '\0';
        return result;
    }
    
    size_t out_len = ((len + 2) / 3) * 4;
    char* result = (char*)malloc(out_len + 1);
    if (!result) return NULL;
    
    size_t i = 0, j = 0;
    for (; i < len - 2; i += 3) {
        result[j++] = base64_chars[(str[i] >> 2) & 0x3F];
        result[j++] = base64_chars[((str[i] & 0x3) << 4) | ((str[i+1] & 0xF0) >> 4)];
        result[j++] = base64_chars[((str[i+1] & 0xF) << 2) | ((str[i+2] & 0xC0) >> 6)];
        result[j++] = base64_chars[str[i+2] & 0x3F];
    }
    
    if (i < len) {
        result[j++] = base64_chars[(str[i] >> 2) & 0x3F];
        if (i == len - 1) {
            result[j++] = base64_chars[((str[i] & 0x3) << 4)];
            result[j++] = '=';
        } else {
            result[j++] = base64_chars[((str[i] & 0x3) << 4) | ((str[i+1] & 0xF0) >> 4)];
            result[j++] = base64_chars[((str[i+1] & 0xF) << 2)];
        }
        result[j++] = '=';
    }
    
    result[j] = '\0';
    return result;
}

static int base64_char_value(char c) {
    if (c >= 'A' && c <= 'Z') return c - 'A';
    if (c >= 'a' && c <= 'z') return c - 'a' + 26;
    if (c >= '0' && c <= '9') return c - '0' + 52;
    if (c == '+') return 62;
    if (c == '/') return 63;
    return -1;
}

char* omni_decode_base64(const char* str) {
    if (!str) return NULL;
    size_t len = strlen(str);
    if (len == 0) {
        char* result = (char*)malloc(1);
        if (result) result[0] = '\0';
        return result;
    }
    
    // Calculate output length
    size_t padding = 0;
    if (len > 0 && str[len-1] == '=') padding++;
    if (len > 1 && str[len-2] == '=') padding++;
    size_t out_len = (len * 3) / 4 - padding;
    
    char* result = (char*)malloc(out_len + 1);
    if (!result) return NULL;
    
    size_t i = 0, j = 0;
    for (; i < len - 4; i += 4) {
        int v1 = base64_char_value(str[i]);
        int v2 = base64_char_value(str[i+1]);
        int v3 = base64_char_value(str[i+2]);
        int v4 = base64_char_value(str[i+3]);
        
        if (v1 < 0 || v2 < 0 || v3 < 0 || v4 < 0) {
            free(result);
            return NULL;
        }
        
        result[j++] = (v1 << 2) | (v2 >> 4);
        result[j++] = ((v2 & 0xF) << 4) | (v3 >> 2);
        result[j++] = ((v3 & 0x3) << 6) | v4;
    }
    
    // Handle last 4 characters
    if (i < len) {
        int v1 = base64_char_value(str[i]);
        int v2 = base64_char_value(str[i+1]);
        if (v1 < 0 || v2 < 0) {
            free(result);
            return NULL;
        }
        result[j++] = (v1 << 2) | (v2 >> 4);
        
        if (i + 2 < len && str[i+2] != '=') {
            int v3 = base64_char_value(str[i+2]);
            if (v3 < 0) {
                free(result);
                return NULL;
            }
            result[j++] = ((v2 & 0xF) << 4) | (v3 >> 2);
            
            if (i + 3 < len && str[i+3] != '=') {
                int v4 = base64_char_value(str[i+3]);
                if (v4 < 0) {
                    free(result);
                    return NULL;
                }
                result[j++] = ((v3 & 0x3) << 6) | v4;
            }
        }
    }
    
    result[j] = '\0';
    return result;
}

// Regex functions using POSIX regex
int32_t omni_string_matches(const char* str, const char* pattern) {
    if (!str || !pattern) return 0;
    
    regex_t regex;
    int ret = regcomp(&regex, pattern, REG_EXTENDED | REG_NOSUB);
    if (ret != 0) {
        return 0; // Invalid pattern
    }
    
    ret = regexec(&regex, str, 0, NULL, 0);
    regfree(&regex);
    
    return (ret == 0) ? 1 : 0;
}

char* omni_string_find_match(const char* str, const char* pattern) {
    if (!str || !pattern) return NULL;
    
    regex_t regex;
    int ret = regcomp(&regex, pattern, REG_EXTENDED);
    if (ret != 0) {
        return NULL; // Invalid pattern
    }
    
    regmatch_t matches[1];
    ret = regexec(&regex, str, 1, matches, 0);
    
    if (ret == 0 && matches[0].rm_so >= 0) {
        size_t match_len = matches[0].rm_eo - matches[0].rm_so;
        char* result = (char*)malloc(match_len + 1);
        if (result) {
            memcpy(result, str + matches[0].rm_so, match_len);
            result[match_len] = '\0';
        }
        regfree(&regex);
        return result;
    }
    
    regfree(&regex);
    return NULL;
}

char* omni_string_find_all_matches(const char* str, const char* pattern, int32_t* count) {
    if (!str || !pattern || !count) return NULL;
    
    *count = 0;
    regex_t regex;
    int ret = regcomp(&regex, pattern, REG_EXTENDED);
    if (ret != 0) {
        return NULL; // Invalid pattern
    }
    
    // Count matches first
    regmatch_t match;
    const char* search_start = str;
    int match_count = 0;
    while (regexec(&regex, search_start, 1, &match, 0) == 0 && match.rm_so >= 0) {
        match_count++;
        search_start += match.rm_eo;
    }
    
    if (match_count == 0) {
        regfree(&regex);
        char* result = (char*)malloc(1);
        if (result) result[0] = '\0';
        return result;
    }
    
    // Allocate array: each match is stored as "start:end," format
    // For simplicity, return a comma-separated string of match positions
    // In a real implementation, this would return an array
    size_t buf_size = match_count * 32; // Rough estimate
    char* result = (char*)malloc(buf_size);
    if (!result) {
        regfree(&regex);
        return NULL;
    }
    
    *count = match_count;
    search_start = str;
    size_t pos = 0;
    int found = 0;
    while (regexec(&regex, search_start, 1, &match, 0) == 0 && match.rm_so >= 0 && pos < buf_size - 1) {
        if (found > 0) {
            result[pos++] = ',';
        }
        int start_pos = (int)(search_start - str) + match.rm_so;
        int end_pos = (int)(search_start - str) + match.rm_eo;
        int written = snprintf(result + pos, buf_size - pos, "%d:%d", start_pos, end_pos);
        if (written > 0) pos += written;
        search_start += match.rm_eo;
        found++;
    }
    result[pos] = '\0';
    
    regfree(&regex);
    return result;
}

char* omni_string_replace_regex(const char* str, const char* pattern, const char* replacement) {
    if (!str || !pattern || !replacement) return NULL;
    
    regex_t regex;
    int ret = regcomp(&regex, pattern, REG_EXTENDED);
    if (ret != 0) {
        return NULL; // Invalid pattern
    }
    
    regmatch_t matches[1];
    const char* search_start = str;
    size_t result_size = strlen(str) + strlen(replacement) + 1;
    char* result = (char*)malloc(result_size);
    if (!result) {
        regfree(&regex);
        return NULL;
    }
    
    size_t result_pos = 0;
    
    while (regexec(&regex, search_start, 1, matches, 0) == 0 && matches[0].rm_so >= 0) {
        // Copy text before match
        size_t before_len = matches[0].rm_so;
        if (result_pos + before_len + strlen(replacement) + 1 >= result_size) {
            result_size = result_pos + before_len + strlen(replacement) + 100;
            result = (char*)realloc(result, result_size);
            if (!result) {
                regfree(&regex);
                return NULL;
            }
        }
        memcpy(result + result_pos, search_start, before_len);
        result_pos += before_len;
        
        // Copy replacement
        size_t repl_len = strlen(replacement);
        memcpy(result + result_pos, replacement, repl_len);
        result_pos += repl_len;
        
        search_start += matches[0].rm_eo;
    }
    
    // Copy remaining text
    size_t remaining = strlen(search_start);
    if (result_pos + remaining + 1 >= result_size) {
        result_size = result_pos + remaining + 1;
        result = (char*)realloc(result, result_size);
        if (!result) {
            regfree(&regex);
            return NULL;
        }
    }
    memcpy(result + result_pos, search_start, remaining);
    result_pos += remaining;
    result[result_pos] = '\0';
    
    regfree(&regex);
    return result;
}

// Time functions
int64_t omni_time_now_unix(void) {
    return (int64_t)time(NULL);
}

int64_t omni_time_now_unix_nano(void) {
#ifdef _WIN32
    FILETIME ft;
    GetSystemTimeAsFileTime(&ft);
    ULARGE_INTEGER uli;
    uli.LowPart = ft.dwLowDateTime;
    uli.HighPart = ft.dwHighDateTime;
    // Windows file time is in 100-nanosecond intervals since 1601-01-01
    // Convert to Unix epoch (seconds since 1970-01-01)
    return (int64_t)((uli.QuadPart / 10000000ULL) - 11644473600ULL) * 1000000000ULL + 
           (int64_t)((uli.QuadPart % 10000000ULL) * 100);
#else
    struct timespec ts;
    clock_gettime(CLOCK_REALTIME, &ts);
    return (int64_t)ts.tv_sec * 1000000000LL + (int64_t)ts.tv_nsec;
#endif
}

void omni_time_sleep_seconds(double seconds) {
    if (seconds <= 0.0) return;
#ifdef _WIN32
    Sleep((DWORD)(seconds * 1000.0));
#else
    struct timespec req;
    req.tv_sec = (time_t)seconds;
    req.tv_nsec = (long)((seconds - req.tv_sec) * 1000000000.0);
    nanosleep(&req, NULL);
#endif
}

void omni_time_sleep_milliseconds(int32_t milliseconds) {
    if (milliseconds <= 0) return;
    omni_time_sleep_seconds(milliseconds / 1000.0);
}

int32_t omni_time_zone_offset(void) {
    time_t now = time(NULL);
    // Use portable calculation method that works on all platforms
    // Calculate offset by comparing local and UTC time
    struct tm local_tm, utc_tm;
    struct tm* local = localtime_r(&now, &local_tm);
    struct tm* utc = gmtime_r(&now, &utc_tm);
    if (!local || !utc) return 0;
    return (int32_t)(mktime(&local_tm) - mktime(&utc_tm));
}

const char* omni_time_zone_name(void) {
    static char tz_name[64];
    const char* tz = getenv("TZ");
    if (tz) {
        strncpy(tz_name, tz, sizeof(tz_name) - 1);
        tz_name[sizeof(tz_name) - 1] = '\0';
        return tz_name;
    }
    return "UTC";
}

// Command-line argument functions
static char** omni_args_array = NULL;
static int32_t omni_args_count_val = 0;

void omni_args_init(int argc, char** argv) {
    omni_args_array = argv;
    omni_args_count_val = argc;
}

int32_t omni_args_count(void) {
    return omni_args_count_val;
}

const char* omni_args_get(int32_t index) {
    if (!omni_args_array || index < 0 || index >= omni_args_count_val) {
        return NULL;
    }
    return omni_args_array[index];
}

int32_t omni_args_has_flag(const char* name) {
    if (!name || !omni_args_array) return 0;
    char flag[256];
    snprintf(flag, sizeof(flag), "--%s", name);
    for (int32_t i = 1; i < omni_args_count_val; i++) {
        if (strcmp(omni_args_array[i], flag) == 0) {
            return 1;
        }
    }
    return 0;
}

const char* omni_args_get_flag(const char* name, const char* default_value) {
    if (!name || !omni_args_array) return default_value;
    char flag[256];
    snprintf(flag, sizeof(flag), "--%s=", name);
    size_t flag_len = strlen(flag);
    
    for (int32_t i = 1; i < omni_args_count_val; i++) {
        if (strncmp(omni_args_array[i], flag, flag_len) == 0) {
            return omni_args_array[i] + flag_len;
        }
        if (strcmp(omni_args_array[i], flag) == 0 && i + 1 < omni_args_count_val) {
            return omni_args_array[i + 1];
        }
    }
    return default_value;
}

const char* omni_args_positional(int32_t index, const char* default_value) {
    if (!omni_args_array) return default_value;
    int32_t pos_idx = 0;
    for (int32_t i = 1; i < omni_args_count_val; i++) {
        if (omni_args_array[i][0] != '-') {
            if (pos_idx == index) {
                return omni_args_array[i];
            }
            pos_idx++;
        }
    }
    return default_value;
}

// Process ID functions
int32_t omni_getpid(void) {
#ifdef _WIN32
    return (int32_t)GetCurrentProcessId();
#else
    return (int32_t)getpid();
#endif
}

int32_t omni_getppid(void) {
#ifdef _WIN32
    // Windows doesn't have a direct equivalent
    return 0;
#else
    return (int32_t)getppid();
#endif
}

// Entry point - this will be implemented by the generated code
// The generated code will provide the omni_main function

// ============================================================================
// Map Implementation
// ============================================================================

// Simple hash map implementation for OmniLang maps
typedef struct omni_map_entry {
    void* key;
    void* value;
    struct omni_map_entry* next;
} omni_map_entry_t;

struct omni_map {
    omni_map_entry_t** buckets;
    int32_t bucket_count;
    int32_t size;
};

// Simple hash function for strings
static uint32_t hash_string(const char* str) {
    uint32_t hash = 5381;
    int c;
    while ((c = *str++)) {
        hash = ((hash << 5) + hash) + c;
    }
    return hash;
}

// Simple hash function for integers
static uint32_t hash_int(int32_t value) {
    return (uint32_t)value;
}

omni_map_t* omni_map_create() {
    omni_map_t* map = (omni_map_t*)malloc(sizeof(omni_map_t));
    if (!map) return NULL;
    
    map->bucket_count = 16; // Start with 16 buckets
    map->size = 0;
    map->buckets = (omni_map_entry_t**)calloc(map->bucket_count, sizeof(omni_map_entry_t*));
    if (!map->buckets) {
        free(map);
        return NULL;
    }
    
    return map;
}

// Rehash the map when load factor exceeds threshold (0.75)
// This function rehashes all entries into a new bucket array with double the size
static void omni_map_rehash(omni_map_t* map, int is_string_key) {
    if (!map || map->size == 0) return;
    
    int32_t old_bucket_count = map->bucket_count;
    omni_map_entry_t** old_buckets = map->buckets;
    
    // Double the bucket count
    map->bucket_count *= 2;
    map->buckets = (omni_map_entry_t**)calloc(map->bucket_count, sizeof(omni_map_entry_t*));
    if (!map->buckets) {
        // Rehashing failed, restore old state
        map->buckets = old_buckets;
        map->bucket_count = old_bucket_count;
        return;
    }
    
    // Rehash all entries
    for (int32_t i = 0; i < old_bucket_count; i++) {
        omni_map_entry_t* entry = old_buckets[i];
        while (entry) {
            omni_map_entry_t* next = entry->next;
            
            // Calculate new bucket
            uint32_t hash;
            if (is_string_key) {
                hash = hash_string((char*)entry->key);
            } else {
                hash = hash_int(*(int32_t*)entry->key);
            }
            int32_t new_bucket = hash % map->bucket_count;
            
            // Insert into new bucket
            entry->next = map->buckets[new_bucket];
            map->buckets[new_bucket] = entry;
            
            entry = next;
        }
    }
    
    free(old_buckets);
}

void omni_map_destroy(omni_map_t* map) {
    if (!map) return;
    
    for (int32_t i = 0; i < map->bucket_count; i++) {
        omni_map_entry_t* entry = map->buckets[i];
        while (entry) {
            omni_map_entry_t* next = entry->next;
            free(entry->key);
            free(entry->value);
            free(entry);
            entry = next;
        }
    }
    
    free(map->buckets);
    free(map);
}

// Map utility functions
int32_t omni_map_keys_string_int(omni_map_t* map, char** keys_buffer, int32_t buffer_size) {
    if (!map || !keys_buffer || buffer_size <= 0) return 0;
    
    int32_t count = 0;
    for (int32_t i = 0; i < map->bucket_count && count < buffer_size; i++) {
        omni_map_entry_t* entry = map->buckets[i];
        while (entry && count < buffer_size) {
            keys_buffer[count] = strdup((char*)entry->key);
            count++;
            entry = entry->next;
        }
    }
    return count;
}

int32_t omni_map_values_string_int(omni_map_t* map, int32_t* values_buffer, int32_t buffer_size) {
    if (!map || !values_buffer || buffer_size <= 0) return 0;
    
    int32_t count = 0;
    for (int32_t i = 0; i < map->bucket_count && count < buffer_size; i++) {
        omni_map_entry_t* entry = map->buckets[i];
        while (entry && count < buffer_size) {
            values_buffer[count] = *(int32_t*)entry->value;
            count++;
            entry = entry->next;
        }
    }
    return count;
}

omni_map_t* omni_map_copy_string_int(omni_map_t* map) {
    if (!map) return NULL;
    
    omni_map_t* new_map = omni_map_create();
    if (!new_map) return NULL;
    
    // Copy all entries
    for (int32_t i = 0; i < map->bucket_count; i++) {
        omni_map_entry_t* entry = map->buckets[i];
        while (entry) {
            omni_map_put_string_int(new_map, (char*)entry->key, *(int32_t*)entry->value);
            entry = entry->next;
        }
    }
    
    return new_map;
}

omni_map_t* omni_map_merge_string_int(omni_map_t* a, omni_map_t* b) {
    if (!a && !b) return NULL;
    if (!a) return omni_map_copy_string_int(b);
    if (!b) return omni_map_copy_string_int(a);
    
    omni_map_t* merged = omni_map_copy_string_int(a);
    if (!merged) return NULL;
    
    // Add all entries from b (will overwrite duplicates)
    for (int32_t i = 0; i < b->bucket_count; i++) {
        omni_map_entry_t* entry = b->buckets[i];
        while (entry) {
            omni_map_put_string_int(merged, (char*)entry->key, *(int32_t*)entry->value);
            entry = entry->next;
        }
    }
    
    return merged;
}

void omni_map_put_string_int(omni_map_t* map, const char* key, int32_t value) {
    if (!map) return;
    
    uint32_t hash = hash_string(key);
    int32_t bucket = hash % map->bucket_count;
    
    // Check if key already exists
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (strcmp((char*)entry->key, key) == 0) {
            // Update existing value
            *(int32_t*)entry->value = value;
            return;
        }
        entry = entry->next;
    }
    
    // Rehash if load factor exceeds 0.75
    if (map->size * 4 >= map->bucket_count * 3) {
        omni_map_rehash(map, 1); // 1 = string key
        bucket = hash % map->bucket_count; // Recalculate bucket after rehash
    }
    
    // Create new entry
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    entry->key = malloc(strlen(key) + 1);
    entry->value = malloc(sizeof(int32_t));
    if (!entry->key || !entry->value) {
        free(entry->key);
        free(entry->value);
        free(entry);
        return;
    }
    
    strcpy((char*)entry->key, key);
    *(int32_t*)entry->value = value;
    entry->next = map->buckets[bucket];
    map->buckets[bucket] = entry;
    map->size++;
}

void omni_map_put_int_int(omni_map_t* map, int32_t key, int32_t value) {
    if (!map) return;
    
    uint32_t hash = hash_int(key);
    int32_t bucket = hash % map->bucket_count;
    
    // Check if key already exists
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (*(int32_t*)entry->key == key) {
            // Update existing value
            *(int32_t*)entry->value = value;
            return;
        }
        entry = entry->next;
    }
    
    // Rehash if load factor exceeds 0.75
    if (map->size * 4 >= map->bucket_count * 3) {
        omni_map_rehash(map, 0); // 0 = int key
        bucket = hash % map->bucket_count; // Recalculate bucket after rehash
    }
    
    // Create new entry
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    entry->key = malloc(sizeof(int32_t));
    entry->value = malloc(sizeof(int32_t));
    if (!entry->key || !entry->value) {
        free(entry->key);
        free(entry->value);
        free(entry);
        return;
    }
    
    *(int32_t*)entry->key = key;
    *(int32_t*)entry->value = value;
    entry->next = map->buckets[bucket];
    map->buckets[bucket] = entry;
    map->size++;
}

int32_t omni_map_get_string_int(omni_map_t* map, const char* key) {
    if (!map) return 0;
    
    uint32_t hash = hash_string(key);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (strcmp((char*)entry->key, key) == 0) {
            return *(int32_t*)entry->value;
        }
        entry = entry->next;
    }
    
    return 0; // Key not found, return default value
}

int32_t omni_map_get_int_int(omni_map_t* map, int32_t key) {
    if (!map) return 0;
    
    uint32_t hash = hash_int(key);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (*(int32_t*)entry->key == key) {
            return *(int32_t*)entry->value;
        }
        entry = entry->next;
    }
    
    return 0; // Key not found, return default value
}

int32_t omni_map_contains_string(omni_map_t* map, const char* key) {
    if (!map) return 0;
    
    uint32_t hash = hash_string(key);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (strcmp((char*)entry->key, key) == 0) {
            return 1; // Found
        }
        entry = entry->next;
    }
    
    return 0; // Not found
}

int32_t omni_map_contains_int(omni_map_t* map, int32_t key) {
    if (!map) return 0;
    
    uint32_t hash = hash_int(key);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (*(int32_t*)entry->key == key) {
            return 1; // Found
        }
        entry = entry->next;
    }
    
    return 0; // Not found
}

int32_t omni_map_size(omni_map_t* map) {
    return map ? map->size : 0;
}

// Additional map put operations
void omni_map_put_string_string(omni_map_t* map, const char* key, const char* value) {
    if (!map) return;
    
    uint32_t hash = hash_string(key);
    int32_t bucket = hash % map->bucket_count;
    
    // Check if key already exists
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (strcmp((char*)entry->key, key) == 0) {
            // Update existing value - free old value before allocating new one
            free(entry->value);
            entry->value = malloc(strlen(value) + 1);
            if (entry->value) {
                strcpy((char*)entry->value, value);
            }
            // Key stays the same, no need to free/reallocate
            return;
        }
        entry = entry->next;
    }
    
    // Rehash if load factor exceeds 0.75
    if (map->size * 4 >= map->bucket_count * 3) {
        omni_map_rehash(map, 1); // 1 = string key
        bucket = hash % map->bucket_count; // Recalculate bucket after rehash
    }
    
    // Create new entry
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    entry->key = malloc(strlen(key) + 1);
    entry->value = malloc(strlen(value) + 1);
    if (!entry->key || !entry->value) {
        free(entry->key);
        free(entry->value);
        free(entry);
        return;
    }
    
    strcpy((char*)entry->key, key);
    strcpy((char*)entry->value, value);
    entry->next = map->buckets[bucket];
    map->buckets[bucket] = entry;
    map->size++;
}

void omni_map_put_string_float(omni_map_t* map, const char* key, double value) {
    if (!map) return;
    
    uint32_t hash = hash_string(key);
    int32_t bucket = hash % map->bucket_count;
    
    // Check if key already exists
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (strcmp((char*)entry->key, key) == 0) {
            *(double*)entry->value = value;
            return;
        }
        entry = entry->next;
    }
    
    // Rehash if load factor exceeds 0.75
    if (map->size * 4 >= map->bucket_count * 3) {
        omni_map_rehash(map, 1); // 1 = string key
        bucket = hash % map->bucket_count; // Recalculate bucket after rehash
    }
    
    // Create new entry
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    entry->key = malloc(strlen(key) + 1);
    entry->value = malloc(sizeof(double));
    if (!entry->key || !entry->value) {
        free(entry->key);
        free(entry->value);
        free(entry);
        return;
    }
    
    strcpy((char*)entry->key, key);
    *(double*)entry->value = value;
    entry->next = map->buckets[bucket];
    map->buckets[bucket] = entry;
    map->size++;
}

void omni_map_put_string_bool(omni_map_t* map, const char* key, int32_t value) {
    if (!map) return;
    
    uint32_t hash = hash_string(key);
    int32_t bucket = hash % map->bucket_count;
    
    // Check if key already exists
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (strcmp((char*)entry->key, key) == 0) {
            *(int32_t*)entry->value = value;
            return;
        }
        entry = entry->next;
    }
    
    // Rehash if load factor exceeds 0.75
    if (map->size * 4 >= map->bucket_count * 3) {
        omni_map_rehash(map, 1); // 1 = string key
        bucket = hash % map->bucket_count; // Recalculate bucket after rehash
    }
    
    // Create new entry
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    entry->key = malloc(strlen(key) + 1);
    entry->value = malloc(sizeof(int32_t));
    if (!entry->key || !entry->value) {
        free(entry->key);
        free(entry->value);
        free(entry);
        return;
    }
    
    strcpy((char*)entry->key, key);
    *(int32_t*)entry->value = value;
    entry->next = map->buckets[bucket];
    map->buckets[bucket] = entry;
    map->size++;
}

void omni_map_put_int_string(omni_map_t* map, int32_t key, const char* value) {
    if (!map) return;
    
    uint32_t hash = hash_int(key);
    int32_t bucket = hash % map->bucket_count;
    
    // Check if key already exists
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (*(int32_t*)entry->key == key) {
            free(entry->value);
            entry->value = malloc(strlen(value) + 1);
            if (entry->value) {
                strcpy((char*)entry->value, value);
            }
            return;
        }
        entry = entry->next;
    }
    
    // Rehash if load factor exceeds 0.75
    if (map->size * 4 >= map->bucket_count * 3) {
        omni_map_rehash(map, 0); // 0 = int key
        bucket = hash % map->bucket_count; // Recalculate bucket after rehash
    }
    
    // Create new entry
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    entry->key = malloc(sizeof(int32_t));
    entry->value = malloc(strlen(value) + 1);
    if (!entry->key || !entry->value) {
        free(entry->key);
        free(entry->value);
        free(entry);
        return;
    }
    
    *(int32_t*)entry->key = key;
    strcpy((char*)entry->value, value);
    entry->next = map->buckets[bucket];
    map->buckets[bucket] = entry;
    map->size++;
}

void omni_map_put_int_float(omni_map_t* map, int32_t key, double value) {
    if (!map) return;
    
    uint32_t hash = hash_int(key);
    int32_t bucket = hash % map->bucket_count;
    
    // Check if key already exists
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (*(int32_t*)entry->key == key) {
            *(double*)entry->value = value;
            return;
        }
        entry = entry->next;
    }
    
    // Rehash if load factor exceeds 0.75
    if (map->size * 4 >= map->bucket_count * 3) {
        omni_map_rehash(map, 0); // 0 = int key
        bucket = hash % map->bucket_count; // Recalculate bucket after rehash
    }
    
    // Create new entry
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    entry->key = malloc(sizeof(int32_t));
    entry->value = malloc(sizeof(double));
    if (!entry->key || !entry->value) {
        free(entry->key);
        free(entry->value);
        free(entry);
        return;
    }
    
    *(int32_t*)entry->key = key;
    *(double*)entry->value = value;
    entry->next = map->buckets[bucket];
    map->buckets[bucket] = entry;
    map->size++;
}

void omni_map_put_int_bool(omni_map_t* map, int32_t key, int32_t value) {
    if (!map) return;
    
    uint32_t hash = hash_int(key);
    int32_t bucket = hash % map->bucket_count;
    
    // Check if key already exists
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (*(int32_t*)entry->key == key) {
            *(int32_t*)entry->value = value;
            return;
        }
        entry = entry->next;
    }
    
    // Rehash if load factor exceeds 0.75
    if (map->size * 4 >= map->bucket_count * 3) {
        omni_map_rehash(map, 0); // 0 = int key
        bucket = hash % map->bucket_count; // Recalculate bucket after rehash
    }
    
    // Create new entry
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    entry->key = malloc(sizeof(int32_t));
    entry->value = malloc(sizeof(int32_t));
    if (!entry->key || !entry->value) {
        free(entry->key);
        free(entry->value);
        free(entry);
        return;
    }
    
    *(int32_t*)entry->key = key;
    *(int32_t*)entry->value = value;
    entry->next = map->buckets[bucket];
    map->buckets[bucket] = entry;
    map->size++;
}

// Additional map get operations
const char* omni_map_get_string_string(omni_map_t* map, const char* key) {
    if (!map) return NULL;
    
    uint32_t hash = hash_string(key);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (strcmp((char*)entry->key, key) == 0) {
            return (const char*)entry->value;
        }
        entry = entry->next;
    }
    
    return NULL; // Key not found
}

double omni_map_get_string_float(omni_map_t* map, const char* key) {
    if (!map) return 0.0;
    
    uint32_t hash = hash_string(key);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (strcmp((char*)entry->key, key) == 0) {
            return *(double*)entry->value;
        }
        entry = entry->next;
    }
    
    return 0.0; // Key not found
}

int32_t omni_map_get_string_bool(omni_map_t* map, const char* key) {
    if (!map) return 0;
    
    uint32_t hash = hash_string(key);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (strcmp((char*)entry->key, key) == 0) {
            return *(int32_t*)entry->value;
        }
        entry = entry->next;
    }
    
    return 0; // Key not found
}

const char* omni_map_get_int_string(omni_map_t* map, int32_t key) {
    if (!map) return NULL;
    
    uint32_t hash = hash_int(key);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (*(int32_t*)entry->key == key) {
            return (const char*)entry->value;
        }
        entry = entry->next;
    }
    
    return NULL; // Key not found
}

double omni_map_get_int_float(omni_map_t* map, int32_t key) {
    if (!map) return 0.0;
    
    uint32_t hash = hash_int(key);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (*(int32_t*)entry->key == key) {
            return *(double*)entry->value;
        }
        entry = entry->next;
    }
    
    return 0.0; // Key not found
}

int32_t omni_map_get_int_bool(omni_map_t* map, int32_t key) {
    if (!map) return 0;
    
    uint32_t hash = hash_int(key);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (*(int32_t*)entry->key == key) {
            return *(int32_t*)entry->value;
        }
        entry = entry->next;
    }
    
    return 0; // Key not found
}

// Map delete operations
void omni_map_delete_string(omni_map_t* map, const char* key) {
    if (!map) return;
    
    uint32_t hash = hash_string(key);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    omni_map_entry_t* prev = NULL;
    
    while (entry) {
        if (strcmp((char*)entry->key, key) == 0) {
            if (prev) {
                prev->next = entry->next;
            } else {
                map->buckets[bucket] = entry->next;
            }
            free(entry->key);
            free(entry->value);
            free(entry);
            map->size--;
            return;
        }
        prev = entry;
        entry = entry->next;
    }
}

void omni_map_delete_int(omni_map_t* map, int32_t key) {
    if (!map) return;
    
    uint32_t hash = hash_int(key);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    omni_map_entry_t* prev = NULL;
    
    while (entry) {
        if (*(int32_t*)entry->key == key) {
            if (prev) {
                prev->next = entry->next;
            } else {
                map->buckets[bucket] = entry->next;
            }
            free(entry->key);
            free(entry->value);
            free(entry);
            map->size--;
            return;
        }
        prev = entry;
        entry = entry->next;
    }
}

// ============================================================================
// Struct Implementation
// ============================================================================

// Simple struct implementation for OmniLang structs
// NOTE: This implementation only supports primitive field types (string, int, float, bool).
// Nested structs and arrays as field values are not supported and will cause
// memory leaks or crashes. For nested structs, the runtime would need to:
// 1. Store struct values as omni_struct_t* pointers
// 2. Recursively free nested structs in omni_struct_destroy
// 3. Handle type checking for nested struct access
// This is a known limitation that requires architectural changes to fix.
typedef struct omni_struct_field {
    char* name;
    void* value;
    int32_t value_type; // 0=string, 1=int, 2=float, 3=bool
    struct omni_struct_field* next;
} omni_struct_field_t;

struct omni_struct {
    omni_struct_field_t* fields;
};

omni_struct_t* omni_struct_create() {
    omni_struct_t* struct_ptr = (omni_struct_t*)malloc(sizeof(omni_struct_t));
    if (!struct_ptr) return NULL;
    
    struct_ptr->fields = NULL;
    return struct_ptr;
}

void omni_struct_destroy(omni_struct_t* struct_ptr) {
    if (!struct_ptr) return;
    
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        omni_struct_field_t* next = field->next;
        free(field->name);
        if (field->value_type == 0) { // string
            free((char*)field->value);
        } else {
            free(field->value);
        }
        free(field);
        field = next;
    }
    
    free(struct_ptr);
}

void omni_struct_set_string_field(omni_struct_t* struct_ptr, const char* field_name, const char* value) {
    if (!struct_ptr) return;
    if (!value) return; // Skip NULL values
    
    // Check if field already exists
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        if (strcmp(field->name, field_name) == 0) {
            // Update existing field
            if (field->value_type == 0) { // string
                free((char*)field->value);
            } else {
                free(field->value);
            }
            field->value = malloc(strlen(value) + 1);
            if (field->value) {
                strcpy((char*)field->value, value);
                field->value_type = 0; // string
            }
            return;
        }
        field = field->next;
    }
    
    // Create new field
    field = (omni_struct_field_t*)malloc(sizeof(omni_struct_field_t));
    if (!field) return;
    
    field->name = malloc(strlen(field_name) + 1);
    field->value = malloc(strlen(value) + 1);
    if (!field->name || !field->value) {
        free(field->name);
        free(field->value);
        free(field);
        return;
    }
    
    strcpy(field->name, field_name);
    strcpy((char*)field->value, value);
    field->value_type = 0; // string
    field->next = struct_ptr->fields;
    struct_ptr->fields = field;
}

void omni_struct_set_int_field(omni_struct_t* struct_ptr, const char* field_name, int32_t value) {
    if (!struct_ptr) return;
    
    // Check if field already exists
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        if (strcmp(field->name, field_name) == 0) {
            // Update existing field
            if (field->value_type == 0) { // string
                free((char*)field->value);
            } else {
                free(field->value);
            }
            field->value = malloc(sizeof(int32_t));
            if (field->value) {
                *(int32_t*)field->value = value;
                field->value_type = 1; // int
            }
            return;
        }
        field = field->next;
    }
    
    // Create new field
    field = (omni_struct_field_t*)malloc(sizeof(omni_struct_field_t));
    if (!field) return;
    
    field->name = malloc(strlen(field_name) + 1);
    field->value = malloc(sizeof(int32_t));
    if (!field->name || !field->value) {
        free(field->name);
        free(field->value);
        free(field);
        return;
    }
    
    strcpy(field->name, field_name);
    *(int32_t*)field->value = value;
    field->value_type = 1; // int
    field->next = struct_ptr->fields;
    struct_ptr->fields = field;
}

void omni_struct_set_float_field(omni_struct_t* struct_ptr, const char* field_name, double value) {
    if (!struct_ptr) return;
    
    // Check if field already exists
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        if (strcmp(field->name, field_name) == 0) {
            // Update existing field
            if (field->value_type == 0) { // string
                free((char*)field->value);
            } else {
                free(field->value);
            }
            field->value = malloc(sizeof(double));
            if (field->value) {
                *(double*)field->value = value;
                field->value_type = 2; // float
            }
            return;
        }
        field = field->next;
    }
    
    // Create new field
    field = (omni_struct_field_t*)malloc(sizeof(omni_struct_field_t));
    if (!field) return;
    
    field->name = malloc(strlen(field_name) + 1);
    field->value = malloc(sizeof(double));
    if (!field->name || !field->value) {
        free(field->name);
        free(field->value);
        free(field);
        return;
    }
    
    strcpy(field->name, field_name);
    *(double*)field->value = value;
    field->value_type = 2; // float
    field->next = struct_ptr->fields;
    struct_ptr->fields = field;
}

void omni_struct_set_bool_field(omni_struct_t* struct_ptr, const char* field_name, int32_t value) {
    if (!struct_ptr) return;
    
    // Check if field already exists
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        if (strcmp(field->name, field_name) == 0) {
            // Update existing field
            if (field->value_type == 0) { // string
                free((char*)field->value);
            } else {
                free(field->value);
            }
            field->value = malloc(sizeof(int32_t));
            if (field->value) {
                *(int32_t*)field->value = value;
                field->value_type = 3; // bool
            }
            return;
        }
        field = field->next;
    }
    
    // Create new field
    field = (omni_struct_field_t*)malloc(sizeof(omni_struct_field_t));
    if (!field) return;
    
    field->name = malloc(strlen(field_name) + 1);
    field->value = malloc(sizeof(int32_t));
    if (!field->name || !field->value) {
        free(field->name);
        free(field->value);
        free(field);
        return;
    }
    
    strcpy(field->name, field_name);
    *(int32_t*)field->value = value;
    field->value_type = 3; // bool
    field->next = struct_ptr->fields;
    struct_ptr->fields = field;
}

const char* omni_struct_get_string_field(omni_struct_t* struct_ptr, const char* field_name) {
    if (!struct_ptr) return "";
    
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        if (strcmp(field->name, field_name) == 0 && field->value_type == 0) {
            return (const char*)field->value;
        }
        field = field->next;
    }
    
    return ""; // Field not found, return default value
}

int32_t omni_struct_get_int_field(omni_struct_t* struct_ptr, const char* field_name) {
    if (!struct_ptr) return 0;
    
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        if (strcmp(field->name, field_name) == 0 && field->value_type == 1) {
            return *(int32_t*)field->value;
        }
        field = field->next;
    }
    
    return 0; // Field not found, return default value
}

double omni_struct_get_float_field(omni_struct_t* struct_ptr, const char* field_name) {
    if (!struct_ptr) return 0.0;
    
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        if (strcmp(field->name, field_name) == 0 && field->value_type == 2) {
            return *(double*)field->value;
        }
        field = field->next;
    }
    
    return 0.0; // Field not found, return default value
}

int32_t omni_struct_get_bool_field(omni_struct_t* struct_ptr, const char* field_name) {
    if (!struct_ptr) return 0;
    
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        if (strcmp(field->name, field_name) == 0 && field->value_type == 3) {
            return *(int32_t*)field->value;
        }
        field = field->next;
    }
    
    return 0; // Field not found, return default value
}

// Promise/Async support (simplified synchronous implementation)
omni_promise_t* omni_promise_create_int(int32_t value) {
    omni_promise_t* promise = (omni_promise_t*)malloc(sizeof(omni_promise_t));
    if (!promise) return NULL;
    promise->type = 0; // int
    promise->value = malloc(sizeof(int32_t));
    if (!promise->value) {
        free(promise);
        return NULL;
    }
    *(int32_t*)promise->value = value;
    promise->done = 1;
    return promise;
}

omni_promise_t* omni_promise_create_string(const char* value) {
    omni_promise_t* promise = (omni_promise_t*)malloc(sizeof(omni_promise_t));
    if (!promise) return NULL;
    promise->type = 1; // string
    if (value) {
        promise->value = strdup(value);
    } else {
        promise->value = strdup("");
    }
    promise->done = 1;
    return promise;
}

omni_promise_t* omni_promise_create_float(double value) {
    omni_promise_t* promise = (omni_promise_t*)malloc(sizeof(omni_promise_t));
    if (!promise) return NULL;
    promise->type = 2; // float
    promise->value = malloc(sizeof(double));
    if (!promise->value) {
        free(promise);
        return NULL;
    }
    *(double*)promise->value = value;
    promise->done = 1;
    return promise;
}

omni_promise_t* omni_promise_create_bool(int32_t value) {
    omni_promise_t* promise = (omni_promise_t*)malloc(sizeof(omni_promise_t));
    if (!promise) return NULL;
    promise->type = 3; // bool
    promise->value = malloc(sizeof(int32_t));
    if (!promise->value) {
        free(promise);
        return NULL;
    }
    *(int32_t*)promise->value = value;
    promise->done = 1;
    return promise;
}

int32_t omni_await_int(omni_promise_t* promise) {
    if (!promise || !promise->done || promise->type != 0) {
        return 0;
    }
    return *(int32_t*)promise->value;
}

// NOTE: This function returns a newly allocated copy of the string stored in the promise.
// The caller is responsible for freeing the returned string using free().
// This ensures the string remains valid even after the promise is freed.
char* omni_await_string(omni_promise_t* promise) {
    if (!promise || !promise->done || promise->type != 1) {
        char* empty = (char*)malloc(1);
        if (empty) {
            empty[0] = '\0';
        }
        return empty;
    }
    const char* str = (const char*)promise->value;
    if (!str) {
        char* empty = (char*)malloc(1);
        if (empty) {
            empty[0] = '\0';
        }
        return empty;
    }
    return strdup(str);
}

double omni_await_float(omni_promise_t* promise) {
    if (!promise || !promise->done || promise->type != 2) {
        return 0.0;
    }
    return *(double*)promise->value;
}

int32_t omni_await_bool(omni_promise_t* promise) {
    if (!promise || !promise->done || promise->type != 3) {
        return 0;
    }
    return *(int32_t*)promise->value;
}

// NOTE: This function frees the promise and all its associated memory.
// For string promises, the string returned by omni_await_string becomes invalid after this call.
// Callers must copy the string (e.g., using strdup) if they need it after freeing the promise.
void omni_promise_free(omni_promise_t* promise) {
    if (!promise) return;
    if (promise->value) {
        // For string promises, value is a strdup'd string that needs freeing
        // For other types, value is a malloc'd buffer that needs freeing
        free(promise->value);
    }
    free(promise);
}

// File I/O convenience functions (for async operations)
// NOTE: Returns a newly allocated string - caller must free it using free()
// This function allocates memory that must be freed by the caller to avoid leaks.
char* omni_read_file(const char* path) {
    if (!path) {
        return NULL;
    }
    
    FILE* file = fopen(path, "r");
    if (!file) {
        // Return NULL on error instead of leaking strdup("")
        return NULL;
    }
    
    fseek(file, 0, SEEK_END);
    long size = ftell(file);
    fseek(file, 0, SEEK_SET);
    
    // Handle empty files
    if (size <= 0) {
        fclose(file);
        char* empty = (char*)malloc(1);
        if (empty) {
            empty[0] = '\0';
        }
        return empty;
    }
    
    char* buffer = (char*)malloc(size + 1);
    if (!buffer) {
        fclose(file);
        return NULL;
    }
    
    size_t read = fread(buffer, 1, size, file);
    buffer[read] = '\0';
    fclose(file);
    
    return buffer;
}

int32_t omni_write_file(const char* path, const char* content) {
    if (!path) {
        fprintf(stderr, "ERROR: omni_write_file called with NULL path\n");
        return 0;
    }
    if (!content) {
        fprintf(stderr, "ERROR: omni_write_file called with NULL content\n");
        return 0;
    }
    
    FILE* file = fopen(path, "w");
    if (!file) {
        return 0;
    }
    
    size_t len = strlen(content);
    size_t written = fwrite(content, 1, len, file);
    fclose(file);
    
    return (written == len) ? 1 : 0;
}

int32_t omni_append_file(const char* path, const char* content) {
    if (!path) {
        fprintf(stderr, "ERROR: omni_append_file called with NULL path\n");
        return 0;
    }
    if (!content) {
        fprintf(stderr, "ERROR: omni_append_file called with NULL content\n");
        return 0;
    }
    
    FILE* file = fopen(path, "a");
    if (!file) {
        return 0;
    }
    
    size_t len = strlen(content);
    size_t written = fwrite(content, 1, len, file);
    fclose(file);
    
    return (written == len) ? 1 : 0;
}

// ============================================================================
// Collection Data Structures Implementation
// ============================================================================

// Set implementation (using map internally)
struct omni_set {
    omni_map_t* map; // Use map to store set elements (key = element, value = 1)
};

omni_set_t* omni_set_create() {
    omni_set_t* set = (omni_set_t*)malloc(sizeof(omni_set_t));
    if (!set) return NULL;
    set->map = omni_map_create();
    if (!set->map) {
        free(set);
        return NULL;
    }
    return set;
}

void omni_set_destroy(omni_set_t* set) {
    if (!set) return;
    if (set->map) omni_map_destroy(set->map);
    free(set);
}

int32_t omni_set_add(omni_set_t* set, int32_t element) {
    if (!set || !set->map) return 0;
    omni_map_put_int_int(set->map, element, 1);
    return 1;
}

int32_t omni_set_remove(omni_set_t* set, int32_t element) {
    if (!set || !set->map) return 0;
    if (omni_map_contains_int(set->map, element)) {
        omni_map_delete_int(set->map, element);
        return 1;
    }
    return 0;
}

int32_t omni_set_contains(omni_set_t* set, int32_t element) {
    if (!set || !set->map) return 0;
    return omni_map_contains_int(set->map, element);
}

int32_t omni_set_size(omni_set_t* set) {
    if (!set || !set->map) return 0;
    return omni_map_size(set->map);
}

void omni_set_clear(omni_set_t* set) {
    if (!set || !set->map) return;
    // Recreate the map to clear it
    omni_map_destroy(set->map);
    set->map = omni_map_create();
}

omni_set_t* omni_set_union(omni_set_t* a, omni_set_t* b) {
    if (!a && !b) return NULL;
    omni_set_t* result = omni_set_create();
    if (!result) return NULL;
    
    // Add all elements from a
    if (a && a->map) {
        for (int32_t i = 0; i < a->map->bucket_count; i++) {
            omni_map_entry_t* entry = a->map->buckets[i];
            while (entry) {
                omni_set_add(result, *(int32_t*)entry->key);
                entry = entry->next;
            }
        }
    }
    
    // Add all elements from b
    if (b && b->map) {
        for (int32_t i = 0; i < b->map->bucket_count; i++) {
            omni_map_entry_t* entry = b->map->buckets[i];
            while (entry) {
                omni_set_add(result, *(int32_t*)entry->key);
                entry = entry->next;
            }
        }
    }
    
    return result;
}

omni_set_t* omni_set_intersection(omni_set_t* a, omni_set_t* b) {
    if (!a || !b) return omni_set_create();
    omni_set_t* result = omni_set_create();
    if (!result) return NULL;
    
    // Add elements that are in both sets
    if (a->map && b->map) {
        for (int32_t i = 0; i < a->map->bucket_count; i++) {
            omni_map_entry_t* entry = a->map->buckets[i];
            while (entry) {
                int32_t element = *(int32_t*)entry->key;
                if (omni_set_contains(b, element)) {
                    omni_set_add(result, element);
                }
                entry = entry->next;
            }
        }
    }
    
    return result;
}

omni_set_t* omni_set_difference(omni_set_t* a, omni_set_t* b) {
    if (!a) return omni_set_create();
    omni_set_t* result = omni_set_create();
    if (!result) return NULL;
    
    // Add elements from a that are not in b
    if (a->map) {
        for (int32_t i = 0; i < a->map->bucket_count; i++) {
            omni_map_entry_t* entry = a->map->buckets[i];
            while (entry) {
                int32_t element = *(int32_t*)entry->key;
                if (!b || !omni_set_contains(b, element)) {
                    omni_set_add(result, element);
                }
                entry = entry->next;
            }
        }
    }
    
    return result;
}

// Queue implementation (FIFO using linked list)
typedef struct omni_queue_node {
    int32_t value;
    struct omni_queue_node* next;
} omni_queue_node_t;

struct omni_queue {
    omni_queue_node_t* front;
    omni_queue_node_t* rear;
    int32_t size;
};

omni_queue_t* omni_queue_create() {
    omni_queue_t* queue = (omni_queue_t*)malloc(sizeof(omni_queue_t));
    if (!queue) return NULL;
    queue->front = NULL;
    queue->rear = NULL;
    queue->size = 0;
    return queue;
}

void omni_queue_destroy(omni_queue_t* queue) {
    if (!queue) return;
    omni_queue_clear(queue);
    free(queue);
}

void omni_queue_enqueue(omni_queue_t* queue, int32_t element) {
    if (!queue) return;
    omni_queue_node_t* node = (omni_queue_node_t*)malloc(sizeof(omni_queue_node_t));
    if (!node) return;
    node->value = element;
    node->next = NULL;
    
    if (queue->rear == NULL) {
        queue->front = queue->rear = node;
    } else {
        queue->rear->next = node;
        queue->rear = node;
    }
    queue->size++;
}

int32_t omni_queue_dequeue(omni_queue_t* queue) {
    if (!queue || queue->front == NULL) return 0;
    omni_queue_node_t* node = queue->front;
    int32_t value = node->value;
    queue->front = queue->front->next;
    if (queue->front == NULL) {
        queue->rear = NULL;
    }
    free(node);
    queue->size--;
    return value;
}

int32_t omni_queue_peek(omni_queue_t* queue) {
    if (!queue || queue->front == NULL) return 0;
    return queue->front->value;
}

int32_t omni_queue_is_empty(omni_queue_t* queue) {
    return (!queue || queue->front == NULL) ? 1 : 0;
}

int32_t omni_queue_size(omni_queue_t* queue) {
    return queue ? queue->size : 0;
}

void omni_queue_clear(omni_queue_t* queue) {
    if (!queue) return;
    while (queue->front != NULL) {
        omni_queue_dequeue(queue);
    }
}

// Stack implementation (LIFO using linked list)
typedef struct omni_stack_node {
    int32_t value;
    struct omni_stack_node* next;
} omni_stack_node_t;

struct omni_stack {
    omni_stack_node_t* top;
    int32_t size;
};

omni_stack_t* omni_stack_create() {
    omni_stack_t* stack = (omni_stack_t*)malloc(sizeof(omni_stack_t));
    if (!stack) return NULL;
    stack->top = NULL;
    stack->size = 0;
    return stack;
}

void omni_stack_destroy(omni_stack_t* stack) {
    if (!stack) return;
    omni_stack_clear(stack);
    free(stack);
}

void omni_stack_push(omni_stack_t* stack, int32_t element) {
    if (!stack) return;
    omni_stack_node_t* node = (omni_stack_node_t*)malloc(sizeof(omni_stack_node_t));
    if (!node) return;
    node->value = element;
    node->next = stack->top;
    stack->top = node;
    stack->size++;
}

int32_t omni_stack_pop(omni_stack_t* stack) {
    if (!stack || stack->top == NULL) return 0;
    omni_stack_node_t* node = stack->top;
    int32_t value = node->value;
    stack->top = node->next;
    free(node);
    stack->size--;
    return value;
}

int32_t omni_stack_peek(omni_stack_t* stack) {
    if (!stack || stack->top == NULL) return 0;
    return stack->top->value;
}

int32_t omni_stack_is_empty(omni_stack_t* stack) {
    return (!stack || stack->top == NULL) ? 1 : 0;
}

int32_t omni_stack_size(omni_stack_t* stack) {
    return stack ? stack->size : 0;
}

void omni_stack_clear(omni_stack_t* stack) {
    if (!stack) return;
    while (stack->top != NULL) {
        omni_stack_pop(stack);
    }
}

// Priority queue implementation (max-heap using array)
#define OMNI_PQ_MAX_SIZE 1024
typedef struct omni_priority_queue_node {
    int32_t element;
    int32_t priority;
} omni_priority_queue_node_t;

struct omni_priority_queue {
    omni_priority_queue_node_t* heap;
    int32_t size;
    int32_t capacity;
};

static void omni_pq_heapify_up(omni_priority_queue_node_t* heap, int32_t index) {
    while (index > 0) {
        int32_t parent = (index - 1) / 2;
        if (heap[parent].priority >= heap[index].priority) break;
        // Swap
        omni_priority_queue_node_t temp = heap[parent];
        heap[parent] = heap[index];
        heap[index] = temp;
        index = parent;
    }
}

static void omni_pq_heapify_down(omni_priority_queue_node_t* heap, int32_t size, int32_t index) {
    while (1) {
        int32_t left = 2 * index + 1;
        int32_t right = 2 * index + 2;
        int32_t largest = index;
        
        if (left < size && heap[left].priority > heap[largest].priority) {
            largest = left;
        }
        if (right < size && heap[right].priority > heap[largest].priority) {
            largest = right;
        }
        if (largest == index) break;
        
        // Swap
        omni_priority_queue_node_t temp = heap[index];
        heap[index] = heap[largest];
        heap[largest] = temp;
        index = largest;
    }
}

omni_priority_queue_t* omni_priority_queue_create() {
    omni_priority_queue_t* pq = (omni_priority_queue_t*)malloc(sizeof(omni_priority_queue_t));
    if (!pq) return NULL;
    pq->capacity = OMNI_PQ_MAX_SIZE;
    pq->size = 0;
    pq->heap = (omni_priority_queue_node_t*)malloc(pq->capacity * sizeof(omni_priority_queue_node_t));
    if (!pq->heap) {
        free(pq);
        return NULL;
    }
    return pq;
}

void omni_priority_queue_destroy(omni_priority_queue_t* pq) {
    if (!pq) return;
    free(pq->heap);
    free(pq);
}

void omni_priority_queue_insert(omni_priority_queue_t* pq, int32_t element, int32_t priority) {
    if (!pq || pq->size >= pq->capacity) return;
    pq->heap[pq->size].element = element;
    pq->heap[pq->size].priority = priority;
    omni_pq_heapify_up(pq->heap, pq->size);
    pq->size++;
}

int32_t omni_priority_queue_extract_max(omni_priority_queue_t* pq) {
    if (!pq || pq->size == 0) return 0;
    int32_t max_element = pq->heap[0].element;
    pq->heap[0] = pq->heap[pq->size - 1];
    pq->size--;
    if (pq->size > 0) {
        omni_pq_heapify_down(pq->heap, pq->size, 0);
    }
    return max_element;
}

int32_t omni_priority_queue_peek(omni_priority_queue_t* pq) {
    if (!pq || pq->size == 0) return 0;
    return pq->heap[0].element;
}

int32_t omni_priority_queue_is_empty(omni_priority_queue_t* pq) {
    return (!pq || pq->size == 0) ? 1 : 0;
}

int32_t omni_priority_queue_size(omni_priority_queue_t* pq) {
    return pq ? pq->size : 0;
}

// Linked list implementation
typedef struct omni_linked_list_node {
    int32_t value;
    struct omni_linked_list_node* next;
} omni_linked_list_node_t;

struct omni_linked_list {
    omni_linked_list_node_t* head;
    int32_t size;
};

omni_linked_list_t* omni_linked_list_create() {
    omni_linked_list_t* ll = (omni_linked_list_t*)malloc(sizeof(omni_linked_list_t));
    if (!ll) return NULL;
    ll->head = NULL;
    ll->size = 0;
    return ll;
}

void omni_linked_list_destroy(omni_linked_list_t* ll) {
    if (!ll) return;
    omni_linked_list_clear(ll);
    free(ll);
}

void omni_linked_list_append(omni_linked_list_t* ll, int32_t element) {
    if (!ll) return;
    omni_linked_list_node_t* node = (omni_linked_list_node_t*)malloc(sizeof(omni_linked_list_node_t));
    if (!node) return;
    node->value = element;
    node->next = NULL;
    
    if (ll->head == NULL) {
        ll->head = node;
    } else {
        omni_linked_list_node_t* current = ll->head;
        while (current->next != NULL) {
            current = current->next;
        }
        current->next = node;
    }
    ll->size++;
}

void omni_linked_list_prepend(omni_linked_list_t* ll, int32_t element) {
    if (!ll) return;
    omni_linked_list_node_t* node = (omni_linked_list_node_t*)malloc(sizeof(omni_linked_list_node_t));
    if (!node) return;
    node->value = element;
    node->next = ll->head;
    ll->head = node;
    ll->size++;
}

int32_t omni_linked_list_insert(omni_linked_list_t* ll, int32_t index, int32_t element) {
    if (!ll || index < 0 || index > ll->size) return 0;
    if (index == 0) {
        omni_linked_list_prepend(ll, element);
        return 1;
    }
    
    omni_linked_list_node_t* node = (omni_linked_list_node_t*)malloc(sizeof(omni_linked_list_node_t));
    if (!node) return 0;
    node->value = element;
    
    omni_linked_list_node_t* current = ll->head;
    for (int32_t i = 0; i < index - 1; i++) {
        current = current->next;
    }
    node->next = current->next;
    current->next = node;
    ll->size++;
    return 1;
}

int32_t omni_linked_list_remove(omni_linked_list_t* ll, int32_t index) {
    if (!ll || index < 0 || index >= ll->size) return 0;
    
    omni_linked_list_node_t* to_remove;
    if (index == 0) {
        to_remove = ll->head;
        ll->head = ll->head->next;
    } else {
        omni_linked_list_node_t* current = ll->head;
        for (int32_t i = 0; i < index - 1; i++) {
            current = current->next;
        }
        to_remove = current->next;
        current->next = to_remove->next;
    }
    free(to_remove);
    ll->size--;
    return 1;
}

int32_t omni_linked_list_get(omni_linked_list_t* ll, int32_t index) {
    if (!ll || index < 0 || index >= ll->size) return 0;
    omni_linked_list_node_t* current = ll->head;
    for (int32_t i = 0; i < index; i++) {
        current = current->next;
    }
    return current->value;
}

int32_t omni_linked_list_set(omni_linked_list_t* ll, int32_t index, int32_t element) {
    if (!ll || index < 0 || index >= ll->size) return 0;
    omni_linked_list_node_t* current = ll->head;
    for (int32_t i = 0; i < index; i++) {
        current = current->next;
    }
    current->value = element;
    return 1;
}

int32_t omni_linked_list_size(omni_linked_list_t* ll) {
    return ll ? ll->size : 0;
}

int32_t omni_linked_list_is_empty(omni_linked_list_t* ll) {
    return (!ll || ll->head == NULL) ? 1 : 0;
}

void omni_linked_list_clear(omni_linked_list_t* ll) {
    if (!ll) return;
    while (ll->head != NULL) {
        omni_linked_list_remove(ll, 0);
    }
}

// Binary tree implementation (BST)
typedef struct omni_binary_tree_node {
    int32_t value;
    struct omni_binary_tree_node* left;
    struct omni_binary_tree_node* right;
} omni_binary_tree_node_t;

struct omni_binary_tree {
    omni_binary_tree_node_t* root;
    int32_t size;
};

static omni_binary_tree_node_t* omni_bt_node_create(int32_t value) {
    omni_binary_tree_node_t* node = (omni_binary_tree_node_t*)malloc(sizeof(omni_binary_tree_node_t));
    if (!node) return NULL;
    node->value = value;
    node->left = NULL;
    node->right = NULL;
    return node;
}

static void omni_bt_node_destroy(omni_binary_tree_node_t* node) {
    if (!node) return;
    omni_bt_node_destroy(node->left);
    omni_bt_node_destroy(node->right);
    free(node);
}

omni_binary_tree_t* omni_binary_tree_create() {
    omni_binary_tree_t* bt = (omni_binary_tree_t*)malloc(sizeof(omni_binary_tree_t));
    if (!bt) return NULL;
    bt->root = NULL;
    bt->size = 0;
    return bt;
}

void omni_binary_tree_destroy(omni_binary_tree_t* bt) {
    if (!bt) return;
    omni_bt_node_destroy(bt->root);
    free(bt);
}

static omni_binary_tree_node_t* omni_bt_insert_recursive(omni_binary_tree_node_t* node, int32_t value) {
    if (node == NULL) {
        return omni_bt_node_create(value);
    }
    if (value < node->value) {
        node->left = omni_bt_insert_recursive(node->left, value);
    } else if (value > node->value) {
        node->right = omni_bt_insert_recursive(node->right, value);
    }
    return node;
}

void omni_binary_tree_insert(omni_binary_tree_t* bt, int32_t element) {
    if (!bt) return;
    bt->root = omni_bt_insert_recursive(bt->root, element);
    bt->size++;
}

static int32_t omni_bt_search_recursive(omni_binary_tree_node_t* node, int32_t value) {
    if (node == NULL) return 0;
    if (value == node->value) return 1;
    if (value < node->value) {
        return omni_bt_search_recursive(node->left, value);
    }
    return omni_bt_search_recursive(node->right, value);
}

int32_t omni_binary_tree_search(omni_binary_tree_t* bt, int32_t element) {
    if (!bt || !bt->root) return 0;
    return omni_bt_search_recursive(bt->root, element);
}

static omni_binary_tree_node_t* omni_bt_find_min(omni_binary_tree_node_t* node) {
    while (node && node->left != NULL) {
        node = node->left;
    }
    return node;
}

static omni_binary_tree_node_t* omni_bt_remove_recursive(omni_binary_tree_node_t* node, int32_t value) {
    if (node == NULL) return NULL;
    
    if (value < node->value) {
        node->left = omni_bt_remove_recursive(node->left, value);
    } else if (value > node->value) {
        node->right = omni_bt_remove_recursive(node->right, value);
    } else {
        if (node->left == NULL) {
            omni_binary_tree_node_t* temp = node->right;
            free(node);
            return temp;
        } else if (node->right == NULL) {
            omni_binary_tree_node_t* temp = node->left;
            free(node);
            return temp;
        }
        
        omni_binary_tree_node_t* temp = omni_bt_find_min(node->right);
        node->value = temp->value;
        node->right = omni_bt_remove_recursive(node->right, temp->value);
    }
    return node;
}

int32_t omni_binary_tree_remove(omni_binary_tree_t* bt, int32_t element) {
    if (!bt || !bt->root) return 0;
    if (omni_binary_tree_search(bt, element)) {
        bt->root = omni_bt_remove_recursive(bt->root, element);
        bt->size--;
        return 1;
    }
    return 0;
}

int32_t omni_binary_tree_size(omni_binary_tree_t* bt) {
    if (!bt) return 0;
    return bt->size;
}

int32_t omni_binary_tree_is_empty(omni_binary_tree_t* bt) {
    return (!bt || bt->root == NULL) ? 1 : 0;
}

void omni_binary_tree_clear(omni_binary_tree_t* bt) {
    if (!bt) return;
    omni_bt_node_destroy(bt->root);
    bt->root = NULL;
    bt->size = 0;
}

// ============================================================================
// Network Functions Implementation
// ============================================================================

// IP address functions
omni_ip_address_t* omni_ip_parse(const char* ip_str) {
    if (!ip_str) return NULL;
    omni_ip_address_t* ip = (omni_ip_address_t*)malloc(sizeof(omni_ip_address_t));
    if (!ip) return NULL;
    
    strncpy(ip->address, ip_str, sizeof(ip->address) - 1);
    ip->address[sizeof(ip->address) - 1] = '\0';
    
    // Simple IPv4 detection (contains dots and no colons)
    ip->is_ipv4 = (strchr(ip_str, '.') != NULL && strchr(ip_str, ':') == NULL) ? 1 : 0;
    ip->is_ipv6 = (strchr(ip_str, ':') != NULL) ? 1 : 0;
    
    return ip;
}

int32_t omni_ip_is_valid(const char* ip_str) {
    if (!ip_str) return 0;
    // Basic validation - check for IPv4 format (dotted decimal)
    int dot_count = 0;
    int digit_count = 0;
    for (const char* p = ip_str; *p; p++) {
        if (*p == '.') {
            dot_count++;
            digit_count = 0;
        } else if (*p >= '0' && *p <= '9') {
            digit_count++;
            if (digit_count > 3) return 0;
        } else if (*p == ':') {
            // IPv6 format
            return 1; // Basic check - assume valid if contains colon
        } else {
            return 0;
        }
    }
    return (dot_count == 3) ? 1 : 0;
}

int32_t omni_ip_is_private(omni_ip_address_t* ip) {
    if (!ip || !ip->is_ipv4) return 0;
    // Check for private IP ranges: 10.x.x.x, 172.16-31.x.x, 192.168.x.x
    if (strncmp(ip->address, "10.", 3) == 0) return 1;
    if (strncmp(ip->address, "172.16.", 7) == 0 || strncmp(ip->address, "172.17.", 7) == 0 ||
        strncmp(ip->address, "172.18.", 7) == 0 || strncmp(ip->address, "172.19.", 7) == 0 ||
        strncmp(ip->address, "172.20.", 7) == 0 || strncmp(ip->address, "172.21.", 7) == 0 ||
        strncmp(ip->address, "172.22.", 7) == 0 || strncmp(ip->address, "172.23.", 7) == 0 ||
        strncmp(ip->address, "172.24.", 7) == 0 || strncmp(ip->address, "172.25.", 7) == 0 ||
        strncmp(ip->address, "172.26.", 7) == 0 || strncmp(ip->address, "172.27.", 7) == 0 ||
        strncmp(ip->address, "172.28.", 7) == 0 || strncmp(ip->address, "172.29.", 7) == 0 ||
        strncmp(ip->address, "172.30.", 7) == 0 || strncmp(ip->address, "172.31.", 7) == 0) return 1;
    if (strncmp(ip->address, "192.168.", 8) == 0) return 1;
    return 0;
}

int32_t omni_ip_is_loopback(omni_ip_address_t* ip) {
    if (!ip || !ip->is_ipv4) return 0;
    return (strncmp(ip->address, "127.", 4) == 0) ? 1 : 0;
}

char* omni_ip_to_string(omni_ip_address_t* ip) {
    if (!ip) return NULL;
    return strdup(ip->address);
}

// URL functions
omni_url_t* omni_url_parse(const char* url_str) {
    if (!url_str) return NULL;
    omni_url_t* url = (omni_url_t*)malloc(sizeof(omni_url_t));
    if (!url) return NULL;
    
    // Initialize with defaults
    strncpy(url->scheme, "http", sizeof(url->scheme) - 1);
    strncpy(url->host, "localhost", sizeof(url->host) - 1);
    url->port = 80;
    strncpy(url->path, "/", sizeof(url->path) - 1);
    url->query[0] = '\0';
    url->fragment[0] = '\0';
    
    // Basic parsing
    const char* scheme_end = strstr(url_str, "://");
    if (scheme_end) {
        size_t scheme_len = scheme_end - url_str;
        if (scheme_len < sizeof(url->scheme)) {
            strncpy(url->scheme, url_str, scheme_len);
            url->scheme[scheme_len] = '\0';
        }
        url_str = scheme_end + 3;
    }
    
    // Extract host and port
    const char* path_start = strchr(url_str, '/');
    const char* port_start = strchr(url_str, ':');
    if (port_start && (!path_start || port_start < path_start)) {
        size_t host_len = port_start - url_str;
        if (host_len < sizeof(url->host)) {
            strncpy(url->host, url_str, host_len);
            url->host[host_len] = '\0';
        }
        sscanf(port_start + 1, "%d", &url->port);
        if (path_start) {
            strncpy(url->path, path_start, sizeof(url->path) - 1);
        }
    } else if (path_start) {
        size_t host_len = path_start - url_str;
        if (host_len < sizeof(url->host)) {
            strncpy(url->host, url_str, host_len);
            url->host[host_len] = '\0';
        }
        strncpy(url->path, path_start, sizeof(url->path) - 1);
    } else {
        strncpy(url->host, url_str, sizeof(url->host) - 1);
    }
    
    return url;
}

char* omni_url_to_string(omni_url_t* url) {
    if (!url) return NULL;
    char* result = (char*)malloc(1024);
    if (!result) return NULL;
    
    snprintf(result, 1024, "%s://%s", url->scheme, url->host);
    if (url->port > 0 && url->port != 80 && url->port != 443) {
        char temp[1024];
        snprintf(temp, 1024, "%s:%d%s", result, url->port, url->path);
        strncpy(result, temp, 1024);
    } else {
        strncat(result, url->path, 1024 - strlen(result) - 1);
    }
    if (url->query[0] != '\0') {
        strncat(result, "?", 1024 - strlen(result) - 1);
        strncat(result, url->query, 1024 - strlen(result) - 1);
    }
    if (url->fragment[0] != '\0') {
        strncat(result, "#", 1024 - strlen(result) - 1);
        strncat(result, url->fragment, 1024 - strlen(result) - 1);
    }
    return result;
}

int32_t omni_url_is_valid(const char* url_str) {
    if (!url_str) return 0;
    // Basic validation - check for scheme://host format
    return (strstr(url_str, "://") != NULL) ? 1 : 0;
}

// DNS functions (stub implementations - would need actual DNS library)
omni_ip_address_t** omni_dns_lookup(const char* hostname, int32_t* count) {
    if (!hostname || !count) return NULL;
    *count = 0;
    // Stub: would need getaddrinfo or similar
    return NULL;
}

char* omni_dns_reverse_lookup(omni_ip_address_t* ip) {
    if (!ip) return NULL;
    // Stub: would need getnameinfo or similar
    return strdup("");
}

// HTTP client functions (stub implementations - would need HTTP library like libcurl)
omni_http_response_t* omni_http_get(const char* url) {
    if (!url) return NULL;
    omni_http_response_t* resp = (omni_http_response_t*)malloc(sizeof(omni_http_response_t));
    if (!resp) return NULL;
    resp->status_code = 200;
    strncpy(resp->status_text, "OK", sizeof(resp->status_text) - 1);
    resp->headers = omni_map_create();
    resp->body = strdup("");
    return resp;
}

omni_http_response_t* omni_http_post(const char* url, const char* body) {
    (void)body; // Unused in stub implementation
    return omni_http_get(url); // Stub
}

omni_http_response_t* omni_http_put(const char* url, const char* body) {
    (void)body; // Unused in stub implementation
    return omni_http_get(url); // Stub
}

omni_http_response_t* omni_http_delete(const char* url) {
    return omni_http_get(url); // Stub
}

omni_http_response_t* omni_http_request(omni_http_request_t* req) {
    if (!req) return NULL;
    return omni_http_get(req->url);
}

void omni_http_response_destroy(omni_http_response_t* resp) {
    if (!resp) return;
    if (resp->headers) omni_map_destroy(resp->headers);
    if (resp->body) free(resp->body);
    free(resp);
}

int32_t omni_http_response_is_success(omni_http_response_t* resp) {
    if (!resp) return 0;
    return (resp->status_code >= 200 && resp->status_code < 300) ? 1 : 0;
}

int32_t omni_http_response_is_client_error(omni_http_response_t* resp) {
    if (!resp) return 0;
    return (resp->status_code >= 400 && resp->status_code < 500) ? 1 : 0;
}

int32_t omni_http_response_is_server_error(omni_http_response_t* resp) {
    if (!resp) return 0;
    return (resp->status_code >= 500 && resp->status_code < 600) ? 1 : 0;
}

char* omni_http_response_get_header(omni_http_response_t* resp, const char* name) {
    if (!resp || !resp->headers || !name) return NULL;
    const char* value = omni_map_get_string_string(resp->headers, name);
    return value ? strdup(value) : NULL;
}

omni_http_request_t* omni_http_request_create(const char* method, const char* url) {
    if (!method || !url) return NULL;
    omni_http_request_t* req = (omni_http_request_t*)malloc(sizeof(omni_http_request_t));
    if (!req) return NULL;
    strncpy(req->method, method, sizeof(req->method) - 1);
    strncpy(req->url, url, sizeof(req->url) - 1);
    req->headers = omni_map_create();
    req->body = NULL;
    return req;
}

void omni_http_request_set_header(omni_http_request_t* req, const char* name, const char* value) {
    if (!req || !req->headers || !name || !value) return;
    omni_map_put_string_string(req->headers, name, value);
}

void omni_http_request_set_body(omni_http_request_t* req, const char* body) {
    if (!req) return;
    if (req->body) free(req->body);
    req->body = body ? strdup(body) : NULL;
}

char* omni_http_request_get_header(omni_http_request_t* req, const char* name) {
    if (!req || !req->headers || !name) return NULL;
    const char* value = omni_map_get_string_string(req->headers, name);
    return value ? strdup(value) : NULL;
}

void omni_http_request_destroy(omni_http_request_t* req) {
    if (!req) return;
    if (req->headers) omni_map_destroy(req->headers);
    if (req->body) free(req->body);
    free(req);
}

// Socket functions (stub implementations - would need socket library)
#ifdef _WIN32
#include <winsock2.h>
#include <ws2tcpip.h>
#else
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <netdb.h>
#include <unistd.h>
#endif

int32_t omni_socket_create() {
#ifdef _WIN32
    WSADATA wsa;
    if (WSAStartup(MAKEWORD(2, 2), &wsa) != 0) return -1;
    return socket(AF_INET, SOCK_STREAM, 0);
#else
    return socket(AF_INET, SOCK_STREAM, 0);
#endif
}

int32_t omni_socket_connect(int32_t socket, const char* address, int32_t port) {
    if (socket < 0 || !address) return 0;
    struct sockaddr_in addr;
    memset(&addr, 0, sizeof(addr));
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    inet_pton(AF_INET, address, &addr.sin_addr);
    return (connect(socket, (struct sockaddr*)&addr, sizeof(addr)) == 0) ? 1 : 0;
}

int32_t omni_socket_bind(int32_t socket, const char* address, int32_t port) {
    if (socket < 0 || !address) return 0;
    struct sockaddr_in addr;
    memset(&addr, 0, sizeof(addr));
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    inet_pton(AF_INET, address, &addr.sin_addr);
    return (bind(socket, (struct sockaddr*)&addr, sizeof(addr)) == 0) ? 1 : 0;
}

int32_t omni_socket_listen(int32_t socket, int32_t backlog) {
    if (socket < 0) return 0;
    return (listen(socket, backlog) == 0) ? 1 : 0;
}

int32_t omni_socket_accept(int32_t socket) {
    if (socket < 0) return -1;
    struct sockaddr_in addr;
    socklen_t len = sizeof(addr);
    return accept(socket, (struct sockaddr*)&addr, &len);
}

int32_t omni_socket_send(int32_t socket, const char* data) {
    if (socket < 0 || !data) return -1;
    return send(socket, data, strlen(data), 0);
}

int32_t omni_socket_receive(int32_t socket, char* buffer, int32_t buffer_size) {
    if (socket < 0 || !buffer || buffer_size <= 0) return -1;
    ssize_t received = recv(socket, buffer, (size_t)(buffer_size - 1), 0);
    if (received > 0) {
        buffer[received] = '\0';
    }
    return (int32_t)received;
}

int32_t omni_socket_close(int32_t socket) {
    if (socket < 0) return 0;
#ifdef _WIN32
    return (closesocket(socket) == 0) ? 1 : 0;
#else
    return (close(socket) == 0) ? 1 : 0;
#endif
}

// Network utility functions
int32_t omni_network_is_connected() {
    // Stub: would need to check network interface status
    return 1; // Assume connected
}

omni_ip_address_t* omni_network_get_local_ip() {
    omni_ip_address_t* ip = (omni_ip_address_t*)malloc(sizeof(omni_ip_address_t));
    if (!ip) return NULL;
    strncpy(ip->address, "127.0.0.1", sizeof(ip->address) - 1);
    ip->is_ipv4 = 1;
    ip->is_ipv6 = 0;
    return ip;
}

int32_t omni_network_ping(const char* host) {
    if (!host) return 0;
    // Stub: would need ICMP ping implementation
    return 0;
}

// ============================================================================
// Coverage Tracking Infrastructure
// ============================================================================

#define OMNI_COVERAGE_MAX_ENTRIES 10000
#define OMNI_COVERAGE_MAX_FUNCTION_NAME 256
#define OMNI_COVERAGE_MAX_FILE_PATH 512

typedef struct {
    char function_name[OMNI_COVERAGE_MAX_FUNCTION_NAME];
    char file_path[OMNI_COVERAGE_MAX_FILE_PATH];
    int32_t line_number;
    int32_t call_count;
} omni_coverage_entry_t;

static struct {
    omni_coverage_entry_t entries[OMNI_COVERAGE_MAX_ENTRIES];
    int32_t count;
    int32_t enabled;
#ifdef _WIN32
    CRITICAL_SECTION mutex;
    int mutex_initialized;
#else
    pthread_mutex_t mutex;
#endif
} omni_coverage_state = {
    .count = 0,
    .enabled = 0,
#ifdef _WIN32
    .mutex_initialized = 0
#else
    .mutex = PTHREAD_MUTEX_INITIALIZER
#endif
};

void omni_coverage_init(void) {
#ifdef _WIN32
    if (!omni_coverage_state.mutex_initialized) {
        InitializeCriticalSection(&omni_coverage_state.mutex);
        omni_coverage_state.mutex_initialized = 1;
    }
    EnterCriticalSection(&omni_coverage_state.mutex);
#else
    pthread_mutex_lock(&omni_coverage_state.mutex);
#endif
    
    omni_coverage_state.count = 0;
    omni_coverage_state.enabled = 1;
    
#ifdef _WIN32
    LeaveCriticalSection(&omni_coverage_state.mutex);
#else
    pthread_mutex_unlock(&omni_coverage_state.mutex);
#endif
}

void omni_coverage_set_enabled(int32_t enabled) {
#ifdef _WIN32
    if (!omni_coverage_state.mutex_initialized) {
        InitializeCriticalSection(&omni_coverage_state.mutex);
        omni_coverage_state.mutex_initialized = 1;
    }
    EnterCriticalSection(&omni_coverage_state.mutex);
#else
    pthread_mutex_lock(&omni_coverage_state.mutex);
#endif
    
    omni_coverage_state.enabled = enabled;
    
#ifdef _WIN32
    LeaveCriticalSection(&omni_coverage_state.mutex);
#else
    pthread_mutex_unlock(&omni_coverage_state.mutex);
#endif
}

int32_t omni_coverage_is_enabled(void) {
    return omni_coverage_state.enabled;
}

void omni_coverage_record(const char* function_name, const char* file_path, int32_t line_number) {
    if (!omni_coverage_state.enabled || !function_name) {
        return;
    }
    
#ifdef _WIN32
    if (!omni_coverage_state.mutex_initialized) {
        InitializeCriticalSection(&omni_coverage_state.mutex);
        omni_coverage_state.mutex_initialized = 1;
    }
    EnterCriticalSection(&omni_coverage_state.mutex);
#else
    pthread_mutex_lock(&omni_coverage_state.mutex);
#endif
    
    // Check if entry already exists
    int32_t found = 0;
    for (int32_t i = 0; i < omni_coverage_state.count; i++) {
        if (strcmp(omni_coverage_state.entries[i].function_name, function_name) == 0 &&
            (file_path == NULL || strcmp(omni_coverage_state.entries[i].file_path, file_path) == 0) &&
            omni_coverage_state.entries[i].line_number == line_number) {
            omni_coverage_state.entries[i].call_count++;
            found = 1;
            break;
        }
    }
    
    // Add new entry if not found and we have space
    if (!found && omni_coverage_state.count < OMNI_COVERAGE_MAX_ENTRIES) {
        omni_coverage_entry_t* entry = &omni_coverage_state.entries[omni_coverage_state.count];
        strncpy(entry->function_name, function_name, OMNI_COVERAGE_MAX_FUNCTION_NAME - 1);
        entry->function_name[OMNI_COVERAGE_MAX_FUNCTION_NAME - 1] = '\0';
        
        if (file_path) {
            strncpy(entry->file_path, file_path, OMNI_COVERAGE_MAX_FILE_PATH - 1);
            entry->file_path[OMNI_COVERAGE_MAX_FILE_PATH - 1] = '\0';
        } else {
            entry->file_path[0] = '\0';
        }
        
        entry->line_number = line_number;
        entry->call_count = 1;
        omni_coverage_state.count++;
    }
    
#ifdef _WIN32
    LeaveCriticalSection(&omni_coverage_state.mutex);
#else
    pthread_mutex_unlock(&omni_coverage_state.mutex);
#endif
}

void omni_coverage_reset(void) {
#ifdef _WIN32
    if (!omni_coverage_state.mutex_initialized) {
        InitializeCriticalSection(&omni_coverage_state.mutex);
        omni_coverage_state.mutex_initialized = 1;
    }
    EnterCriticalSection(&omni_coverage_state.mutex);
#else
    pthread_mutex_lock(&omni_coverage_state.mutex);
#endif
    
    omni_coverage_state.count = 0;
    
#ifdef _WIN32
    LeaveCriticalSection(&omni_coverage_state.mutex);
#else
    pthread_mutex_unlock(&omni_coverage_state.mutex);
#endif
}

char* omni_coverage_export(void) {
    // Estimate buffer size: each entry needs ~200 bytes for JSON
    size_t buffer_size = 1024 + (omni_coverage_state.count * 300);
    char* json = malloc(buffer_size);
    if (!json) {
        return NULL;
    }
    
#ifdef _WIN32
    if (!omni_coverage_state.mutex_initialized) {
        InitializeCriticalSection(&omni_coverage_state.mutex);
        omni_coverage_state.mutex_initialized = 1;
    }
    EnterCriticalSection(&omni_coverage_state.mutex);
#else
    pthread_mutex_lock(&omni_coverage_state.mutex);
#endif
    
    size_t pos = 0;
    pos += snprintf(json + pos, buffer_size - pos, "{\"entries\":[");
    
    for (int32_t i = 0; i < omni_coverage_state.count; i++) {
        if (i > 0) {
            pos += snprintf(json + pos, buffer_size - pos, ",");
        }
        omni_coverage_entry_t* entry = &omni_coverage_state.entries[i];
        pos += snprintf(json + pos, buffer_size - pos,
            "{\"function\":\"%s\",\"file\":\"%s\",\"line\":%d,\"count\":%d}",
            entry->function_name,
            entry->file_path,
            entry->line_number,
            entry->call_count);
        
        if (pos >= buffer_size - 100) {
            // Resize buffer if needed
            buffer_size *= 2;
            char* new_json = realloc(json, buffer_size);
            if (!new_json) {
                free(json);
#ifdef _WIN32
                LeaveCriticalSection(&omni_coverage_state.mutex);
#else
                pthread_mutex_unlock(&omni_coverage_state.mutex);
#endif
                return NULL;
            }
            json = new_json;
        }
    }
    
    pos += snprintf(json + pos, buffer_size - pos, "]}");
    
#ifdef _WIN32
    LeaveCriticalSection(&omni_coverage_state.mutex);
#else
    pthread_mutex_unlock(&omni_coverage_state.mutex);
#endif
    
    return json;
}
