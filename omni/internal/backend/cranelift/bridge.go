//go:build !darwin && !windows
// +build !darwin,!windows

package cranelift

/*
#cgo LDFLAGS: -L${SRCDIR}/../../../native/clift/target/release -lomni_clift
#include <stdlib.h>
int omni_clift_compile_json(const char*);
int omni_clift_compile_to_object(const char*, const char*);
int omni_clift_compile_to_object_with_opt(const char*, const char*, const char*);
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/omni-lang/omni/internal/mir"
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

// CompileModuleToObject compiles a MIR module to a native object file.
func CompileModuleToObject(module *mir.Module, outputPath string) error {
	// Convert MIR module to JSON
	jsonData, err := module.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert MIR to JSON: %w", err)
	}

	// Compile to object file
	return CompileToObject(string(jsonData), outputPath)
}

// CompileToObjectWithOpt compiles MIR to a native object file with optimization level.
func CompileToObjectWithOpt(json string, outputPath string, optLevel string) error {
	if json == "" {
		return fmt.Errorf("mir payload required")
	}
	if outputPath == "" {
		return fmt.Errorf("output path required")
	}
	if optLevel == "" {
		optLevel = "speed" // Default optimization level
	}
	cJson := C.CString(json)
	defer C.free(unsafe.Pointer(cJson))
	cOutput := C.CString(outputPath)
	defer C.free(unsafe.Pointer(cOutput))
	cOptLevel := C.CString(optLevel)
	defer C.free(unsafe.Pointer(cOptLevel))
	rc := C.omni_clift_compile_to_object_with_opt(cJson, cOutput, cOptLevel)
	if rc != 0 {
		return fmt.Errorf("cranelift object compilation failed: %d", int(rc))
	}
	return nil
}

// CompileModuleToObjectWithOpt compiles a MIR module to a native object file with optimization level.
func CompileModuleToObjectWithOpt(module *mir.Module, outputPath string, optLevel string) error {
	// Convert MIR module to JSON
	jsonData, err := module.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert MIR to JSON: %w", err)
	}

	// Compile to object file with optimization
	return CompileToObjectWithOpt(string(jsonData), outputPath, optLevel)
}
