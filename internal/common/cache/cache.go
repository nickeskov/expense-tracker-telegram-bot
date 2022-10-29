package cache

import "github.com/pkg/errors"

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
