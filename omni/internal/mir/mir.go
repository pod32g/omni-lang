package mir

// Module captures the mid-level SSA representation that backs OmniLang code
// generation. The bootstrap SSA graph only exposes a label so tests can evolve
// incrementally.
type Module struct {
	Name string
}

// New constructs a placeholder MIR module with the supplied name.
func New(name string) Module {
	return Module{Name: name}
}
