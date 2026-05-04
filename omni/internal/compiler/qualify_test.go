package compiler

import (
	"testing"

	"github.com/omni-lang/omni/internal/ast"
)

// TestQualifyTypeRefs_Recursive verifies that nested generic type arguments
// (e.g. array<IPAddress>) are qualified, not just the top-level Name. Before
// this fix, std.network's dns_lookup return type was cloned as
// array<IPAddress> instead of array<std.network.IPAddress>.
func TestQualifyTypeRefs_Recursive(t *testing.T) {
	types := map[string]bool{"IPAddress": true, "URL": true}

	// array<IPAddress>
	in := &ast.TypeExpr{
		Name: "array",
		Args: []*ast.TypeExpr{{Name: "IPAddress"}},
	}
	got := qualifyTypeRefs(in, types, "std.network")
	if got.Name != "array" {
		t.Errorf("top-level Name = %q, want %q", got.Name, "array")
	}
	if len(got.Args) != 1 || got.Args[0].Name != "std.network.IPAddress" {
		t.Errorf("Args[0].Name = %q, want %q", got.Args[0].Name, "std.network.IPAddress")
	}
	// Original must be untouched.
	if in.Args[0].Name != "IPAddress" {
		t.Errorf("input mutated: Args[0].Name = %q", in.Args[0].Name)
	}

	// map<string,URL> — element type qualified, key untouched.
	in2 := &ast.TypeExpr{
		Name: "map",
		Args: []*ast.TypeExpr{{Name: "string"}, {Name: "URL"}},
	}
	got2 := qualifyTypeRefs(in2, types, "std.network")
	if got2.Args[0].Name != "string" {
		t.Errorf("map key qualified unexpectedly: %q", got2.Args[0].Name)
	}
	if got2.Args[1].Name != "std.network.URL" {
		t.Errorf("map value Name = %q, want std.network.URL", got2.Args[1].Name)
	}

	// Nil safe.
	if qualifyTypeRefs(nil, types, "x") != nil {
		t.Error("nil input should produce nil output")
	}

	// Type not in set: passthrough.
	in3 := &ast.TypeExpr{Name: "int"}
	got3 := qualifyTypeRefs(in3, types, "std.network")
	if got3.Name != "int" {
		t.Errorf("non-module type qualified: %q", got3.Name)
	}
}

func TestQualifyFuncDeclTypes_NestedReturn(t *testing.T) {
	types := map[string]bool{"IPAddress": true}
	decl := &ast.FuncDecl{
		Name: "dns_lookup",
		Params: []ast.Param{
			{Name: "hostname", Type: &ast.TypeExpr{Name: "string"}},
		},
		Return: &ast.TypeExpr{
			Name: "array",
			Args: []*ast.TypeExpr{{Name: "IPAddress"}},
		},
	}
	qualifyFuncDeclTypes(decl, types, "std.network")
	if decl.Return.Args[0].Name != "std.network.IPAddress" {
		t.Errorf("return inner type = %q, want std.network.IPAddress", decl.Return.Args[0].Name)
	}
	if decl.Params[0].Type.Name != "string" {
		t.Errorf("string param incorrectly qualified: %q", decl.Params[0].Type.Name)
	}
}

func TestCollectModuleTypeNames(t *testing.T) {
	mod := &ast.Module{
		Decls: []ast.Decl{
			&ast.StructDecl{Name: "IPAddress"},
			&ast.StructDecl{Name: "URL"},
			&ast.EnumDecl{Name: "Protocol"},
			&ast.TypeAliasDecl{Name: "Headers"},
			&ast.FuncDecl{Name: "ip_parse"},
		},
	}
	got := collectModuleTypeNames(mod)
	for _, want := range []string{"IPAddress", "URL", "Protocol", "Headers"} {
		if !got[want] {
			t.Errorf("expected %q in type set", want)
		}
	}
	if got["ip_parse"] {
		t.Error("function name leaked into type set")
	}
}
