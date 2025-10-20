//go:build darwin || windows
// +build darwin windows

package compiler

import (
	"fmt"

	"github.com/omni-lang/omni/internal/mir"
)

func compileCranelift(cfg Config, emit string, mod *mir.Module) error {
	return fmt.Errorf("cranelift backend not available for cross-compilation to %s", emit)
}
