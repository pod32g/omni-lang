package compiler_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/omni-lang/omni/internal/compiler"
)

const sampleProgram = "func fortyTwo():int => 42\n"

func TestCompileVMEmitMIRWritesDefaultOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "simple.omni")
	if err := os.WriteFile(input, []byte(sampleProgram), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	cfg := compiler.Config{
		InputPath: input,
		Backend:   "vm",
		Emit:      "mir",
	}
	if err := compiler.Compile(cfg); err != nil {
		t.Fatalf("compile: %v", err)
	}

	output := filepath.Join(dir, "simple.mir")
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	expected := "func fortyTwo():int\n  block entry:\n    %0 = const.int 42:int\n    ret %0\n"
	if string(data) != expected {
		t.Fatalf("unexpected MIR output\nexpected:\n%s\nactual:\n%s", expected, string(data))
	}
}

func TestCompileVMEmitMIRCustomOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "main.omni")
	if err := os.WriteFile(input, []byte(sampleProgram), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	output := filepath.Join(dir, "out", "program.mir")
	cfg := compiler.Config{
		InputPath:  input,
		OutputPath: output,
		Backend:    "vm",
		Emit:       "mir",
	}
	if err := compiler.Compile(cfg); err != nil {
		t.Fatalf("compile: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("expected MIR data to be written")
	}
}

func TestCompileVMRejectsUnsupportedEmit(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "simple.omni")
	if err := os.WriteFile(input, []byte(sampleProgram), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	cfg := compiler.Config{
		InputPath: input,
		Backend:   "vm",
		Emit:      "obj",
	}
	err := compiler.Compile(cfg)
	if err == nil {
		t.Fatalf("expected error for unsupported emit option")
	}
	if !strings.Contains(err.Error(), "emit option") {
		t.Fatalf("unexpected error: %v", err)
	}
}
