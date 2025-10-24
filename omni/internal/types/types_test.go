package types

import (
	"testing"
)

func TestNewPrimitive(t *testing.T) {
	primitive := NewPrimitive(KindInt)

	if primitive.Kind != KindInt {
		t.Errorf("Expected primitive kind 'int', got '%s'", primitive.Kind)
	}
}

func TestType(t *testing.T) {
	primitive := Type{
		Kind: KindString,
	}

	if primitive.Kind != KindString {
		t.Errorf("Expected primitive kind 'string', got '%s'", primitive.Kind)
	}
}

func TestKindConstants(t *testing.T) {
	tests := []struct {
		name     string
		kind     Kind
		expected string
	}{
		{"int", KindInt, "int"},
		{"long", KindLong, "long"},
		{"byte", KindByte, "byte"},
		{"float", KindFloat, "float"},
		{"double", KindDouble, "double"},
		{"bool", KindBool, "bool"},
		{"char", KindChar, "char"},
		{"string", KindString, "string"},
		{"void", KindVoid, "void"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.kind) != tt.expected {
				t.Errorf("Expected kind '%s', got '%s'", tt.expected, string(tt.kind))
			}
		})
	}
}

func TestTypeCreation(t *testing.T) {
	// Test creating different types
	intType := NewPrimitive(KindInt)
	if intType.Kind != KindInt {
		t.Errorf("Expected int type, got %s", intType.Kind)
	}

	stringType := NewPrimitive(KindString)
	if stringType.Kind != KindString {
		t.Errorf("Expected string type, got %s", stringType.Kind)
	}

	boolType := NewPrimitive(KindBool)
	if boolType.Kind != KindBool {
		t.Errorf("Expected bool type, got %s", boolType.Kind)
	}
}
