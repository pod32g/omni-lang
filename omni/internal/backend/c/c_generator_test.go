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
							Op: "return",
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
		if !strings.Contains(result, "int main()") {
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
						Op: "return",
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

		generator := NewCGenerator(module)
		generator.generateBlock(block)

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
			Op: "return",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: 1, Type: "int"},
			},
		}

		generator := NewCGenerator(module)
		generator.generateTerminator(terminator)

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
		if !strings.Contains(result, "int main()") {
			t.Error("Expected main function signature")
		}
	})
}
