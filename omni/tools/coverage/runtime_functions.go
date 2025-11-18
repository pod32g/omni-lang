package coverage

import (
	"strings"
)

// RuntimeFunction represents a function that is wired to the runtime
type RuntimeFunction struct {
	OmniLangName string // e.g., "std.io.print"
	RuntimeName  string // e.g., "omni_print_string"
	Module       string // e.g., "std.io"
	Function     string // e.g., "print"
}

// GetRuntimeWiredFunctions returns a map of all runtime-wired functions
// This is extracted from the C backend's hasRuntimeImplementation map
func GetRuntimeWiredFunctions() map[string]*RuntimeFunction {
	// This map is based on the runtimeImplMap in c_generator.go
	// We only include functions that are actually wired to runtime
	funcs := make(map[string]*RuntimeFunction)

	// I/O functions
	addFunction(funcs, "std.io.print", "omni_print_string", "std.io", "print")
	addFunction(funcs, "std.io.println", "omni_println_string", "std.io", "println")
	addFunction(funcs, "io.print", "omni_print_string", "std.io", "print")
	addFunction(funcs, "io.println", "omni_println_string", "std.io", "println")
	addFunction(funcs, "std.io.read_line", "omni_read_line", "std.io", "read_line")
	addFunction(funcs, "io.read_line", "omni_read_line", "std.io", "read_line")

	// String functions
	stringFuncs := []struct {
		omniName, runtimeName, funcName string
	}{
		{"std.string.length", "omni_strlen", "length"},
		{"std.string.concat", "omni_strcat", "concat"},
		{"std.string.substring", "omni_substring", "substring"},
		{"std.string.char_at", "omni_char_at", "char_at"},
		{"std.string.starts_with", "omni_starts_with", "starts_with"},
		{"std.string.ends_with", "omni_ends_with", "ends_with"},
		{"std.string.contains", "omni_contains", "contains"},
		{"std.string.index_of", "omni_index_of", "index_of"},
		{"std.string.last_index_of", "omni_last_index_of", "last_index_of"},
		{"std.string.trim", "omni_trim", "trim"},
		{"std.string.to_upper", "omni_to_upper", "to_upper"},
		{"std.string.to_lower", "omni_to_lower", "to_lower"},
		{"std.string.equals", "omni_string_equals", "equals"},
		{"std.string.compare", "omni_string_compare", "compare"},
	}

	for _, f := range stringFuncs {
		addFunction(funcs, f.omniName, f.runtimeName, "std.string", f.funcName)
		// Also add without std. prefix
		shortName := strings.TrimPrefix(f.omniName, "std.")
		addFunction(funcs, shortName, f.runtimeName, "std.string", f.funcName)
	}

	// Math functions
	mathFuncs := []struct {
		omniName, runtimeName, funcName string
	}{
		{"std.math.abs", "omni_abs", "abs"},
		{"std.math.max", "omni_max", "max"},
		{"std.math.min", "omni_min", "min"},
		{"std.math.pow", "omni_pow", "pow"},
		{"std.math.sqrt", "omni_sqrt", "sqrt"},
		{"std.math.floor", "omni_floor", "floor"},
		{"std.math.ceil", "omni_ceil", "ceil"},
		{"std.math.round", "omni_round", "round"},
		{"std.math.gcd", "omni_gcd", "gcd"},
		{"std.math.lcm", "omni_lcm", "lcm"},
	}

	for _, f := range mathFuncs {
		addFunction(funcs, f.omniName, f.runtimeName, "std.math", f.funcName)
	}

	// Array functions
	addFunction(funcs, "std.array.length", "omni_array_length", "std.array", "length")
	addFunction(funcs, "array.length", "omni_array_length", "std.array", "length")

	// File functions (runtime-wired ones)
	fileFuncs := []struct {
		omniName, runtimeName, funcName string
	}{
		{"std.file.open", "omni_file_open", "open"},
		{"std.file.close", "omni_file_close", "close"},
		{"std.file.read", "omni_file_read", "read"},
		{"std.file.write", "omni_file_write", "write"},
		{"std.file.seek", "omni_file_seek", "seek"},
		{"std.file.tell", "omni_file_tell", "tell"},
		{"std.file.exists", "omni_file_exists", "exists"},
		{"std.file.size", "omni_file_size", "size"},
	}

	for _, f := range fileFuncs {
		addFunction(funcs, f.omniName, f.runtimeName, "std.file", f.funcName)
	}

	// OS functions (runtime-wired ones)
	osFuncs := []struct {
		omniName, runtimeName, funcName string
	}{
		{"std.os.exit", "omni_exit", "exit"},
		{"std.os.getenv", "omni_getenv", "getenv"},
		{"std.os.setenv", "omni_setenv", "setenv"},
		{"std.os.remove", "omni_remove", "remove"},
	}

	for _, f := range osFuncs {
		addFunction(funcs, f.omniName, f.runtimeName, "std.os", f.funcName)
	}

	// Map/Collections functions (runtime-wired ones)
	// Note: Many collections functions are runtime-wired, but we'll focus on the main ones
	mapFuncs := []struct {
		omniName, runtimeName, funcName string
	}{
		{"std.collections.map.create", "omni_map_create", "create"},
		{"std.collections.map.destroy", "omni_map_destroy", "destroy"},
		{"std.collections.map.put", "omni_map_put_string_int", "put"}, // Simplified
		{"std.collections.map.get", "omni_map_get_string_int", "get"}, // Simplified
		{"std.collections.map.size", "omni_map_size", "size"},
	}

	for _, f := range mapFuncs {
		addFunction(funcs, f.omniName, f.runtimeName, "std.collections", f.funcName)
	}

	return funcs
}

func addFunction(funcs map[string]*RuntimeFunction, omniName, runtimeName, module, funcName string) {
	funcs[omniName] = &RuntimeFunction{
		OmniLangName: omniName,
		RuntimeName:  runtimeName,
		Module:       module,
		Function:     funcName,
	}
}

// IsRuntimeWired checks if a function name is wired to the runtime
func IsRuntimeWired(functionName string) bool {
	funcs := GetRuntimeWiredFunctions()
	_, exists := funcs[functionName]
	return exists
}

// GetRuntimeFunction returns the RuntimeFunction for a given OmniLang function name
func GetRuntimeFunction(functionName string) *RuntimeFunction {
	funcs := GetRuntimeWiredFunctions()
	return funcs[functionName]
}
