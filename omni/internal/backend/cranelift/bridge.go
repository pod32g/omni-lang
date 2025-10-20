package cranelift

/*
#cgo LDFLAGS: -L${SRCDIR}/../../../native/clift/target/release -lomni_clift
#include <stdlib.h>
int omni_clift_compile_json(const char*);
int omni_clift_compile_to_object(const char*, const char*);
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// CompileMIRJSON delegates MIR emission to the native Cranelift bridge.
func CompileMIRJSON(json string) error {
	if json == "" {
		return fmt.Errorf("mir payload required")
	}
	c := C.CString(json)
	defer C.free(unsafe.Pointer(c))
	rc := C.omni_clift_compile_json(c)
	if rc != 0 {
		return fmt.Errorf("cranelift compile failed: %d", int(rc))
	}
	return nil
}

// CompileToObject compiles MIR to a native object file.
func CompileToObject(json string, outputPath string) error {
	if json == "" {
		return fmt.Errorf("mir payload required")
	}
	if outputPath == "" {
		return fmt.Errorf("output path required")
	}
	cJson := C.CString(json)
	defer C.free(unsafe.Pointer(cJson))
	cOutput := C.CString(outputPath)
	defer C.free(unsafe.Pointer(cOutput))
	rc := C.omni_clift_compile_to_object(cJson, cOutput)
	if rc != 0 {
		return fmt.Errorf("cranelift object compilation failed: %d", int(rc))
	}
	return nil
}
