package compiler

import (
	"sync"
)

// StringPool provides string interning to reduce memory usage and improve comparison performance.
type StringPool struct {
	mu   sync.RWMutex
	pool map[string]string
}

// NewStringPool creates a new string pool.
func NewStringPool() *StringPool {
	return &StringPool{
		pool: make(map[string]string),
	}
}

// Intern returns an interned version of the string, reusing existing instances when possible.
func (sp *StringPool) Intern(s string) string {
	if s == "" {
		return ""
	}

	sp.mu.RLock()
	if interned, exists := sp.pool[s]; exists {
		sp.mu.RUnlock()
		return interned
	}
	sp.mu.RUnlock()

	sp.mu.Lock()
	defer sp.mu.Unlock()

	// Double-check after acquiring write lock
	if interned, exists := sp.pool[s]; exists {
		return interned
	}

	// Store the string itself as the interned version
	sp.pool[s] = s
	return s
}

// Size returns the number of interned strings.
func (sp *StringPool) Size() int {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return len(sp.pool)
}

// Clear removes all interned strings.
func (sp *StringPool) Clear() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.pool = make(map[string]string)
}
