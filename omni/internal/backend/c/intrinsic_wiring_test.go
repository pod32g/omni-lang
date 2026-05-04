package cbackend

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/omni-lang/omni/internal/mir"
)

// TestStdNetworkIntrinsicWiring is a static cross-check between
// `std/network/network.omni` and the C-backend + VM dispatch tables.
//
// std.network groups its functions into three categories via doc-
// comment markers (see the file header):
//
//   [INTRINSIC]    must be wired on BOTH the C backend and the VM.
//                  The OmniLang body is a fail-loud panic.
//   [INTRINSIC-C]  must be wired on the C backend; the OmniLang body
//                  is the VM-side implementation, so the VM dispatch
//                  is optional.
//   [PURE]         not an intrinsic; nothing to check.
//
// This test parses network.omni for those markers and asserts the
// dispatch tables agree. If a registry entry is dropped (or a marker
// is added without a registry entry), CI fails here long before a
// user program would silently fall through to the panic body.
func TestStdNetworkIntrinsicWiring(t *testing.T) {
	src := readNetworkOmni(t)
	funcs := parseAnnotatedFunctions(t, src)

	if len(funcs) == 0 {
		t.Fatal("found no annotated functions in network.omni — parser broken or markers removed")
	}

	// Read the VM dispatch source once so we can grep it per-function.
	vmSrc := readVMSource(t)

	// Build a CGenerator just to invoke the unexported methods; we
	// don't run codegen, only the lookup tables.
	g := NewCGenerator(&mir.Module{})

	for _, fn := range funcs {
		fn := fn
		t.Run(fn.Name, func(t *testing.T) {
			qname := "std.network." + fn.Name

			switch fn.Marker {
			case "INTRINSIC", "INTRINSIC-C":
				// C-backend wiring: name mapping must produce something
				// other than the default `<dotted_name_with_underscores>`
				// fallback, and isRuntimeProvidedFunction must agree.
				mapped := g.mapFunctionName(qname)
				expectedFallback := strings.ReplaceAll(qname, ".", "_")
				if mapped == expectedFallback {
					t.Errorf("C backend: std.network.%s marked [%s] but mapFunctionName falls back to %q (no runtime mapping registered)",
						fn.Name, fn.Marker, mapped)
				}
				if !g.isRuntimeProvidedFunction(qname) {
					t.Errorf("C backend: std.network.%s marked [%s] but isRuntimeProvidedFunction returns false (the OmniLang stub body would emit and shadow the runtime symbol)",
						fn.Name, fn.Marker)
				}
				if !strings.HasPrefix(mapped, "omni_") {
					t.Errorf("C backend: std.network.%s mapped to %q which doesn't follow the omni_* convention",
						fn.Name, mapped)
				}

				// VM wiring is required for [INTRINSIC] (both backends),
				// optional for [INTRINSIC-C] (the OmniLang body is the
				// VM-side fallback).
				if fn.Marker == "INTRINSIC" {
					if !vmHasDispatch(vmSrc, qname) {
						t.Errorf("VM: std.network.%s marked [INTRINSIC] but vm.go has no `case %q` — the OmniLang fail-loud body would run on omnir",
							fn.Name, qname)
					}
				}
			case "PURE":
				// Nothing to check; the OmniLang body IS the implementation.
			default:
				t.Fatalf("unknown marker %q on std.network.%s — update the validator", fn.Marker, fn.Name)
			}
		})
	}
}

// annotatedFunc captures a function declaration in network.omni along
// with its category marker.
type annotatedFunc struct {
	Name   string
	Marker string // "INTRINSIC", "INTRINSIC-C", or "PURE"
}

// parseAnnotatedFunctions scans network.omni for `func NAME(...)`
// definitions whose preceding doc-comment block contains one of the
// category markers. Comments and definitions can be separated by
// blank lines, but the marker must appear before the next blank-line
// gap, so we accumulate comments in a window that resets on blanks.
func parseAnnotatedFunctions(t *testing.T, src string) []annotatedFunc {
	t.Helper()
	funcRE := regexp.MustCompile(`^func\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	markerRE := regexp.MustCompile(`\[(INTRINSIC|INTRINSIC-C|PURE)\]`)

	lines := strings.Split(src, "\n")
	var window []string
	var found []annotatedFunc
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		// Reset the window on blank lines or section-header rules.
		if trim == "" {
			window = nil
			continue
		}
		if strings.HasPrefix(trim, "//") {
			window = append(window, trim)
			continue
		}
		// Non-blank, non-comment line. If it's a func, look for a
		// marker in the accumulated comment window.
		if m := funcRE.FindStringSubmatch(trim); m != nil {
			name := m[1]
			// Skip the internal helper itself.
			if strings.HasPrefix(name, "_") {
				window = nil
				continue
			}
			marker := ""
			for _, c := range window {
				if mm := markerRE.FindStringSubmatch(c); mm != nil {
					marker = mm[1]
					break
				}
			}
			if marker == "" {
				t.Errorf("std.network.%s has no [INTRINSIC] / [INTRINSIC-C] / [PURE] marker in its doc comment; add one or the validator can't classify it", name)
				marker = "INTRINSIC" // assume strict for the rest of the test
			}
			found = append(found, annotatedFunc{Name: name, Marker: marker})
		}
		window = nil
	}
	return found
}

// vmHasDispatch returns true if vm.go contains a `case "<qname>":`
// branch in its intrinsic switch. We don't try to verify what the
// branch does — only that it intercepts the call before falling
// through to the OmniLang body.
func vmHasDispatch(vmSrc, qname string) bool {
	needle := `case "` + qname + `":`
	return strings.Contains(vmSrc, needle)
}

func readNetworkOmni(t *testing.T) string {
	t.Helper()
	// Test cwd is .../omni/internal/backend/c. network.omni is at
	// .../omni/std/network/network.omni — four levels up.
	path, err := filepath.Abs(filepath.Join("..", "..", "..", "std", "network", "network.omni"))
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func readVMSource(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("..", "..", "vm", "vm.go"))
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
