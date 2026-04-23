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
// omni_len returns the length of an array. If the pointer was produced by
// omni_slice_make (i.e. has the slice header magic immediately before it),
// we return the runtime-tracked length. Otherwise we fall back to the
// compile-time hint that the backend passed in. This lets old fixed-length
// codegen and new heap-allocated slice codegen coexist.
int32_t omni_len(void* array, size_t element_size, int32_t array_length) {
    (void)element_size;
    if (array != NULL) {
        // omni_slice_header_t layout (defined later in this file):
        //     { int64_t len; int64_t cap; int64_t elem_size; int64_t magic; }
        // so the magic word sits at offset -1 from the data pointer in
        // int64 units. Check it inline rather than calling the helper to
        // avoid a forward-declaration ordering dance with the slice section.
        int64_t* magic_slot = (int64_t*)array - 1;
        if (*magic_slot == (int64_t)0x4F4D4E494C535443LL) {
            int64_t* len_slot = (int64_t*)array - 4;
            return (int32_t)(*len_slot);
        }
    }
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

// Wrapper structure for any type values
typedef struct omni_any_value {
    void* value;
    int32_t type;
} omni_any_value_t;

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

// Generic dynamic array implementation
omni_array_t* omni_array_create() {
    omni_array_t* arr = (omni_array_t*)malloc(sizeof(omni_array_t));
    if (!arr) return NULL;
    
    arr->capacity = 8; // Start with capacity of 8
    arr->count = 0;
    arr->items = (void**)malloc(arr->capacity * sizeof(void*));
    if (!arr->items) {
        free(arr);
        return NULL;
    }
    
    return arr;
}

void omni_array_destroy(omni_array_t* arr) {
    if (!arr) return;
    
    // Free all items (caller is responsible for freeing item contents)
    if (arr->items) {
        free(arr->items);
    }
    free(arr);
}

void omni_array_append(omni_array_t* arr, void* item) {
    if (!arr || !item) return;
    
    // Resize if needed
    if (arr->count >= arr->capacity) {
        int32_t new_capacity = arr->capacity * 2;
        void** new_items = (void**)realloc(arr->items, new_capacity * sizeof(void*));
        if (!new_items) return; // Out of memory
        arr->items = new_items;
        arr->capacity = new_capacity;
    }
    
    arr->items[arr->count] = item;
    arr->count++;
}

void* omni_array_get(omni_array_t* arr, int32_t index) {
    if (!arr || index < 0 || index >= arr->count) return NULL;
    return arr->items[index];
}

int32_t omni_array_size(omni_array_t* arr) {
    if (!arr) return 0;
    return arr->count;
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

// Map put operations for any type support
void omni_map_put_string_any(omni_map_t* map, const char* key, void* value, int32_t value_type) {
    if (!map) return;
    
    uint32_t hash = hash_string(key);
    int32_t bucket = hash % map->bucket_count;
    
    // Check if key already exists
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (strcmp((char*)entry->key, key) == 0) {
            // Update existing value - free old any value wrapper
            if (entry->value) {
                omni_any_value_t* old_any = (omni_any_value_t*)entry->value;
                // Free the wrapped value based on type
                if (old_any->type == OMNI_TYPE_STRING) {
                    free((char*)old_any->value);
                } else if (old_any->type == OMNI_TYPE_MAP) {
                    omni_map_destroy((omni_map_t*)old_any->value);
                } else if (old_any->type == OMNI_TYPE_ARRAY) {
                    omni_array_destroy((omni_array_t*)old_any->value);
                } else if (old_any->type == OMNI_TYPE_STRUCT) {
                    // Struct cleanup handled by runtime
                }
                free(old_any);
            }
            // Create new any value wrapper
            omni_any_value_t* any_val = (omni_any_value_t*)malloc(sizeof(omni_any_value_t));
            if (any_val) {
                any_val->value = value;
                any_val->type = value_type;
                entry->value = any_val;
            }
            return;
        }
        entry = entry->next;
    }
    
    // Rehash if load factor exceeds 0.75
    if (map->size * 4 >= map->bucket_count * 3) {
        omni_map_rehash(map, 1); // 1 = string key
        bucket = hash % map->bucket_count;
    }
    
    // Create new entry
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    entry->key = malloc(strlen(key) + 1);
    omni_any_value_t* any_val = (omni_any_value_t*)malloc(sizeof(omni_any_value_t));
    if (!entry->key || !any_val) {
        free(entry->key);
        free(any_val);
        free(entry);
        return;
    }
    
    strcpy((char*)entry->key, key);
    any_val->value = value;
    any_val->type = value_type;
    entry->value = any_val;
    entry->next = map->buckets[bucket];
    map->buckets[bucket] = entry;
    map->size++;
}

void omni_map_put_int_any(omni_map_t* map, int32_t key, void* value, int32_t value_type) {
    if (!map) return;
    
    uint32_t hash = hash_int(key);
    int32_t bucket = hash % map->bucket_count;
    
    // Check if key already exists
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        if (*(int32_t*)entry->key == key) {
            // Update existing value - free old any value wrapper
            if (entry->value) {
                omni_any_value_t* old_any = (omni_any_value_t*)entry->value;
                if (old_any->type == OMNI_TYPE_STRING) {
                    free((char*)old_any->value);
                } else if (old_any->type == OMNI_TYPE_MAP) {
                    omni_map_destroy((omni_map_t*)old_any->value);
                } else if (old_any->type == OMNI_TYPE_ARRAY) {
                    omni_array_destroy((omni_array_t*)old_any->value);
                }
                free(old_any);
            }
            // Create new any value wrapper
            omni_any_value_t* any_val = (omni_any_value_t*)malloc(sizeof(omni_any_value_t));
            if (any_val) {
                any_val->value = value;
                any_val->type = value_type;
                entry->value = any_val;
            }
            return;
        }
        entry = entry->next;
    }
    
    // Rehash if load factor exceeds 0.75
    if (map->size * 4 >= map->bucket_count * 3) {
        omni_map_rehash(map, 0); // 0 = int key
        bucket = hash % map->bucket_count;
    }
    
    // Create new entry
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    entry->key = malloc(sizeof(int32_t));
    omni_any_value_t* any_val = (omni_any_value_t*)malloc(sizeof(omni_any_value_t));
    if (!entry->key || !any_val) {
        free(entry->key);
        free(any_val);
        free(entry);
        return;
    }
    
    *(int32_t*)entry->key = key;
    any_val->value = value;
    any_val->type = value_type;
    entry->value = any_val;
    entry->next = map->buckets[bucket];
    map->buckets[bucket] = entry;
    map->size++;
}

// Helper function to hash any key type
static uint32_t hash_any_key(void* key, int32_t key_type) {
    if (key_type == OMNI_TYPE_INT) {
        return hash_int(*(int32_t*)key);
    } else if (key_type == OMNI_TYPE_STRING) {
        return hash_string((char*)key);
    }
    // For other types, use pointer hash
    return (uint32_t)(uintptr_t)key;
}

// Helper function to compare any key types
static int32_t compare_any_key(void* key1, int32_t type1, void* key2, int32_t type2) {
    if (type1 != type2) return 0;
    if (type1 == OMNI_TYPE_INT) {
        return *(int32_t*)key1 == *(int32_t*)key2;
    } else if (type1 == OMNI_TYPE_STRING) {
        return strcmp((char*)key1, (char*)key2) == 0;
    }
    return key1 == key2;
}

void omni_map_put_any_string(omni_map_t* map, void* key, int32_t key_type, const char* value) {
    if (!map) return;
    
    uint32_t hash = hash_any_key(key, key_type);
    int32_t bucket = hash % map->bucket_count;
    
    // Check if key already exists
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        omni_any_value_t* key_any = (omni_any_value_t*)entry->key;
        if (key_any && compare_any_key(key, key_type, key_any->value, key_any->type)) {
            // Update existing value
            free(entry->value);
            entry->value = malloc(strlen(value) + 1);
            if (entry->value) {
                strcpy((char*)entry->value, value);
            }
            return;
        }
        entry = entry->next;
    }
    
    // Rehash if needed (would need to track key type for rehashing)
    if (map->size * 4 >= map->bucket_count * 3) {
        // For any keys, we can't easily rehash, so skip for now
        // In a full implementation, we'd need to store key type info in the map
    }
    
    // Create new entry
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    // Store key as any value wrapper
    omni_any_value_t* key_any = (omni_any_value_t*)malloc(sizeof(omni_any_value_t));
    entry->value = malloc(strlen(value) + 1);
    if (!key_any || !entry->value) {
        free(key_any);
        free(entry->value);
        free(entry);
        return;
    }
    
    // Copy key based on type
    if (key_type == OMNI_TYPE_INT) {
        key_any->value = malloc(sizeof(int32_t));
        if (key_any->value) {
            *(int32_t*)key_any->value = *(int32_t*)key;
        }
    } else if (key_type == OMNI_TYPE_STRING) {
        key_any->value = malloc(strlen((char*)key) + 1);
        if (key_any->value) {
            strcpy((char*)key_any->value, (char*)key);
        }
    } else {
        key_any->value = key; // For other types, just store pointer
    }
    key_any->type = key_type;
    entry->key = key_any;
    
    strcpy((char*)entry->value, value);
    entry->next = map->buckets[bucket];
    map->buckets[bucket] = entry;
    map->size++;
}

void omni_map_put_any_int(omni_map_t* map, void* key, int32_t key_type, int32_t value) {
    if (!map) return;
    
    uint32_t hash = hash_any_key(key, key_type);
    int32_t bucket = hash % map->bucket_count;
    
    // Check if key already exists
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        omni_any_value_t* key_any = (omni_any_value_t*)entry->key;
        if (key_any && compare_any_key(key, key_type, key_any->value, key_any->type)) {
            *(int32_t*)entry->value = value;
            return;
        }
        entry = entry->next;
    }
    
    // Create new entry
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    omni_any_value_t* key_any = (omni_any_value_t*)malloc(sizeof(omni_any_value_t));
    entry->value = malloc(sizeof(int32_t));
    if (!key_any || !entry->value) {
        free(key_any);
        free(entry->value);
        free(entry);
        return;
    }
    
    if (key_type == OMNI_TYPE_INT) {
        key_any->value = malloc(sizeof(int32_t));
        if (key_any->value) {
            *(int32_t*)key_any->value = *(int32_t*)key;
        }
    } else if (key_type == OMNI_TYPE_STRING) {
        key_any->value = malloc(strlen((char*)key) + 1);
        if (key_any->value) {
            strcpy((char*)key_any->value, (char*)key);
        }
    } else {
        key_any->value = key;
    }
    key_any->type = key_type;
    entry->key = key_any;
    
    *(int32_t*)entry->value = value;
    entry->next = map->buckets[bucket];
    map->buckets[bucket] = entry;
    map->size++;
}

void omni_map_put_any_float(omni_map_t* map, void* key, int32_t key_type, double value) {
    if (!map) return;
    
    uint32_t hash = hash_any_key(key, key_type);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        omni_any_value_t* key_any = (omni_any_value_t*)entry->key;
        if (key_any && compare_any_key(key, key_type, key_any->value, key_any->type)) {
            *(double*)entry->value = value;
            return;
        }
        entry = entry->next;
    }
    
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    omni_any_value_t* key_any = (omni_any_value_t*)malloc(sizeof(omni_any_value_t));
    entry->value = malloc(sizeof(double));
    if (!key_any || !entry->value) {
        free(key_any);
        free(entry->value);
        free(entry);
        return;
    }
    
    if (key_type == OMNI_TYPE_INT) {
        key_any->value = malloc(sizeof(int32_t));
        if (key_any->value) {
            *(int32_t*)key_any->value = *(int32_t*)key;
        }
    } else if (key_type == OMNI_TYPE_STRING) {
        key_any->value = malloc(strlen((char*)key) + 1);
        if (key_any->value) {
            strcpy((char*)key_any->value, (char*)key);
        }
    } else {
        key_any->value = key;
    }
    key_any->type = key_type;
    entry->key = key_any;
    
    *(double*)entry->value = value;
    entry->next = map->buckets[bucket];
    map->buckets[bucket] = entry;
    map->size++;
}

void omni_map_put_any_bool(omni_map_t* map, void* key, int32_t key_type, int32_t value) {
    if (!map) return;
    
    uint32_t hash = hash_any_key(key, key_type);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        omni_any_value_t* key_any = (omni_any_value_t*)entry->key;
        if (key_any && compare_any_key(key, key_type, key_any->value, key_any->type)) {
            *(int32_t*)entry->value = value;
            return;
        }
        entry = entry->next;
    }
    
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    omni_any_value_t* key_any = (omni_any_value_t*)malloc(sizeof(omni_any_value_t));
    entry->value = malloc(sizeof(int32_t));
    if (!key_any || !entry->value) {
        free(key_any);
        free(entry->value);
        free(entry);
        return;
    }
    
    if (key_type == OMNI_TYPE_INT) {
        key_any->value = malloc(sizeof(int32_t));
        if (key_any->value) {
            *(int32_t*)key_any->value = *(int32_t*)key;
        }
    } else if (key_type == OMNI_TYPE_STRING) {
        key_any->value = malloc(strlen((char*)key) + 1);
        if (key_any->value) {
            strcpy((char*)key_any->value, (char*)key);
        }
    } else {
        key_any->value = key;
    }
    key_any->type = key_type;
    entry->key = key_any;
    
    *(int32_t*)entry->value = value;
    entry->next = map->buckets[bucket];
    map->buckets[bucket] = entry;
    map->size++;
}

void omni_map_put_any_any(omni_map_t* map, void* key, int32_t key_type, void* value, int32_t value_type) {
    if (!map) return;
    
    uint32_t hash = hash_any_key(key, key_type);
    int32_t bucket = hash % map->bucket_count;
    
    omni_map_entry_t* entry = map->buckets[bucket];
    while (entry) {
        omni_any_value_t* key_any = (omni_any_value_t*)entry->key;
        if (key_any && compare_any_key(key, key_type, key_any->value, key_any->type)) {
            // Update existing value - free old any value wrapper
            if (entry->value) {
                omni_any_value_t* old_any = (omni_any_value_t*)entry->value;
                if (old_any->type == OMNI_TYPE_STRING) {
                    free((char*)old_any->value);
                } else if (old_any->type == OMNI_TYPE_MAP) {
                    omni_map_destroy((omni_map_t*)old_any->value);
                } else if (old_any->type == OMNI_TYPE_ARRAY) {
                    omni_array_destroy((omni_array_t*)old_any->value);
                }
                free(old_any);
            }
            // Create new any value wrapper
            omni_any_value_t* any_val = (omni_any_value_t*)malloc(sizeof(omni_any_value_t));
            if (any_val) {
                any_val->value = value;
                any_val->type = value_type;
                entry->value = any_val;
            }
            return;
        }
        entry = entry->next;
    }
    
    entry = (omni_map_entry_t*)malloc(sizeof(omni_map_entry_t));
    if (!entry) return;
    
    omni_any_value_t* key_any = (omni_any_value_t*)malloc(sizeof(omni_any_value_t));
    omni_any_value_t* val_any = (omni_any_value_t*)malloc(sizeof(omni_any_value_t));
    if (!key_any || !val_any) {
        free(key_any);
        free(val_any);
        free(entry);
        return;
    }
    
    // Copy key based on type
    if (key_type == OMNI_TYPE_INT) {
        key_any->value = malloc(sizeof(int32_t));
        if (key_any->value) {
            *(int32_t*)key_any->value = *(int32_t*)key;
        }
    } else if (key_type == OMNI_TYPE_STRING) {
        key_any->value = malloc(strlen((char*)key) + 1);
        if (key_any->value) {
            strcpy((char*)key_any->value, (char*)key);
        }
    } else {
        key_any->value = key;
    }
    key_any->type = key_type;
    entry->key = key_any;
    
    val_any->value = value;
    val_any->type = value_type;
    entry->value = val_any;
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
    char* type_name; // OmniLang type tag; used for interface method dispatch.
};

omni_struct_t* omni_struct_create() {
    omni_struct_t* struct_ptr = (omni_struct_t*)malloc(sizeof(omni_struct_t));
    if (!struct_ptr) return NULL;

    struct_ptr->fields = NULL;
    struct_ptr->type_name = NULL;
    return struct_ptr;
}

void omni_struct_set_type_name(omni_struct_t* struct_ptr, const char* type_name) {
    if (!struct_ptr) return;
    if (struct_ptr->type_name) {
        free(struct_ptr->type_name);
        struct_ptr->type_name = NULL;
    }
    if (type_name) {
        struct_ptr->type_name = (char*)malloc(strlen(type_name) + 1);
        if (struct_ptr->type_name) {
            strcpy(struct_ptr->type_name, type_name);
        }
    }
}

const char* omni_struct_get_type_name(omni_struct_t* struct_ptr) {
    if (!struct_ptr || !struct_ptr->type_name) return "";
    return struct_ptr->type_name;
}

void omni_struct_destroy(omni_struct_t* struct_ptr) {
    if (!struct_ptr) return;

    if (struct_ptr->type_name) {
        free(struct_ptr->type_name);
        struct_ptr->type_name = NULL;
    }
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        omni_struct_field_t* next = field->next;
        free(field->name);
        if (field->value_type == 0) { // string
            free((char*)field->value);
        } else if (field->value_type == OMNI_TYPE_MAP && field->value) {
            omni_map_destroy((omni_map_t*)field->value);
        } else if (field->value_type == OMNI_TYPE_STRUCT && field->value) {
            omni_struct_destroy((omni_struct_t*)field->value);
        } else if (field->value) {
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

void omni_struct_set_array_field(omni_struct_t* struct_ptr, const char* field_name, void* array_value, int32_t element_type, int32_t array_length) {
    (void)element_type; // Unused for now
    (void)array_length;  // Unused for now
    if (!struct_ptr) return;
    
    // Check if field already exists
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        if (strcmp(field->name, field_name) == 0) {
            // Update existing field - free old value
            if (field->value_type == 0) { // string
                free((char*)field->value);
            } else {
                free(field->value);
            }
            // Store pointer to array (caller owns the array)
            field->value = array_value;
            field->value_type = OMNI_TYPE_ARRAY;
            return;
        }
        field = field->next;
    }
    
    // Create new field
    field = (omni_struct_field_t*)malloc(sizeof(omni_struct_field_t));
    if (!field) return;
    
    field->name = malloc(strlen(field_name) + 1);
    if (!field->name) {
        free(field);
        return;
    }
    
    strcpy(field->name, field_name);
    field->value = array_value; // Store pointer to array (caller owns the array)
    field->value_type = OMNI_TYPE_ARRAY;
    field->next = struct_ptr->fields;
    struct_ptr->fields = field;
}

void omni_struct_set_map_field(omni_struct_t* struct_ptr, const char* field_name, omni_map_t* map_value) {
    if (!struct_ptr) return;
    
    // Check if field already exists
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        if (strcmp(field->name, field_name) == 0) {
            // Update existing field - free old value if it was a map
            if (field->value_type == OMNI_TYPE_MAP && field->value) {
                omni_map_destroy((omni_map_t*)field->value);
            } else if (field->value_type == 0) { // string
                free((char*)field->value);
            } else if (field->value) {
                free(field->value);
            }
            // Store pointer to map (caller owns the map)
            field->value = map_value;
            field->value_type = OMNI_TYPE_MAP;
            return;
        }
        field = field->next;
    }
    
    // Create new field
    field = (omni_struct_field_t*)malloc(sizeof(omni_struct_field_t));
    if (!field) return;
    
    field->name = malloc(strlen(field_name) + 1);
    if (!field->name) {
        free(field);
        return;
    }
    
    strcpy(field->name, field_name);
    field->value = map_value; // Store pointer to map (caller owns the map)
    field->value_type = OMNI_TYPE_MAP;
    field->next = struct_ptr->fields;
    struct_ptr->fields = field;
}

omni_struct_t* omni_struct_get_struct_field(omni_struct_t* struct_ptr, const char* field_name) {
    if (!struct_ptr) return NULL;
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        if (strcmp(field->name, field_name) == 0 && field->value_type == OMNI_TYPE_STRUCT) {
            return (omni_struct_t*)field->value;
        }
        field = field->next;
    }
    return NULL;
}

void omni_struct_set_struct_field(omni_struct_t* struct_ptr, const char* field_name, omni_struct_t* struct_value) {
    if (!struct_ptr) return;
    
    // Check if field already exists
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        if (strcmp(field->name, field_name) == 0) {
            // Update existing field - free old value if it was a struct
            if (field->value_type == OMNI_TYPE_STRUCT && field->value) {
                omni_struct_destroy((omni_struct_t*)field->value);
            } else if (field->value_type == 0) { // string
                free((char*)field->value);
            } else if (field->value) {
                free(field->value);
            }
            // Store pointer to struct (caller owns the struct)
            field->value = struct_value;
            field->value_type = OMNI_TYPE_STRUCT;
            return;
        }
        field = field->next;
    }
    
    // Create new field
    field = (omni_struct_field_t*)malloc(sizeof(omni_struct_field_t));
    if (!field) return;
    
    field->name = malloc(strlen(field_name) + 1);
    if (!field->name) {
        free(field);
        return;
    }
    
    strcpy(field->name, field_name);
    field->value = struct_value; // Store pointer to struct (caller owns the struct)
    field->value_type = OMNI_TYPE_STRUCT;
    field->next = struct_ptr->fields;
    struct_ptr->fields = field;
}

void omni_struct_set_null_field(omni_struct_t* struct_ptr, const char* field_name) {
    if (!struct_ptr) return;
    
    // Check if field already exists
    omni_struct_field_t* field = struct_ptr->fields;
    while (field) {
        if (strcmp(field->name, field_name) == 0) {
            // Update existing field - free old value
            if (field->value_type == OMNI_TYPE_MAP && field->value) {
                omni_map_destroy((omni_map_t*)field->value);
            } else if (field->value_type == OMNI_TYPE_STRUCT && field->value) {
                omni_struct_destroy((omni_struct_t*)field->value);
            } else if (field->value_type == 0) { // string
                free((char*)field->value);
            } else if (field->value) {
                free(field->value);
            }
            // Set to null
            field->value = NULL;
            field->value_type = 0; // Use 0 for null/void
            return;
        }
        field = field->next;
    }
    
    // Create new field with null value
    field = (omni_struct_field_t*)malloc(sizeof(omni_struct_field_t));
    if (!field) return;
    
    field->name = malloc(strlen(field_name) + 1);
    if (!field->name) {
        free(field);
        return;
    }
    
    strcpy(field->name, field_name);
    field->value = NULL;
    field->value_type = 0; // Use 0 for null/void
    field->next = struct_ptr->fields;
    struct_ptr->fields = field;
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

// Network includes (needed for DNS and HTTP functions)
#ifdef _WIN32
#include <winsock2.h>
#include <ws2tcpip.h>
#include <iphlpapi.h>
#include <icmpapi.h>
#include <fcntl.h>
#include <sys/select.h>
#ifndef NI_MAXHOST
#define NI_MAXHOST 1025
#endif
#else
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <netdb.h>
#include <ifaddrs.h>
#include <net/if.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/select.h>
#ifndef NI_MAXHOST
#define NI_MAXHOST 1025
#endif
#ifndef IFF_LOOPBACK
#define IFF_LOOPBACK 0x8
#endif
#ifndef IFF_UP
#define IFF_UP 0x1
#endif
#endif

// Forward declaration
static int32_t omni_network_ping_tcp_fallback(const char* host);

// DNS functions
omni_ip_address_t** omni_dns_lookup(const char* hostname, int32_t* count) {
    if (!hostname || !count) return NULL;
    *count = 0;
    
    struct addrinfo hints, *result = NULL, *rp = NULL;
    memset(&hints, 0, sizeof(struct addrinfo));
    hints.ai_family = AF_UNSPEC;  // Support both IPv4 and IPv6
    hints.ai_socktype = SOCK_STREAM;
    hints.ai_flags = AI_ALL | AI_V4MAPPED;
    
    int status = getaddrinfo(hostname, NULL, &hints, &result);
    if (status != 0) {
        return NULL;
    }
    
    // Count addresses
    int addr_count = 0;
    for (rp = result; rp != NULL; rp = rp->ai_next) {
        addr_count++;
    }
    
    if (addr_count == 0) {
        freeaddrinfo(result);
        return NULL;
    }
    
    // Allocate array of IP address pointers
    omni_ip_address_t** ip_array = (omni_ip_address_t**)malloc(addr_count * sizeof(omni_ip_address_t*));
    if (!ip_array) {
        freeaddrinfo(result);
        return NULL;
    }
    
    // Convert addresses
    int idx = 0;
    for (rp = result; rp != NULL && idx < addr_count; rp = rp->ai_next) {
        omni_ip_address_t* ip = (omni_ip_address_t*)malloc(sizeof(omni_ip_address_t));
        if (!ip) {
            // Free already allocated IPs
            for (int i = 0; i < idx; i++) {
                free(ip_array[i]);
            }
            free(ip_array);
            freeaddrinfo(result);
            return NULL;
        }
        
        char ip_str[INET6_ADDRSTRLEN];
        if (rp->ai_family == AF_INET) {
            struct sockaddr_in* sin = (struct sockaddr_in*)rp->ai_addr;
            inet_ntop(AF_INET, &sin->sin_addr, ip_str, INET6_ADDRSTRLEN);
            ip->is_ipv4 = 1;
            ip->is_ipv6 = 0;
        } else if (rp->ai_family == AF_INET6) {
            struct sockaddr_in6* sin6 = (struct sockaddr_in6*)rp->ai_addr;
            inet_ntop(AF_INET6, &sin6->sin6_addr, ip_str, INET6_ADDRSTRLEN);
            ip->is_ipv4 = 0;
            ip->is_ipv6 = 1;
        } else {
            free(ip);
            continue;
        }
        
        strncpy(ip->address, ip_str, sizeof(ip->address) - 1);
        ip->address[sizeof(ip->address) - 1] = '\0';
        ip_array[idx++] = ip;
    }
    
    freeaddrinfo(result);
    *count = idx;
    return ip_array;
}

char* omni_dns_reverse_lookup(omni_ip_address_t* ip) {
    if (!ip) return NULL;
    
    struct sockaddr_storage sa;
    memset(&sa, 0, sizeof(sa));
    socklen_t salen;
    
    if (ip->is_ipv4) {
        struct sockaddr_in* sin = (struct sockaddr_in*)&sa;
        sin->sin_family = AF_INET;
        if (inet_pton(AF_INET, ip->address, &sin->sin_addr) != 1) {
            return strdup("");
        }
        salen = sizeof(struct sockaddr_in);
    } else if (ip->is_ipv6) {
        struct sockaddr_in6* sin6 = (struct sockaddr_in6*)&sa;
        sin6->sin6_family = AF_INET6;
        if (inet_pton(AF_INET6, ip->address, &sin6->sin6_addr) != 1) {
            return strdup("");
        }
        salen = sizeof(struct sockaddr_in6);
    } else {
        return strdup("");
    }
    
    char hostname[NI_MAXHOST];
    int status = getnameinfo((struct sockaddr*)&sa, salen, hostname, NI_MAXHOST, NULL, 0, 0);
    if (status != 0) {
        return strdup("");
    }
    
    return strdup(hostname);
}

// HTTP client helper functions
#ifdef HAVE_LIBCURL
#include <curl/curl.h>

// Check if libcurl is available at runtime
static int32_t omni_http_has_libcurl() {
    return 1; // If compiled with HAVE_LIBCURL, libcurl is available
}

// Structure to hold HTTP response data for libcurl
typedef struct {
    char* data;
    size_t size;
    omni_map_t* headers;
    int32_t status_code;
    char status_text[64];
} omni_curl_response_t;

// Callback for libcurl to write response body
static size_t omni_curl_write_callback(void* contents, size_t size, size_t nmemb, void* userp) {
    size_t realsize = size * nmemb;
    omni_curl_response_t* resp = (omni_curl_response_t*)userp;
    
    char* ptr = (char*)realloc(resp->data, resp->size + realsize + 1);
    if (!ptr) return 0;
    
    resp->data = ptr;
    memcpy(&(resp->data[resp->size]), contents, realsize);
    resp->size += realsize;
    resp->data[resp->size] = '\0';
    
    return realsize;
}

// Callback for libcurl to write headers
static size_t omni_curl_header_callback(char* buffer, size_t size, size_t nitems, void* userp) {
    size_t realsize = size * nitems;
    omni_curl_response_t* resp = (omni_curl_response_t*)userp;
    
    // Parse header line (format: "Header-Name: value\r\n")
    char* colon = strchr(buffer, ':');
    if (colon) {
        *colon = '\0';
        char* name = buffer;
        char* value = colon + 1;
        // Skip leading whitespace
        while (*value == ' ' || *value == '\t') value++;
        // Remove trailing \r\n
        char* end = value + strlen(value);
        while (end > value && (end[-1] == '\r' || end[-1] == '\n')) {
            end--;
            *end = '\0';
        }
        if (name && value) {
            omni_map_put_string_string(resp->headers, name, value);
        }
        *colon = ':'; // Restore for safety
    }
    
    return realsize;
}

// Perform HTTP request using libcurl
static omni_http_response_t* omni_http_via_libcurl(const char* method, const char* url, 
                                                    omni_map_t* headers, const char* body) {
    CURL* curl = curl_easy_init();
    if (!curl) return NULL;
    
    omni_curl_response_t curl_resp;
    memset(&curl_resp, 0, sizeof(curl_resp));
    curl_resp.headers = omni_map_create();
    curl_resp.status_code = 0;
    
    curl_easy_setopt(curl, CURLOPT_URL, url);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, omni_curl_write_callback);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, (void*)&curl_resp);
    curl_easy_setopt(curl, CURLOPT_HEADERFUNCTION, omni_curl_header_callback);
    curl_easy_setopt(curl, CURLOPT_HEADERDATA, (void*)&curl_resp);
    curl_easy_setopt(curl, CURLOPT_FOLLOWLOCATION, 1L);
    
    // Set HTTP method
    if (strcmp(method, "POST") == 0) {
        curl_easy_setopt(curl, CURLOPT_POST, 1L);
        if (body) {
            curl_easy_setopt(curl, CURLOPT_POSTFIELDS, body);
        }
    } else if (strcmp(method, "PUT") == 0) {
        curl_easy_setopt(curl, CURLOPT_CUSTOMREQUEST, "PUT");
        if (body) {
            curl_easy_setopt(curl, CURLOPT_POSTFIELDS, body);
        }
    } else if (strcmp(method, "DELETE") == 0) {
        curl_easy_setopt(curl, CURLOPT_CUSTOMREQUEST, "DELETE");
    }
    
    // Add custom headers
    struct curl_slist* header_list = NULL;
    if (headers) {
        // Iterate through headers map and add to curl
        // Note: This is simplified - full implementation would iterate map
        // For now, we'll set common headers manually if needed
    }
    if (header_list) {
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, header_list);
    }
    
    CURLcode res = curl_easy_perform(curl);
    
    if (res == CURLE_OK) {
        long http_code = 0;
        curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &http_code);
        curl_resp.status_code = (int32_t)http_code;
        
        // Get status text from response code
        if (http_code == 200) {
            strncpy(curl_resp.status_text, "OK", sizeof(curl_resp.status_text) - 1);
        } else if (http_code == 404) {
            strncpy(curl_resp.status_text, "Not Found", sizeof(curl_resp.status_text) - 1);
        } else if (http_code == 500) {
            strncpy(curl_resp.status_text, "Internal Server Error", sizeof(curl_resp.status_text) - 1);
        } else {
            snprintf(curl_resp.status_text, sizeof(curl_resp.status_text), "Status %ld", http_code);
        }
    } else {
        curl_easy_cleanup(curl);
        if (header_list) curl_slist_free_all(header_list);
        if (curl_resp.data) free(curl_resp.data);
        omni_map_destroy(curl_resp.headers);
        return NULL;
    }
    
    curl_easy_cleanup(curl);
    if (header_list) curl_slist_free_all(header_list);
    
    // Create response structure
    omni_http_response_t* resp = (omni_http_response_t*)malloc(sizeof(omni_http_response_t));
    if (!resp) {
        if (curl_resp.data) free(curl_resp.data);
        omni_map_destroy(curl_resp.headers);
        return NULL;
    }
    
    resp->status_code = curl_resp.status_code;
    strncpy(resp->status_text, curl_resp.status_text, sizeof(resp->status_text) - 1);
    resp->headers = curl_resp.headers;
    resp->body = curl_resp.data ? curl_resp.data : strdup("");
    
    return resp;
}
#else
// Helper function to check for libcurl availability (unused when libcurl not available)
__attribute__((unused)) static int32_t omni_http_has_libcurl() {
    return 0;
}
#endif

// Parse URL to extract host, port, and path
static int omni_http_parse_url(const char* url, char* host, int* port, char* path) {
    if (!url || !host || !port || !path) return 0;
    
    // Find scheme://
    const char* scheme_end = strstr(url, "://");
    if (!scheme_end) return 0;
    
    const char* host_start = scheme_end + 3;
    
    // Find port and path
    const char* path_start = strchr(host_start, '/');
    const char* port_start = strchr(host_start, ':');
    
    // Extract host
    size_t host_len;
    if (port_start && (!path_start || port_start < path_start)) {
        host_len = port_start - host_start;
        *port = atoi(port_start + 1);
    } else if (path_start) {
        host_len = path_start - host_start;
        *port = 80; // Default HTTP port
    } else {
        host_len = strlen(host_start);
        *port = 80;
    }
    
    if (host_len >= 256) host_len = 255;
    strncpy(host, host_start, host_len);
    host[host_len] = '\0';
    
    // Extract path
    if (path_start) {
        strncpy(path, path_start, 511);
        path[511] = '\0';
    } else {
        strcpy(path, "/");
    }
    
    // Adjust port for HTTPS
    if (strncmp(url, "https://", 8) == 0 && *port == 80) {
        *port = 443;
    }
    
    return 1;
}

// Perform HTTP request using raw sockets
static omni_http_response_t* omni_http_via_socket(const char* method, const char* url,
                                                   omni_map_t* headers, const char* body) {
    (void)headers; // Headers not yet implemented in raw socket version
    char host[256];
    int port;
    char path[512];
    
    if (!omni_http_parse_url(url, host, &port, path)) {
        return NULL;
    }
    
    // Resolve hostname to IP
    struct addrinfo hints, *result = NULL;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_INET;
    hints.ai_socktype = SOCK_STREAM;
    
    char port_str[16];
    snprintf(port_str, sizeof(port_str), "%d", port);
    
    if (getaddrinfo(host, port_str, &hints, &result) != 0) {
        return NULL;
    }
    
    // Create socket and connect
    int sock = socket(AF_INET, SOCK_STREAM, 0);
    if (sock < 0) {
        freeaddrinfo(result);
        return NULL;
    }
    
    if (connect(sock, result->ai_addr, result->ai_addrlen) != 0) {
#ifdef _WIN32
        closesocket(sock);
#else
        close(sock);
#endif
        freeaddrinfo(result);
        return NULL;
    }
    
    freeaddrinfo(result);
    
    // Build HTTP request
    char request[4096];
    snprintf(request, sizeof(request), "%s %s HTTP/1.1\r\n", method, path);
    strcat(request, "Host: ");
    strncat(request, host, sizeof(request) - strlen(request) - 1);
    strcat(request, "\r\n");
    strcat(request, "Connection: close\r\n");
    
    if (body) {
        char content_length[64];
        snprintf(content_length, sizeof(content_length), "Content-Length: %zu\r\n", strlen(body));
        strcat(request, content_length);
    }
    
    // Add custom headers
    // Note: Full implementation would iterate headers map
    
    strcat(request, "\r\n");
    
    if (body) {
        strncat(request, body, sizeof(request) - strlen(request) - 1);
    }
    
    // Send request
    if (send(sock, request, strlen(request), 0) < 0) {
        close(sock);
        return NULL;
    }
    
    // Read response
    char response_buffer[8192];
    ssize_t total_received = 0;
    ssize_t received;
    
    while ((received = recv(sock, response_buffer + total_received, 
                            sizeof(response_buffer) - total_received - 1, 0)) > 0) {
        total_received += received;
        if (total_received >= (ssize_t)(sizeof(response_buffer) - 1)) break;
    }
    
#ifdef _WIN32
    closesocket(sock);
#else
    close(sock);
#endif
    
    if (total_received <= 0) {
        return NULL;
    }
    
    response_buffer[total_received] = '\0';
    
    // Parse response
    omni_http_response_t* resp = (omni_http_response_t*)malloc(sizeof(omni_http_response_t));
    if (!resp) return NULL;
    
    resp->headers = omni_map_create();
    resp->status_code = 200;
    strncpy(resp->status_text, "OK", sizeof(resp->status_text) - 1);
    
    // Parse status line
    char* line = response_buffer;
    char* status_line_end = strstr(line, "\r\n");
    if (status_line_end) {
        *status_line_end = '\0';
        // Parse "HTTP/1.1 200 OK"
        char* status_code_start = strchr(line, ' ');
        if (status_code_start) {
            status_code_start++;
            resp->status_code = atoi(status_code_start);
            char* status_text_start = strchr(status_code_start, ' ');
            if (status_text_start) {
                status_text_start++;
                strncpy(resp->status_text, status_text_start, sizeof(resp->status_text) - 1);
            }
        }
        line = status_line_end + 2;
    }
    
    // Parse headers
    char* header_end = strstr(line, "\r\n\r\n");
    if (header_end) {
        *header_end = '\0';
        // Simple header parsing (one header per line)
        char* header_line = line;
        while (*header_line) {
            char* header_line_end = strstr(header_line, "\r\n");
            if (header_line_end) {
                *header_line_end = '\0';
            }
            char* colon = strchr(header_line, ':');
            if (colon) {
                *colon = '\0';
                char* name = header_line;
                char* value = colon + 1;
                while (*value == ' ' || *value == '\t') value++;
                omni_map_put_string_string(resp->headers, name, value);
                *colon = ':';
            }
            if (!header_line_end) break;
            header_line = header_line_end + 2;
        }
        line = header_end + 4;
    }
    
    // Body is everything after headers
    resp->body = strdup(line);
    
    return resp;
}

// HTTP client functions
omni_http_response_t* omni_http_get(const char* url) {
    if (!url) return NULL;
    
    // Try libcurl first, fallback to raw sockets
#ifdef HAVE_LIBCURL
    if (omni_http_has_libcurl()) {
        omni_http_response_t* resp = omni_http_via_libcurl("GET", url, NULL, NULL);
        if (resp) return resp;
    }
#endif
    
    return omni_http_via_socket("GET", url, NULL, NULL);
}

omni_http_response_t* omni_http_post(const char* url, const char* body) {
    if (!url) return NULL;
    
#ifdef HAVE_LIBCURL
    if (omni_http_has_libcurl()) {
        omni_http_response_t* resp = omni_http_via_libcurl("POST", url, NULL, body);
        if (resp) return resp;
    }
#endif
    
    return omni_http_via_socket("POST", url, NULL, body);
}

omni_http_response_t* omni_http_put(const char* url, const char* body) {
    if (!url) return NULL;
    
#ifdef HAVE_LIBCURL
    if (omni_http_has_libcurl()) {
        omni_http_response_t* resp = omni_http_via_libcurl("PUT", url, NULL, body);
        if (resp) return resp;
    }
#endif
    
    return omni_http_via_socket("PUT", url, NULL, body);
}

omni_http_response_t* omni_http_delete(const char* url) {
    if (!url) return NULL;
    
#ifdef HAVE_LIBCURL
    if (omni_http_has_libcurl()) {
        omni_http_response_t* resp = omni_http_via_libcurl("DELETE", url, NULL, NULL);
        if (resp) return resp;
    }
#endif
    
    return omni_http_via_socket("DELETE", url, NULL, NULL);
}

omni_http_response_t* omni_http_request(omni_http_request_t* req) {
    if (!req) return NULL;
    
#ifdef HAVE_LIBCURL
    if (omni_http_has_libcurl()) {
        omni_http_response_t* resp = omni_http_via_libcurl(req->method, req->url, req->headers, req->body);
        if (resp) return resp;
    }
#endif
    
    return omni_http_via_socket(req->method, req->url, req->headers, req->body);
}

// HTTP server functions
omni_http_request_t* omni_http_parse_request(const char* raw_request) {
    if (!raw_request) return NULL;
    
    omni_http_request_t* req = (omni_http_request_t*)malloc(sizeof(omni_http_request_t));
    if (!req) return NULL;
    
    req->headers = omni_map_create();
    req->body = NULL;
    memset(req->method, 0, sizeof(req->method));
    memset(req->url, 0, sizeof(req->url));
    
    // Make a copy to avoid modifying the original
    char* request = strdup(raw_request);
    if (!request) {
        free(req);
        return NULL;
    }
    
    // Parse request line: "METHOD /path?query HTTP/1.1\r\n"
    char* line = request;
    char* request_line_end = strstr(line, "\r\n");
    if (!request_line_end) {
        free(request);
        omni_map_destroy(req->headers);
        free(req);
        return NULL;
    }
    
    *request_line_end = '\0';
    
    // Extract method
    char* method_end = strchr(line, ' ');
    if (!method_end) {
        free(request);
        omni_map_destroy(req->headers);
        free(req);
        return NULL;
    }
    size_t method_len = method_end - line;
    if (method_len >= sizeof(req->method)) method_len = sizeof(req->method) - 1;
    strncpy(req->method, line, method_len);
    req->method[method_len] = '\0';
    
    // Extract URL (path + query)
    char* url_start = method_end + 1;
    while (*url_start == ' ') url_start++;
    char* url_end = strchr(url_start, ' ');
    if (!url_end) {
        free(request);
        omni_map_destroy(req->headers);
        free(req);
        return NULL;
    }
    size_t url_len = url_end - url_start;
    if (url_len >= sizeof(req->url)) url_len = sizeof(req->url) - 1;
    strncpy(req->url, url_start, url_len);
    req->url[url_len] = '\0';
    
    line = request_line_end + 2;
    
    // Parse headers
    char* header_end = strstr(line, "\r\n\r\n");
    if (header_end) {
        *header_end = '\0';
        char* header_line = line;
        while (*header_line) {
            char* header_line_end = strstr(header_line, "\r\n");
            if (header_line_end) {
                *header_line_end = '\0';
            }
            char* colon = strchr(header_line, ':');
            if (colon) {
                *colon = '\0';
                char* name = header_line;
                char* value = colon + 1;
                while (*value == ' ' || *value == '\t') value++;
                omni_map_put_string_string(req->headers, name, value);
                *colon = ':';
            }
            if (!header_line_end) break;
            header_line = header_line_end + 2;
        }
        line = header_end + 4;
    }
    
    // Body is everything after headers
    if (*line) {
        req->body = strdup(line);
    }
    
    free(request);
    return req;
}

char* omni_http_build_response(omni_http_response_t* resp) {
    if (!resp) return NULL;
    
    // Calculate size needed
    size_t size = 128; // Status line
    if (resp->headers) {
        // Estimate header size
        size += 512;
    }
    if (resp->body) {
        size += strlen(resp->body);
    }
    size += 64; // Extra buffer
    
    char* response = (char*)malloc(size);
    if (!response) return NULL;
    
    // Build status line: "HTTP/1.1 200 OK\r\n"
    int pos = snprintf(response, size, "HTTP/1.1 %d %s\r\n", 
                       resp->status_code, resp->status_text);
    
    // Add headers
    if (resp->headers) {
        // Iterate through headers map and add each header
        // Note: This is a simplified version - full implementation would iterate the map
        // For now, we'll add Content-Length if body exists
        if (resp->body) {
            pos += snprintf(response + pos, size - pos, "Content-Length: %zu\r\n", 
                           strlen(resp->body));
        }
        // Add other headers would go here
    }
    
    // End of headers
    pos += snprintf(response + pos, size - pos, "\r\n");
    
    // Add body
    if (resp->body) {
        if (pos + strlen(resp->body) < size) {
            strcpy(response + pos, resp->body);
            pos += strlen(resp->body);
        }
    }
    
    return response;
}

void omni_http_parse_query(const char* query_string, omni_map_t* params) {
    if (!query_string || !params) return;
    
    char* query = strdup(query_string);
    if (!query) return;
    
    char* token = strtok(query, "&");
    while (token) {
        char* equals = strchr(token, '=');
        if (equals) {
            *equals = '\0';
            char* key = token;
            char* value = equals + 1;
            
            // URL decode key and value
            char* decoded_key = omni_decode_url(key);
            char* decoded_value = omni_decode_url(value);
            
            if (decoded_key && decoded_value) {
                omni_map_put_string_string(params, decoded_key, decoded_value);
                free(decoded_key);
                free(decoded_value);
            }
            *equals = '=';
        }
        token = strtok(NULL, "&");
    }
    
    free(query);
}

int32_t omni_http_match_path(const char* pattern, const char* path, omni_map_t* params) {
    if (!pattern || !path || !params) return 0;
    
    // Simple pattern matching: /user/:id matches /user/123
    // Supports :param, *wildcard, and ?optional
    
    const char* p = pattern;
    const char* s = path;
    
    while (*p && *s) {
        if (*p == ':') {
            // Parameter: :id
            p++; // Skip ':'
            char param_name[64] = {0};
            int i = 0;
            while (*p && *p != '/' && *p != '?' && *p != '*' && i < 63) {
                param_name[i++] = *p++;
            }
            param_name[i] = '\0';
            
            // Extract parameter value
            char param_value[256] = {0};
            int j = 0;
            while (*s && *s != '/' && *s != '?' && j < 255) {
                param_value[j++] = *s++;
            }
            param_value[j] = '\0';
            
            // Store parameter
            omni_map_put_string_string(params, param_name, param_value);
        } else if (*p == '*') {
            // Wildcard: matches rest of path
            char wildcard_value[512] = {0};
            strncpy(wildcard_value, s, 511);
            omni_map_put_string_string(params, "*", wildcard_value);
            return 1; // Match
        } else if (*p == '?') {
            // Optional segment
            p++; // Skip '?'
            // Try to match optional segment
            if (*p == '/') p++;
            // Skip optional segment in path if present
            if (*s == '/') {
                while (*s && *s != '/') s++;
                if (*s == '/') s++;
            }
        } else if (*p == *s) {
            // Literal match
            p++;
            s++;
        } else {
            // Mismatch
            return 0;
        }
    }
    
    // Both must be at end (or pattern has optional trailing parts)
    while (*p == '?' || *p == '/') p++;
    return (*p == '\0' && (*s == '\0' || *s == '?'));
}

// Context functions (for std.web framework)
const char* omni_context_param(omni_struct_t* ctx, const char* name) {
    if (!ctx || !name) return "";
    // Stub implementation - would extract from ctx.params map
    (void)ctx;
    return "";
}

void* omni_context_body_json(omni_struct_t* ctx) {
    if (!ctx) return NULL;
    const char* body = omni_struct_get_string_field(ctx, "body");
    if (!body || !*body) return NULL;
    return omni_json_parse(body);
}

omni_struct_t* omni_context_text(omni_struct_t* ctx, const char* text) {
    if (!ctx || !text) return ctx;
    // Stub implementation - would set response body and headers
    (void)text;
    return ctx;
}

omni_struct_t* omni_context_json(omni_struct_t* ctx, void* data) {
    if (!ctx) return ctx;
    char* json_str = omni_json_stringify(data, 0, 0);
    if (json_str) free(json_str);
    return ctx;
}

omni_struct_t* omni_context_file(omni_struct_t* ctx, const char* path) {
    if (!ctx || !path) return ctx;
    int32_t size = 0;
    char* data = omni_file_read_binary(path, &size);
    if (data) free(data);
    return ctx;
}

// Additional context functions (stubs)
const char* omni_context_query(omni_struct_t* ctx, const char* name) {
    (void)ctx; (void)name;
    return "";
}

omni_map_t* omni_context_query_all(omni_struct_t* ctx) {
    (void)ctx;
    return omni_map_create();
}

const char* omni_context_header(omni_struct_t* ctx, const char* name) {
    (void)ctx; (void)name;
    return "";
}

void omni_context_set_header(omni_struct_t* ctx, const char* name, const char* value) {
    (void)ctx; (void)name; (void)value;
}

void omni_context_status(omni_struct_t* ctx, int32_t code) {
    (void)ctx; (void)code;
}

omni_struct_t* omni_context_html(omni_struct_t* ctx, const char* html) {
    (void)html;
    return ctx;
}

omni_struct_t* omni_context_redirect(omni_struct_t* ctx, const char* url, int32_t code) {
    (void)url; (void)code;
    return ctx;
}

omni_struct_t* omni_context_cookie(omni_struct_t* ctx, const char* name, const char* value, omni_map_t* options) {
    (void)name; (void)value; (void)options;
    return ctx;
}

const char* omni_context_get_cookie(omni_struct_t* ctx, const char* name) {
    (void)ctx; (void)name;
    return "";
}

const char* omni_context_body(omni_struct_t* ctx) {
    (void)ctx;
    return "";
}

omni_struct_t* omni_context_set_state(omni_struct_t* ctx, const char* key, void* value, int32_t value_type) {
    (void)key; (void)value; (void)value_type;
    return ctx;
}

void* omni_context_get_state(omni_struct_t* ctx, const char* key, int32_t* value_type) {
    (void)ctx; (void)key;
    if (value_type) *value_type = OMNI_TYPE_ANY;
    return NULL;
}

// Parse request body as form-urlencoded data. Stub returns empty map.
omni_map_t* omni_context_body_form(omni_struct_t* ctx) {
    (void)ctx;
    return omni_map_create();
}

// Uploaded files stub. Returns empty array.
omni_array_t* omni_context_files(omni_struct_t* ctx) {
    (void)ctx;
    return omni_array_create();
}

// WebSocket stubs
void omni_server_websocket(omni_server_t* server, const char* pattern, void* handler) {
    (void)server; (void)pattern; (void)handler;
}

int32_t omni_websocket_send(void* conn, const char* data, int32_t len) {
    (void)conn; (void)data; (void)len;
    return 0;
}

int32_t omni_websocket_receive(void* conn, char* buf, int32_t max_len) {
    (void)conn; (void)buf; (void)max_len;
    return 0;
}

void omni_websocket_close(void* conn) {
    (void)conn;
}

// Server routing functions (for std.web framework)
void omni_server_get(omni_server_t* server, const char* pattern, void* handler) {
    (void)server; (void)pattern; (void)handler;
}

void omni_server_post(omni_server_t* server, const char* pattern, void* handler) {
    (void)server; (void)pattern; (void)handler;
}

void omni_server_put(omni_server_t* server, const char* pattern, void* handler) {
    (void)server; (void)pattern; (void)handler;
}

void omni_server_delete(omni_server_t* server, const char* pattern, void* handler) {
    (void)server; (void)pattern; (void)handler;
}

void omni_server_patch(omni_server_t* server, const char* pattern, void* handler) {
    (void)server; (void)pattern; (void)handler;
}

void omni_server_all(omni_server_t* server, const char* pattern, void* handler) {
    (void)server; (void)pattern; (void)handler;
}

void omni_server_route(omni_server_t* server, const char* method, const char* pattern, void* handler) {
    (void)server; (void)method; (void)pattern; (void)handler;
}

omni_struct_t* omni_server_group(omni_server_t* server, const char* prefix) {
    (void)server; (void)prefix;
    return omni_struct_create();
}

void omni_server_use(omni_server_t* server, void* middleware) {
    (void)server; (void)middleware;
}

void omni_server_use_before(omni_server_t* server, void* middleware) {
    (void)server; (void)middleware;
}

void omni_server_use_after(omni_server_t* server, void* middleware) {
    (void)server; (void)middleware;
}

void omni_group_get(omni_struct_t* group, const char* pattern, void* handler) {
    (void)group; (void)pattern; (void)handler;
}

void omni_group_post(omni_struct_t* group, const char* pattern, void* handler) {
    (void)group; (void)pattern; (void)handler;
}

void omni_group_use(omni_struct_t* group, void* middleware) {
    (void)group; (void)middleware;
}

// Middleware functions (stubs)
omni_struct_t* omni_middleware_logger(omni_struct_t* ctx) {
    return ctx;
}

omni_struct_t* omni_middleware_cors(omni_struct_t* ctx, omni_map_t* options) {
    (void)options;
    return ctx;
}

omni_struct_t* omni_middleware_json_parser(omni_struct_t* ctx) {
    return ctx;
}

omni_struct_t* omni_middleware_form_parser(omni_struct_t* ctx) {
    return ctx;
}

omni_struct_t* omni_middleware_multipart_parser_impl(omni_struct_t* ctx) {
    return ctx;
}

void* omni_middleware_multipart_parser(omni_struct_t* ctx, int32_t max_size) {
    (void)ctx; (void)max_size;
    return (void*)omni_middleware_multipart_parser_impl;
}

omni_struct_t* omni_middleware_static_impl(omni_struct_t* ctx) {
    return ctx;
}

void* omni_middleware_static(omni_server_t* server, const char* path, const char* dir, omni_map_t* options) {
    (void)server; (void)path; (void)dir; (void)options;
    return (void*)omni_middleware_static_impl;
}

// Template functions (stubs)
char* omni_template_render(const char* template, omni_map_t* data) {
    (void)data;
    if (!template) return NULL;
    size_t len = strlen(template);
    char* result = (char*)malloc(len + 1);
    if (result) strcpy(result, template);
    return result;
}

char* omni_template_load(const char* path) {
    return omni_read_file(path);
}

void omni_template_cache_enable(int32_t enable) {
    (void)enable;
}

// Validation functions (stubs)
omni_map_t* omni_validate_request(omni_struct_t* ctx, omni_map_t* rules) {
    (void)ctx; (void)rules;
    return omni_map_create();
}

// Test client functions (stubs)
void* omni_test_client_create(omni_server_t* server) {
    (void)server;
    return NULL;
}

void* omni_test_client_get(void* client, const char* path, omni_map_t* headers) {
    (void)client; (void)path; (void)headers;
    return NULL;
}

void* omni_test_client_post(void* client, const char* path, const char* body, omni_map_t* headers) {
    (void)client; (void)path; (void)body; (void)headers;
    return NULL;
}

int32_t omni_test_response_status(void* resp) {
    (void)resp;
    return 0;
}

const char* omni_test_response_body(void* resp) {
    (void)resp;
    return "";
}

omni_map_t* omni_test_response_headers(void* resp) {
    (void)resp;
    return omni_map_create();
}

void* omni_test_response_json(void* resp) {
    (void)resp;
    return NULL;
}

// Memory management functions (stubs)
void omni_panic(const char* message) {
    if (message) {
        fprintf(stderr, "PANIC: %s\n", message);
    } else {
        fprintf(stderr, "PANIC: unknown error\n");
    }
    exit(1);
}

// Note: omni_malloc, omni_free, and omni_realloc are already defined earlier in this file

// JSON parser helper structure
typedef struct {
    const char* json;
    size_t pos;
    size_t len;
} json_parser_t;

static void json_skip_whitespace(json_parser_t* parser) {
    while (parser->pos < parser->len && 
           (parser->json[parser->pos] == ' ' || 
            parser->json[parser->pos] == '\t' || 
            parser->json[parser->pos] == '\n' || 
            parser->json[parser->pos] == '\r')) {
        parser->pos++;
    }
}

static void* json_parse_value(json_parser_t* parser);

static void* json_parse_string(json_parser_t* parser) {
    if (parser->json[parser->pos] != '"') return NULL;
    parser->pos++;
    
    size_t start = parser->pos;
    while (parser->pos < parser->len && parser->json[parser->pos] != '"') {
        if (parser->json[parser->pos] == '\\' && parser->pos + 1 < parser->len) {
            parser->pos += 2; // Skip escape sequence
        } else {
            parser->pos++;
        }
    }
    
    if (parser->pos >= parser->len) return NULL;
    
    size_t len = parser->pos - start;
    char* str = (char*)malloc(len + 1);
    if (!str) return NULL;
    
    // Copy and unescape
    size_t j = 0;
    for (size_t i = start; i < parser->pos; i++) {
        if (parser->json[i] == '\\' && i + 1 < parser->pos) {
            i++;
            switch (parser->json[i]) {
                case 'n': str[j++] = '\n'; break;
                case 't': str[j++] = '\t'; break;
                case 'r': str[j++] = '\r'; break;
                case '\\': str[j++] = '\\'; break;
                case '"': str[j++] = '"'; break;
                default: str[j++] = parser->json[i]; break;
            }
        } else {
            str[j++] = parser->json[i];
        }
    }
    str[j] = '\0';
    
    parser->pos++; // Skip closing quote
    return str;
}

static void* json_parse_number(json_parser_t* parser) {
    size_t start = parser->pos;
    int is_float = 0;
    
    if (parser->json[parser->pos] == '-') parser->pos++;
    while (parser->pos < parser->len && 
           (parser->json[parser->pos] >= '0' && parser->json[parser->pos] <= '9')) {
        parser->pos++;
    }
    if (parser->pos < parser->len && parser->json[parser->pos] == '.') {
        is_float = 1;
        parser->pos++;
        while (parser->pos < parser->len && 
               (parser->json[parser->pos] >= '0' && parser->json[parser->pos] <= '9')) {
            parser->pos++;
        }
    }
    
    size_t len = parser->pos - start;
    char* num_str = (char*)malloc(len + 1);
    if (!num_str) return NULL;
    strncpy(num_str, parser->json + start, len);
    num_str[len] = '\0';
    
    if (is_float) {
        double* d = (double*)malloc(sizeof(double));
        if (d) *d = atof(num_str);
        free(num_str);
        return d;
    } else {
        int32_t* i = (int32_t*)malloc(sizeof(int32_t));
        if (i) *i = atoi(num_str);
        free(num_str);
        return i;
    }
}

static void* json_parse_object(json_parser_t* parser) {
    if (parser->json[parser->pos] != '{') return NULL;
    parser->pos++;
    
    omni_map_t* map = omni_map_create();
    if (!map) return NULL;
    
    json_skip_whitespace(parser);
    
    if (parser->json[parser->pos] == '}') {
        parser->pos++;
        return map;
    }
    
    while (parser->pos < parser->len) {
        json_skip_whitespace(parser);
        
        // Parse key
        void* key_ptr = json_parse_string(parser);
        if (!key_ptr) {
            omni_map_destroy(map);
            return NULL;
        }
        char* key = (char*)key_ptr;
        
        json_skip_whitespace(parser);
        if (parser->json[parser->pos] != ':') {
            free(key);
            omni_map_destroy(map);
            return NULL;
        }
        parser->pos++;
        
        json_skip_whitespace(parser);
        
        // Parse value
        void* value = json_parse_value(parser);
        if (!value) {
            free(key);
            omni_map_destroy(map);
            return NULL;
        }
        
        // Store in map (simplified - assumes string values for now)
        if (value) {
            char* value_str = (char*)value;
            omni_map_put_string_string(map, key, value_str);
            free(value_str);
        }
        free(key);
        
        json_skip_whitespace(parser);
        if (parser->json[parser->pos] == '}') {
            parser->pos++;
            break;
        }
        if (parser->json[parser->pos] != ',') {
            omni_map_destroy(map);
            return NULL;
        }
        parser->pos++;
    }
    
    return map;
}

static void* json_parse_array(json_parser_t* parser) {
    if (parser->json[parser->pos] != '[') return NULL;
    parser->pos++;
    
    omni_array_t* arr = omni_array_create();
    if (!arr) return NULL;
    
    json_skip_whitespace(parser);
    
    if (parser->json[parser->pos] == ']') {
        parser->pos++;
        return arr;
    }
    
    while (parser->pos < parser->len) {
        json_skip_whitespace(parser);
        
        void* value = json_parse_value(parser);
        if (!value) {
            omni_array_destroy(arr);
            return NULL;
        }
        
        // Add to array (value is already a pointer)
        if (value) {
            omni_array_append(arr, value);
        }
        
        json_skip_whitespace(parser);
        if (parser->json[parser->pos] == ']') {
            parser->pos++;
            break;
        }
        if (parser->json[parser->pos] != ',') {
            omni_array_destroy(arr);
            return NULL;
        }
        parser->pos++;
    }
    
    return arr;
}

static void* json_parse_value(json_parser_t* parser) {
    json_skip_whitespace(parser);
    
    if (parser->pos >= parser->len) return NULL;
    
    char c = parser->json[parser->pos];
    
    if (c == '"') {
        return json_parse_string(parser);
    } else if (c == '{') {
        return json_parse_object(parser);
    } else if (c == '[') {
        return json_parse_array(parser);
    } else if (c == '-' || (c >= '0' && c <= '9')) {
        return json_parse_number(parser);
    } else if (c == 't' && parser->pos + 3 < parser->len && 
               strncmp(parser->json + parser->pos, "true", 4) == 0) {
        parser->pos += 4;
        int32_t* b = (int32_t*)malloc(sizeof(int32_t));
        if (b) *b = 1;
        return b;
    } else if (c == 'f' && parser->pos + 4 < parser->len && 
               strncmp(parser->json + parser->pos, "false", 5) == 0) {
        parser->pos += 5;
        int32_t* b = (int32_t*)malloc(sizeof(int32_t));
        if (b) *b = 0;
        return b;
    } else if (c == 'n' && parser->pos + 3 < parser->len && 
               strncmp(parser->json + parser->pos, "null", 4) == 0) {
        parser->pos += 4;
        return NULL; // NULL value
    }
    
    return NULL;
}

void* omni_json_parse(const char* json_str) {
    if (!json_str) return NULL;
    
    json_parser_t parser;
    parser.json = json_str;
    parser.pos = 0;
    parser.len = strlen(json_str);
    
    return json_parse_value(&parser);
}

// JSON stringifier
static void json_stringify_value(char* buffer, size_t* pos, size_t size, void* value, int32_t value_type, int32_t pretty, int32_t indent) {
    (void)pretty;  // Unused for now
    (void)indent;  // Unused for now
    if (!value) {
        *pos += snprintf(buffer + *pos, size - *pos, "null");
        return;
    }
    
    switch (value_type) {
        case 0: { // int
            int32_t* i = (int32_t*)value;
            *pos += snprintf(buffer + *pos, size - *pos, "%d", *i);
            break;
        }
        case 1: { // string
            char* str = (char*)value;
            *pos += snprintf(buffer + *pos, size - *pos, "\"%s\"", str);
            break;
        }
        case 2: { // float
            double* d = (double*)value;
            *pos += snprintf(buffer + *pos, size - *pos, "%.6f", *d);
            break;
        }
        case 3: { // bool
            int32_t* b = (int32_t*)value;
            *pos += snprintf(buffer + *pos, size - *pos, *b ? "true" : "false");
            break;
        }
        case 4: { // map
            (void)value; // Map handling not yet implemented
            // omni_map_t* map = (omni_map_t*)value;
            *pos += snprintf(buffer + *pos, size - *pos, "{");
            // Simplified - would iterate map
            *pos += snprintf(buffer + *pos, size - *pos, "}");
            break;
        }
        case 5: { // array
            (void)value; // Array handling not yet implemented
            // omni_array_t* arr = (omni_array_t*)value;
            *pos += snprintf(buffer + *pos, size - *pos, "[");
            // Simplified - would iterate array
            *pos += snprintf(buffer + *pos, size - *pos, "]");
            break;
        }
        default:
            *pos += snprintf(buffer + *pos, size - *pos, "null");
            break;
    }
}

char* omni_json_stringify(void* value, int32_t value_type, int32_t pretty) {
    if (!value && value_type != 0) {
        char* result = (char*)malloc(5);
        if (result) strcpy(result, "null");
        return result;
    }
    
    size_t size = 1024;
    char* buffer = (char*)malloc(size);
    if (!buffer) return NULL;
    
    size_t pos = 0;
    json_stringify_value(buffer, &pos, size, value, value_type, pretty, 0);
    buffer[pos] = '\0';
    
    return buffer;
}

// Form data parsing
void omni_http_parse_form_urlencoded(const char* body, omni_map_t* params) {
    if (!body || !params) return;
    
    // Same as query string parsing
    omni_http_parse_query(body, params);
}

// Multipart form data structure for files
typedef struct {
    char name[256];
    char filename[256];
    char content_type[128];
    int32_t size;
    char* data;
    char path[512];
} omni_uploaded_file_t;

void omni_http_parse_multipart(const char* body, const char* boundary, omni_map_t* fields, omni_array_t* files) {
    if (!body || !boundary || !fields) return;
    
    char boundary_str[256];
    snprintf(boundary_str, sizeof(boundary_str), "--%s", boundary);
    
    const char* part_start = strstr(body, boundary_str);
    if (!part_start) return;
    
    part_start += strlen(boundary_str);
    
    while (part_start && *part_start) {
        // Skip CRLF
        if (*part_start == '\r') part_start++;
        if (*part_start == '\n') part_start++;
        
        // Parse headers
        const char* header_end = strstr(part_start, "\r\n\r\n");
        if (!header_end) break;
        
        // Extract Content-Disposition
        const char* disp_start = strstr(part_start, "Content-Disposition:");
        if (disp_start && disp_start < header_end) {
            // Parse name and filename
            char name[256] = {0};
            char filename[256] = {0};
            
            const char* name_start = strstr(disp_start, "name=\"");
            if (name_start) {
                name_start += 6;
                const char* name_end = strchr(name_start, '"');
                if (name_end) {
                    size_t name_len = name_end - name_start;
                    if (name_len < sizeof(name)) {
                        strncpy(name, name_start, name_len);
                    }
                }
            }
            
            const char* filename_start = strstr(disp_start, "filename=\"");
            if (filename_start) {
                filename_start += 10;
                const char* filename_end = strchr(filename_start, '"');
                if (filename_end) {
                    size_t filename_len = filename_end - filename_start;
                    if (filename_len < sizeof(filename)) {
                        strncpy(filename, filename_start, filename_len);
                    }
                }
            }
            
            // Extract Content-Type if present
            char content_type[128] = {0};
            const char* type_start = strstr(part_start, "Content-Type:");
            if (type_start && type_start < header_end) {
                type_start += 13;
                while (*type_start == ' ') type_start++;
                const char* type_end = strstr(type_start, "\r\n");
                if (type_end) {
                    size_t type_len = type_end - type_start;
                    if (type_len < sizeof(content_type)) {
                        strncpy(content_type, type_start, type_len);
                    }
                }
            }
            
            // Get part data (between headers and next boundary)
            const char* data_start = header_end + 4;
            const char* data_end = strstr(data_start, boundary_str);
            if (!data_end) {
                // Last part - find final boundary
                data_end = strstr(data_start, "\r\n--");
                if (!data_end) data_end = data_start + strlen(data_start);
            }
            
            size_t data_size = data_end - data_start;
            // Remove trailing CRLF
            while (data_size > 0 && (data_start[data_size - 1] == '\n' || data_start[data_size - 1] == '\r')) {
                data_size--;
            }
            
            if (filename[0]) {
                // This is a file upload
                if (files) {
                    omni_uploaded_file_t* file = (omni_uploaded_file_t*)malloc(sizeof(omni_uploaded_file_t));
                    if (file) {
                        strncpy(file->name, name, sizeof(file->name) - 1);
                        strncpy(file->filename, filename, sizeof(file->filename) - 1);
                        strncpy(file->content_type, content_type, sizeof(file->content_type) - 1);
                        file->size = (int32_t)data_size;
                        file->data = (char*)malloc(data_size + 1);
                        if (file->data) {
                            memcpy(file->data, data_start, data_size);
                            file->data[data_size] = '\0';
                        }
                        file->path[0] = '\0';
                        omni_array_append(files, file);
                    }
                }
            } else {
                // This is a form field
                char* value = (char*)malloc(data_size + 1);
                if (value) {
                    memcpy(value, data_start, data_size);
                    value[data_size] = '\0';
                    omni_map_put_string_string(fields, name, value);
                    free(value);
                }
            }
        }
        
        // Move to next part
        part_start = strstr(part_start, boundary_str);
        if (part_start) {
            part_start += strlen(boundary_str);
            // Check if this is the final boundary
            if (part_start[0] == '-' && part_start[1] == '-') {
                break; // End of multipart data
            }
        } else {
            break;
        }
    }
}

char* omni_file_upload_save(const char* data, int32_t size, const char* filename, const char* upload_dir) {
    if (!data || size <= 0 || !filename || !upload_dir) return NULL;
    
    // Generate unique filename if needed
    char full_path[1024];
    snprintf(full_path, sizeof(full_path), "%s/%s", upload_dir, filename);
    
    // Create directory if it doesn't exist
    #ifdef _WIN32
    _mkdir(upload_dir);
    #else
    mkdir(upload_dir, 0755);
    #endif
    
    // Write file
    FILE* f = fopen(full_path, "wb");
    if (!f) return NULL;
    
    size_t written = fwrite(data, 1, size, f);
    fclose(f);
    
    if (written != (size_t)size) {
        return NULL;
    }
    
    char* result = strdup(full_path);
    return result;
}

int32_t omni_file_upload_validate(const char* filename, int32_t size, const char* allowed_types, int32_t max_size) {
    if (!filename) return 0;
    
    // Check size
    if (max_size > 0 && size > max_size) return 0;
    
    // Check file extension/type
    if (allowed_types && *allowed_types) {
        // Get file extension
        const char* ext = strrchr(filename, '.');
        if (!ext) return 0;
        ext++; // Skip '.'
        
        // Check if extension is in allowed_types (comma-separated list)
        char* types = strdup(allowed_types);
        if (!types) return 0;
        
        char* token = strtok(types, ",");
        int found = 0;
        while (token) {
            // Trim whitespace
            while (*token == ' ') token++;
            char* token_end = token + strlen(token) - 1;
            while (token_end > token && *token_end == ' ') {
                *token_end = '\0';
                token_end--;
            }
            
            if (strcasecmp(ext, token) == 0) {
                found = 1;
                break;
            }
            token = strtok(NULL, ",");
        }
        
        free(types);
        return found;
    }
    
    return 1; // No restrictions
}

// Static file serving functions
char* omni_file_read_binary(const char* path, int32_t* size) {
    if (!path || !size) return NULL;
    
    FILE* f = fopen(path, "rb");
    if (!f) return NULL;
    
    // Get file size
    fseek(f, 0, SEEK_END);
    long file_size = ftell(f);
    fseek(f, 0, SEEK_SET);
    
    if (file_size < 0) {
        fclose(f);
        return NULL;
    }
    
    // Allocate buffer
    char* buffer = (char*)malloc(file_size + 1);
    if (!buffer) {
        fclose(f);
        return NULL;
    }
    
    // Read file
    size_t read_size = fread(buffer, 1, file_size, f);
    fclose(f);
    
    if (read_size != (size_t)file_size) {
        free(buffer);
        return NULL;
    }
    
    buffer[file_size] = '\0';
    *size = (int32_t)file_size;
    return buffer;
}

const char* omni_file_get_mime_type(const char* filename) {
    if (!filename) return "application/octet-stream";
    
    const char* ext = strrchr(filename, '.');
    if (!ext) return "application/octet-stream";
    ext++; // Skip '.'
    
    // Common MIME types
    if (strcasecmp(ext, "html") == 0 || strcasecmp(ext, "htm") == 0) {
        return "text/html";
    } else if (strcasecmp(ext, "css") == 0) {
        return "text/css";
    } else if (strcasecmp(ext, "js") == 0) {
        return "application/javascript";
    } else if (strcasecmp(ext, "json") == 0) {
        return "application/json";
    } else if (strcasecmp(ext, "png") == 0) {
        return "image/png";
    } else if (strcasecmp(ext, "jpg") == 0 || strcasecmp(ext, "jpeg") == 0) {
        return "image/jpeg";
    } else if (strcasecmp(ext, "gif") == 0) {
        return "image/gif";
    } else if (strcasecmp(ext, "svg") == 0) {
        return "image/svg+xml";
    } else if (strcasecmp(ext, "pdf") == 0) {
        return "application/pdf";
    } else if (strcasecmp(ext, "txt") == 0) {
        return "text/plain";
    } else if (strcasecmp(ext, "xml") == 0) {
        return "application/xml";
    } else if (strcasecmp(ext, "zip") == 0) {
        return "application/zip";
    } else if (strcasecmp(ext, "ico") == 0) {
        return "image/x-icon";
    }
    
    return "application/octet-stream";
}

int32_t omni_file_get_size(const char* path) {
    if (!path) return -1;
    
    struct stat st;
    if (stat(path, &st) != 0) return -1;
    
    if (S_ISREG(st.st_mode)) {
        return (int32_t)st.st_size;
    }
    
    return -1;
}

// Response compression functions
// Note: Full gzip implementation would require zlib library
// This is a simplified version
char* omni_http_compress_gzip(const char* data, int32_t len, int32_t* compressed_len) {
    if (!data || len <= 0 || !compressed_len) return NULL;
    
    // Simplified compression - in production, would use zlib
    // For now, return a copy of the data (no compression)
    // This allows the framework to work, compression can be added later with zlib
    
    char* compressed = (char*)malloc(len + 1);
    if (!compressed) return NULL;
    
    memcpy(compressed, data, len);
    compressed[len] = '\0';
    *compressed_len = len;
    
    return compressed;
}

char* omni_http_decompress_gzip(const char* compressed, int32_t len, int32_t* decompressed_len) {
    if (!compressed || len <= 0 || !decompressed_len) return NULL;
    
    // Simplified decompression - in production, would use zlib
    // For now, return a copy of the data (no decompression)
    
    char* decompressed = (char*)malloc(len + 1);
    if (!decompressed) return NULL;
    
    memcpy(decompressed, compressed, len);
    decompressed[len] = '\0';
    *decompressed_len = len;
    
    return decompressed;
}

// Validation and sanitization functions
int32_t omni_validate_string(const char* value, const char* pattern, int32_t min_len, int32_t max_len) {
    if (!value) return 0;
    
    int32_t len = (int32_t)strlen(value);
    
    // Check length constraints
    if (min_len > 0 && len < min_len) return 0;
    if (max_len > 0 && len > max_len) return 0;
    
    // Check pattern (regex) if provided
    if (pattern && *pattern) {
        return omni_string_matches(value, pattern);
    }
    
    return 1; // Valid
}

int32_t omni_validate_int(const char* value, int32_t min, int32_t max) {
    if (!value) return 0;
    
    char* endptr;
    errno = 0;
    long num = strtol(value, &endptr, 10);
    
    // Check for conversion errors
    if (errno != 0 || *endptr != '\0' || endptr == value) {
        return 0;
    }
    
    // Check range
    if (num < min || num > max) {
        return 0;
    }
    
    return 1; // Valid
}

int32_t omni_validate_email(const char* email) {
    if (!email) return 0;
    
    // Basic email validation: contains @ and .
    const char* at = strchr(email, '@');
    if (!at) return 0;
    
    const char* dot = strchr(at, '.');
    if (!dot) return 0;
    
    // Check that @ comes before .
    if (dot < at) return 0;
    
    // Check that there's at least one character before @
    if (at == email) return 0;
    
    // Check that there's at least one character after @ and before .
    if (dot - at <= 1) return 0;
    
    // Check that there's at least one character after .
    if (dot[1] == '\0') return 0;
    
    return 1; // Valid
}

int32_t omni_validate_url(const char* url) {
    if (!url) return 0;
    
    // Basic URL validation: starts with http:// or https://
    if (strncmp(url, "http://", 7) != 0 && strncmp(url, "https://", 8) != 0) {
        return 0;
    }
    
    // Check that there's at least one character after the scheme
    const char* host_start = strstr(url, "://");
    if (!host_start) return 0;
    host_start += 3;
    
    if (*host_start == '\0') return 0;
    
    return 1; // Valid
}

char* omni_sanitize_html(const char* html) {
    if (!html) return NULL;
    
    size_t len = strlen(html);
    char* sanitized = (char*)malloc(len * 6 + 1); // Worst case: all chars need escaping
    if (!sanitized) return NULL;
    
    size_t j = 0;
    for (size_t i = 0; i < len; i++) {
        switch (html[i]) {
            case '<':
                strcpy(sanitized + j, "&lt;");
                j += 4;
                break;
            case '>':
                strcpy(sanitized + j, "&gt;");
                j += 4;
                break;
            case '&':
                strcpy(sanitized + j, "&amp;");
                j += 5;
                break;
            case '"':
                strcpy(sanitized + j, "&quot;");
                j += 6;
                break;
            case '\'':
                strcpy(sanitized + j, "&#39;");
                j += 5;
                break;
            default:
                sanitized[j++] = html[i];
                break;
        }
    }
    sanitized[j] = '\0';
    
    return sanitized;
}

char* omni_sanitize_sql(const char* sql) {
    if (!sql) return NULL;
    
    size_t len = strlen(sql);
    char* escaped = (char*)malloc(len * 2 + 1); // Worst case: all chars need escaping
    if (!escaped) return NULL;
    
    size_t j = 0;
    for (size_t i = 0; i < len; i++) {
        switch (sql[i]) {
            case '\'':
                escaped[j++] = '\'';
                escaped[j++] = '\'';
                break;
            case '\\':
                escaped[j++] = '\\';
                escaped[j++] = '\\';
                break;
            case '\0':
                escaped[j++] = '\\';
                escaped[j++] = '0';
                break;
            case '\n':
                escaped[j++] = '\\';
                escaped[j++] = 'n';
                break;
            case '\r':
                escaped[j++] = '\\';
                escaped[j++] = 'r';
                break;
            case '\t':
                escaped[j++] = '\\';
                escaped[j++] = 't';
                break;
            default:
                escaped[j++] = sql[i];
                break;
        }
    }
    escaped[j] = '\0';
    
    return escaped;
}

// WebSocket functions
// WebSocket handshake response structure
typedef struct {
    char accept_key[128];
    char response[512];
} omni_websocket_handshake_t;

// Simple SHA1 implementation (simplified - full implementation would use OpenSSL or similar)
static void simple_sha1(const char* input, char* output) {
    // Simplified SHA1 - in production, would use proper SHA1 implementation
    // For now, create a basic hash-like string
    size_t len = strlen(input);
    snprintf(output, 41, "%08x%08x%08x%08x%08x", 
             (unsigned int)(len * 0x12345678),
             (unsigned int)(len * 0x87654321),
             (unsigned int)(len * 0xABCDEF00),
             (unsigned int)(len * 0x00FEDCBA),
             (unsigned int)(len * 0x11223344));
}

char* omni_websocket_handshake(const char* request_headers) {
    if (!request_headers) return NULL;
    
    // Extract Sec-WebSocket-Key from headers
    const char* key_start = strstr(request_headers, "Sec-WebSocket-Key:");
    if (!key_start) return NULL;
    
    key_start += 18; // Skip "Sec-WebSocket-Key:"
    while (*key_start == ' ' || *key_start == '\t') key_start++;
    
    const char* key_end = strstr(key_start, "\r\n");
    if (!key_end) return NULL;
    
    size_t key_len = key_end - key_start;
    char key[256] = {0};
    if (key_len >= sizeof(key)) key_len = sizeof(key) - 1;
    strncpy(key, key_start, key_len);
    key[key_len] = '\0';
    
    // WebSocket magic string
    const char* magic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11";
    
    // Concatenate key + magic
    char combined[512];
    snprintf(combined, sizeof(combined), "%s%s", key, magic);
    
    // SHA1 hash (simplified)
    char hash[41];
    simple_sha1(combined, hash);
    
    // Base64 encode (using existing function)
    char* encoded = omni_encode_base64(hash);
    if (!encoded) return NULL;
    
    // Build response
    char* response = (char*)malloc(512);
    if (!response) {
        free(encoded);
        return NULL;
    }
    
    snprintf(response, 512,
             "HTTP/1.1 101 Switching Protocols\r\n"
             "Upgrade: websocket\r\n"
             "Connection: Upgrade\r\n"
             "Sec-WebSocket-Accept: %s\r\n"
             "\r\n",
             encoded);
    
    free(encoded);
    return response;
}

// WebSocket frame structure
typedef struct {
    int32_t fin;
    int32_t opcode;
    int32_t mask;
    int32_t payload_len;
    char* payload;
} omni_websocket_frame_t;

char* omni_websocket_frame_create(const char* data, int32_t len, int32_t opcode, int32_t mask) {
    if (!data || len < 0) return NULL;
    
    // Calculate frame size
    size_t frame_size = 2; // Base header
    if (len > 125) {
        if (len > 65535) {
            frame_size += 8; // 64-bit length
        } else {
            frame_size += 2; // 16-bit length
        }
    }
    if (mask) {
        frame_size += 4; // Masking key
    }
    frame_size += len; // Payload
    
    char* frame = (char*)malloc(frame_size);
    if (!frame) return NULL;
    
    size_t pos = 0;
    
    // First byte: FIN (1 bit) + RSV (3 bits) + Opcode (4 bits)
    frame[pos++] = (char)(0x80 | (opcode & 0x0F)); // FIN=1, opcode
    
    // Second byte: MASK (1 bit) + Payload length (7 bits)
    if (len < 126) {
        frame[pos++] = (char)((mask ? 0x80 : 0x00) | len);
    } else if (len < 65536) {
        frame[pos++] = (char)((mask ? 0x80 : 0x00) | 126);
        frame[pos++] = (char)((len >> 8) & 0xFF);
        frame[pos++] = (char)(len & 0xFF);
    } else {
        frame[pos++] = (char)((mask ? 0x80 : 0x00) | 127);
        // 64-bit length (simplified - use 32-bit)
        frame[pos++] = 0;
        frame[pos++] = 0;
        frame[pos++] = 0;
        frame[pos++] = 0;
        frame[pos++] = (char)((len >> 24) & 0xFF);
        frame[pos++] = (char)((len >> 16) & 0xFF);
        frame[pos++] = (char)((len >> 8) & 0xFF);
        frame[pos++] = (char)(len & 0xFF);
    }
    
    // Masking key (if needed)
    if (mask) {
        uint32_t mask_key = 0x12345678; // Simplified - would be random
        frame[pos++] = (char)((mask_key >> 24) & 0xFF);
        frame[pos++] = (char)((mask_key >> 16) & 0xFF);
        frame[pos++] = (char)((mask_key >> 8) & 0xFF);
        frame[pos++] = (char)(mask_key & 0xFF);
        
        // Mask payload
        for (int32_t i = 0; i < len; i++) {
            frame[pos++] = (char)(data[i] ^ ((mask_key >> ((i % 4) * 8)) & 0xFF));
        }
    } else {
        // Copy payload
        memcpy(frame + pos, data, len);
        pos += len;
    }
    
    return frame;
}

void* omni_websocket_frame_parse(const char* frame, int32_t len) {
    if (!frame || len < 2) return NULL;
    
    omni_websocket_frame_t* parsed = (omni_websocket_frame_t*)malloc(sizeof(omni_websocket_frame_t));
    if (!parsed) return NULL;
    
    // Parse first byte
    parsed->fin = (frame[0] & 0x80) ? 1 : 0;
    parsed->opcode = frame[0] & 0x0F;
    
    // Parse second byte
    parsed->mask = (frame[1] & 0x80) ? 1 : 0;
    int32_t payload_len = frame[1] & 0x7F;
    
    size_t pos = 2;
    
    // Extended length
    if (payload_len == 126) {
        if (len < 4) {
            free(parsed);
            return NULL;
        }
        payload_len = (frame[2] << 8) | frame[3];
        pos = 4;
    } else if (payload_len == 127) {
        if (len < 10) {
            free(parsed);
            return NULL;
        }
        // Simplified - use 32-bit length
        payload_len = (frame[6] << 24) | (frame[7] << 16) | (frame[8] << 8) | frame[9];
        pos = 10;
    }
    
    // Masking key
    uint32_t mask_key = 0;
    if (parsed->mask) {
        if ((int32_t)len < (int32_t)(pos + 4)) {
            free(parsed);
            return NULL;
        }
        mask_key = (frame[pos] << 24) | (frame[pos+1] << 16) | (frame[pos+2] << 8) | frame[pos+3];
        pos += 4;
    }
    
    // Payload
    if ((int32_t)len < (int32_t)(pos + payload_len)) {
        free(parsed);
        return NULL;
    }
    
    parsed->payload_len = payload_len;
    parsed->payload = (char*)malloc(payload_len + 1);
    if (!parsed->payload) {
        free(parsed);
        return NULL;
    }
    
    if (parsed->mask) {
        // Unmask payload
        for (int32_t i = 0; i < payload_len; i++) {
            parsed->payload[i] = (char)(frame[pos + i] ^ ((mask_key >> ((i % 4) * 8)) & 0xFF));
        }
    } else {
        memcpy(parsed->payload, frame + pos, payload_len);
    }
    parsed->payload[payload_len] = '\0';
    
    return parsed;
}

// Server concurrency and connection management
struct omni_connection_pool {
    int32_t* sockets;
    int32_t max_connections;
    int32_t current_connections;
    int32_t* in_use;
};

omni_connection_pool_t* omni_server_connection_pool_create(int32_t max_connections) {
    if (max_connections <= 0) return NULL;
    
    omni_connection_pool_t* pool = (omni_connection_pool_t*)malloc(sizeof(omni_connection_pool_t));
    if (!pool) return NULL;
    
    pool->sockets = (int32_t*)malloc(max_connections * sizeof(int32_t));
    pool->in_use = (int32_t*)malloc(max_connections * sizeof(int32_t));
    if (!pool->sockets || !pool->in_use) {
        if (pool->sockets) free(pool->sockets);
        if (pool->in_use) free(pool->in_use);
        free(pool);
        return NULL;
    }
    
    pool->max_connections = max_connections;
    pool->current_connections = 0;
    for (int32_t i = 0; i < max_connections; i++) {
        pool->in_use[i] = 0;
    }
    
    return pool;
}

int32_t omni_server_connection_pool_acquire(omni_connection_pool_t* pool) {
    if (!pool) return -1;
    
    // Find available socket
    for (int32_t i = 0; i < pool->max_connections; i++) {
        if (!pool->in_use[i]) {
            pool->in_use[i] = 1;
            return pool->sockets[i];
        }
    }
    
    return -1; // No available connections
}

void omni_server_connection_pool_release(omni_connection_pool_t* pool, int32_t socket) {
    if (!pool) return;
    
    // Find and release socket
    for (int32_t i = 0; i < pool->max_connections; i++) {
        if (pool->sockets[i] == socket && pool->in_use[i]) {
            pool->in_use[i] = 0;
            break;
        }
    }
}

struct omni_thread_pool {
    int32_t num_threads;
    // Simplified - full implementation would use pthreads
};

omni_thread_pool_t* omni_server_thread_pool_create(int32_t num_threads) {
    if (num_threads <= 0) return NULL;
    
    omni_thread_pool_t* pool = (omni_thread_pool_t*)malloc(sizeof(omni_thread_pool_t));
    if (!pool) return NULL;
    
    pool->num_threads = num_threads;
    
    return pool;
}

void omni_server_thread_pool_submit(omni_thread_pool_t* pool, void (*task)(void*), void* arg) {
    if (!pool || !task) return;
    
    // Simplified - execute task directly
    // Full implementation would queue task and execute in thread pool
    task(arg);
}

// Server timeouts and limits
static int32_t global_max_request_size = 10 * 1024 * 1024; // 10MB default
static int32_t global_max_headers_size = 8192; // 8KB default

void omni_server_set_timeout(int32_t socket, int32_t timeout_seconds) {
    if (socket < 0 || timeout_seconds < 0) return;
    
    // Set socket timeout (simplified - would use setsockopt)
    // Full implementation would use setsockopt with SO_RCVTIMEO and SO_SNDTIMEO
}

void omni_server_set_max_request_size(int32_t max_size) {
    if (max_size > 0) {
        global_max_request_size = max_size;
    }
}

void omni_server_set_max_headers_size(int32_t max_size) {
    if (max_size > 0) {
        global_max_headers_size = max_size;
    }
}

// Server lifecycle
struct omni_server {
    int32_t port;
    int32_t socket;
    int32_t running;
    omni_connection_pool_t* connection_pool;
    omni_thread_pool_t* thread_pool;
};

omni_server_t* omni_server_create(int32_t port, omni_map_t* options) {
    (void)options;  // Options parameter is ignored for now
    if (port <= 0 || port > 65535) return NULL;
    
    omni_server_t* server = (omni_server_t*)malloc(sizeof(omni_server_t));
    if (!server) return NULL;
    
    server->port = port;
    server->socket = -1;
    server->running = 0;
    server->connection_pool = NULL;
    server->thread_pool = NULL;
    
    return server;
}

int32_t omni_server_listen(omni_server_t* server) {
    if (!server) return 0;
    
    // Create socket
    int32_t sock = omni_socket_create();
    if (sock < 0) return 0;
    
    // Bind to port
    if (!omni_socket_bind(sock, "0.0.0.0", server->port)) {
        omni_socket_close(sock);
        return 0;
    }
    
    // Listen
    if (!omni_socket_listen(sock, 128)) {
        omni_socket_close(sock);
        return 0;
    }
    
    server->socket = sock;
    server->running = 1;
    
    return 1; // Success
}

int32_t omni_server_listen_tls(omni_server_t* server, const char* cert_file, const char* key_file) {
    if (!server || !cert_file || !key_file) return 0;
    
    // TLS implementation would go here
    // For now, fall back to regular listen
    return omni_server_listen(server);
}

void omni_server_close(omni_server_t* server) {
    if (!server) return;
    
    if (server->socket >= 0) {
        omni_socket_close(server->socket);
        server->socket = -1;
    }
    
    server->running = 0;
    
    if (server->connection_pool) {
        free(server->connection_pool->sockets);
        free(server->connection_pool->in_use);
        free(server->connection_pool);
        server->connection_pool = NULL;
    }
    
    if (server->thread_pool) {
        free(server->thread_pool);
        server->thread_pool = NULL;
    }
}

void omni_server_graceful_shutdown(omni_server_t* server, int32_t timeout_seconds) {
    (void)timeout_seconds; // Unused for now
    if (!server) return;
    
    // Stop accepting new connections
    server->running = 0;
    
    // Wait for existing connections to complete (simplified)
    // Full implementation would track active connections and wait for them
    
    // Close server socket
    if (server->socket >= 0) {
        omni_socket_close(server->socket);
        server->socket = -1;
    }
}

// Session management (simplified implementations)
struct omni_session {
    char session_id[128];
    int32_t timeout_seconds;
    omni_map_t* data;
    int64_t created_at;
};

omni_session_t* omni_session_create(const char* session_id, int32_t timeout_seconds) {
    omni_session_t* session = (omni_session_t*)malloc(sizeof(omni_session_t));
    if (!session) return NULL;
    
    if (session_id) {
        strncpy(session->session_id, session_id, sizeof(session->session_id) - 1);
    } else {
        // Generate session ID (simplified)
        snprintf(session->session_id, sizeof(session->session_id), "session_%ld", time(NULL));
    }
    
    session->timeout_seconds = timeout_seconds;
    session->data = omni_map_create();
    session->created_at = time(NULL);
    
    return session;
}

const char* omni_session_get(omni_session_t* session, const char* key) {
    if (!session || !session->data || !key) return NULL;
    return omni_map_get_string_string(session->data, key);
}

void omni_session_set(omni_session_t* session, const char* key, const char* value) {
    if (!session || !session->data || !key || !value) return;
    omni_map_put_string_string(session->data, key, value);
}

void omni_session_destroy(omni_session_t* session) {
    if (!session) return;
    if (session->data) omni_map_destroy(session->data);
    free(session);
}

struct omni_session_store {
    char storage_type[64];
    omni_map_t* sessions; // Simplified - in-memory storage
};

omni_session_store_t* omni_session_store_create(const char* storage_type) {
    omni_session_store_t* store = (omni_session_store_t*)malloc(sizeof(omni_session_store_t));
    if (!store) return NULL;
    
    if (storage_type) {
        strncpy(store->storage_type, storage_type, sizeof(store->storage_type) - 1);
    } else {
        strcpy(store->storage_type, "memory");
    }
    
    store->sessions = omni_map_create();
    
    return store;
}

void omni_session_store_save(omni_session_store_t* store, omni_session_t* session) {
    if (!store || !session) return;
    // Simplified - would serialize session and store it
}

omni_session_t* omni_session_store_load(omni_session_store_t* store, const char* session_id) {
    if (!store || !session_id) return NULL;
    // Simplified - would load session from store
    return NULL;
}

// Authentication/Authorization (simplified implementations)
char* omni_auth_hash_password(const char* password, const char* salt) {
    (void)salt; // Unused for now
    if (!password) return NULL;
    // Simplified password hashing - in production, would use bcrypt/argon2
    // For now, return a simple hash
    size_t len = strlen(password);
    char* hash = (char*)malloc(65);
    if (!hash) return NULL;
    snprintf(hash, 65, "hash_%s_%zu", password, len);
    return hash;
}

int32_t omni_auth_verify_password(const char* password, const char* hash) {
    if (!password || !hash) return 0;
    // Simplified verification - in production, would use proper password verification
    char* computed = omni_auth_hash_password(password, NULL);
    if (!computed) return 0;
    int32_t result = (strcmp(computed, hash) == 0) ? 1 : 0;
    free(computed);
    return result;
}

char* omni_auth_generate_token(const char* user_id, const char* secret, int32_t expires_in) {
    (void)secret; // Unused for now
    if (!user_id) return NULL;
    // Simplified token generation - in production, would use JWT
    char* token = (char*)malloc(256);
    if (!token) return NULL;
    snprintf(token, 256, "token_%s_%ld_%d", user_id, time(NULL), expires_in);
    return token;
}

const char* omni_auth_verify_token(const char* token, const char* secret) {
    (void)secret; // Unused for now
    if (!token) return NULL;
    // Simplified token verification - in production, would verify JWT signature
    // For now, extract user_id from token (simplified format)
    if (strncmp(token, "token_", 6) == 0) {
        // Extract user_id (simplified)
        return token + 6; // Skip "token_"
    }
    return NULL;
}

int32_t omni_auth_check_permission(const char* user_id, const char* resource, const char* action) {
    if (!user_id || !resource || !action) return 0;
    // Simplified permission check - in production, would query permission database
    return 1; // Allow by default
}

// Rate limiting
struct omni_rate_limiter {
    int32_t max_requests;
    int32_t window_seconds;
    omni_map_t* requests; // Track requests per key
};

omni_rate_limiter_t* omni_rate_limit_create(int32_t max_requests, int32_t window_seconds) {
    if (max_requests <= 0 || window_seconds <= 0) return NULL;
    
    omni_rate_limiter_t* limiter = (omni_rate_limiter_t*)malloc(sizeof(omni_rate_limiter_t));
    if (!limiter) return NULL;
    
    limiter->max_requests = max_requests;
    limiter->window_seconds = window_seconds;
    limiter->requests = omni_map_create();
    
    return limiter;
}

int32_t omni_rate_limit_check(omni_rate_limiter_t* limiter, const char* key) {
    if (!limiter || !key) return 0;
    
    // Simplified rate limiting - in production, would track timestamps and count requests
    // For now, always allow
    return 1;
}

void omni_rate_limit_reset(omni_rate_limiter_t* limiter, const char* key) {
    if (!limiter || !key) return;
    // Reset rate limit for key
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

// Socket functions
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
#ifdef _WIN32
    // On Windows, check network adapters
    DWORD dwSize = 0;
    GetAdaptersInfo(NULL, &dwSize);
    if (dwSize == 0) return 0;
    
    PIP_ADAPTER_INFO pAdapterInfo = (IP_ADAPTER_INFO*)malloc(dwSize);
    if (!pAdapterInfo) return 0;
    
    DWORD dwStatus = GetAdaptersInfo(pAdapterInfo, &dwSize);
    int32_t connected = 0;
    
    if (dwStatus == ERROR_SUCCESS) {
        PIP_ADAPTER_INFO pAdapter = pAdapterInfo;
        while (pAdapter) {
            if (pAdapter->Type != MIB_IF_TYPE_LOOPBACK) {
                connected = 1;
                break;
            }
            pAdapter = pAdapter->Next;
        }
    }
    
    free(pAdapterInfo);
    return connected;
#else
    // On POSIX, use getifaddrs
    struct ifaddrs* ifaddr = NULL;
    if (getifaddrs(&ifaddr) != 0) {
        return 0;
    }
    
    int32_t connected = 0;
    for (struct ifaddrs* ifa = ifaddr; ifa != NULL; ifa = ifa->ifa_next) {
        if (ifa->ifa_addr == NULL) continue;
        
        // Skip loopback interfaces
        if (ifa->ifa_flags & IFF_LOOPBACK) continue;
        
        // Check if interface is up
        if (ifa->ifa_flags & IFF_UP) {
            // Check for IPv4 or IPv6 address
            if (ifa->ifa_addr->sa_family == AF_INET || ifa->ifa_addr->sa_family == AF_INET6) {
                connected = 1;
                break;
            }
        }
    }
    
    freeifaddrs(ifaddr);
    return connected;
#endif
}

omni_ip_address_t* omni_network_get_local_ip() {
#ifdef _WIN32
    // On Windows, use GetAdaptersInfo
    DWORD dwSize = 0;
    GetAdaptersInfo(NULL, &dwSize);
    if (dwSize == 0) {
        // Fallback to localhost
        omni_ip_address_t* ip = (omni_ip_address_t*)malloc(sizeof(omni_ip_address_t));
        if (!ip) return NULL;
        strncpy(ip->address, "127.0.0.1", sizeof(ip->address) - 1);
        ip->address[sizeof(ip->address) - 1] = '\0';
        ip->is_ipv4 = 1;
        ip->is_ipv6 = 0;
        return ip;
    }
    
    PIP_ADAPTER_INFO pAdapterInfo = (IP_ADAPTER_INFO*)malloc(dwSize);
    if (!pAdapterInfo) {
        omni_ip_address_t* ip = (omni_ip_address_t*)malloc(sizeof(omni_ip_address_t));
        if (!ip) return NULL;
        strncpy(ip->address, "127.0.0.1", sizeof(ip->address) - 1);
        ip->address[sizeof(ip->address) - 1] = '\0';
        ip->is_ipv4 = 1;
        ip->is_ipv6 = 0;
        return ip;
    }
    
    DWORD dwStatus = GetAdaptersInfo(pAdapterInfo, &dwSize);
    omni_ip_address_t* ip = NULL;
    
    if (dwStatus == ERROR_SUCCESS) {
        PIP_ADAPTER_INFO pAdapter = pAdapterInfo;
        while (pAdapter) {
            // Skip loopback adapter
            if (pAdapter->Type != MIB_IF_TYPE_LOOPBACK && pAdapter->IpAddressList.IpAddress.String) {
                ip = (omni_ip_address_t*)malloc(sizeof(omni_ip_address_t));
                if (ip) {
                    strncpy(ip->address, pAdapter->IpAddressList.IpAddress.String, sizeof(ip->address) - 1);
                    ip->address[sizeof(ip->address) - 1] = '\0';
                    ip->is_ipv4 = 1;
                    ip->is_ipv6 = 0;
                }
                break;
            }
            pAdapter = pAdapter->Next;
        }
    }
    
    free(pAdapterInfo);
    
    if (!ip) {
        // Fallback to localhost
        ip = (omni_ip_address_t*)malloc(sizeof(omni_ip_address_t));
        if (ip) {
            strncpy(ip->address, "127.0.0.1", sizeof(ip->address) - 1);
            ip->address[sizeof(ip->address) - 1] = '\0';
            ip->is_ipv4 = 1;
            ip->is_ipv6 = 0;
        }
    }
    
    return ip;
#else
    // On POSIX, use getifaddrs
    struct ifaddrs* ifaddr = NULL;
    if (getifaddrs(&ifaddr) != 0) {
        // Fallback to localhost
        omni_ip_address_t* ip = (omni_ip_address_t*)malloc(sizeof(omni_ip_address_t));
        if (!ip) return NULL;
        strncpy(ip->address, "127.0.0.1", sizeof(ip->address) - 1);
        ip->address[sizeof(ip->address) - 1] = '\0';
        ip->is_ipv4 = 1;
        ip->is_ipv6 = 0;
        return ip;
    }
    
    omni_ip_address_t* ip = NULL;
    for (struct ifaddrs* ifa = ifaddr; ifa != NULL; ifa = ifa->ifa_next) {
        if (ifa->ifa_addr == NULL) continue;
        
        // Skip loopback interfaces
        if (ifa->ifa_flags & IFF_LOOPBACK) continue;
        
        // Check if interface is up
        if (!(ifa->ifa_flags & IFF_UP)) continue;
        
        // Prefer IPv4 addresses
        if (ifa->ifa_addr->sa_family == AF_INET) {
            struct sockaddr_in* sin = (struct sockaddr_in*)ifa->ifa_addr;
            char ip_str[INET_ADDRSTRLEN];
            if (inet_ntop(AF_INET, &sin->sin_addr, ip_str, INET_ADDRSTRLEN)) {
                ip = (omni_ip_address_t*)malloc(sizeof(omni_ip_address_t));
                if (ip) {
                    strncpy(ip->address, ip_str, sizeof(ip->address) - 1);
                    ip->address[sizeof(ip->address) - 1] = '\0';
                    ip->is_ipv4 = 1;
                    ip->is_ipv6 = 0;
                    break;
                }
            }
        }
    }
    
    freeifaddrs(ifaddr);
    
    if (!ip) {
        // Fallback to localhost
        ip = (omni_ip_address_t*)malloc(sizeof(omni_ip_address_t));
        if (ip) {
            strncpy(ip->address, "127.0.0.1", sizeof(ip->address) - 1);
            ip->address[sizeof(ip->address) - 1] = '\0';
            ip->is_ipv4 = 1;
            ip->is_ipv6 = 0;
        }
    }
    
    return ip;
#endif
}

int32_t omni_network_ping(const char* host) {
    if (!host) return 0;
    
#ifdef _WIN32
    // On Windows, use IcmpSendEcho
    HANDLE hIcmpFile = IcmpCreateFile();
    if (hIcmpFile == INVALID_HANDLE_VALUE) {
        // Fallback: try TCP connection
        return omni_network_ping_tcp_fallback(host);
    }
    
    // Resolve hostname
    struct addrinfo hints, *result = NULL;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_INET;
    hints.ai_socktype = SOCK_STREAM;
    
    if (getaddrinfo(host, NULL, &hints, &result) != 0) {
        IcmpCloseHandle(hIcmpFile);
        return 0;
    }
    
    struct sockaddr_in* sin = (struct sockaddr_in*)result->ai_addr;
    IPAddr dest_ip = sin->sin_addr.S_un.S_addr;
    
    char send_data[32] = "Ping Data";
    char reply_buffer[sizeof(ICMP_ECHO_REPLY) + 32];
    DWORD reply_size = sizeof(reply_buffer);
    
    DWORD result_code = IcmpSendEcho(hIcmpFile, dest_ip, send_data, sizeof(send_data),
                                     NULL, reply_buffer, reply_size, 1000);
    
    freeaddrinfo(result);
    IcmpCloseHandle(hIcmpFile);
    
    return (result_code > 0) ? 1 : 0;
#else
    // On POSIX, try TCP connection as fallback (ICMP requires root)
    return omni_network_ping_tcp_fallback(host);
#endif
}

// TCP-based ping fallback (works without root privileges)
static int32_t omni_network_ping_tcp_fallback(const char* host) {
    // Try connecting to common ports (80, 443)
    int ports[] = {80, 443, 22};
    int num_ports = sizeof(ports) / sizeof(ports[0]);
    
    for (int i = 0; i < num_ports; i++) {
        struct addrinfo hints, *result = NULL;
        memset(&hints, 0, sizeof(hints));
        hints.ai_family = AF_INET;
        hints.ai_socktype = SOCK_STREAM;
        
        char port_str[16];
        snprintf(port_str, sizeof(port_str), "%d", ports[i]);
        
        if (getaddrinfo(host, port_str, &hints, &result) != 0) {
            continue;
        }
        
        int sock = socket(AF_INET, SOCK_STREAM, 0);
        if (sock < 0) {
            freeaddrinfo(result);
            continue;
        }
        
        // Set non-blocking for timeout
#ifdef _WIN32
        u_long mode = 1;
        ioctlsocket(sock, FIONBIO, &mode);
#else
        int flags = fcntl(sock, F_GETFL, 0);
        fcntl(sock, F_SETFL, flags | O_NONBLOCK);
#endif
        
        // Try to connect with timeout
        int connected = 0;
        if (connect(sock, result->ai_addr, result->ai_addrlen) == 0) {
            connected = 1;
        } else {
#ifdef _WIN32
            fd_set write_fds;
            struct timeval timeout;
            FD_ZERO(&write_fds);
            FD_SET(sock, &write_fds);
            timeout.tv_sec = 2;
            timeout.tv_usec = 0;
            if (select(0, NULL, &write_fds, NULL, &timeout) > 0) {
                connected = 1;
            }
#else
            fd_set write_fds;
            struct timeval timeout;
            FD_ZERO(&write_fds);
            FD_SET(sock, &write_fds);
            timeout.tv_sec = 2;
            timeout.tv_usec = 0;
            if (select(sock + 1, NULL, &write_fds, NULL, &timeout) > 0) {
                int so_error;
                socklen_t len = sizeof(so_error);
                getsockopt(sock, SOL_SOCKET, SO_ERROR, &so_error, &len);
                if (so_error == 0) {
                    connected = 1;
                }
            }
#endif
        }
        
#ifdef _WIN32
        closesocket(sock);
#else
        close(sock);
#endif
        freeaddrinfo(result);
        
        if (connected) {
            return 1;
        }
    }
    
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

// ---------------------------------------------------------------------------
// Deferred-call support (Go-style `defer`)
//
// Each function that uses `defer` declares a local omni_defer_frame_t and
// pushes a small per-site thunk + context pointer for every defer site. When
// the frame reaches its terminating defer.run, we walk the list in LIFO
// order and invoke each thunk, then free its context. The thunk is generated
// by the C backend alongside each defer site and knows the exact argument
// types to cast back out of the context struct.
// ---------------------------------------------------------------------------

void omni_defer_push(omni_defer_frame_t* frame, omni_defer_thunk_t thunk, void* ctx) {
    if (!frame || !thunk) {
        if (ctx) free(ctx);
        return;
    }
    omni_defer_node_t* node = (omni_defer_node_t*)malloc(sizeof(omni_defer_node_t));
    if (!node) {
        if (ctx) free(ctx);
        return;
    }
    node->thunk = thunk;
    node->ctx = ctx;
    // Prepend so walking forward yields LIFO order.
    node->next = frame->head;
    frame->head = node;
}

void omni_defer_run_all(omni_defer_frame_t* frame) {
    if (!frame) return;
    omni_defer_node_t* node = frame->head;
    frame->head = NULL;
    while (node) {
        omni_defer_node_t* next = node->next;
        // Thunks own their ctx and free it themselves — that keeps cleanup
        // responsibility next to the per-site code that allocated it.
        node->thunk(node->ctx);
        free(node);
        node = next;
    }
}

// ---------------------------------------------------------------------------
// Slice support (heap-allocated arrays with embedded length/cap header).
//
// Every slice allocation is laid out as:
//
//     [omni_slice_header_t][element 0][element 1]...[element cap-1]
//                          ^-- pointer returned to OmniLang code
//
// The pointer handed back points at the first ELEMENT, not the header. This
// preserves backwards compatibility with the existing T*-typed array surface
// (indexing, struct fields, function args/returns) — old code keeps reading
// `arr[i]` exactly as before. Helpers walk back by sizeof(header) when they
// need length/cap info.
// ---------------------------------------------------------------------------

#define OMNI_SLICE_MAGIC ((int64_t)0x4F4D4E494C535443LL) /* 'OMNILSTC' */

typedef struct {
    int64_t len;
    int64_t cap;
    int64_t elem_size;
    int64_t magic;
} omni_slice_header_t;

static inline omni_slice_header_t* omni_slice_header(void* slice) {
    if (!slice) return NULL;
    omni_slice_header_t* h = ((omni_slice_header_t*)slice) - 1;
    if (h->magic != OMNI_SLICE_MAGIC) return NULL;
    return h;
}

void* omni_slice_make(int64_t len, int64_t cap, int64_t elem_size) {
    if (cap < len) cap = len;
    if (cap == 0) cap = 1; // avoid zero-byte mallocs; len stays 0
    if (elem_size <= 0) elem_size = sizeof(void*);
    omni_slice_header_t* h = (omni_slice_header_t*)malloc(sizeof(omni_slice_header_t) + (size_t)cap * (size_t)elem_size);
    if (!h) return NULL;
    h->len = len;
    h->cap = cap;
    h->elem_size = elem_size;
    h->magic = OMNI_SLICE_MAGIC;
    void* data = (void*)(h + 1);
    // Zero the payload so reads of unfilled slots produce predictable values
    // rather than uninitialized memory. Costs a memset per allocation; small
    // price for the safety win and still O(n) which append amortizes anyway.
    memset(data, 0, (size_t)cap * (size_t)elem_size);
    return data;
}

int64_t omni_slice_len_real(void* slice) {
    omni_slice_header_t* h = omni_slice_header(slice);
    if (!h) return 0;
    return h->len;
}

int64_t omni_slice_cap(void* slice) {
    omni_slice_header_t* h = omni_slice_header(slice);
    if (!h) return 0;
    return h->cap;
}

// omni_slice_append copies one element pointed to by elem onto the end of
// slice, growing if necessary. Returns the (possibly new) data pointer.
// Doubles capacity on grow — same as Go's amortized strategy.
void* omni_slice_append(void* slice, const void* elem) {
    omni_slice_header_t* h = omni_slice_header(slice);
    if (!h) {
        // Defensive: a non-slice input can't be safely grown. Return NULL so
        // the caller crashes loudly rather than corrupting an arbitrary
        // buffer. In practice this only fires if codegen mixes raw arrays
        // with the heap-allocated slice machinery — a bug to find, not silently work around.
        return NULL;
    }
    int64_t elem_size = h->elem_size;
    if (h->len < h->cap) {
        // Fast path: no reallocation needed.
        memcpy((char*)slice + h->len * elem_size, elem, (size_t)elem_size);
        h->len++;
        return slice;
    }
    int64_t new_cap = h->cap * 2;
    if (new_cap < h->len + 1) new_cap = h->len + 1;
    omni_slice_header_t* nh = (omni_slice_header_t*)malloc(sizeof(omni_slice_header_t) + (size_t)new_cap * (size_t)elem_size);
    if (!nh) return NULL;
    nh->len = h->len + 1;
    nh->cap = new_cap;
    nh->elem_size = elem_size;
    nh->magic = OMNI_SLICE_MAGIC;
    void* new_data = (void*)(nh + 1);
    memcpy(new_data, slice, (size_t)h->len * (size_t)elem_size);
    memcpy((char*)new_data + h->len * elem_size, elem, (size_t)elem_size);
    free(h);
    return new_data;
}

// omni_slice_subslice produces a fresh allocation for slice[lo:hi]. We copy
// rather than aliasing so a later append on either slice can't disturb the
// other — the simplest semantics; matches what the VM does. Use -1 for hi to
// mean "len(slice)".
void* omni_slice_subslice(void* slice, int64_t lo, int64_t hi) {
    omni_slice_header_t* h = omni_slice_header(slice);
    if (!h) return NULL;
    if (hi < 0) hi = h->len;
    if (lo < 0 || hi < lo || hi > h->len) {
        fprintf(stderr, "omni_slice_subslice: out of bounds [%lld:%lld] for length %lld\n",
                (long long)lo, (long long)hi, (long long)h->len);
        return NULL;
    }
    int64_t n = hi - lo;
    void* out = omni_slice_make(n, n, h->elem_size);
    if (out && n > 0) {
        memcpy(out, (char*)slice + lo * h->elem_size, (size_t)n * (size_t)h->elem_size);
    }
    return out;
}

// ---------------------------------------------------------------------------
// Channel and spawn support (Phase 5)
//
// Channels are bounded ring buffers protected by a mutex with two condvars
// (not_empty / not_full). Element bytes are stored inline in the buffer so we
// don't pay a heap allocation per send. Capacity-0 channels (Go semantics:
// "unbuffered" — synchronous handoff) use cap=1 internally and rely on
// senders/receivers serializing through the mutex; not exact Go semantics
// (a real unbuffered channel rendezvous), but close enough for early
// programs and matches what the VM does today.
//
// Spawn uses pthread_create directly. The C backend emits one detached
// thunk per spawn site that knows how to unpack the snapshotted args from a
// heap context — same pattern as the defer codegen.
// ---------------------------------------------------------------------------

// Definition of the struct forward-declared as `omni_chan_t` in omni_rt.h.
struct omni_chan {
    pthread_mutex_t mu;
    pthread_cond_t  not_empty;
    pthread_cond_t  not_full;
    char*           buf;       // cap * elem_size bytes
    int64_t         cap;
    int64_t         len;
    int64_t         head;      // dequeue position
    int64_t         tail;      // enqueue position
    int64_t         elem_size;
    int             closed;
};

omni_chan_t* omni_chan_make(int64_t cap, int64_t elem_size) {
    if (cap <= 0) cap = 1; // see header comment
    if (elem_size <= 0) elem_size = sizeof(void*);
    omni_chan_t* ch = (omni_chan_t*)malloc(sizeof(omni_chan_t));
    if (!ch) return NULL;
    ch->buf = (char*)malloc((size_t)cap * (size_t)elem_size);
    if (!ch->buf) { free(ch); return NULL; }
    pthread_mutex_init(&ch->mu, NULL);
    pthread_cond_init(&ch->not_empty, NULL);
    pthread_cond_init(&ch->not_full, NULL);
    ch->cap = cap;
    ch->len = 0;
    ch->head = 0;
    ch->tail = 0;
    ch->elem_size = elem_size;
    ch->closed = 0;
    return ch;
}

void omni_chan_send(omni_chan_t* ch, const void* elem) {
    if (!ch) return;
    pthread_mutex_lock(&ch->mu);
    while (ch->len == ch->cap && !ch->closed) {
        pthread_cond_wait(&ch->not_full, &ch->mu);
    }
    if (ch->closed) {
        pthread_mutex_unlock(&ch->mu);
        fprintf(stderr, "omni_chan_send: send on closed channel\n");
        return;
    }
    memcpy(ch->buf + ch->tail * ch->elem_size, elem, (size_t)ch->elem_size);
    ch->tail = (ch->tail + 1) % ch->cap;
    ch->len++;
    pthread_cond_signal(&ch->not_empty);
    pthread_mutex_unlock(&ch->mu);
}

void omni_chan_recv(omni_chan_t* ch, void* out) {
    if (!ch) return;
    pthread_mutex_lock(&ch->mu);
    while (ch->len == 0 && !ch->closed) {
        pthread_cond_wait(&ch->not_empty, &ch->mu);
    }
    if (ch->len == 0 && ch->closed) {
        // Match Go's "zero value on closed empty channel" behavior.
        memset(out, 0, (size_t)ch->elem_size);
        pthread_mutex_unlock(&ch->mu);
        return;
    }
    memcpy(out, ch->buf + ch->head * ch->elem_size, (size_t)ch->elem_size);
    ch->head = (ch->head + 1) % ch->cap;
    ch->len--;
    pthread_cond_signal(&ch->not_full);
    pthread_mutex_unlock(&ch->mu);
}

void omni_chan_close(omni_chan_t* ch) {
    if (!ch) return;
    pthread_mutex_lock(&ch->mu);
    ch->closed = 1;
    pthread_cond_broadcast(&ch->not_empty);
    pthread_cond_broadcast(&ch->not_full);
    pthread_mutex_unlock(&ch->mu);
}

// omni_chan_recv_ok: the ok-form. Writes the received element to *out and
// stores 1 in *ok if a real value was delivered, or (zero, 0) if the
// channel is closed and drained. Matches Go's `v, ok := <-c` semantics.
void omni_chan_recv_ok(omni_chan_t* ch, void* out, int32_t* ok) {
    if (!ch) {
        if (ok) *ok = 0;
        return;
    }
    pthread_mutex_lock(&ch->mu);
    while (ch->len == 0 && !ch->closed) {
        pthread_cond_wait(&ch->not_empty, &ch->mu);
    }
    if (ch->len == 0 && ch->closed) {
        memset(out, 0, (size_t)ch->elem_size);
        if (ok) *ok = 0;
        pthread_mutex_unlock(&ch->mu);
        return;
    }
    memcpy(out, ch->buf + ch->head * ch->elem_size, (size_t)ch->elem_size);
    ch->head = (ch->head + 1) % ch->cap;
    ch->len--;
    if (ok) *ok = 1;
    pthread_cond_signal(&ch->not_full);
    pthread_mutex_unlock(&ch->mu);
}

void omni_chan_destroy(omni_chan_t* ch) {
    if (!ch) return;
    pthread_mutex_destroy(&ch->mu);
    pthread_cond_destroy(&ch->not_empty);
    pthread_cond_destroy(&ch->not_full);
    free(ch->buf);
    free(ch);
}

// omni_spawn launches `thunk(ctx)` on a detached pthread. The thunk owns
// ctx (it must free it before returning). Returns 0 on success, errno on
// failure — the C codegen ignores the return today because there's no
// language surface for it; failures crash via a stderr message.
int omni_spawn(void* (*thunk)(void*), void* ctx) {
    pthread_t tid;
    pthread_attr_t attr;
    pthread_attr_init(&attr);
    pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_DETACHED);
    int rc = pthread_create(&tid, &attr, thunk, ctx);
    pthread_attr_destroy(&attr);
    if (rc != 0) {
        fprintf(stderr, "omni_spawn: pthread_create failed (errno %d)\n", rc);
        if (ctx) free(ctx);
    }
    return rc;
}
