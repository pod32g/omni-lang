package passes

import (
	"testing"

	"github.com/omni-lang/omni/internal/mir"
)

func TestVerifyDetectsMissingTerminator(t *testing.T) {
	fn := mir.NewFunction("broken", "void", nil)
	fn.NewBlock("entry")
	mod := &mir.Module{Functions: []*mir.Function{fn}}
	if err := Verify(mod); err == nil {
		t.Fatalf("expected verifier error for missing terminator")
	}
}
