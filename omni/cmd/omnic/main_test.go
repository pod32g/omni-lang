package main

import (
	"flag"
	"testing"
)

func TestStringFlag(t *testing.T) {
	// Test stringFlag creation
	flag := newStringFlag("default")
	if flag == nil {
		t.Fatal("newStringFlag returned nil")
	}

	if flag.value != "default" {
		t.Errorf("Expected value 'default', got '%s'", flag.value)
	}

	if flag.set != false {
		t.Errorf("Expected set to be false initially, got %t", flag.set)
	}
}

func TestStringFlagString(t *testing.T) {
	// Test String method
	flag := newStringFlag("test")
	result := flag.String()
	if result != "test" {
		t.Errorf("Expected 'test', got '%s'", result)
	}
}

func TestStringFlagSet(t *testing.T) {
	// Test Set method
	flag := newStringFlag("default")

	err := flag.Set("newvalue")
	if err != nil {
		t.Errorf("Set returned error: %v", err)
	}

	if flag.value != "newvalue" {
		t.Errorf("Expected value 'newvalue', got '%s'", flag.value)
	}

	if flag.set != true {
		t.Errorf("Expected set to be true after Set, got %t", flag.set)
	}
}

func TestStringFlagSetMultiple(t *testing.T) {
	// Test setting flag multiple times
	flag := newStringFlag("default")

	// Set first value
	err := flag.Set("first")
	if err != nil {
		t.Errorf("Set returned error: %v", err)
	}

	if flag.value != "first" {
		t.Errorf("Expected value 'first', got '%s'", flag.value)
	}

	// Set second value
	err = flag.Set("second")
	if err != nil {
		t.Errorf("Set returned error: %v", err)
	}

	if flag.value != "second" {
		t.Errorf("Expected value 'second', got '%s'", flag.value)
	}

	if flag.set != true {
		t.Errorf("Expected set to be true, got %t", flag.set)
	}
}

func TestVersionConstants(t *testing.T) {
	// Test version constants
	if Version == "" {
		t.Error("Version should not be empty")
	}

	if BuildTime == "" {
		t.Error("BuildTime should not be empty")
	}
}

func TestStringFlagImplementsFlagValue(t *testing.T) {
	// Test that stringFlag implements flag.Value interface
	var flag flag.Value = newStringFlag("test")
	if flag == nil {
		t.Fatal("stringFlag should implement flag.Value")
	}

	// Test String method
	result := flag.String()
	if result != "test" {
		t.Errorf("Expected 'test', got '%s'", result)
	}

	// Test Set method
	err := flag.Set("newvalue")
	if err != nil {
		t.Errorf("Set returned error: %v", err)
	}

	// Verify the value was set
	result = flag.String()
	if result != "newvalue" {
		t.Errorf("Expected 'newvalue', got '%s'", result)
	}
}

func TestStringFlagEdgeCases(t *testing.T) {
	// Test edge cases
	flag := newStringFlag("")

	// Test setting empty string
	err := flag.Set("")
	if err != nil {
		t.Errorf("Set returned error: %v", err)
	}

	if flag.value != "" {
		t.Errorf("Expected empty string, got '%s'", flag.value)
	}

	if flag.set != true {
		t.Errorf("Expected set to be true, got %t", flag.set)
	}

	// Test setting string with special characters
	err = flag.Set("test with spaces and symbols!@#$%")
	if err != nil {
		t.Errorf("Set returned error: %v", err)
	}

	if flag.value != "test with spaces and symbols!@#$%" {
		t.Errorf("Expected special string, got '%s'", flag.value)
	}
}
