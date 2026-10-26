package pin

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash/fnv"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// PinCache is a read-through cache for pin metadata. It stores repository data only -
// presigned download URLs are generated per request so they never outlive their expiry.
// Single pins are invalidated on write; lists are not and rely on a short TTL instead.
type PinCache interface {
	// GetPin reports a hit with a nil pin when the id is cached as not found
	GetPin(ctx context.Context, id string) (*PinResponse, bool)
	SetPin(ctx context.Context, pin *PinResponse)
	SetPinNotFound(ctx context.Context, id string)
	DeletePin(ctx context.Context, id string)
	GetPins(ctx context.Context, filters PinFilters) ([]PinResponse, int, bool)
	SetPins(ctx context.Context, filters PinFilters, pins []PinResponse, count int)
}

// notFoundMarker is stored in place of a pin to cache a miss. It isn't valid JSON,
// so it can never collide with an encoded pin.
const notFoundMarker = "\x00notfound"

type redisPinCache struct {
	client      *redis.Client
	pinTTL      time.Duration
	listTTL     time.Duration
	notFoundTTL time.Duration
}

type cachedPinList struct {
	Pins  []PinResponse `json:"pins"`
	Count int           `json:"count"`
}

func NewRedisPinCache(client *redis.Client, pinTTL, listTTL, notFoundTTL time.Duration) PinCache {
	return &redisPinCache{client: client, pinTTL: pinTTL, listTTL: listTTL, notFoundTTL: notFoundTTL}
}

func (c *redisPinCache) GetPin(ctx context.Context, id string) (*PinResponse, bool) {
	key := pinKey(id)
	data, ok := c.getRaw(ctx, key)
	if !ok {
		return nil, false
	}
	if string(data) == notFoundMarker {
		return nil, true
	}
	var p PinResponse
	if !c.decode(key, data, &p) {
		return nil, false
	}
	return &p, true
}

func (c *redisPinCache) SetPin(ctx context.Context, pin *PinResponse) {
	c.set(ctx, pinKey(pin.Id.String()), pin, c.pinTTL)
}

func (c *redisPinCache) SetPinNotFound(ctx context.Context, id string) {
	key := pinKey(id)
	if err := c.client.Set(ctx, key, notFoundMarker, c.notFoundTTL).Err(); err != nil {
		log.Printf("pin cache set %s: %v", key, err)
	}
}

// DeletePin fails open: if the delete fails the entry stays stale until its TTL expires
func (c *redisPinCache) DeletePin(ctx context.Context, id string) {
	key := pinKey(id)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		log.Printf("pin cache delete %s: %v", key, err)
	}
}

func (c *redisPinCache) GetPins(ctx context.Context, filters PinFilters) ([]PinResponse, int, bool) {
	key, ok := pinListKey(filters)
	if !ok {
		return nil, 0, false
	}
	var l cachedPinList
	if !c.get(ctx, key, &l) {
		return nil, 0, false
	}
	return l.Pins, l.Count, true
}

func (c *redisPinCache) SetPins(ctx context.Context, filters PinFilters, pins []PinResponse, count int) {
	key, ok := pinListKey(filters)
	if !ok {
		return
	}
	c.set(ctx, key, cachedPinList{Pins: pins, Count: count}, c.listTTL)
}

// Helpers

// get fails open: any Redis or decoding error is treated as a miss so reads fall back to the DB
func (c *redisPinCache) get(ctx context.Context, key string, dst any) bool {
	data, ok := c.getRaw(ctx, key)
	return ok && c.decode(key, data, dst)
}

func (c *redisPinCache) getRaw(ctx context.Context, key string) ([]byte, bool) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Printf("pin cache get %s: %v", key, err)
		}
		return nil, false
	}
	return data, true
}

func (c *redisPinCache) decode(key string, data []byte, dst any) bool {
	if err := json.Unmarshal(data, dst); err != nil {
		log.Printf("pin cache decode %s: %v", key, err)
		return false
	}
	return true
}

func (c *redisPinCache) set(ctx context.Context, key string, val any, ttl time.Duration) {
	data, err := json.Marshal(val)
	if err != nil {
		log.Printf("pin cache encode %s: %v", key, err)
		return
	}
	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Printf("pin cache set %s: %v", key, err)
	}
}

func pinKey(id string) string {
	return "pin:" + strings.ToLower(id)
}

// pinListKey hashing the filters bounds the key's length
// Slices order matters
func pinListKey(filters PinFilters) (string, bool) {
	h := fnv.New128a()
	if err := json.NewEncoder(h).Encode(filters); err != nil {
		return "", false
	}
	return "pins:list:" + hex.EncodeToString(h.Sum(nil)), true
}

// noopPinCache disables caching: every lookup misses and writes are dropped.
// Used where caching is irrelevant, e.g. tests that don't exercise it.
type noopPinCache struct{}

func (noopPinCache) GetPin(context.Context, string) (*PinResponse, bool) { return nil, false }
func (noopPinCache) SetPin(context.Context, *PinResponse)                {}
func (noopPinCache) SetPinNotFound(context.Context, string)              {}
func (noopPinCache) DeletePin(context.Context, string)                   {}
func (noopPinCache) GetPins(context.Context, PinFilters) ([]PinResponse, int, bool) {
	return nil, 0, false
}
func (noopPinCache) SetPins(context.Context, PinFilters, []PinResponse, int) {}
