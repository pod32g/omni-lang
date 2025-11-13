//go:build cgo
// +build cgo

package cbackend

/*
#cgo CFLAGS: -I${SRCDIR}/../../../runtime
#cgo linux  LDFLAGS: -lm
#cgo darwin LDFLAGS: -lm
#include <stdlib.h>
#include "omni_rt.h"
#include "omni_rt.c"
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
