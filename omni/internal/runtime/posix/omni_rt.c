#include <stdio.h>

void omni_print(const char* s) {
    if (s == NULL) {
        return;
    }
    fputs(s, stdout);
}
