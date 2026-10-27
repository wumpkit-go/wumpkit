package shardedmap_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wumpkit/wumpkit/internal/shardedmap"
)

type customKey struct {
	ID   uint64
	Name string
}

func TestShardedMap_BasicCRUD(t *testing.T) {
	m := shardedmap.New[string, int](16)

	m.Set("apple", 10)
	m.Set("banana", 20)

	val, ok := m.Get("apple")
	assert.True(t, ok)
	assert.Equal(t, 10, val)

	assert.True(t, m.Has("banana"))
	assert.False(t, m.Has("cherry"))

	assert.Equal(t, 2, m.Len())

	old, ok := m.Delete("apple")
	assert.True(t, ok)
	assert.Equal(t, 10, old)
	assert.False(t, m.Has("apple"))
	assert.Equal(t, 1, m.Len())
}

func TestShardedMap_CustomComparableKey(t *testing.T) {
	m := shardedmap.New[customKey, string](8)

	k1 := customKey{ID: 100, Name: "alpha"}
	k2 := customKey{ID: 200, Name: "beta"}

	m.Set(k1, "first")
	m.Set(k2, "second")

	v1, ok := m.Get(k1)
	assert.True(t, ok)
	assert.Equal(t, "first", v1)

	v2, ok := m.Get(k2)
	assert.True(t, ok)
	assert.Equal(t, "second", v2)

	assert.Equal(t, 2, m.Len())
}

func TestShardedMap_Compute(t *testing.T) {
	m := shardedmap.New[string, int](4)

	res := m.Compute("counter", func(val int, exists bool) int {
		assert.False(t, exists)
		return val + 1
	})
	assert.Equal(t, 1, res)

	res = m.Compute("counter", func(val int, exists bool) int {
		assert.True(t, exists)
		return val + 1
	})
	assert.Equal(t, 2, res)
}

func TestShardedMap_ClearAndRange(t *testing.T) {
	m := shardedmap.New[string, int](16)
	m.Set("k1", 100)
	m.Set("k2", 200)
	m.Set("k3", 300)

	assert.Equal(t, 3, m.Len())

	collected := make(map[string]int)
	m.Range(func(k string, v int) bool {
		collected[k] = v
		return true
	})

	assert.Equal(t, 3, len(collected))
	assert.Equal(t, 100, collected["k1"])
	assert.Equal(t, 200, collected["k2"])
	assert.Equal(t, 300, collected["k3"])

	m.Clear()
	assert.Equal(t, 0, m.Len())
	assert.False(t, m.Has("k1"))
}

func TestShardedMap_ConcurrentAccess(t *testing.T) {
	m := shardedmap.New[int, int](32)
	var wg sync.WaitGroup
	workers := 20
	opsPerWorker := 500

	for w := range workers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := range opsPerWorker {
				key := (workerID * 1000) + i
				m.Set(key, i)
				_, _ = m.Get(key)
				m.Has(key)
				m.Compute(key, func(v int, exists bool) int {
					return v + 1
				})
			}
		}(w)
	}

	wg.Wait()
	assert.Equal(t, workers*opsPerWorker, m.Len())
}
