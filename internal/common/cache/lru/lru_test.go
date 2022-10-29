package lru

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	cachePkg "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/common/cache"
)

func newStringCache(capacity int, keyval ...string) (*Cache[string, string], error) {
	cache, err := New[string, string](capacity)
	if err != nil {
		return nil, err
	}
	for _, s := range keyval {
		if err := cache.Set(s, s); err != nil {
			return nil, err
		}
	}
	return cache, nil
}

func TestCache_Get(t *testing.T) {
	tests := []struct {
		cap    int
		keyVal []string
	}{
		{cap: 2, keyVal: []string{"1", "2", "3", "4"}},
		{cap: 10, keyVal: []string{"1", "2", "3", "4"}},
	}
	for i := range tests {
		test := tests[i]
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			cache, err := newStringCache(test.cap, test.keyVal...)
			require.NoError(t, err)
			keyVal := test.keyVal
			if test.cap < len(keyVal) {
				for i := 0; i < test.cap; i++ {
					_, ok, err := cache.Get(keyVal[i])
					require.NoError(t, err)
					require.False(t, ok)
				}
				keyVal = keyVal[test.cap:]
			}
			for _, s := range keyVal {
				val, ok, err := cache.Get(s)
				require.NoError(t, err)
				require.True(t, ok)
				require.Equal(t, s, val)
				front := cache.l.Front().Value.(*item[string, string])
				require.Equal(t, s, front.Key)
				require.Equal(t, s, front.Value)
			}
		})
	}
}

func TestCache_Set(t *testing.T) {
	tests := []struct {
		cap    int
		keyVal []string
		err    error
	}{
		{cap: 2, keyVal: []string{"1", "2", "3", "4", "1"}},
		{cap: 1, keyVal: []string{"3", "4", "1"}},
		{cap: 3, keyVal: []string{"1", "2", "3", "4", "1", "2"}},
		{cap: 0, keyVal: []string{"1"}, err: cachePkg.ErrEvictionFailed},
		{cap: -1, keyVal: nil, err: cachePkg.ErrInvalidCapacity},
	}
	for i := range tests {
		test := tests[i]
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			cache, err := newStringCache(test.cap, test.keyVal...)
			if test.err != nil {
				require.ErrorIs(t, err, test.err)
				return
			}
			require.NoError(t, err)

			keyVal := test.keyVal

			back := keyVal[len(keyVal)-test.cap]
			backV := cache.l.Back().Value.(*item[string, string]).Value
			require.Equal(t, back, backV)

			front := keyVal[len(keyVal)-1]
			frontV := cache.l.Front().Value.(*item[string, string]).Value
			require.Equal(t, front, frontV)

			for _, s := range keyVal[len(keyVal)-test.cap:] {
				v := cache.m[s].Value.(*item[string, string]).Value
				require.Equal(t, s, v)
			}
		})
	}
}

func TestCache_Drop(t *testing.T) {
	tests := []struct {
		cap      int
		keyVal   []string
		dropElem string
		result   bool
		err      string
	}{
		{cap: 2, keyVal: []string{"1", "2", "3", "4", "1"}, dropElem: "1", result: true},
		{cap: 3, keyVal: []string{"1", "2", "3", "4", "1", "2"}, dropElem: "4", result: true},
		{cap: 1, keyVal: []string{"3", "4", "1"}, dropElem: "4", result: false},
		{cap: 1, keyVal: []string{"3"}, dropElem: "3", result: true},
	}
	for i := range tests {
		test := tests[i]
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			cache, err := newStringCache(test.cap, test.keyVal...)
			require.NoError(t, err)

			ok, err := cache.Drop(test.dropElem)
			require.NoError(t, err)
			require.Equal(t, test.result, ok)
			require.NotContains(t, cache.m, test.dropElem)
			for i := cache.l.Back(); i != nil && i.Next() != nil; i = i.Next() {
				it := i.Value.(*item[string, string])
				require.NotEqual(t, test.dropElem, it.Key)
				require.NotEqual(t, test.dropElem, it.Value)
			}
		})
	}
}

func TestCache_Len(t *testing.T) {
	tests := []struct {
		cap    int
		len    int
		keyVal []string
	}{
		{cap: 9, len: 6, keyVal: []string{"1", "2", "3", "4", "5", "6"}},
		{cap: 9, len: 4, keyVal: []string{"1", "2", "3", "4", "1", "2"}},
		{cap: 2, len: 2, keyVal: []string{"1", "2", "3", "4", "1"}},
		{cap: 2, len: 1, keyVal: []string{"1", "1", "1", "1", "1"}},
		{cap: 1, len: 1, keyVal: []string{"3", "4", "1"}},
		{cap: 1, len: 0, keyVal: nil},
	}
	for i := range tests {
		test := tests[i]
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			cache, err := newStringCache(test.cap, test.keyVal...)
			require.NoError(t, err)

			l, err := cache.Len()
			require.NoError(t, err)
			require.Equal(t, test.len, l)
			require.Len(t, cache.m, test.len)
			require.Equal(t, test.len, cache.l.Len())
		})
	}
}

func TestCache_Clear(t *testing.T) {
	tests := []struct {
		cap    int
		keyVal []string
	}{
		{cap: 10, keyVal: []string{"1", "2", "3", "4", "1", "2"}},
		{cap: 2, keyVal: []string{"1", "2", "3", "4"}},
		{cap: 1, keyVal: nil},
	}
	for i := range tests {
		test := tests[i]
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			cache, err := newStringCache(len(test.keyVal), test.keyVal...)
			require.NoError(t, err)

			err = cache.Clear()
			require.NoError(t, err)

			l, err := cache.Len()
			require.NoError(t, err)
			require.Zero(t, l)
			require.Empty(t, cache.m)
			require.Zero(t, cache.l.Len())
		})
	}
}
