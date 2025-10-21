#include "omni_rt.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int32_t omni_main() {
int32_t v1 = 1;
int32_t v2 = 2;
int32_t v3 = 3;
int32_t v4 = 4;
int32_t v5 = 5;
int32_t v0[] = {v1, v2, v3, v4, v5};
int32_t v6 = 0;
int32_t v7 = v0[v6];
int32_t v8 = 2;
int32_t v9 = v0[v8];
int32_t v11 = 10;
int32_t v12 = 20;
int32_t v13 = 30;
int32_t v14 = 40;
int32_t v15 = 50;
int32_t v10[] = {v11, v12, v13, v14, v15};
int32_t v16 = 0;
int32_t v17 = 0;
int32_t v18 = v10[v17];
int32_t v19 = v16 + v18;
v16 = v19;
int32_t v21 = 2;
int32_t v22 = v10[v21];
int32_t v23 = v16 + v22;
v16 = v23;
int32_t v25 = 4;
int32_t v26 = v10[v25];
int32_t v27 = v16 + v26;
v16 = v27;
int32_t v30 = 5;
int32_t v31 = 10;
int32_t v32 = 15;
int32_t v33 = 20;
int32_t v29[] = {v30, v31, v32, v33};
int32_t v34 = 0;
int32_t v35 = 0;
int32_t v36 = 4;
goto range_loop_header_0;
range_loop_header_0:
int32_t v37 = (v35 < v36) ? 1 : 0;
if (v37) {
goto range_loop_body_1;
} else {
goto range_loop_exit_2;
}
range_loop_body_1:
int32_t v38 = v29[v35];
int32_t v39 = v34 + v38;
v34 = v39;
int32_t v41 = v35 + 1;
v35 = v41;
goto range_loop_header_0;
range_loop_exit_2:
int32_t v43 = v7 + v9;
int32_t v44 = v43 + v16;
int32_t v45 = v44 + v34;
return v45;
}

int main() {
int32_t result = omni_main();
printf("OmniLang program result: %d\n", result);
return (int)result;
}
