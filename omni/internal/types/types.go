package types

// Kind encodes the primitive type family classification.
type Kind string

const (
	KindInt    Kind = "int"
	KindLong   Kind = "long"
	KindByte   Kind = "byte"
	KindFloat  Kind = "float"
	KindDouble Kind = "double"
	KindBool   Kind = "bool"
	KindChar   Kind = "char"
	KindString Kind = "string"
	KindVoid   Kind = "void"
)

// Type captures the eventual type system representation. The bootstrap version
// only stores the kind for diagnostics and planning.
type Type struct {
	Kind Kind
}

// NewPrimitive constructs a type from a primitive kind.
func NewPrimitive(kind Kind) Type {
	return Type{Kind: kind}
}
