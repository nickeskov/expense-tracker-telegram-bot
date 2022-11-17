package lru

import (
	"container/list"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/common/cache"
)

type item[K comparable, V any] struct {
	key   K
	value V
}

func (i *item[K, V]) Value() V {
	return i.value
}

func newItem[K comparable, V any](k K, v V) *item[K, V] {
	return &item[K, V]{key: k, value: v}
}

type Cache[K comparable, V any] struct {
	m   map[K]*list.Element
	l   *list.List
	cap int
}

func New[K comparable, V any](capacity int) (*Cache[K, V], error) {
	if capacity < 0 {
		return nil, errors.Wrap(cache.ErrInvalidCapacity, "capacity is negative")
	}
	return &Cache[K, V]{
		m:   make(map[K]*list.Element), // TODO: consider capacity reserving
		l:   list.New(),
		cap: capacity,
	}, nil
}

func (c *Cache[K, V]) Get(key K) (cache.Item[V], bool, error) {
	v, ok := c.moveToFrontAndGet(key)
	if !ok {
		return nil, false, nil
	}
	return v, ok, nil
}

func (c *Cache[K, V]) Set(key K, value V) error {
	if it, ok := c.moveToFrontAndGet(key); ok {
		it.value = value
		return nil
	}
	if l := c.len(); l >= c.cap {
		if evicted := c.evictBack(); !evicted {
			return errors.Wrap(cache.ErrEvictionFailed, "length is zero")
		}
	}
	c.pushFront(key, value)
	return nil
}

func (c *Cache[K, V]) Drop(key K) (bool, error) {
	return c.drop(key), nil
}

func (c *Cache[K, V]) Len() (int, error) {
	return c.len(), nil
}

func (c *Cache[K, V]) Clear() error {
	c.m = make(map[K]*list.Element) // TODO: consider capacity reserving
	c.l.Init()
	return nil
}

func (c *Cache[K, V]) len() int {
	return len(c.m)
}

func (c *Cache[K, V]) drop(key K) bool {
	elem, ok := c.m[key]
	if !ok {
		return false
	}
	c.l.Remove(elem)
	delete(c.m, key)
	return true
}

func (c *Cache[K, V]) pushFront(key K, value V) {
	r := newItem(key, value)
	elem := c.l.PushFront(r)
	c.m[key] = elem
}

func (c *Cache[K, V]) evictBack() bool {
	if c.l.Len() == 0 {
		return false
	}
	back := c.l.Back()
	it := back.Value.(*item[K, V])
	delete(c.m, it.key)
	c.l.Remove(back)
	return true
}

func (c *Cache[K, V]) moveToFrontAndGet(key K) (*item[K, V], bool) {
	elem, ok := c.m[key]
	if !ok {
		return nil, false
	}
	c.l.MoveToFront(elem)
	return elem.Value.(*item[K, V]), true
}
