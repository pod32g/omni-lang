//go:build cgo
// +build cgo

package cbackend

/*
#cgo CFLAGS: -I/Users/pod32g/Documents/code/omni-lang/omni/runtime
#cgo linux  LDFLAGS: -L/Users/pod32g/Documents/code/omni-lang/omni/runtime/posix -lomni_rt -Wl,-rpath,/Users/pod32g/Documents/code/omni-lang/omni/runtime/posix
#cgo darwin LDFLAGS: -L/Users/pod32g/Documents/code/omni-lang/omni/runtime/posix -lomni_rt -Wl,-rpath,/Users/pod32g/Documents/code/omni-lang/omni/runtime/posix
#include <stdlib.h>
#include "omni_rt.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// Print proxies to the runtime printing primitive.
func Print(s string) error {
	if s == "" {
		return fmt.Errorf("print requires non-empty input")
	}
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	C.omni_print_string(cs)
	return nil
}
