package pin

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPinTTL      = 5 * time.Minute
	testListTTL     = time.Minute
	testNotFoundTTL = 30 * time.Second
)

func newTestCache(t *testing.T) (PinCache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })
	return NewRedisPinCache(client, testPinTTL, testListTTL, testNotFoundTTL), mr
}

func TestPinCache_Pin(t *testing.T) {
	ctx := context.Background()
	p := newTestPin(uuid.New(), uuid.New(), "confirmed")

	t.Run("miss", func(t *testing.T) {
		cache, _ := newTestCache(t)
		got, ok := cache.GetPin(ctx, p.Id.String())
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("round trip", func(t *testing.T) {
		cache, mr := newTestCache(t)
		cache.SetPin(ctx, p)

		got, ok := cache.GetPin(ctx, p.Id.String())
		require.True(t, ok)
		assert.Equal(t, p, got)
		assert.Equal(t, testPinTTL, mr.TTL(pinKey(p.Id.String())))
	})

	t.Run("id lookup is case insensitive", func(t *testing.T) {
		cache, _ := newTestCache(t)
		cache.SetPin(ctx, p)

		_, ok := cache.GetPin(ctx, strings.ToUpper(p.Id.String()))
		assert.True(t, ok)
	})

	t.Run("expires after ttl", func(t *testing.T) {
		cache, mr := newTestCache(t)
		cache.SetPin(ctx, p)
		mr.FastForward(testPinTTL)

		_, ok := cache.GetPin(ctx, p.Id.String())
		assert.False(t, ok)
	})

	t.Run("corrupt value is a miss", func(t *testing.T) {
		cache, mr := newTestCache(t)
		require.NoError(t, mr.Set(pinKey(p.Id.String()), "not json"))

		_, ok := cache.GetPin(ctx, p.Id.String())
		assert.False(t, ok)
	})

	t.Run("not found round trip", func(t *testing.T) {
		cache, mr := newTestCache(t)
		cache.SetPinNotFound(ctx, p.Id.String())

		got, ok := cache.GetPin(ctx, p.Id.String())
		assert.True(t, ok)
		assert.Nil(t, got)
		assert.Equal(t, testNotFoundTTL, mr.TTL(pinKey(p.Id.String())))
	})

	t.Run("not found expires after ttl", func(t *testing.T) {
		cache, mr := newTestCache(t)
		cache.SetPinNotFound(ctx, p.Id.String())
		mr.FastForward(testNotFoundTTL)

		_, ok := cache.GetPin(ctx, p.Id.String())
		assert.False(t, ok)
	})

	t.Run("set overwrites not found", func(t *testing.T) {
		cache, _ := newTestCache(t)
		cache.SetPinNotFound(ctx, p.Id.String())
		cache.SetPin(ctx, p)

		got, ok := cache.GetPin(ctx, p.Id.String())
		require.True(t, ok)
		assert.Equal(t, p, got)
	})

	t.Run("delete", func(t *testing.T) {
		cache, mr := newTestCache(t)
		cache.SetPin(ctx, p)
		cache.DeletePin(ctx, strings.ToUpper(p.Id.String()))

		assert.False(t, mr.Exists(pinKey(p.Id.String())))
		_, ok := cache.GetPin(ctx, p.Id.String())
		assert.False(t, ok)
	})

	t.Run("delete missing key is harmless", func(t *testing.T) {
		cache, _ := newTestCache(t)
		assert.NotPanics(t, func() { cache.DeletePin(ctx, p.Id.String()) })
	})
}

func TestPinCache_Pins(t *testing.T) {
	ctx := context.Background()
	owner := uuid.New()
	pins := []PinResponse{
		*newTestPin(uuid.New(), owner, "confirmed"),
		*newTestPin(uuid.New(), owner, "confirmed"),
	}
	filters := PinFilters{SortBy: "created_at", SortOrder: "desc", Limit: 20}

	t.Run("miss", func(t *testing.T) {
		cache, _ := newTestCache(t)
		_, _, ok := cache.GetPins(ctx, filters)
		assert.False(t, ok)
	})

	t.Run("round trip keeps count", func(t *testing.T) {
		cache, mr := newTestCache(t)
		cache.SetPins(ctx, filters, pins, 42)

		got, count, ok := cache.GetPins(ctx, filters)
		require.True(t, ok)
		assert.Equal(t, pins, got)
		assert.Equal(t, 42, count)

		key, _ := pinListKey(filters)
		assert.Equal(t, testListTTL, mr.TTL(key))
	})

	t.Run("different filters use different keys", func(t *testing.T) {
		cache, _ := newTestCache(t)
		cache.SetPins(ctx, filters, pins, len(pins))

		other := filters
		other.Offset = 20
		_, _, ok := cache.GetPins(ctx, other)
		assert.False(t, ok)
	})

	t.Run("expires after ttl", func(t *testing.T) {
		cache, mr := newTestCache(t)
		cache.SetPins(ctx, filters, pins, len(pins))
		mr.FastForward(testListTTL)

		_, _, ok := cache.GetPins(ctx, filters)
		assert.False(t, ok)
	})
}

func TestPinCache_RedisDown(t *testing.T) {
	ctx := context.Background()
	cache, mr := newTestCache(t)
	mr.Close()
	p := newTestPin(uuid.New(), uuid.New(), "confirmed")

	assert.NotPanics(t, func() {
		cache.SetPin(ctx, p)
		cache.SetPinNotFound(ctx, p.Id.String())
		cache.DeletePin(ctx, p.Id.String())
		cache.SetPins(ctx, PinFilters{}, []PinResponse{*p}, 1)
	})
	_, ok := cache.GetPin(ctx, p.Id.String())
	assert.False(t, ok)
	_, _, ok = cache.GetPins(ctx, PinFilters{})
	assert.False(t, ok)
}

func TestNoopPinCache(t *testing.T) {
	ctx := context.Background()
	var cache PinCache = noopPinCache{}
	p := newTestPin(uuid.New(), uuid.New(), "confirmed")

	// Writes are dropped, so lookups miss even right after a set
	cache.SetPin(ctx, p)
	cache.SetPinNotFound(ctx, p.Id.String())
	cache.DeletePin(ctx, p.Id.String())
	cache.SetPins(ctx, PinFilters{}, []PinResponse{*p}, 1)
	_, ok := cache.GetPin(ctx, p.Id.String())
	assert.False(t, ok)
	_, _, ok = cache.GetPins(ctx, PinFilters{})
	assert.False(t, ok)
}
