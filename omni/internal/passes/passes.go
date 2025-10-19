package passes

import "github.com/omni-lang/omni/internal/mir"

// Pipeline owns an ordered set of MIR passes. While unimplemented, the
// structure provides the scaffolding needed by future tests.
type Pipeline struct {
	Name string
}

// NewPipeline constructs a placeholder pipeline.
func NewPipeline(name string) Pipeline {
	return Pipeline{Name: name}
}

// Run currently returns the input module verbatim. Future work will thread the
// MIR allocator, SSA builder and opt passes through this entry point.
func (p Pipeline) Run(mod mir.Module) (mir.Module, error) {
	return mod, nil
}
