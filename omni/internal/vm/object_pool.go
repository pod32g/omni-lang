package vm

import (
	"strings"
	"sync"
)

// ObjectPool provides a simple object pool for reducing allocations.
type ObjectPool struct {
	pool sync.Pool
}

// NewObjectPool creates a new object pool with a factory function.
func NewObjectPool(factory func() interface{}) *ObjectPool {
	return &ObjectPool{
		pool: sync.Pool{
			New: factory,
		},
	}
}

// Get retrieves an object from the pool.
func (p *ObjectPool) Get() interface{} {
	return p.pool.Get()
}

// Put returns an object to the pool.
func (p *ObjectPool) Put(obj interface{}) {
	p.pool.Put(obj)
}

// StringBuilderPool provides a pool of strings.Builder objects.
var StringBuilderPool = NewObjectPool(func() interface{} {
	return &strings.Builder{}
})

// GetStringBuilder retrieves a strings.Builder from the pool.
func GetStringBuilder() *strings.Builder {
	return StringBuilderPool.Get().(*strings.Builder)
}

// PutStringBuilder returns a strings.Builder to the pool.
func PutStringBuilder(builder *strings.Builder) {
	builder.Reset()
	StringBuilderPool.Put(builder)
}
