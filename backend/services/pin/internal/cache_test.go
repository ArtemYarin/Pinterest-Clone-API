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
	testPinTTL  = 5 * time.Minute
	testListTTL = time.Minute
)

func newTestCache(t *testing.T) (PinCache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })
	return NewRedisPinCache(client, testPinTTL, testListTTL), mr
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
	cache.SetPins(ctx, PinFilters{}, []PinResponse{*p}, 1)
	_, ok := cache.GetPin(ctx, p.Id.String())
	assert.False(t, ok)
	_, _, ok = cache.GetPins(ctx, PinFilters{})
	assert.False(t, ok)
}
