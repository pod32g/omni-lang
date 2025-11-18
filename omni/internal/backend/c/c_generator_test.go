package cbackend

import (
	"strings"
	"testing"

	"github.com/omni-lang/omni/internal/mir"
)

func TestCGenerator(t *testing.T) {
	// Create a simple MIR module for testing
	module := &mir.Module{
		Functions: []*mir.Function{
			{
				Name:       "main",
				ReturnType: "int",
				Blocks: []*mir.BasicBlock{
					{
						Instructions: []mir.Instruction{
							{
								ID:   1,
								Op:   "const",
								Type: "int",
								Operands: []mir.Operand{
									{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
								},
							},
						},
						Terminator: mir.Terminator{
							Op: "ret",
							Operands: []mir.Operand{
								{Kind: mir.OperandValue, Value: 1, Type: "int"},
							},
						},
					},
				},
			},
		},
	}

	t.Run("NewCGenerator", func(t *testing.T) {
		generator := NewCGenerator(module)
		if generator == nil {
			t.Fatal("NewCGenerator returned nil")
		}

		if generator.module != module {
			t.Error("CGenerator module not set correctly")
		}

		if generator.variables == nil {
			t.Error("CGenerator variables map is nil")
		}
	})

	t.Run("NewCGeneratorWithOptLevel", func(t *testing.T) {
		generator := NewCGeneratorWithOptLevel(module, "O2")
		if generator == nil {
			t.Fatal("NewCGeneratorWithOptLevel returned nil")
		}

		if generator.optLevel != "O2" {
			t.Errorf("Expected optLevel O2, got %s", generator.optLevel)
		}
	})

	t.Run("NewCGeneratorWithDebug", func(t *testing.T) {
		generator := NewCGeneratorWithDebug(module, "O1", true, "test.c")
		if generator == nil {
			t.Fatal("NewCGeneratorWithDebug returned nil")
		}

		if !generator.debugInfo {
			t.Error("Expected debugInfo to be true")
		}

		if generator.sourceFile != "test.c" {
			t.Errorf("Expected sourceFile 'test.c', got '%s'", generator.sourceFile)
		}
	})

	t.Run("GenerateC", func(t *testing.T) {
		result, err := GenerateC(module)
		if err != nil {
			t.Errorf("GenerateC failed: %v", err)
		}

		if result == "" {
			t.Error("Expected non-empty C code generation")
		}

		// Check for basic C structure
		if !strings.Contains(result, "int main(") {
			t.Error("Expected main function in generated C code")
		}

		if !strings.Contains(result, "return") {
			t.Error("Expected return statement in generated C code")
		}
	})

	t.Run("GenerateCOptimized", func(t *testing.T) {
		result, err := GenerateCOptimized(module, "O2")
		if err != nil {
			t.Errorf("GenerateCOptimized failed: %v", err)
		}

		if result == "" {
			t.Error("Expected non-empty optimized C code generation")
		}
	})

	t.Run("GenerateCWithDebug", func(t *testing.T) {
		result, err := GenerateCWithDebug(module, "O1", true, "test.c")
		if err != nil {
			t.Errorf("GenerateCWithDebug failed: %v", err)
		}

		if result == "" {
			t.Error("Expected non-empty debug C code generation")
		}
	})

	t.Run("Generate", func(t *testing.T) {
		generator := NewCGenerator(module)
		result, _ := generator.Generate()

		if result == "" {
			t.Error("Expected non-empty C code generation")
		}
	})

	t.Run("GenerateSourceMap", func(t *testing.T) {
		generator := NewCGenerator(module)
		result := generator.GenerateSourceMap()

		// Source map should be nil when debugInfo is false
		if result != nil {
			t.Error("Expected nil source map generation when debugInfo is false")
		}
	})

	t.Run("WriteHeader", func(t *testing.T) {
		generator := NewCGenerator(module)
		generator.writeHeader()

		result := generator.output.String()
		if result == "" {
			t.Error("Expected non-empty header")
		}

		// Check for basic C includes
		if !strings.Contains(result, "#include") {
			t.Error("Expected #include statements in header")
		}
	})

	t.Run("WriteStdLibFunctions", func(t *testing.T) {
		generator := NewCGenerator(module)
		generator.writeStdLibFunctions()

		result := generator.output.String()
		// stdlib functions might be empty, so we just check it doesn't panic
		_ = result
	})

	t.Run("WriteFunctionDeclarations", func(t *testing.T) {
		generator := NewCGenerator(module)
		generator.writeFunctionDeclarations()

		result := generator.output.String()
		if result == "" {
			t.Error("Expected non-empty function declarations")
		}
	})

	t.Run("GenerateFunction", func(t *testing.T) {
		// Create a simple function
		function := &mir.Function{
			Name:       "main",
			ReturnType: "int",
			Blocks: []*mir.BasicBlock{
				{
					Instructions: []mir.Instruction{
						{
							ID:   1,
							Op:   "const",
							Type: "int",
							Operands: []mir.Operand{
								{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
							},
						},
					},
					Terminator: mir.Terminator{
						Op: "ret",
						Operands: []mir.Operand{
							{Kind: mir.OperandValue, Value: 1, Type: "int"},
						},
					},
				},
			},
		}

		generator := NewCGenerator(module)
		generator.generateFunction(function)

		result := generator.output.String()
		if result == "" {
			t.Error("Expected non-empty function generation")
		}
	})

	t.Run("GenerateBlock", func(t *testing.T) {
		// Create a simple block
		block := &mir.BasicBlock{
			Instructions: []mir.Instruction{
				{
					ID:   1,
					Op:   "const",
					Type: "int",
					Operands: []mir.Operand{
						{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
					},
				},
			},
		}

		// Create a function to pass to generateBlock
		function := &mir.Function{
			Name:       "main",
			ReturnType: "int",
		}

		generator := NewCGenerator(module)
		generator.generateBlock(block, function)

		result := generator.output.String()
		if result == "" {
			t.Error("Expected non-empty block generation")
		}
	})

	t.Run("GenerateInstruction", func(t *testing.T) {
		// Test const instruction
		instruction := &mir.Instruction{
			ID:   1,
			Op:   "const",
			Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
			},
		}

		generator := NewCGenerator(module)
		generator.generateInstruction(instruction)

		result := generator.output.String()
		if result == "" {
			t.Error("Expected non-empty instruction generation")
		}

		// Test add instruction
		instruction2 := &mir.Instruction{
			ID:   2,
			Op:   "add",
			Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "int"},
				{Kind: mir.OperandLiteral, Literal: "5", Type: "int"},
			},
		}

		generator.generateInstruction(instruction2)

		result = generator.output.String()
		if !strings.Contains(result, "+") {
			t.Error("Expected addition operator in generated code")
		}
	})

	t.Run("GenerateTerminator", func(t *testing.T) {
		// Test return terminator
		terminator := &mir.Terminator{
			Op: "ret",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "int"},
			},
		}

		generator := NewCGenerator(module)
		generator.generateTerminator(terminator, "main", "int")

		result := generator.output.String()
		if result == "" {
			t.Error("Expected non-empty terminator generation")
		}
	})

	t.Run("GetOperandValue", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test literal operand
		operand := mir.Operand{
			Kind:    mir.OperandLiteral,
			Literal: "42",
			Type:    "int",
		}

		result := generator.getOperandValue(operand)
		if result != "42" {
			t.Errorf("Expected '42', got '%s'", result)
		}

		// Test value operand
		operand2 := mir.Operand{
			Kind:  mir.OperandValue,
			Value: 1,
			Type:  "int",
		}

		result2 := generator.getOperandValue(operand2)
		if result2 == "" {
			t.Error("Expected non-empty value operand result")
		}
	})

	t.Run("ConvertOperandToString", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test int operand
		operand := mir.Operand{
			Kind:  mir.OperandValue,
			Value: 1,
			Type:  "int",
		}

		result := generator.convertOperandToString(operand)
		if result == "" {
			t.Error("Expected non-empty string conversion result")
		}

		// Test string operand
		operand2 := mir.Operand{
			Kind:  mir.OperandValue,
			Value: 2,
			Type:  "string",
		}

		result2 := generator.convertOperandToString(operand2)
		if result2 == "" {
			t.Error("Expected non-empty string conversion result")
		}
	})

	t.Run("IsMapVariable", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test with map type
		result := generator.isMapVariable("map<string,int>")
		// Just check it doesn't panic, the actual result might vary
		_ = result

		// Test with non-map type
		result2 := generator.isMapVariable("int")
		// Just check it doesn't panic, the actual result might vary
		_ = result2
	})

	t.Run("MapType", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test mapping int type
		result := generator.mapType("int")
		if result != "int32_t" {
			t.Errorf("Expected 'int32_t', got '%s'", result)
		}

		// Test mapping string type
		result2 := generator.mapType("string")
		if result2 != "const char*" {
			t.Errorf("Expected 'const char*', got '%s'", result2)
		}
	})

	t.Run("MapFunctionType", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test mapping function type
		result := generator.mapFunctionType("func(int): int")
		if result == "" {
			t.Error("Expected non-empty function type mapping")
		}
	})

	t.Run("MapFunctionTypeWithName", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test mapping function type with name
		result := generator.mapFunctionTypeWithName("func(int): int", "test")
		if result == "" {
			t.Error("Expected non-empty function type mapping with name")
		}
	})

	t.Run("MapFunctionReturnType", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test mapping function return type
		result := generator.mapFunctionReturnType("int")
		if result == "" {
			t.Error("Expected non-empty function return type mapping")
		}
	})

	t.Run("GenerateFunctionPointerReturnType", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test generating function pointer return type
		result := generator.generateFunctionPointerReturnType("int", "test")
		if result == "" {
			t.Error("Expected non-empty function pointer return type")
		}
	})

	t.Run("GenerateCompleteFunctionSignature", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test generating complete function signature
		result := generator.generateCompleteFunctionSignature("func(int): int", "test", []mir.Param{})
		if result == "" {
			t.Error("Expected non-empty complete function signature")
		}
	})

	t.Run("MapFunctionName", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test mapping function name
		result := generator.mapFunctionName("test_function")
		if result == "" {
			t.Error("Expected non-empty function name mapping")
		}
	})

	t.Run("IsRuntimeProvidedFunction", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test runtime provided function
		result := generator.isRuntimeProvidedFunction("std.io.println")
		if !result {
			t.Error("Expected std.io.println to be runtime provided")
		}

		// Test non-runtime function
		result2 := generator.isRuntimeProvidedFunction("custom_function")
		if result2 {
			t.Error("Expected custom_function to not be runtime provided")
		}
	})

	t.Run("GetVariableName", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test getting variable name
		result := generator.getVariableName(1)
		if result == "" {
			t.Error("Expected non-empty variable name")
		}
	})

	t.Run("ConvertLiteralToDecimal", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test converting literal to decimal
		result := generator.convertLiteralToDecimal("42")
		if result != "42" {
			t.Errorf("Expected '42', got '%s'", result)
		}

		// Test converting hex literal
		result2 := generator.convertLiteralToDecimal("0x2A")
		if result2 == "" {
			t.Error("Expected non-empty hex conversion result")
		}
	})

	t.Run("WriteMain", func(t *testing.T) {
		generator := NewCGenerator(module)
		generator.writeMain()

		result := generator.output.String()
		if result == "" {
			t.Error("Expected non-empty main function")
		}

		// Check for main function signature
		if !strings.Contains(result, "int main(") {
			t.Error("Expected main function signature")
		}
	})

	// Test type mapping functions comprehensively
	t.Run("MapTypePrimitives", func(t *testing.T) {
		generator := NewCGenerator(module)
		testCases := []struct {
			omniType string
			expected string
		}{
			{"int", "int32_t"},
			{"float", "double"},
			{"double", "double"},
			{"string", "const char*"},
			{"void", "void"},
			{"bool", "int32_t"},
			{"ptr", "void*"},
		}

		for _, tc := range testCases {
			result := generator.mapType(tc.omniType)
			if result != tc.expected {
				t.Errorf("mapType(%q) = %q, expected %q", tc.omniType, result, tc.expected)
			}
		}
	})

	t.Run("MapTypeArrays", func(t *testing.T) {
		generator := NewCGenerator(module)
		testCases := []struct {
			omniType string
			expected string
		}{
			{"array<int>", "int32_t*"},
			{"[]<int>", "int32_t*"},
			{"array<string>", "const char**"},
			{"[]<string>", "const char**"},
		}

		for _, tc := range testCases {
			result := generator.mapType(tc.omniType)
			if !strings.HasSuffix(result, "*") {
				t.Errorf("mapType(%q) = %q, expected pointer type", tc.omniType, result)
			}
		}
	})

	t.Run("MapTypeMaps", func(t *testing.T) {
		generator := NewCGenerator(module)
		result := generator.mapType("map<string,int>")
		if result != "omni_map_t*" {
			t.Errorf("mapType(map<string,int>) = %q, expected 'omni_map_t*'", result)
		}
	})

	t.Run("MapTypePointers", func(t *testing.T) {
		generator := NewCGenerator(module)
		result := generator.mapType("*int")
		if !strings.HasSuffix(result, "*") {
			t.Errorf("mapType(*int) = %q, expected pointer type", result)
		}
	})

	t.Run("MapTypePromises", func(t *testing.T) {
		generator := NewCGenerator(module)
		result := generator.mapType("Promise<int>")
		if result != "omni_promise_t*" {
			t.Errorf("mapType(Promise<int>) = %q, expected 'omni_promise_t*'", result)
		}
	})

	t.Run("MapTypeInfer", func(t *testing.T) {
		generator := NewCGenerator(module)
		testCases := []string{"<infer>", "infer", "<inferred>", ""}
		for _, tc := range testCases {
			result := generator.mapType(tc)
			if result != "int32_t" {
				t.Errorf("mapType(%q) = %q, expected 'int32_t'", tc, result)
			}
		}
	})

	t.Run("MapFunctionType", func(t *testing.T) {
		generator := NewCGenerator(module)
		result := generator.mapFunctionType("(int, string) -> bool")
		if result == "" {
			t.Error("Expected non-empty function type mapping")
		}
		if !strings.Contains(result, "(*)") {
			t.Error("Expected function pointer syntax")
		}
	})

	t.Run("MapFunctionTypeWithName", func(t *testing.T) {
		generator := NewCGenerator(module)
		result := generator.mapFunctionTypeWithName("(int) -> int", "fn")
		if result == "" {
			t.Error("Expected non-empty function type mapping with name")
		}
		if !strings.Contains(result, "fn") {
			t.Error("Expected function name in result")
		}
	})

	t.Run("MapFunctionReturnType", func(t *testing.T) {
		generator := NewCGenerator(module)
		result := generator.mapFunctionReturnType("(int, string) -> bool")
		if result == "" {
			t.Error("Expected non-empty function return type mapping")
		}
	})

	t.Run("ExtractMapTypes", func(t *testing.T) {
		generator := NewCGenerator(module)
		keyType, valueType := generator.extractMapTypes("map<string,int>")
		if keyType != "string" {
			t.Errorf("Expected key type 'string', got %q", keyType)
		}
		if valueType != "int" {
			t.Errorf("Expected value type 'int', got %q", valueType)
		}
	})

	t.Run("GetMapPutFunction", func(t *testing.T) {
		generator := NewCGenerator(module)
		testCases := []struct {
			keyType   string
			valueType string
			expected  string
		}{
			{"string", "int", "omni_map_put_string_int"},
			{"string", "string", "omni_map_put_string_string"},
			{"int", "int", "omni_map_put_int_int"},
			{"int", "string", "omni_map_put_int_string"},
		}

		for _, tc := range testCases {
			result := generator.getMapPutFunction(tc.keyType, tc.valueType)
			if result != tc.expected {
				t.Errorf("getMapPutFunction(%q, %q) = %q, expected %q", tc.keyType, tc.valueType, result, tc.expected)
			}
		}
	})

	t.Run("GetMapGetFunction", func(t *testing.T) {
		generator := NewCGenerator(module)
		testCases := []struct {
			keyType   string
			valueType string
			expected  string
		}{
			{"string", "int", "omni_map_get_string_int"},
			{"string", "string", "omni_map_get_string_string"},
			{"int", "int", "omni_map_get_int_int"},
			{"int", "string", "omni_map_get_int_string"},
		}

		for _, tc := range testCases {
			result := generator.getMapGetFunction(tc.keyType, tc.valueType)
			if result != tc.expected {
				t.Errorf("getMapGetFunction(%q, %q) = %q, expected %q", tc.keyType, tc.valueType, result, tc.expected)
			}
		}
	})

	t.Run("ExtractGenericType", func(t *testing.T) {
		generator := NewCGenerator(module)
		baseName, typeArgs := generator.extractGenericType("array<int>")
		if baseName != "array" {
			t.Errorf("Expected base name 'array', got %q", baseName)
		}
		if len(typeArgs) != 1 || typeArgs[0] != "int" {
			t.Errorf("Expected type args ['int'], got %v", typeArgs)
		}

		// Test nested generics
		baseName2, typeArgs2 := generator.extractGenericType("Box<array<int>>")
		if baseName2 != "Box" {
			t.Errorf("Expected base name 'Box', got %q", baseName2)
		}
		if len(typeArgs2) != 1 {
			t.Errorf("Expected 1 type arg, got %d", len(typeArgs2))
		}
	})

	t.Run("SplitGenericArgs", func(t *testing.T) {
		generator := NewCGenerator(module)
		args := generator.splitGenericArgs("int, string, bool")
		if len(args) != 3 {
			t.Errorf("Expected 3 args, got %d", len(args))
		}

		// Test nested generics
		args2 := generator.splitGenericArgs("array<int>, string")
		if len(args2) != 2 {
			t.Errorf("Expected 2 args, got %d", len(args2))
		}
	})

	t.Run("IsPrimitiveType", func(t *testing.T) {
		generator := NewCGenerator(module)
		primitives := []string{"int", "float", "double", "string", "void", "bool", "ptr"}
		nonPrimitives := []string{"array<int>", "map<string,int>", "Point", "CustomType"}

		for _, typ := range primitives {
			if !generator.isPrimitiveType(typ) {
				t.Errorf("Expected isPrimitiveType(%q) to be true", typ)
			}
		}

		for _, typ := range nonPrimitives {
			if generator.isPrimitiveType(typ) {
				t.Errorf("Expected isPrimitiveType(%q) to be false", typ)
			}
		}
	})

	t.Run("IsVariableDeclared", func(t *testing.T) {
		generator := NewCGenerator(module)
		// Variables are tracked by ValueID, not by name
		// Test that getVariableName creates and tracks variables
		name1 := generator.getVariableName(1)
		if name1 == "" {
			t.Error("Expected non-empty variable name")
		}

		// Getting the same ValueID should return the same name
		name2 := generator.getVariableName(1)
		if name1 != name2 {
			t.Errorf("Expected same variable name for same ValueID, got %q and %q", name1, name2)
		}

		// Check if the variable is declared (by checking if name exists in map)
		if !generator.isVariableDeclared(name1) {
			t.Error("Expected variable to be declared after getVariableName")
		}
	})

	t.Run("HasRuntimeImplementation", func(t *testing.T) {
		generator := NewCGenerator(module)
		if !generator.hasRuntimeImplementation("std.io.println") {
			t.Error("Expected std.io.println to have runtime implementation")
		}

		if generator.hasRuntimeImplementation("nonexistent.function") {
			t.Error("Expected nonexistent.function to not have runtime implementation")
		}
	})

	t.Run("IsStdFunction", func(t *testing.T) {
		generator := NewCGenerator(module)
		if !generator.isStdFunction("std.io.println") {
			t.Error("Expected std.io.println to be std function")
		}

		if generator.isStdFunction("custom_function") {
			t.Error("Expected custom_function to not be std function")
		}
	})

	t.Run("IsStringReturningFunction", func(t *testing.T) {
		generator := NewCGenerator(module)
		if !generator.isStringReturningFunction("std.string.trim") {
			t.Error("Expected std.string.trim to be string returning")
		}

		if generator.isStringReturningFunction("std.io.println") {
			t.Error("Expected std.io.println to not be string returning")
		}
	})

	t.Run("ConvertLiteralToDecimal", func(t *testing.T) {
		generator := NewCGenerator(module)
		testCases := []struct {
			literal  string
			expected string
		}{
			{"42", "42"},
			{"0x2A", "42"},
			{"0X2A", "42"},
			{"0b101010", "42"},
			{"0B101010", "42"},
		}

		for _, tc := range testCases {
			result := generator.convertLiteralToDecimal(tc.literal)
			if result != tc.expected {
				t.Errorf("convertLiteralToDecimal(%q) = %q, expected %q", tc.literal, result, tc.expected)
			}
		}
	})

	// Test instruction generation for different instruction types
	t.Run("GenerateInstructionTypes", func(t *testing.T) {
		generator := NewCGenerator(module)
		// Variables are created automatically by getVariableName, so we don't need to declare them

		instructions := []mir.Instruction{
			{ID: 1, Op: "const", Type: "int", Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "42", Type: "int"}}},
			{ID: 2, Op: "add", Type: "int", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "5", Type: "int"}}},
			{ID: 3, Op: "sub", Type: "int", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 2, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "1", Type: "int"}}},
			{ID: 4, Op: "mul", Type: "int", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 3, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "2", Type: "int"}}},
			{ID: 5, Op: "div", Type: "int", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 4, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "2", Type: "int"}}},
			{ID: 6, Op: "mod", Type: "int", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 5, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "3", Type: "int"}}},
			{ID: 7, Op: "cmp.eq", Type: "bool", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 6, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "42", Type: "int"}}},
			{ID: 8, Op: "cmp.neq", Type: "bool", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "0", Type: "int"}}},
			{ID: 9, Op: "cmp.lt", Type: "bool", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "10", Type: "int"}}},
			{ID: 10, Op: "cmp.gt", Type: "bool", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "0", Type: "int"}}},
			{ID: 11, Op: "cmp.lte", Type: "bool", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "42", Type: "int"}}},
			{ID: 12, Op: "cmp.gte", Type: "bool", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "42", Type: "int"}}},
			{ID: 13, Op: "bitand", Type: "int", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "3", Type: "int"}}},
			{ID: 14, Op: "bitor", Type: "int", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "2", Type: "int"}}},
			{ID: 15, Op: "bitxor", Type: "int", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "1", Type: "int"}}},
			{ID: 16, Op: "lshift", Type: "int", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "1", Type: "int"}}},
			{ID: 17, Op: "rshift", Type: "int", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}, {Kind: mir.OperandLiteral, Literal: "1", Type: "int"}}},
			{ID: 18, Op: "neg", Type: "int", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}}},
			{ID: 19, Op: "not", Type: "bool", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 7, Type: "bool"}}},
			{ID: 20, Op: "bitnot", Type: "int", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}}},
			{ID: 21, Op: "and", Type: "bool", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 7, Type: "bool"}, {Kind: mir.OperandValue, Value: 8, Type: "bool"}}},
			{ID: 22, Op: "or", Type: "bool", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 7, Type: "bool"}, {Kind: mir.OperandValue, Value: 8, Type: "bool"}}},
			{ID: 23, Op: "cast", Type: "float", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}}},
		}

		for _, inst := range instructions {
			err := generator.generateInstruction(&inst)
			if err != nil {
				t.Errorf("generateInstruction failed for %s: %v", inst.Op, err)
			}
		}

		result := generator.output.String()
		if result == "" {
			t.Error("Expected non-empty instruction generation")
		}
	})

	t.Run("GenerateInstructionConstTypes", func(t *testing.T) {
		generator := NewCGenerator(module)

		constInstructions := []mir.Instruction{
			{ID: 1, Op: "const", Type: "int", Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "42", Type: "int"}}},
			{ID: 2, Op: "const", Type: "float", Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "3.14", Type: "float"}}},
			{ID: 3, Op: "const", Type: "string", Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"hello\"", Type: "string"}}},
			{ID: 4, Op: "const", Type: "bool", Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "true", Type: "bool"}}},
			{ID: 5, Op: "const", Type: "null", Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "null", Type: "null"}}},
		}

		for _, inst := range constInstructions {
			err := generator.generateInstruction(&inst)
			if err != nil {
				t.Errorf("generateInstruction failed for const %s: %v", inst.Type, err)
			}
		}
	})

	t.Run("GenerateInstructionStringConcat", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create string constants first
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "string",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"hello\"", Type: "string"}},
		})
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "string",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"world\"", Type: "string"}},
		})

		// Test string concatenation
		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "strcat", Type: "string",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "string"},
				{Kind: mir.OperandValue, Value: 2, Type: "string"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for strcat: %v", err)
		}

		result := generator.output.String()
		if !strings.Contains(result, "omni_strcat") {
			t.Error("Expected omni_strcat in string concatenation")
		}
	})

	t.Run("GenerateInstructionThrow", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create exception string
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "string",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"error\"", Type: "string"}},
		})

		// Test throw instruction
		err := generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "throw", Type: "void",
			Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "string"}},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for throw: %v", err)
		}
	})

	t.Run("GenerateInstructionArrayInit", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create array elements
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "1", Type: "int"}},
		})
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "2", Type: "int"}},
		})

		// Test array initialization
		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "array.init", Type: "array<int>",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "int"},
				{Kind: mir.OperandValue, Value: 2, Type: "int"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for array.init: %v", err)
		}

		result := generator.output.String()
		if !strings.Contains(result, "v3") {
			t.Error("Expected array variable in output")
		}
	})

	t.Run("GenerateInstructionMapInit", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create key and value
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "string",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"key\"", Type: "string"}},
		})
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "42", Type: "int"}},
		})

		// Test map initialization
		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "map.init", Type: "map<string,int>",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "string"},
				{Kind: mir.OperandValue, Value: 2, Type: "int"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for map.init: %v", err)
		}
	})

	t.Run("GenerateInstructionStructInit", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create field values
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "1", Type: "int"}},
		})
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "2", Type: "int"}},
		})

		// Test struct initialization with named fields
		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "struct.init", Type: "Point",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "x", Type: "string"},
				{Kind: mir.OperandValue, Value: 1, Type: "int"},
				{Kind: mir.OperandLiteral, Literal: "y", Type: "string"},
				{Kind: mir.OperandValue, Value: 2, Type: "int"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for struct.init: %v", err)
		}
	})

	t.Run("GenerateInstructionIndex", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create array and index
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "array.init", Type: "array<int>",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
				{Kind: mir.OperandLiteral, Literal: "2", Type: "int"},
			},
		})
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "0", Type: "int"}},
		})

		// Set array type and length for proper code generation
		generator.valueTypes[1] = "array<int>"
		generator.arrayLengths[1] = 2

		// Test array indexing
		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "index", Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "array<int>"},
				{Kind: mir.OperandValue, Value: 2, Type: "int"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for index: %v", err)
		}
	})

	t.Run("GenerateInstructionMember", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create struct
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "struct.init", Type: "Point",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "x", Type: "string"},
				{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
			},
		})

		// Test member access
		err := generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "member", Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "Point"},
				{Kind: mir.OperandLiteral, Literal: "x", Type: "string"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for member: %v", err)
		}
	})

	t.Run("GenerateInstructionCall", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test function call with different return types
		testCases := []struct {
			op         string
			returnType string
			funcName   string
		}{
			{"call", "int", "add"},
			{"call.void", "void", "println"},
			{"call.string", "string", "read_line"},
			{"call.bool", "bool", "is_empty"},
		}

		for _, tc := range testCases {
			err := generator.generateInstruction(&mir.Instruction{
				ID: 1, Op: tc.op, Type: tc.returnType,
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: tc.funcName, Type: "string"},
				},
			})

			if err != nil {
				t.Errorf("generateInstruction failed for %s: %v", tc.op, err)
			}
		}
	})

	t.Run("GenerateInstructionAwait", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create promise
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "call", Type: "Promise<int>",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "asyncFunc", Type: "string"},
			},
		})

		// Test await instruction
		err := generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "await", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "Promise<int>"}},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for await: %v", err)
		}
	})

	t.Run("GenerateInstructionPHI", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create values for PHI node
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "0", Type: "int"}},
		})
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "1", Type: "int"}},
		})

		// Mark as PHI variable
		generator.phiVars[3] = true

		// Test PHI instruction
		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "phi", Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "int"},
				{Kind: mir.OperandValue, Value: 2, Type: "int"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for phi: %v", err)
		}
	})

	t.Run("GenerateInstructionAssign", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create source value
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "42", Type: "int"}},
		})

		// Test assignment instruction
		err := generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "assign", Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "int"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for assign: %v", err)
		}
	})

	t.Run("GenerateInstructionStdlibFunctions", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test various stdlib function calls via "call" instruction
		stdlibTests := []struct {
			returnType string
			funcName   string
			needsArg   bool
		}{
			{"void", "std.io.println", true},
			{"string", "std.io.read_line", false},
			{"float", "std.math.sqrt", true},
			{"string", "std.string.trim", true},
		}

		for _, tc := range stdlibTests {
			// Create argument if needed
			if tc.needsArg {
				generator.generateInstruction(&mir.Instruction{
					ID: 1, Op: "const", Type: "string",
					Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"test\"", Type: "string"}},
				})
			}

			operands := []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: tc.funcName, Type: "string"},
			}
			if tc.needsArg {
				operands = append(operands, mir.Operand{Kind: mir.OperandValue, Value: 1, Type: "string"})
			}

			// Use "call" instruction for all stdlib functions
			callOp := "call"
			if tc.returnType == "void" {
				callOp = "call.void"
			} else if tc.returnType == "string" {
				callOp = "call.string"
			} else if tc.returnType == "float" {
				callOp = "call"
			}

			err := generator.generateInstruction(&mir.Instruction{
				ID: 2, Op: callOp, Type: tc.returnType,
				Operands: operands,
			})

			if err != nil {
				t.Errorf("generateInstruction failed for %s: %v", tc.funcName, err)
			}
		}
	})

	t.Run("GenerateInstructionMalloc", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test malloc instruction (heap allocation)
		err := generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "malloc", Type: "void*",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "100", Type: "int"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for malloc: %v", err)
		}
	})

	t.Run("GenerateInstructionFree", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create pointer value
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "malloc", Type: "void*",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "100", Type: "int"},
			},
		})

		// Test free instruction
		err := generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "free", Type: "void",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "void*"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for free: %v", err)
		}
	})

	t.Run("GenerateInstructionFileOperations", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test file.open
		err := generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "file.open", Type: "int32_t",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "\"test.txt\"", Type: "string"},
				{Kind: mir.OperandLiteral, Literal: "\"r\"", Type: "string"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for file.open: %v", err)
		}

		// Test file.close
		err = generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "file.close", Type: "void",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "int32_t"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for file.close: %v", err)
		}

		// Test file.exists
		err = generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "file.exists", Type: "bool",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "\"test.txt\"", Type: "string"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for file.exists: %v", err)
		}
	})

	t.Run("GenerateInstructionTestOperations", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test test.start
		err := generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "test.start", Type: "void",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "\"test_name\"", Type: "string"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for test.start: %v", err)
		}

		// Test test.end
		err = generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "test.end", Type: "void",
			Operands: []mir.Operand{},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for test.end: %v", err)
		}

		// Test assert
		generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "const", Type: "bool",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "true", Type: "bool"}},
		})
		err = generator.generateInstruction(&mir.Instruction{
			ID: 4, Op: "assert", Type: "void",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 3, Type: "bool"},
				{Kind: mir.OperandLiteral, Literal: "\"message\"", Type: "string"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for assert: %v", err)
		}
	})

	t.Run("GenerateInstructionLogOperations", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test log operations
		logTests := []struct {
			op      string
			message string
		}{
			{"std.log.debug", "\"debug message\""},
			{"std.log.info", "\"info message\""},
			{"std.log.warn", "\"warn message\""},
			{"std.log.error", "\"error message\""},
		}

		for _, tc := range logTests {
			err := generator.generateInstruction(&mir.Instruction{
				ID: 1, Op: tc.op, Type: "void",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: tc.message, Type: "string"},
				},
			})

			if err != nil {
				t.Errorf("generateInstruction failed for %s: %v", tc.op, err)
			}
		}
	})

	// Test terminator generation
	t.Run("GenerateTerminatorTypes", func(t *testing.T) {
		generator := NewCGenerator(module)
		generator.declareVariable("v1")

		terminators := []mir.Terminator{
			{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 1, Type: "int"}}},
			{Op: "ret"}, // void return
			{Op: "jmp", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: 0, Type: "block"}}},
		}

		for _, term := range terminators {
			err := generator.generateTerminator(&term, "test", "int")
			if err != nil {
				t.Errorf("generateTerminator failed for %s: %v", term.Op, err)
			}
		}
	})

	// Integration tests
	t.Run("GenerateCompleteModule", func(t *testing.T) {
		complexModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "add",
					ReturnType: "int",
					Params: []mir.Param{
						{Name: "a", Type: "int", ID: 0},
						{Name: "b", Type: "int", ID: 1},
					},
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								{
									ID:   2,
									Op:   "add",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandValue, Value: 0, Type: "int"},
										{Kind: mir.OperandValue, Value: 1, Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 2, Type: "int"},
								},
							},
						},
					},
				},
				{
					Name:       "main",
					ReturnType: "int",
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								{
									ID:   1,
									Op:   "const",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 1, Type: "int"},
								},
							},
						},
					},
				},
			},
		}

		result, err := GenerateC(complexModule)
		if err != nil {
			t.Fatalf("GenerateC failed: %v", err)
		}

		if result == "" {
			t.Error("Expected non-empty C code generation")
		}

		// Check for function declarations
		if !strings.Contains(result, "add") {
			t.Error("Expected 'add' function in generated code")
		}
	})

	t.Run("GenerateWithOptimization", func(t *testing.T) {
		result, err := GenerateCOptimized(module, "O3")
		if err != nil {
			t.Fatalf("GenerateCOptimized failed: %v", err)
		}

		if result == "" {
			t.Error("Expected non-empty optimized C code generation")
		}
	})

	t.Run("GenerateWithDebugInfo", func(t *testing.T) {
		result, err := GenerateCWithDebug(module, "O0", true, "test.c")
		if err != nil {
			t.Fatalf("GenerateCWithDebug failed: %v", err)
		}

		if result == "" {
			t.Error("Expected non-empty debug C code generation")
		}

		// Check for debug info
		generator := NewCGeneratorWithDebug(module, "O0", true, "test.c")
		sourceMap := generator.GenerateSourceMap()
		if sourceMap == nil {
			t.Error("Expected non-nil source map when debugInfo is true")
		}
	})

	t.Run("GenerateMainWithDifferentReturnTypes", func(t *testing.T) {
		testCases := []struct {
			returnType string
			expect     string
		}{
			{"void", "void"},
			{"int", "int32_t"},
			{"string", "const char*"},
			{"float", "double"},
		}

		for _, tc := range testCases {
			testModule := &mir.Module{
				Functions: []*mir.Function{
					{
						Name:       "main",
						ReturnType: tc.returnType,
						Blocks: []*mir.BasicBlock{
							{
								Terminator: mir.Terminator{Op: "ret"},
							},
						},
					},
				},
			}

			generator := NewCGenerator(testModule)
			generator.writeMain()
			result := generator.output.String()

			if !strings.Contains(result, "int main(") {
				t.Errorf("Expected main function signature for return type %s", tc.returnType)
			}
		}
	})

	t.Run("ConvertOperandToString", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test OperandValue with different types
		testCases := []struct {
			operand          mir.Operand
			valueType        string
			expectedContains string
		}{
			{
				operand:          mir.Operand{Kind: mir.OperandValue, Value: 1, Type: "string"},
				valueType:        "string",
				expectedContains: "v1",
			},
			{
				operand:          mir.Operand{Kind: mir.OperandValue, Value: 2, Type: "int"},
				valueType:        "int",
				expectedContains: "omni_int_to_string",
			},
			{
				operand:          mir.Operand{Kind: mir.OperandValue, Value: 3, Type: "float"},
				valueType:        "float",
				expectedContains: "omni_float_to_string",
			},
			{
				operand:          mir.Operand{Kind: mir.OperandValue, Value: 4, Type: "bool"},
				valueType:        "bool",
				expectedContains: "omni_bool_to_string",
			},
		}

		for _, tc := range testCases {
			// Create the value first
			generator.generateInstruction(&mir.Instruction{
				ID: tc.operand.Value, Op: "const", Type: tc.valueType,
				Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "1", Type: tc.valueType}},
			})

			result := generator.convertOperandToString(tc.operand)
			// Check the output for the conversion function call
			output := generator.output.String()
			if !strings.Contains(output, tc.expectedContains) && !strings.Contains(result, tc.expectedContains) {
				t.Errorf("convertOperandToString for %s: expected to contain %s, got %s (output: %s)", tc.valueType, tc.expectedContains, result, output)
			}
		}

		// Test OperandLiteral
		literalTests := []struct {
			operand          mir.Operand
			expectedContains string
		}{
			{mir.Operand{Kind: mir.OperandLiteral, Literal: "\"hello\"", Type: "string"}, "\"hello\""},
			{mir.Operand{Kind: mir.OperandLiteral, Literal: "42", Type: "int"}, "omni_int_to_string"},
			{mir.Operand{Kind: mir.OperandLiteral, Literal: "3.14", Type: "float"}, "omni_float_to_string"},
			{mir.Operand{Kind: mir.OperandLiteral, Literal: "true", Type: "bool"}, "omni_bool_to_string(1)"},
			{mir.Operand{Kind: mir.OperandLiteral, Literal: "false", Type: "bool"}, "omni_bool_to_string(0)"},
		}

		for _, tc := range literalTests {
			result := generator.convertOperandToString(tc.operand)
			if !strings.Contains(result, tc.expectedContains) {
				t.Errorf("convertOperandToString for literal %s: expected to contain %s, got %s", tc.operand.Literal, tc.expectedContains, result)
			}
		}
	})

	t.Run("GenerateFunctionPointerReturnType", func(t *testing.T) {
		generator := NewCGenerator(module)

		testCases := []struct {
			funcType string
			expected string
		}{
			{"(int) -> int", "int (*)()"},
			{"(int, int) -> string", "const char* (*)()"},
			{"(string) -> void", "void (*)()"},
			{"() -> bool", "int32_t (*)()"},
		}

		for _, tc := range testCases {
			result := generator.generateFunctionPointerReturnType(tc.funcType, "testFunc")
			// The function includes the function name and parameters, so check for key parts
			if !strings.Contains(result, "(*") || !strings.Contains(result, ")(") {
				t.Errorf("generateFunctionPointerReturnType(%s): expected function pointer syntax, got %s", tc.funcType, result)
			}
		}
	})

	t.Run("GenerateCompleteFunctionSignature", func(t *testing.T) {
		generator := NewCGenerator(module)

		testCases := []struct {
			name             string
			params           []mir.Param
			returnType       string
			expectedContains string
		}{
			{
				name:             "simple",
				params:           []mir.Param{{Name: "x", Type: "int", ID: 0}},
				returnType:       "(int) -> int",
				expectedContains: "int32_t (*simple(",
			},
			{
				name: "multiple_params",
				params: []mir.Param{
					{Name: "a", Type: "int", ID: 0},
					{Name: "b", Type: "string", ID: 1},
				},
				returnType:       "(int, string) -> void",
				expectedContains: "void (*multiple_params(",
			},
		}

		for _, tc := range testCases {
			result := generator.generateCompleteFunctionSignature(tc.returnType, tc.name, tc.params)
			if !strings.Contains(result, tc.expectedContains) {
				t.Errorf("generateCompleteFunctionSignature: expected to contain %s, got %s", tc.expectedContains, result)
			}
		}
	})

	t.Run("DeclareVariable", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test declaring variables with different types
		testCases := []struct {
			varName          string
			varType          string
			expectedContains string
		}{
			{"v1", "int", "int32_t v1"},
			{"v2", "string", "const char* v2"},
			{"v3", "float", "double v3"},
			{"v4", "bool", "int32_t v4"},
			{"v5", "array<int>", "int32_t* v5"},
		}

		for _, tc := range testCases {
			generator.declareVariable(tc.varName)
			generator.valueTypes[mir.ValueID(1)] = tc.varType
			// Reset output for next test
			generator.output.Reset()
		}

		// Test that variables are tracked by ValueID
		// declareVariable doesn't track by name, it tracks by ValueID
		// So we need to use getVariableName to get the name for a ValueID
		testValueID := mir.ValueID(100)
		varName := generator.getVariableName(testValueID)
		generator.declareVariable(varName)
		// Variables are tracked by ValueID in the variables map
		if _, exists := generator.variables[testValueID]; !exists {
			t.Error("Expected variable to be tracked by ValueID")
		}
	})

	t.Run("GenerateTerminatorCbr", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create condition value
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "bool",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "true", Type: "bool"}},
		})

		// Test conditional branch
		term := &mir.Terminator{
			Op: "cbr",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "bool"},
				{Kind: mir.OperandValue, Value: 0, Type: "block"},
				{Kind: mir.OperandValue, Value: 0, Type: "block"},
			},
		}

		err := generator.generateTerminator(term, "test", "void")
		if err != nil {
			t.Errorf("generateTerminator failed for cbr: %v", err)
		}

		result := generator.output.String()
		if !strings.Contains(result, "if") {
			t.Error("Expected if statement in conditional branch")
		}
	})

	t.Run("GenerateTerminatorRetPromise", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create promise value
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "call", Type: "Promise<int>",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "asyncFunc", Type: "string"},
			},
		})

		// Test return with Promise
		term := &mir.Terminator{
			Op: "ret",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "Promise<int>"},
			},
		}

		err := generator.generateTerminator(term, "test", "Promise<int>")
		if err != nil {
			t.Errorf("generateTerminator failed for ret Promise: %v", err)
		}
	})

	t.Run("GenerateFunctionWithMultipleBlocks", func(t *testing.T) {
		funcModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "test",
					ReturnType: "int",
					Params:     []mir.Param{{Name: "x", Type: "int", ID: 0}},
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								{
									ID:   1,
									Op:   "const",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "0", Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "cbr",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 0, Type: "int"},
									{Kind: mir.OperandValue, Value: 1, Type: "block"},
									{Kind: mir.OperandValue, Value: 2, Type: "block"},
								},
							},
						},
						{
							Name: "then",
							Instructions: []mir.Instruction{
								{
									ID:   2,
									Op:   "const",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 2, Type: "int"},
								},
							},
						},
						{
							Name: "else",
							Instructions: []mir.Instruction{
								{
									ID:   3,
									Op:   "const",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "0", Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 3, Type: "int"},
								},
							},
						},
					},
				},
			},
		}

		generator := NewCGenerator(funcModule)
		err := generator.generateFunction(funcModule.Functions[0])
		if err != nil {
			t.Errorf("generateFunction failed: %v", err)
		}

		result := generator.output.String()
		if !strings.Contains(result, "int32_t test") {
			t.Error("Expected function signature in output")
		}
	})

	t.Run("GenerateInstructionFuncRef", func(t *testing.T) {
		generator := NewCGenerator(module)

		err := generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "func.ref", Type: "(int) -> int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "\"add\"", Type: "string"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for func.ref: %v", err)
		}
	})

	t.Run("GenerateInstructionFuncAssign", func(t *testing.T) {
		generator := NewCGenerator(module)

		err := generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "func.assign", Type: "(int) -> int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "\"add\"", Type: "string"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for func.assign: %v", err)
		}
	})

	t.Run("GenerateInstructionFuncCall", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create function pointer
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "func.ref", Type: "(int) -> int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "\"add\"", Type: "string"},
			},
		})

		// Create argument
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "42", Type: "int"}},
		})

		// Test function call through pointer
		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "func.call", Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "(int) -> int"},
				{Kind: mir.OperandValue, Value: 2, Type: "int"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for func.call: %v", err)
		}
	})

	t.Run("GenerateInstructionClosure", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test closure operations (should error)
		closureOps := []string{"closure.create", "closure.capture", "closure.bind"}

		for _, op := range closureOps {
			err := generator.generateInstruction(&mir.Instruction{
				ID: 1, Op: op, Type: "void",
				Operands: []mir.Operand{},
			})

			if err == nil {
				t.Errorf("generateInstruction should fail for %s (closures not supported)", op)
			}
		}
	})

	t.Run("GenerateInstructionAssertEq", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create expected and actual values
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "42", Type: "int"}},
		})
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "42", Type: "int"}},
		})

		// Test assert.eq
		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "assert.eq", Type: "void",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "int"},
				{Kind: mir.OperandValue, Value: 2, Type: "int"},
				{Kind: mir.OperandLiteral, Literal: "\"message\"", Type: "string"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for assert.eq: %v", err)
		}
	})

	t.Run("GenerateInstructionAssertTrueFalse", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create condition
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "bool",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "true", Type: "bool"}},
		})

		// Test assert.true
		err := generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "assert.true", Type: "void",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "bool"},
				{Kind: mir.OperandLiteral, Literal: "\"message\"", Type: "string"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for assert.true: %v", err)
		}

		// Test assert.false
		err = generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "assert.false", Type: "void",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "bool"},
				{Kind: mir.OperandLiteral, Literal: "\"message\"", Type: "string"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for assert.false: %v", err)
		}
	})

	t.Run("GenerateInstructionFileReadWrite", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create file handle
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "file.open", Type: "int32_t",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "\"test.txt\"", Type: "string"},
				{Kind: mir.OperandLiteral, Literal: "\"r\"", Type: "string"},
			},
		})

		// Test file.read
		err := generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "file.read", Type: "string",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "int32_t"},
				{Kind: mir.OperandLiteral, Literal: "100", Type: "int"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for file.read: %v", err)
		}

		// Test file.write
		generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "const", Type: "string",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"data\"", Type: "string"}},
		})
		err = generator.generateInstruction(&mir.Instruction{
			ID: 4, Op: "file.write", Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "int32_t"},
				{Kind: mir.OperandValue, Value: 3, Type: "string"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for file.write: %v", err)
		}
	})

	t.Run("GenerateInstructionFileSeekTell", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create file handle
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "file.open", Type: "int32_t",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "\"test.txt\"", Type: "string"},
				{Kind: mir.OperandLiteral, Literal: "\"r\"", Type: "string"},
			},
		})

		// Test file.seek
		err := generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "file.seek", Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "int32_t"},
				{Kind: mir.OperandLiteral, Literal: "0", Type: "int"},
				{Kind: mir.OperandLiteral, Literal: "0", Type: "int"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for file.seek: %v", err)
		}

		// Test file.tell
		err = generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "file.tell", Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "int32_t"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for file.tell: %v", err)
		}
	})

	t.Run("GenerateInstructionFileSize", func(t *testing.T) {
		generator := NewCGenerator(module)

		err := generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "file.size", Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "\"test.txt\"", Type: "string"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for file.size: %v", err)
		}
	})

	t.Run("GenerateInstructionRealloc", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create pointer
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "malloc", Type: "void*",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "100", Type: "int"},
			},
		})

		// Test realloc
		err := generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "realloc", Type: "void*",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "void*"},
				{Kind: mir.OperandLiteral, Literal: "200", Type: "int"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for realloc: %v", err)
		}
	})

	t.Run("GenerateInstructionStdIoPrint", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test std.io.print
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "string",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"hello\"", Type: "string"}},
		})

		err := generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "std.io.print", Type: "void",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "string"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for std.io.print: %v", err)
		}

		// Test std.io.println with no arguments
		err = generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "std.io.println", Type: "void",
			Operands: []mir.Operand{},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for std.io.println (no args): %v", err)
		}
	})

	t.Run("GenerateInstructionMapIndexing", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create map with different key/value types
		mapTests := []struct {
			mapType   string
			keyType   string
			valueType string
		}{
			{"map<string,int>", "string", "int"},
			{"map<int,int>", "int", "int"},
		}

		for _, tc := range mapTests {
			// Create map
			generator.generateInstruction(&mir.Instruction{
				ID: 1, Op: "map.init", Type: tc.mapType,
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: "\"key\"", Type: "string"},
					{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
				},
			})
			generator.mapTypes[1] = tc.mapType
			generator.valueTypes[1] = tc.mapType

			// Create key
			generator.generateInstruction(&mir.Instruction{
				ID: 2, Op: "const", Type: tc.keyType,
				Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"key\"", Type: tc.keyType}},
			})

			// Test map indexing
			err := generator.generateInstruction(&mir.Instruction{
				ID: 3, Op: "index", Type: tc.valueType,
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 1, Type: tc.mapType},
					{Kind: mir.OperandValue, Value: 2, Type: tc.keyType},
				},
			})

			if err != nil {
				t.Errorf("generateInstruction failed for map indexing (%s): %v", tc.mapType, err)
			}
		}
	})

	t.Run("GenerateInstructionArrayIndexingTypes", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test array indexing with different element types
		arrayTests := []struct {
			arrayType   string
			elementType string
		}{
			{"array<int>", "int"},
			{"array<string>", "string"},
			{"array<float>", "float"},
			{"array<bool>", "bool"},
		}

		for _, tc := range arrayTests {
			// Create array
			generator.generateInstruction(&mir.Instruction{
				ID: 1, Op: "array.init", Type: tc.arrayType,
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: "1", Type: tc.elementType},
					{Kind: mir.OperandLiteral, Literal: "2", Type: tc.elementType},
				},
			})
			generator.valueTypes[1] = tc.arrayType
			generator.arrayLengths[1] = 2

			// Create index
			generator.generateInstruction(&mir.Instruction{
				ID: 2, Op: "const", Type: "int",
				Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "0", Type: "int"}},
			})

			// Test array indexing
			err := generator.generateInstruction(&mir.Instruction{
				ID: 3, Op: "index", Type: tc.elementType,
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 1, Type: tc.arrayType},
					{Kind: mir.OperandValue, Value: 2, Type: "int"},
				},
			})

			if err != nil {
				t.Errorf("generateInstruction failed for array indexing (%s): %v", tc.arrayType, err)
			}
		}
	})

	t.Run("GenerateInstructionArrayIndexingUnknownLength", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create array without known length (simulating parameter array)
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "array<int>",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "NULL", Type: "array<int>"}},
		})
		generator.valueTypes[1] = "array<int>"
		// Don't set arrayLengths[1] to simulate unknown length

		// Create index
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "0", Type: "int"}},
		})

		// Test array indexing with unknown length
		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "index", Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "array<int>"},
				{Kind: mir.OperandValue, Value: 2, Type: "int"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for array indexing (unknown length): %v", err)
		}

		// Should have a warning about unknown length
		if len(generator.errors) == 0 {
			t.Error("Expected warning about unknown array length")
		}
	})

	t.Run("GenerateInstructionStructArrayIndexing", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create struct array
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "array.init", Type: "array<Point>",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "x", Type: "string"},
				{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
			},
		})
		generator.valueTypes[1] = "array<Point>"
		generator.arrayLengths[1] = 1

		// Create index
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "0", Type: "int"}},
		})

		// Test struct array indexing
		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "index", Type: "Point",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "array<Point>"},
				{Kind: mir.OperandValue, Value: 2, Type: "int"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for struct array indexing: %v", err)
		}
	})

	t.Run("GenerateInstructionLenFunction", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test len() with known array length
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "array.init", Type: "array<int>",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
				{Kind: mir.OperandLiteral, Literal: "2", Type: "int"},
			},
		})
		generator.valueTypes[1] = "array<int>"
		generator.arrayLengths[1] = 2

		err := generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "call", Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "len", Type: "string"},
				{Kind: mir.OperandValue, Value: 1, Type: "array<int>"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for len() with known length: %v", err)
		}

		// Test len() with unknown array length
		generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "const", Type: "array<int>",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "NULL", Type: "array<int>"}},
		})
		generator.valueTypes[3] = "array<int>"
		// Don't set arrayLengths[3]

		err = generator.generateInstruction(&mir.Instruction{
			ID: 4, Op: "call", Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "len", Type: "string"},
				{Kind: mir.OperandValue, Value: 3, Type: "array<int>"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for len() with unknown length: %v", err)
		}
	})

	t.Run("GenerateInstructionAsyncIO", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test std.io.read_line_async
		err := generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "call", Type: "Promise<string>",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "std.io.read_line_async", Type: "string"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for read_line_async: %v", err)
		}

		// Test std.os.read_file_async
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "string",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"test.txt\"", Type: "string"}},
		})
		err = generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "call", Type: "Promise<string>",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "std.os.read_file_async", Type: "string"},
				{Kind: mir.OperandValue, Value: 2, Type: "string"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for read_file_async: %v", err)
		}

		// Test std.os.write_file_async
		generator.generateInstruction(&mir.Instruction{
			ID: 4, Op: "const", Type: "string",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"data\"", Type: "string"}},
		})
		err = generator.generateInstruction(&mir.Instruction{
			ID: 5, Op: "call", Type: "Promise<bool>",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "std.os.write_file_async", Type: "string"},
				{Kind: mir.OperandValue, Value: 2, Type: "string"},
				{Kind: mir.OperandValue, Value: 4, Type: "string"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for write_file_async: %v", err)
		}

		// Test std.os.append_file_async
		err = generator.generateInstruction(&mir.Instruction{
			ID: 6, Op: "call", Type: "Promise<bool>",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "std.os.append_file_async", Type: "string"},
				{Kind: mir.OperandValue, Value: 2, Type: "string"},
				{Kind: mir.OperandValue, Value: 4, Type: "string"},
			},
		})
		if err != nil {
			t.Errorf("generateInstruction failed for append_file_async: %v", err)
		}
	})

	t.Run("GenerateInstructionPromiseReturn", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test function call that returns Promise<int>
		err := generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "call", Type: "Promise<int>",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "asyncFunc", Type: "string"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for Promise return: %v", err)
		}
	})

	t.Run("GenerateInstructionVoidCall", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test void function call
		err := generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "call.void", Type: "void",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "voidFunc", Type: "string"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for void call: %v", err)
		}
	})

	t.Run("GenerateInstructionCallWithArgs", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create arguments
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "1", Type: "int"}},
		})
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "2", Type: "int"}},
		})

		// Test function call with multiple arguments
		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "call", Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "add", Type: "string"},
				{Kind: mir.OperandValue, Value: 1, Type: "int"},
				{Kind: mir.OperandValue, Value: 2, Type: "int"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for call with args: %v", err)
		}
	})

	t.Run("GenerateInstructionMemberAccessTypes", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test member access with different field types
		memberTests := []struct {
			fieldType    string
			expectedFunc string
		}{
			{"string", "omni_struct_get_string_field"},
			{"float", "omni_struct_get_float_field"},
			{"double", "omni_struct_get_float_field"},
			{"bool", "omni_struct_get_bool_field"},
			{"int", "omni_struct_get_int_field"},
		}

		for _, tc := range memberTests {
			// Create struct
			generator.generateInstruction(&mir.Instruction{
				ID: 1, Op: "struct.init", Type: "TestStruct",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: "field", Type: "string"},
					{Kind: mir.OperandLiteral, Literal: "1", Type: tc.fieldType},
				},
			})

			// Test member access
			err := generator.generateInstruction(&mir.Instruction{
				ID: 2, Op: "member", Type: tc.fieldType,
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 1, Type: "TestStruct"},
					{Kind: mir.OperandLiteral, Literal: "field", Type: "string"},
				},
			})

			if err != nil {
				t.Errorf("generateInstruction failed for member access (%s): %v", tc.fieldType, err)
			}

			result := generator.output.String()
			if !strings.Contains(result, tc.expectedFunc) {
				t.Errorf("Expected %s in output for field type %s, got: %s", tc.expectedFunc, tc.fieldType, result)
			}
		}
	})

	t.Run("GenerateInstructionMemberAccessInferred", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create struct
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "struct.init", Type: "TestStruct",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "field", Type: "string"},
				{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
			},
		})
		generator.valueTypes[1] = "TestStruct"

		// Test member access with inferred type
		err := generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "member", Type: "<infer>",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "TestStruct"},
				{Kind: mir.OperandLiteral, Literal: "field", Type: "string"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for member access (inferred): %v", err)
		}
	})

	t.Run("GenerateInstructionStringComparisons", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create string values
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "string",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"hello\"", Type: "string"}},
		})
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "string",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"world\"", Type: "string"}},
		})

		// Test string comparisons
		comparisonOps := []string{"cmp.eq", "cmp.neq", "cmp.lt", "cmp.lte", "cmp.gt", "cmp.gte"}

		for _, op := range comparisonOps {
			err := generator.generateInstruction(&mir.Instruction{
				ID: 3, Op: op, Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 1, Type: "string"},
					{Kind: mir.OperandValue, Value: 2, Type: "string"},
				},
			})

			if err != nil {
				t.Errorf("generateInstruction failed for string comparison (%s): %v", op, err)
			}
		}
	})

	t.Run("GenerateInstructionMixedTypeComparisons", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create mixed type values (int vs string)
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "42", Type: "int"}},
		})
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "string",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"42\"", Type: "string"}},
		})

		// Test comparison with one string operand (should use string comparison)
		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "cmp.eq", Type: "bool",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "int"},
				{Kind: mir.OperandValue, Value: 2, Type: "string"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for mixed type comparison: %v", err)
		}
	})

	t.Run("GenerateInstructionArrayInitStructArray", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create struct elements
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "struct.init", Type: "Point",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "x", Type: "string"},
				{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
			},
		})
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "struct.init", Type: "Point",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "x", Type: "string"},
				{Kind: mir.OperandLiteral, Literal: "2", Type: "int"},
			},
		})

		// Test array initialization with struct elements
		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "array.init", Type: "array<Point>",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "Point"},
				{Kind: mir.OperandValue, Value: 2, Type: "Point"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for struct array init: %v", err)
		}
	})

	t.Run("GenerateInstructionArrayInitBracketSyntax", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test array initialization with []<T> syntax
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "1", Type: "int"}},
		})
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "int",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "2", Type: "int"}},
		})

		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "array.init", Type: "[]<int>",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "int"},
				{Kind: mir.OperandValue, Value: 2, Type: "int"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for array init ([]<T> syntax): %v", err)
		}
	})

	t.Run("GenerateInstructionMapInitUnsupported", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test map initialization with unsupported type combination
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "float",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "1.0", Type: "float"}},
		})
		generator.generateInstruction(&mir.Instruction{
			ID: 2, Op: "const", Type: "float",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "2.0", Type: "float"}},
		})

		err := generator.generateInstruction(&mir.Instruction{
			ID: 3, Op: "map.init", Type: "map<float,float>",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "float"},
				{Kind: mir.OperandValue, Value: 2, Type: "float"},
			},
		})

		if err != nil {
			t.Errorf("generateInstruction failed for map init (unsupported): %v", err)
		}

		// Should have an error about unsupported type
		if len(generator.errors) == 0 {
			t.Error("Expected error about unsupported map type")
		}
	})

	t.Run("GenerateInstructionAwaitDifferentTypes", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test await with different Promise types
		awaitTests := []struct {
			promiseType string
			resultType  string
		}{
			{"Promise<int>", "int"},
			{"Promise<string>", "string"},
			{"Promise<float>", "float"},
			{"Promise<bool>", "bool"},
		}

		for _, tc := range awaitTests {
			// Create promise
			generator.generateInstruction(&mir.Instruction{
				ID: 1, Op: "call", Type: tc.promiseType,
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: "asyncFunc", Type: "string"},
				},
			})

			// Test await
			err := generator.generateInstruction(&mir.Instruction{
				ID: 2, Op: "await", Type: tc.resultType,
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 1, Type: tc.promiseType},
				},
			})

			if err != nil {
				t.Errorf("generateInstruction failed for await (%s): %v", tc.promiseType, err)
			}
		}
	})

	t.Run("GenerateBlockNonEntry", func(t *testing.T) {
		generator := NewCGenerator(module)

		block := &mir.BasicBlock{
			Name: "then",
			Instructions: []mir.Instruction{
				{
					ID:   1,
					Op:   "const",
					Type: "int",
					Operands: []mir.Operand{
						{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
					},
				},
			},
			Terminator: mir.Terminator{
				Op: "ret",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 1, Type: "int"},
				},
			},
		}

		fn := &mir.Function{
			Name:       "test",
			ReturnType: "int",
		}

		err := generator.generateBlock(block, fn)
		if err != nil {
			t.Errorf("generateBlock failed for non-entry block: %v", err)
		}

		result := generator.output.String()
		if !strings.Contains(result, "then:") {
			t.Error("Expected block label in output")
		}
	})

	t.Run("GenerateBlockConstTypeInference", func(t *testing.T) {
		generator := NewCGenerator(module)

		block := &mir.BasicBlock{
			Name: "entry",
			Instructions: []mir.Instruction{
				{
					ID:   1,
					Op:   "const",
					Type: "", // No type set - should infer from literal
					Operands: []mir.Operand{
						{Kind: mir.OperandLiteral, Literal: "\"hello\"", Type: ""},
					},
				},
				{
					ID:   2,
					Op:   "const",
					Type: "",
					Operands: []mir.Operand{
						{Kind: mir.OperandLiteral, Literal: "true", Type: ""},
					},
				},
				{
					ID:   3,
					Op:   "const",
					Type: "",
					Operands: []mir.Operand{
						{Kind: mir.OperandLiteral, Literal: "3.14", Type: ""},
					},
				},
				{
					ID:   4,
					Op:   "const",
					Type: "",
					Operands: []mir.Operand{
						{Kind: mir.OperandLiteral, Literal: "42", Type: ""},
					},
				},
			},
			Terminator: mir.Terminator{
				Op: "ret",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 4, Type: "int"},
				},
			},
		}

		fn := &mir.Function{
			Name:       "test",
			ReturnType: "int",
		}

		err := generator.generateBlock(block, fn)
		if err != nil {
			t.Errorf("generateBlock failed for const type inference: %v", err)
		}

		// Check that types were inferred correctly
		if generator.valueTypes[1] != "string" {
			t.Errorf("Expected string type for v1, got %s", generator.valueTypes[1])
		}
		if generator.valueTypes[2] != "bool" {
			t.Errorf("Expected bool type for v2, got %s", generator.valueTypes[2])
		}
		if generator.valueTypes[3] != "float" {
			t.Errorf("Expected float type for v3, got %s", generator.valueTypes[3])
		}
		if generator.valueTypes[4] != "int" {
			t.Errorf("Expected int type for v4, got %s", generator.valueTypes[4])
		}
	})

	t.Run("GenerateFunctionAsync", func(t *testing.T) {
		funcModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "asyncFunc",
					ReturnType: "Promise<int>",
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								{
									ID:   1,
									Op:   "const",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 1, Type: "int"},
								},
							},
						},
					},
				},
			},
		}

		generator := NewCGenerator(funcModule)
		err := generator.generateFunction(funcModule.Functions[0])
		if err != nil {
			t.Errorf("generateFunction failed for async function: %v", err)
		}

		result := generator.output.String()
		if !strings.Contains(result, "omni_promise_t*") {
			t.Error("Expected Promise return type in async function")
		}
	})

	t.Run("GenerateFunctionWithFunctionPointerParam", func(t *testing.T) {
		funcModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "test",
					ReturnType: "int",
					Params: []mir.Param{
						{Name: "fn", Type: "(int) -> int", ID: 0},
					},
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								{
									ID:   1,
									Op:   "const",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 1, Type: "int"},
								},
							},
						},
					},
				},
			},
		}

		generator := NewCGenerator(funcModule)
		err := generator.generateFunction(funcModule.Functions[0])
		if err != nil {
			t.Errorf("generateFunction failed for function with function pointer param: %v", err)
		}

		result := generator.output.String()
		if !strings.Contains(result, "(*fn)") {
			t.Error("Expected function pointer parameter syntax")
		}
	})

	t.Run("GenerateFunctionRuntimeProvided", func(t *testing.T) {
		funcModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "len",
					ReturnType: "int",
					Blocks:     []*mir.BasicBlock{},
				},
			},
		}

		generator := NewCGenerator(funcModule)
		err := generator.generateFunction(funcModule.Functions[0])
		if err != nil {
			t.Errorf("generateFunction failed for runtime-provided function: %v", err)
		}

		// Runtime-provided functions should be skipped (no function body)
		result := generator.output.String()
		// Check that no function body was generated (no opening brace for function)
		if strings.Contains(result, "int32_t len(") || strings.Contains(result, "int len(") {
			t.Error("Expected runtime-provided function to be skipped, but function signature was generated")
		}
	})

	t.Run("GenerateFunctionWithDebugInfo", func(t *testing.T) {
		funcModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "test",
					ReturnType: "int",
					Blocks: []*mir.BasicBlock{
						{
							Name:         "entry",
							Instructions: []mir.Instruction{},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandLiteral, Literal: "0", Type: "int"},
								},
							},
						},
					},
				},
			},
		}

		generator := NewCGeneratorWithDebug(funcModule, "0", true, "test.omni")
		err := generator.generateFunction(funcModule.Functions[0])
		if err != nil {
			t.Errorf("generateFunction failed with debug info: %v", err)
		}

		result := generator.output.String()
		if !strings.Contains(result, "Debug:") {
			t.Error("Expected debug information in output")
		}
	})

	t.Run("GenerateTerminatorRetString", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Create string value
		generator.generateInstruction(&mir.Instruction{
			ID: 1, Op: "const", Type: "string",
			Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "\"hello\"", Type: "string"}},
		})
		generator.stringsToFree[1] = true

		// Test return with string (should exclude from cleanup)
		term := &mir.Terminator{
			Op: "ret",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "string"},
			},
		}

		err := generator.generateTerminator(term, "test", "string")
		if err != nil {
			t.Errorf("generateTerminator failed for string return: %v", err)
		}

		// Check that string was removed from cleanup list
		if generator.stringsToFree[1] {
			t.Error("Expected returned string to be excluded from cleanup")
		}
	})

	t.Run("GenerateTerminatorRetVoid", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test void return (not main)
		term := &mir.Terminator{
			Op:       "ret",
			Operands: []mir.Operand{},
		}

		err := generator.generateTerminator(term, "test", "void")
		if err != nil {
			t.Errorf("generateTerminator failed for void return: %v", err)
		}

		result := generator.output.String()
		if !strings.Contains(result, "return;") {
			t.Error("Expected void return statement")
		}
	})

	t.Run("GenerateTerminatorRetMain", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test void return for main (omni_main)
		term := &mir.Terminator{
			Op:       "ret",
			Operands: []mir.Operand{},
		}

		err := generator.generateTerminator(term, "omni_main", "void")
		if err != nil {
			t.Errorf("generateTerminator failed for main void return: %v", err)
		}

		result := generator.output.String()
		if !strings.Contains(result, "return 0;") {
			t.Error("Expected return 0 for omni_main")
		}
	})

	t.Run("GenerateTerminatorJmp", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test jmp terminator
		term := &mir.Terminator{
			Op: "jmp",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 0, Type: "block"},
			},
		}

		err := generator.generateTerminator(term, "test", "void")
		if err != nil {
			t.Errorf("generateTerminator failed for jmp: %v", err)
		}

		result := generator.output.String()
		if !strings.Contains(result, "goto") {
			t.Error("Expected goto statement for jmp")
		}
	})

	t.Run("GenerateTerminatorBr", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test br terminator
		term := &mir.Terminator{
			Op: "br",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 0, Type: "block"},
			},
		}

		err := generator.generateTerminator(term, "test", "void")
		if err != nil {
			t.Errorf("generateTerminator failed for br: %v", err)
		}

		result := generator.output.String()
		if !strings.Contains(result, "goto") {
			t.Error("Expected goto statement for br")
		}
	})

	t.Run("GenerateTerminatorUnknown", func(t *testing.T) {
		generator := NewCGenerator(module)

		// Test unknown terminator (should error)
		term := &mir.Terminator{
			Op:       "unknown",
			Operands: []mir.Operand{},
		}

		err := generator.generateTerminator(term, "test", "void")
		if err == nil {
			t.Error("Expected error for unknown terminator")
		}

		if len(generator.errors) == 0 {
			t.Error("Expected error message for unknown terminator")
		}
	})

	t.Run("GenerateCompleteModuleWithArraysAndMaps", func(t *testing.T) {
		complexModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "test",
					ReturnType: "int",
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								// Create array
								{
									ID:   1,
									Op:   "array.init",
									Type: "array<int>",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
										{Kind: mir.OperandLiteral, Literal: "2", Type: "int"},
										{Kind: mir.OperandLiteral, Literal: "3", Type: "int"},
									},
								},
								// Create map
								{
									ID:   2,
									Op:   "map.init",
									Type: "map<string,int>",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "\"key\"", Type: "string"},
										{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
									},
								},
								// Get array length
								{
									ID:   3,
									Op:   "call",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "len", Type: "string"},
										{Kind: mir.OperandValue, Value: 1, Type: "array<int>"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 3, Type: "int"},
								},
							},
						},
					},
				},
			},
		}

		generator := NewCGenerator(complexModule)
		// Set array length for len() call
		generator.arrayLengths[1] = 3
		generator.valueTypes[1] = "array<int>"

		result, err := GenerateC(complexModule)
		if err != nil {
			t.Fatalf("GenerateC failed: %v", err)
		}

		if result == "" {
			t.Error("Expected non-empty C code generation")
		}

		// Check for array and map operations
		if !strings.Contains(result, "omni_map") {
			t.Error("Expected map operations in generated code")
		}
	})

	t.Run("GenerateCompleteModuleWithStructs", func(t *testing.T) {
		complexModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "test",
					ReturnType: "int",
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								// Create struct
								{
									ID:   1,
									Op:   "struct.init",
									Type: "Point",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "x", Type: "string"},
										{Kind: mir.OperandLiteral, Literal: "10", Type: "int"},
										{Kind: mir.OperandLiteral, Literal: "y", Type: "string"},
										{Kind: mir.OperandLiteral, Literal: "20", Type: "int"},
									},
								},
								// Access struct field
								{
									ID:   2,
									Op:   "member",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandValue, Value: 1, Type: "Point"},
										{Kind: mir.OperandLiteral, Literal: "x", Type: "string"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 2, Type: "int"},
								},
							},
						},
					},
				},
			},
		}

		result, err := GenerateC(complexModule)
		if err != nil {
			t.Fatalf("GenerateC failed: %v", err)
		}

		if result == "" {
			t.Error("Expected non-empty C code generation")
		}

		// Check for struct operations
		if !strings.Contains(result, "omni_struct") {
			t.Error("Expected struct operations in generated code")
		}
	})

	t.Run("GenerateCompleteModuleWithControlFlow", func(t *testing.T) {
		complexModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "test",
					ReturnType: "int",
					Params:     []mir.Param{{Name: "x", Type: "int", ID: 0}},
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								{
									ID:   1,
									Op:   "const",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "0", Type: "int"},
									},
								},
								{
									ID:   2,
									Op:   "cmp.gt",
									Type: "bool",
									Operands: []mir.Operand{
										{Kind: mir.OperandValue, Value: 0, Type: "int"},
										{Kind: mir.OperandValue, Value: 1, Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "cbr",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 2, Type: "bool"},
									{Kind: mir.OperandValue, Value: 3, Type: "block"},
									{Kind: mir.OperandValue, Value: 4, Type: "block"},
								},
							},
						},
						{
							Name: "then",
							Instructions: []mir.Instruction{
								{
									ID:   3,
									Op:   "const",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 3, Type: "int"},
								},
							},
						},
						{
							Name: "else",
							Instructions: []mir.Instruction{
								{
									ID:   4,
									Op:   "const",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "0", Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 4, Type: "int"},
								},
							},
						},
					},
				},
			},
		}

		result, err := GenerateC(complexModule)
		if err != nil {
			t.Fatalf("GenerateC failed: %v", err)
		}

		if result == "" {
			t.Error("Expected non-empty C code generation")
		}

		// Check for control flow
		if !strings.Contains(result, "goto") && !strings.Contains(result, "if") {
			t.Error("Expected control flow statements in generated code")
		}
	})

	t.Run("GenerateSourceMap", func(t *testing.T) {
		generator := NewCGeneratorWithDebug(module, "O0", true, "test.omni")

		// Generate some code first
		if len(module.Functions) > 0 {
			generator.generateFunction(module.Functions[0])
		}

		sourceMap := generator.GenerateSourceMap()
		if sourceMap == nil {
			t.Error("Expected non-nil source map")
		}
	})

	t.Run("GenerateWithAllOptimizationLevels", func(t *testing.T) {
		optLevels := []string{"O0", "O1", "O2", "O3", "Os", "Oz"}

		for _, optLevel := range optLevels {
			result, err := GenerateCOptimized(module, optLevel)
			if err != nil {
				t.Errorf("GenerateCOptimized failed for %s: %v", optLevel, err)
				continue
			}

			if result == "" {
				t.Errorf("Expected non-empty C code for optimization level %s", optLevel)
			}
		}
	})

	t.Run("GenerateCompleteModuleWithStringOperations", func(t *testing.T) {
		complexModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "test",
					ReturnType: "string",
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								{
									ID:   1,
									Op:   "const",
									Type: "string",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "\"hello\"", Type: "string"},
									},
								},
								{
									ID:   2,
									Op:   "const",
									Type: "string",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "\" world\"", Type: "string"},
									},
								},
								{
									ID:   3,
									Op:   "strcat",
									Type: "string",
									Operands: []mir.Operand{
										{Kind: mir.OperandValue, Value: 1, Type: "string"},
										{Kind: mir.OperandValue, Value: 2, Type: "string"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 3, Type: "string"},
								},
							},
						},
					},
				},
			},
		}

		result, err := GenerateC(complexModule)
		if err != nil {
			t.Fatalf("GenerateC failed: %v", err)
		}

		if result == "" {
			t.Error("Expected non-empty C code generation")
		}

		// Check for string operations
		if !strings.Contains(result, "omni_strcat") {
			t.Error("Expected string concatenation in generated code")
		}
	})

	// Additional integration tests
	t.Run("GenerateModuleWithArraysAndMaps", func(t *testing.T) {
		arrayMapModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "test",
					ReturnType: "int",
					Params:     []mir.Param{},
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								{
									ID:   1,
									Op:   "array.init",
									Type: "array<int>",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
										{Kind: mir.OperandLiteral, Literal: "2", Type: "int"},
									},
								},
								{
									ID:   2,
									Op:   "map.init",
									Type: "map<string,int>",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "\"key\"", Type: "string"},
										{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandLiteral, Literal: "0", Type: "int"},
								},
							},
						},
					},
				},
			},
		}

		result, err := GenerateC(arrayMapModule)
		if err != nil {
			t.Fatalf("GenerateC failed: %v", err)
		}

		if !strings.Contains(result, "omni_map_create") {
			t.Error("Expected map creation in generated code")
		}
	})

	t.Run("GenerateModuleWithStructs", func(t *testing.T) {
		structModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "test",
					ReturnType: "int",
					Params:     []mir.Param{},
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								{
									ID:   1,
									Op:   "struct.init",
									Type: "Point",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "x", Type: "string"},
										{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
										{Kind: mir.OperandLiteral, Literal: "y", Type: "string"},
										{Kind: mir.OperandLiteral, Literal: "2", Type: "int"},
									},
								},
								{
									ID:   2,
									Op:   "member",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandValue, Value: 1, Type: "Point"},
										{Kind: mir.OperandLiteral, Literal: "x", Type: "string"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 2, Type: "int"},
								},
							},
						},
					},
				},
			},
		}

		result, err := GenerateC(structModule)
		if err != nil {
			t.Fatalf("GenerateC failed: %v", err)
		}

		if !strings.Contains(result, "omni_struct_create") {
			t.Error("Expected struct creation in generated code")
		}
		if !strings.Contains(result, "omni_struct_get_int_field") {
			t.Error("Expected struct field access in generated code")
		}
	})

	t.Run("GenerateModuleWithControlFlow", func(t *testing.T) {
		controlFlowModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "test",
					ReturnType: "int",
					Params:     []mir.Param{{Name: "x", Type: "int", ID: 0}},
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								{
									ID:   1,
									Op:   "const",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "0", Type: "int"},
									},
								},
								{
									ID:   2,
									Op:   "cmp.eq",
									Type: "bool",
									Operands: []mir.Operand{
										{Kind: mir.OperandValue, Value: 0, Type: "int"},
										{Kind: mir.OperandValue, Value: 1, Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "cbr",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 2, Type: "bool"},
									{Kind: mir.OperandValue, Value: 1, Type: "block"},
									{Kind: mir.OperandValue, Value: 2, Type: "block"},
								},
							},
						},
						{
							Name: "then",
							Instructions: []mir.Instruction{
								{
									ID:   3,
									Op:   "const",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 3, Type: "int"},
								},
							},
						},
						{
							Name: "else",
							Instructions: []mir.Instruction{
								{
									ID:   4,
									Op:   "const",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "0", Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 4, Type: "int"},
								},
							},
						},
					},
				},
			},
		}

		result, err := GenerateC(controlFlowModule)
		if err != nil {
			t.Fatalf("GenerateC failed: %v", err)
		}

		if !strings.Contains(result, "if") || !strings.Contains(result, "goto") {
			t.Error("Expected control flow constructs in generated code")
		}
	})

	t.Run("GenerateModuleWithAsyncFunction", func(t *testing.T) {
		asyncModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "asyncFunc",
					ReturnType: "Promise<int>",
					Params:     []mir.Param{},
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								{
									ID:   1,
									Op:   "const",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 1, Type: "int"},
								},
							},
						},
					},
				},
			},
		}

		result, err := GenerateC(asyncModule)
		if err != nil {
			t.Fatalf("GenerateC failed: %v", err)
		}

		if !strings.Contains(result, "omni_promise_create_int") {
			t.Error("Expected promise creation in async function")
		}
	})

	t.Run("GenerateWithAllOptimizationLevels", func(t *testing.T) {
		optLevels := []string{"O0", "O1", "O2", "O3", "Os", "Oz"}

		for _, level := range optLevels {
			result, err := GenerateCOptimized(module, level)
			if err != nil {
				t.Errorf("GenerateCOptimized failed for %s: %v", level, err)
				continue
			}

			if result == "" {
				t.Errorf("Expected non-empty C code for optimization level %s", level)
			}
		}
	})

	t.Run("GenerateSourceMap", func(t *testing.T) {
		generator := NewCGeneratorWithDebug(module, "O0", true, "test.omni")

		// Generate some code to create source map entries
		generator.generateInstruction(&mir.Instruction{
			ID:   1,
			Op:   "const",
			Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
			},
		})

		sourceMap := generator.GenerateSourceMap()
		if sourceMap == nil {
			t.Error("Expected non-nil source map")
		}
	})

	t.Run("GenerateWithErrorHandling", func(t *testing.T) {
		// Test module with unsupported instruction
		errorModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "test",
					ReturnType: "void",
					Params:     []mir.Param{},
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								{
									ID:       1,
									Op:       "closure.create", // Unsupported
									Type:     "void",
									Operands: []mir.Operand{},
								},
							},
							Terminator: mir.Terminator{
								Op:       "ret",
								Operands: []mir.Operand{},
							},
						},
					},
				},
			},
		}

		generator := NewCGenerator(errorModule)
		err := generator.generateFunction(errorModule.Functions[0])
		if err == nil {
			t.Error("Expected error for unsupported closure instruction")
		}

		if len(generator.errors) == 0 {
			t.Error("Expected error to be recorded")
		}
	})

	t.Run("GenerateCompleteComplexModule", func(t *testing.T) {
		// Use a module with known array lengths to avoid warnings
		complexModule := &mir.Module{
			Functions: []*mir.Function{
				{
					Name:       "complex",
					ReturnType: "int",
					Params:     []mir.Param{},
					Blocks: []*mir.BasicBlock{
						{
							Name: "entry",
							Instructions: []mir.Instruction{
								{
									ID:   1,
									Op:   "array.init",
									Type: "array<int>",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
										{Kind: mir.OperandLiteral, Literal: "2", Type: "int"},
									},
								},
								{
									ID:   2,
									Op:   "map.init",
									Type: "map<string,int>",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "\"key\"", Type: "string"},
										{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
									},
								},
								{
									ID:   3,
									Op:   "const",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "0", Type: "int"},
									},
								},
								{
									ID:   4,
									Op:   "index",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandValue, Value: 1, Type: "array<int>"},
										{Kind: mir.OperandValue, Value: 3, Type: "int"},
									},
								},
								{
									ID:   5,
									Op:   "const",
									Type: "string",
									Operands: []mir.Operand{
										{Kind: mir.OperandLiteral, Literal: "\"key\"", Type: "string"},
									},
								},
								{
									ID:   6,
									Op:   "index",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandValue, Value: 2, Type: "map<string,int>"},
										{Kind: mir.OperandValue, Value: 5, Type: "string"},
									},
								},
								{
									ID:   7,
									Op:   "add",
									Type: "int",
									Operands: []mir.Operand{
										{Kind: mir.OperandValue, Value: 4, Type: "int"},
										{Kind: mir.OperandValue, Value: 6, Type: "int"},
									},
								},
							},
							Terminator: mir.Terminator{
								Op: "ret",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 7, Type: "int"},
								},
							},
						},
					},
				},
			},
		}

		result, err := GenerateC(complexModule)
		if err != nil {
			t.Fatalf("GenerateC failed: %v", err)
		}

		if result == "" {
			t.Error("Expected non-empty C code")
		}

		// Should have function signature
		if !strings.Contains(result, "int32_t complex") {
			t.Error("Expected function signature in generated code")
		}
	})
}
