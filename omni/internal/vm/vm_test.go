package vm

import (
	"strings"
	"testing"

	"github.com/omni-lang/omni/internal/mir"
)

func TestNewObjectPool(t *testing.T) {
	// Test creating a new object pool
	factory := func() interface{} {
		return &strings.Builder{}
	}

	pool := NewObjectPool(factory)
	if pool == nil {
		t.Fatal("NewObjectPool returned nil")
	}
}

func TestObjectPoolGetPut(t *testing.T) {
	// Test object pool get and put operations
	factory := func() interface{} {
		return &strings.Builder{}
	}

	pool := NewObjectPool(factory)
	if pool == nil {
		t.Fatal("NewObjectPool returned nil")
	}

	// Test getting an object
	obj := pool.Get()
	if obj == nil {
		t.Error("Expected non-nil object from pool")
	}

	// Test putting object back
	pool.Put(obj)
}

func TestStringBuilderPool(t *testing.T) {
	// Test string builder pool
	builder := GetStringBuilder()
	if builder == nil {
		t.Fatal("GetStringBuilder returned nil")
	}

	// Test using the builder
	builder.WriteString("test")
	if builder.String() != "test" {
		t.Errorf("Expected 'test', got '%s'", builder.String())
	}

	// Test returning to pool
	PutStringBuilder(builder)

	// Test getting another builder
	builder2 := GetStringBuilder()
	if builder2 == nil {
		t.Fatal("GetStringBuilder returned nil")
	}

	// Test that the builder is reset
	if builder2.String() != "" {
		t.Errorf("Expected empty string, got '%s'", builder2.String())
	}

	PutStringBuilder(builder2)
}

func TestVMInit(t *testing.T) {
	// Test that VM initialization works
	if instructionHandlers == nil {
		t.Fatal("Expected instructionHandlers to be initialized")
	}

	// Test that some basic instruction handlers exist
	expectedHandlers := []string{
		"const", "add", "sub", "mul", "div", "mod",
		"neg", "not", "cmp.eq", "cmp.neq", "and", "or",
	}

	for _, handler := range expectedHandlers {
		if instructionHandlers[handler] == nil {
			t.Errorf("Expected instruction handler '%s' to be registered", handler)
		}
	}
}

func TestInstructionHandlers(t *testing.T) {
	// Test that instruction handlers are properly registered
	if len(instructionHandlers) == 0 {
		t.Error("Expected instruction handlers to be registered")
	}

	// Test specific instruction handlers
	arithmeticOps := []string{"add", "sub", "mul", "div", "mod"}
	for _, op := range arithmeticOps {
		if instructionHandlers[op] == nil {
			t.Errorf("Expected arithmetic instruction handler '%s' to be registered", op)
		}
	}

	bitwiseOps := []string{"bitand", "bitor", "bitxor", "lshift", "rshift"}
	for _, op := range bitwiseOps {
		if instructionHandlers[op] == nil {
			t.Errorf("Expected bitwise instruction handler '%s' to be registered", op)
		}
	}

	comparisonOps := []string{"cmp.eq", "cmp.neq", "cmp.lt", "cmp.lte", "cmp.gt", "cmp.gte"}
	for _, op := range comparisonOps {
		if instructionHandlers[op] == nil {
			t.Errorf("Expected comparison instruction handler '%s' to be registered", op)
		}
	}
}

func TestObjectPoolConcurrency(t *testing.T) {
	// Test object pool with concurrent access
	factory := func() interface{} {
		return &strings.Builder{}
	}

	pool := NewObjectPool(factory)
	if pool == nil {
		t.Fatal("NewObjectPool returned nil")
	}

	// Test concurrent get/put operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Get object from pool
			obj := pool.Get()
			if obj == nil {
				t.Error("Expected non-nil object from pool")
				return
			}

			// Put object back to pool
			pool.Put(obj)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestTCOSelfRecursionDeep verifies the VM's tail-call dispatch loop
// handles a million-deep self-recursive call without growing the Go
// stack. Without TCO, this would either crash on a stack-overflow
// signal or grow the goroutine's stack to many MB.
//
// Module shape (count_down(n, acc) returns acc when n==0, else
// recurses):
//
//	if n == 0 { return acc }
//	return count_down(n - 1, acc + 1)
func TestTCOSelfRecursionDeep(t *testing.T) {
	param0 := mir.Param{Name: "n", Type: "int", ID: mir.ValueID(0)}
	param1 := mir.Param{Name: "acc", Type: "int", ID: mir.ValueID(1)}

	fn := &mir.Function{
		Name:       "count_down",
		ReturnType: "int",
		Params:     []mir.Param{param0, param1},
		Blocks: []*mir.BasicBlock{
			{
				Name: "entry",
				Instructions: []mir.Instruction{
					{ID: mir.ValueID(2), Op: "const", Type: "int",
						Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "0"}}},
					{ID: mir.ValueID(3), Op: "cmp.eq", Type: "bool",
						Operands: []mir.Operand{
							{Kind: mir.OperandValue, Value: mir.ValueID(0), Type: "int"},
							{Kind: mir.OperandValue, Value: mir.ValueID(2), Type: "int"},
						}},
				},
				Terminator: mir.Terminator{Op: "cbr", Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: mir.ValueID(3), Type: "bool"},
					{Kind: mir.OperandLiteral, Literal: "ret_block"},
					{Kind: mir.OperandLiteral, Literal: "rec_block"},
				}},
			},
			{
				Name:       "ret_block",
				Terminator: mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: mir.ValueID(1), Type: "int"}}},
			},
			{
				Name: "rec_block",
				Instructions: []mir.Instruction{
					{ID: mir.ValueID(4), Op: "const", Type: "int",
						Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "1"}}},
					{ID: mir.ValueID(5), Op: "sub", Type: "int",
						Operands: []mir.Operand{
							{Kind: mir.OperandValue, Value: mir.ValueID(0), Type: "int"},
							{Kind: mir.OperandValue, Value: mir.ValueID(4), Type: "int"},
						}},
					{ID: mir.ValueID(6), Op: "const", Type: "int",
						Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "1"}}},
					{ID: mir.ValueID(7), Op: "add", Type: "int",
						Operands: []mir.Operand{
							{Kind: mir.OperandValue, Value: mir.ValueID(1), Type: "int"},
							{Kind: mir.OperandValue, Value: mir.ValueID(6), Type: "int"},
						}},
					{ID: mir.ValueID(8), Op: "call", Type: "int",
						Operands: []mir.Operand{
							{Kind: mir.OperandLiteral, Literal: "count_down"},
							{Kind: mir.OperandValue, Value: mir.ValueID(5), Type: "int"},
							{Kind: mir.OperandValue, Value: mir.ValueID(7), Type: "int"},
						}},
				},
				Terminator: mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: mir.ValueID(8), Type: "int"}}},
			},
		},
	}
	funcs := map[string]*mir.Function{"count_down": fn}

	const depth = 1_000_000
	args := []Result{
		{Type: "int", Value: depth},
		{Type: "int", Value: 0},
	}
	res, err := execFunction(funcs, fn, args)
	if err != nil {
		t.Fatalf("execFunction: %v", err)
	}
	got, ok := res.Value.(int)
	if !ok {
		t.Fatalf("expected int result, got %T", res.Value)
	}
	if got != depth {
		t.Fatalf("expected %d, got %d", depth, got)
	}
}
