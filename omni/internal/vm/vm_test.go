package vm

import (
	"strings"
	"testing"
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
