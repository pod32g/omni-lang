//go:build cgo
// +build cgo

package cbackend

/*
#cgo CFLAGS: -I${SRCDIR}/../../runtime/include
#cgo linux  LDFLAGS: -L${SRCDIR}/../../runtime/posix -lomni_rt
#cgo darwin LDFLAGS: -L${SRCDIR}/../../runtime/posix -lomni_rt
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
	C.omni_print(cs)
	return nil
}
