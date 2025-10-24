package compiler

import (
	"testing"
)

func TestNewStringPool(t *testing.T) {
	// Test creating a new string pool
	pool := NewStringPool()
	if pool == nil {
		t.Fatal("NewStringPool returned nil")
	}
}

func TestStringPoolIntern(t *testing.T) {
	// Test string pooling
	pool := NewStringPool()
	if pool == nil {
		t.Fatal("NewStringPool returned nil")
	}

	// Test interning a string
	str := pool.Intern("test")
	if str == "" {
		t.Error("Expected non-empty string from Intern")
	}

	// Test interning the same string again
	str2 := pool.Intern("test")
	if str != str2 {
		t.Error("Expected same string for same input")
	}
}

func TestStringPoolSize(t *testing.T) {
	// Test string pool size
	pool := NewStringPool()
	if pool == nil {
		t.Fatal("NewStringPool returned nil")
	}

	// Test initial size
	size := pool.Size()
	if size < 0 {
		t.Error("Expected non-negative size")
	}

	// Test size after interning
	pool.Intern("test")
	newSize := pool.Size()
	if newSize <= size {
		t.Error("Expected size to increase after interning")
	}
}

func TestStringPoolClear(t *testing.T) {
	// Test string pool clear
	pool := NewStringPool()
	if pool == nil {
		t.Fatal("NewStringPool returned nil")
	}

	// Test interning before clear
	pool.Intern("test")
	size1 := pool.Size()

	// Test clear
	pool.Clear()
	size2 := pool.Size()

	if size2 >= size1 {
		t.Error("Expected size to decrease after clear")
	}
}

func TestStringPoolMultipleStrings(t *testing.T) {
	// Test string pool with multiple strings
	pool := NewStringPool()
	if pool == nil {
		t.Fatal("NewStringPool returned nil")
	}

	// Test interning multiple strings
	str1 := pool.Intern("hello")
	str2 := pool.Intern("world")
	str3 := pool.Intern("hello") // Duplicate

	if str1 == "" || str2 == "" || str3 == "" {
		t.Error("Expected non-empty strings from Intern")
	}

	if str1 != str3 {
		t.Error("Expected same string for duplicate input")
	}

	if str1 == str2 {
		t.Error("Expected different strings for different input")
	}
}
