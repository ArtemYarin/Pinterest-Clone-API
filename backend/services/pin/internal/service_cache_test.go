package pin

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newTestServiceWithCache wires a pinService with fresh mocks and a miniredis-backed cache
func newTestServiceWithCache(t *testing.T) (*pinService, *mockPinRepository, *mockImageStore, *miniredis.Miniredis) {
	t.Helper()
	svc, repo, store := newTestService(t)
	cache, mr := newTestCache(t)
	svc.cache = cache
	return svc, repo, store, mr
}

func TestPinService_GetPinByID_Cache(t *testing.T) {
	ctx := context.Background()
	pinID := uuid.New()
	owner := uuid.New()

	t.Run("second call is served from cache", func(t *testing.T) {
		svc, repo, store, _ := newTestServiceWithCache(t)
		p := newTestPin(pinID, owner, "confirmed")
		repo.On("GetPinByID", mock.Anything, pinID.String()).Return(p, nil).Once()
		// Download URLs are never cached, so one is generated per call
		store.On("GenerateDownloadURL", mock.Anything, p.Image_url, time.Hour).Return("https://download.example", nil).Twice()

		for range 2 {
			resp, err := svc.GetPinByID(ctx, pinID.String())
			require.NoError(t, err)
			assert.Equal(t, *p, resp.Pin)
			assert.Equal(t, "https://download.example", resp.Download_url)
		}
	})

	t.Run("pending pin is not cached", func(t *testing.T) {
		svc, repo, _, mr := newTestServiceWithCache(t)
		repo.On("GetPinByID", mock.Anything, pinID.String()).Return(newTestPin(pinID, owner, "pending"), nil).Twice()

		for range 2 {
			_, err := svc.GetPinByID(ctx, pinID.String())
			require.NoError(t, err)
		}
		assert.False(t, mr.Exists(pinKey(pinID.String())))
	})

	t.Run("not found is not cached", func(t *testing.T) {
		svc, repo, _, mr := newTestServiceWithCache(t)
		repo.On("GetPinByID", mock.Anything, pinID.String()).Return(nil, errPinNotFound).Once()

		_, err := svc.GetPinByID(ctx, pinID.String())
		require.ErrorIs(t, err, errPinNotFound)
		assert.Empty(t, mr.Keys())
	})

	t.Run("refetches after ttl", func(t *testing.T) {
		svc, repo, store, mr := newTestServiceWithCache(t)
		p := newTestPin(pinID, owner, "confirmed")
		repo.On("GetPinByID", mock.Anything, pinID.String()).Return(p, nil).Twice()
		store.On("GenerateDownloadURL", mock.Anything, p.Image_url, time.Hour).Return("https://download.example", nil).Twice()

		_, err := svc.GetPinByID(ctx, pinID.String())
		require.NoError(t, err)
		mr.FastForward(testPinTTL)
		_, err = svc.GetPinByID(ctx, pinID.String())
		require.NoError(t, err)
	})

	t.Run("falls back to repository when redis is down", func(t *testing.T) {
		svc, repo, store, mr := newTestServiceWithCache(t)
		mr.Close()
		p := newTestPin(pinID, owner, "confirmed")
		repo.On("GetPinByID", mock.Anything, pinID.String()).Return(p, nil).Once()
		store.On("GenerateDownloadURL", mock.Anything, p.Image_url, time.Hour).Return("https://download.example", nil).Once()

		resp, err := svc.GetPinByID(ctx, pinID.String())
		require.NoError(t, err)
		assert.Equal(t, *p, resp.Pin)
	})
}

func TestPinService_GetPins_Cache(t *testing.T) {
	ctx := context.Background()
	owner := uuid.New()
	pins := []PinResponse{
		*newTestPin(uuid.New(), owner, "confirmed"),
		*newTestPin(uuid.New(), owner, "confirmed"),
	}
	filters := PinFilters{SortBy: "created_at", SortOrder: "desc", Limit: 20}

	t.Run("second call is served from cache", func(t *testing.T) {
		svc, repo, store, _ := newTestServiceWithCache(t)
		repo.On("GetPins", mock.Anything, filters).Return(pins, 42, nil).Once()
		store.On("GenerateDownloadURL", mock.Anything, mock.Anything, time.Hour).Return("https://download.example", nil).Times(2 * len(pins))

		for range 2 {
			resp, count, err := svc.GetPins(ctx, filters)
			require.NoError(t, err)
			assert.Equal(t, 42, count)
			require.Len(t, resp, len(pins))
			for i := range pins {
				assert.Equal(t, pins[i], resp[i].Pin)
				assert.Equal(t, "https://download.example", resp[i].Download_url)
			}
		}
	})

	t.Run("different filters hit the repository", func(t *testing.T) {
		svc, repo, store, _ := newTestServiceWithCache(t)
		other := filters
		other.UserID = owner.String()
		repo.On("GetPins", mock.Anything, filters).Return(pins, len(pins), nil).Once()
		repo.On("GetPins", mock.Anything, other).Return(pins[:1], 1, nil).Once()
		store.On("GenerateDownloadURL", mock.Anything, mock.Anything, time.Hour).Return("https://download.example", nil)

		_, count, err := svc.GetPins(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, len(pins), count)

		_, count, err = svc.GetPins(ctx, other)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("empty result is cached", func(t *testing.T) {
		svc, repo, _, _ := newTestServiceWithCache(t)
		repo.On("GetPins", mock.Anything, filters).Return([]PinResponse{}, 0, nil).Once()

		for range 2 {
			resp, count, err := svc.GetPins(ctx, filters)
			require.NoError(t, err)
			assert.Zero(t, count)
			assert.Empty(t, resp)
		}
	})

	t.Run("falls back to repository when redis is down", func(t *testing.T) {
		svc, repo, store, mr := newTestServiceWithCache(t)
		mr.Close()
		repo.On("GetPins", mock.Anything, filters).Return(pins, len(pins), nil).Once()
		store.On("GenerateDownloadURL", mock.Anything, mock.Anything, time.Hour).Return("https://download.example", nil).Times(len(pins))

		resp, count, err := svc.GetPins(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, len(pins), count)
		assert.Len(t, resp, len(pins))
	})
}
