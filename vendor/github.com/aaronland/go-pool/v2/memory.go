package pool

// https://github.com/SimonWaldherr/golang-examples/blob/2be89f3185aded00740a45a64e3c98855193b948/advanced/lifo.go

import (
	"context"
	"sync"
	"sync/atomic"
)

const MEMORY_SCHEME string = "memory"

func init() {
	ctx := context.Background()
	RegisterPool(ctx, MEMORY_SCHEME, NewMemoryPool)
}

type MemoryPool struct {
	Pool
	nodes []any
	count int64
	mutex *sync.Mutex
}

func NewMemoryPool(ctx context.Context, uri string) (Pool, error) {

	mu := new(sync.Mutex)
	nodes := make([]any, 0)

	pl := &MemoryPool{
		mutex: mu,
		nodes: nodes,
		count: 0,
	}

	return pl, nil
}

func (pl *MemoryPool) Length(ctx context.Context) int64 {

	pl.mutex.Lock()
	defer pl.mutex.Unlock()

	return atomic.LoadInt64(&pl.count)
}

func (pl *MemoryPool) Push(ctx context.Context, i any) error {

	pl.mutex.Lock()
	defer pl.mutex.Unlock()

	pl.nodes = append(pl.nodes[:pl.count], i)
	atomic.AddInt64(&pl.count, 1)
	return nil
}

func (pl *MemoryPool) Pop(ctx context.Context) (any, bool) {

	pl.mutex.Lock()
	defer pl.mutex.Unlock()

	if pl.count == 0 {
		return nil, false
	}

	atomic.AddInt64(&pl.count, -1)
	i := pl.nodes[pl.count]

	return i, true
}
