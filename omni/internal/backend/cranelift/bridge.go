package cranelift

import "fmt"

// CompileMIRJSON delegates MIR emission to the native Cranelift bridge. The
// bootstrap implementation returns an informative error so higher layers can
// make progress without native codegen.
func CompileMIRJSON(json string) error {
	if json == "" {
		return fmt.Errorf("mir payload required")
	}
	return fmt.Errorf("cranelift backend: not implemented")
}
