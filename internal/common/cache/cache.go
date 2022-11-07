package cache

import (
	"sync"

	"github.com/pkg/errors"
)

var (
	ErrInvalidCapacity = errors.New("invalid capacity")
	ErrEvictionFailed  = errors.New("eviction failed")
)

type Cache[K any, V any] interface {
	Get(key K) (V, bool, error)
	Set(key K, value V) error
	Drop(key K) (bool, error)
	Clear() error
	Len() (int, error)
}

type ThreadSafeCache[K any, V any, C Cache[K, V]] interface {
	Cache[K, V]
	Map(fn func(cache C) error) error
}

type threadSafeCache[K any, V any, C Cache[K, V]] struct {
	mu    *sync.Mutex
	inner C
}

func NewThreadSafeCache[K any, V any, C Cache[K, V]](inner C) ThreadSafeCache[K, V, C] {
	return &threadSafeCache[K, V, C]{mu: new(sync.Mutex), inner: inner}
}

func (c *threadSafeCache[K, V, C]) Get(key K) (V, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.inner.Get(key)
}

func (c *threadSafeCache[K, V, C]) Set(key K, value V) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.inner.Set(key, value)
}

func (c *threadSafeCache[K, V, C]) Drop(key K) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.inner.Drop(key)
}

func (c *threadSafeCache[K, V, C]) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.inner.Clear()
}

func (c *threadSafeCache[K, V, C]) Len() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.inner.Len()
}

func (c *threadSafeCache[K, V, C]) Map(fn func(cache C) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return fn(c.inner)
}
