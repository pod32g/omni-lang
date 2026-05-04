package vm

import (
	"bytes"
	"encoding/json"
	"strings"
)

// JSON kind discriminator — must stay in sync with std/json/json.omni
// and the runtime omni_json_t tag.
const (
	jsonKindNull   = 0
	jsonKindBool   = 1
	jsonKindInt    = 2
	jsonKindFloat  = 3
	jsonKindString = 4
	jsonKindArray  = 5
	jsonKindObject = 6
)

// jsonNode wraps a parsed JSON tree. Numbers are stored as Go json.Number
// so we can distinguish int-valued from float-valued at access time
// without losing precision. Objects use map[string]interface{} (the
// encoding/json default) so keys are ordered by Go's map iteration —
// good enough for read paths; object_key_at returns an arbitrary stable
// order for now.
// jsonNode is a sharable handle to a JSON value. Wrapping the underlying
// interface{} in a pointer-box lets array_push / object_set replace the
// boxed value (e.g. when append grows the backing array) and have the
// change visible to other handles pointing at the same node.
type jsonNode struct {
	box *interface{}
}

// newJsonNode boxes v so jsonNode handles can mutate the wrapped value.
func newJsonNode(v interface{}) jsonNode {
	return jsonNode{box: &v}
}

// value returns the boxed value or nil when unset.
func (n jsonNode) value() interface{} {
	if n.box == nil {
		return nil
	}
	return *n.box
}

// jsonParse decodes s into a jsonNode. Returns a null-kind node on
// malformed input to match the C runtime's tolerance.
func jsonParse(s string) jsonNode {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var v interface{}
	if err := dec.Decode(&v); err != nil {
		return newJsonNode(nil)
	}
	return newJsonNode(v)
}

// jsonStringify re-encodes a node compactly.
func jsonStringify(n jsonNode) string {
	b, err := json.Marshal(n.value())
	if err != nil {
		return ""
	}
	return string(b)
}

// jsonStringifyPretty re-encodes with two-space indentation.
func jsonStringifyPretty(n jsonNode) string {
	b, err := json.MarshalIndent(n.value(), "", "  ")
	if err != nil {
		return ""
	}
	// json.MarshalIndent returns the bytes already; convert directly.
	var buf bytes.Buffer
	buf.Write(b)
	return buf.String()
}

// jsonKind returns the kind discriminator for n.
func jsonKind(n jsonNode) int {
	switch v := n.value().(type) {
	case nil:
		return jsonKindNull
	case bool:
		return jsonKindBool
	case json.Number:
		if _, err := v.Int64(); err == nil && !strings.ContainsAny(string(v), ".eE") {
			return jsonKindInt
		}
		return jsonKindFloat
	case float64:
		return jsonKindFloat
	case string:
		return jsonKindString
	case []interface{}:
		return jsonKindArray
	case map[string]interface{}:
		return jsonKindObject
	default:
		return jsonKindNull
	}
}

// jsonAsBool extracts a bool, defaulting to false.
func jsonAsBool(n jsonNode) bool {
	if b, ok := n.value().(bool); ok {
		return b
	}
	return false
}

// jsonAsInt truncates numeric values to int.
func jsonAsInt(n jsonNode) int {
	switch v := n.value().(type) {
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i)
		}
		if f, err := v.Float64(); err == nil {
			return int(f)
		}
	case float64:
		return int(v)
	}
	return 0
}

// jsonAsFloat extracts the numeric value as float64.
func jsonAsFloat(n jsonNode) float64 {
	switch v := n.value().(type) {
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f
		}
	case float64:
		return v
	}
	return 0.0
}

// jsonAsString extracts a string, defaulting to "".
func jsonAsString(n jsonNode) string {
	if s, ok := n.value().(string); ok {
		return s
	}
	return ""
}

// jsonObjectGet looks up key, returning a null node if absent or n
// isn't an object.
func jsonObjectGet(n jsonNode, key string) jsonNode {
	if m, ok := n.value().(map[string]interface{}); ok {
		if v, present := m[key]; present {
			return newJsonNode(v)
		}
	}
	return newJsonNode(nil)
}

// jsonObjectHas returns whether n has key.
func jsonObjectHas(n jsonNode, key string) bool {
	if m, ok := n.value().(map[string]interface{}); ok {
		_, present := m[key]
		return present
	}
	return false
}

// jsonObjectSize returns the number of keys.
func jsonObjectSize(n jsonNode) int {
	if m, ok := n.value().(map[string]interface{}); ok {
		return len(m)
	}
	return 0
}

// jsonObjectKeyAt returns the i-th key. Go map iteration is not stable
// across runs, so we sort keys lexicographically before indexing —
// this gives a deterministic order at the cost of differing from the
// C runtime's insertion order. Callers that care about order should
// iterate via object_size + object_key_at consistently within a run.
func jsonObjectKeyAt(n jsonNode, i int) string {
	m, ok := n.value().(map[string]interface{})
	if !ok {
		return ""
	}
	if i < 0 || i >= len(m) {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// stable order
	sortStrings(keys)
	return keys[i]
}

// sortStrings sorts in place — small helper to avoid pulling sort into
// this package's import set just for one call site.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}

// jsonArrayGet returns the i-th element, null on out-of-range / wrong kind.
func jsonArrayGet(n jsonNode, i int) jsonNode {
	if a, ok := n.value().([]interface{}); ok {
		if i >= 0 && i < len(a) {
			return newJsonNode(a[i])
		}
	}
	return newJsonNode(nil)
}

// jsonArrayLen returns len(array) or 0.
func jsonArrayLen(n jsonNode) int {
	if a, ok := n.value().([]interface{}); ok {
		return len(a)
	}
	return 0
}
