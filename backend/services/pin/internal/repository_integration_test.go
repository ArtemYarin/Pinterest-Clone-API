//go:build integration

package pin

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testPool *pgxpool.Pool

// TestMain starts a single Postgres container for the package and applies the
// real migrations, so tests run against the same schema as production.
func TestMain(m *testing.M) {
	os.Exit(runWithPostgres(m))
}

func runWithPostgres(m *testing.M) int {
	ctx := context.Background()

	ctr, err := postgres.Run(ctx, "postgres:18",
		postgres.WithDatabase("pins_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("terminate postgres container: %v", err)
		}
	}()
	if err != nil {
		log.Printf("start postgres container: %v", err)
		return 1
	}

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("postgres connection string: %v", err)
		return 1
	}
	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		log.Printf("connect to postgres: %v", err)
		return 1
	}
	defer testPool.Close()

	if err := applyMigrations(ctx, testPool); err != nil {
		log.Printf("apply migrations: %v", err)
		return 1
	}

	return m.Run()
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	files, err := filepath.Glob(filepath.Join("migrations", "*.up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return err
		}
	}
	return nil
}

// newTestRepository returns a repository over an empty pins table.
func newTestRepository(t *testing.T) PinRepository {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE pins")
	require.NoError(t, err)
	return NewPinRepository(testPool)
}

// seedPin inserts a pin and optionally adjusts columns the repository can't set directly.
func seedPin(t *testing.T, repo PinRepository, owner uuid.UUID, title, status string, likes int, age time.Duration) *PinResponse {
	t.Helper()
	ctx := context.Background()

	p, err := repo.CreatePin(ctx, owner, "pins/"+owner.String()+"/"+uuid.NewString(), CreatePinRequest{Title: title})
	require.NoError(t, err)

	_, err = testPool.Exec(ctx,
		"UPDATE pins SET image_status = $1, likes = $2, created_at = NOW() - $3::interval WHERE id = $4",
		status, likes, age.String(), p.Id)
	require.NoError(t, err)
	return p
}

func TestPinRepository_CreateAndGet(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()
	owner := uuid.New()
	desc := "Golden hour"

	created, err := repo.CreatePin(ctx, owner, "pins/key", CreatePinRequest{Title: "Sunset", Description: &desc})
	require.NoError(t, err)

	got, err := repo.GetPinByID(ctx, created.Id.String())
	require.NoError(t, err)

	assert.Equal(t, owner, got.User_id)
	assert.Equal(t, "Sunset", got.Title)
	assert.Equal(t, "pins/key", got.Image_url)
	assert.Equal(t, &desc, got.Description)
	// Column defaults from the migrations
	assert.Equal(t, "pending", got.Image_status)
	assert.Equal(t, 0, got.Likes)
	assert.WithinDuration(t, time.Now(), got.Created_at, time.Minute)
}

func TestPinRepository_Errors(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()
	existing, err := repo.CreatePin(ctx, uuid.New(), "pins/taken", CreatePinRequest{Title: "Taken"})
	require.NoError(t, err)
	missingID := uuid.New()
	title := "New title"

	tests := []struct {
		name    string
		run     func() error
		wantErr error
	}{
		{
			name: "duplicate image url",
			run: func() error {
				_, err := repo.CreatePin(ctx, uuid.New(), existing.Image_url, CreatePinRequest{Title: "Copy"})
				return err
			},
			wantErr: errImageURLExists,
		},
		{
			name: "get missing pin",
			run: func() error {
				_, err := repo.GetPinByID(ctx, missingID.String())
				return err
			},
			wantErr: errPinNotFound,
		},
		{
			name:    "update missing pin",
			run:     func() error { return repo.UpdatePin(ctx, UpdatePinRequest{Id: missingID, Title: &title}) },
			wantErr: errPinNotFound,
		},
		{
			name:    "update without fields",
			run:     func() error { return repo.UpdatePin(ctx, UpdatePinRequest{Id: existing.Id}) },
			wantErr: errBadRequest,
		},
		{
			name:    "delete missing pin",
			run:     func() error { return repo.DeletePin(ctx, missingID.String()) },
			wantErr: errPinNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, tt.run(), tt.wantErr)
		})
	}
}

func TestPinRepository_GetPins(t *testing.T) {
	repo := newTestRepository(t)
	alice, bob := uuid.New(), uuid.New()
	seedPin(t, repo, alice, "Sunset beach", "confirmed", 3, 0)
	seedPin(t, repo, alice, "City night", "confirmed", 2, 0)
	seedPin(t, repo, bob, "Mountain sunset", "confirmed", 1, 0)
	seedPin(t, repo, bob, "Pending sunset", "pending", 0, 0) // never listed

	byLikes := PinFilters{SortBy: "likes", SortOrder: "asc", Limit: 20}
	with := func(modify func(f *PinFilters)) PinFilters {
		f := byLikes
		modify(&f)
		return f
	}

	tests := []struct {
		name       string
		filters    PinFilters
		wantTitles []string
		wantCount  int
	}{
		{
			name:       "only confirmed pins",
			filters:    byLikes,
			wantTitles: []string{"Mountain sunset", "City night", "Sunset beach"},
			wantCount:  3,
		},
		{
			name:       "search is case-insensitive",
			filters:    with(func(f *PinFilters) { f.Search = "SUNSET" }),
			wantTitles: []string{"Mountain sunset", "Sunset beach"},
			wantCount:  2,
		},
		{
			name:       "filter by user",
			filters:    with(func(f *PinFilters) { f.UserID = alice.String() }),
			wantTitles: []string{"City night", "Sunset beach"},
			wantCount:  2,
		},
		{
			name:       "sort descending",
			filters:    with(func(f *PinFilters) { f.SortOrder = "desc" }),
			wantTitles: []string{"Sunset beach", "City night", "Mountain sunset"},
			wantCount:  3,
		},
		{
			// count reports the total across all pages, not the page size
			name:       "pagination",
			filters:    with(func(f *PinFilters) { f.Limit, f.Offset = 1, 1 }),
			wantTitles: []string{"City night"},
			wantCount:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pins, count, err := repo.GetPins(context.Background(), tt.filters)
			require.NoError(t, err)

			titles := make([]string, len(pins))
			for i, p := range pins {
				titles[i] = p.Title
			}
			assert.Equal(t, tt.wantTitles, titles)
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestPinRepository_UpdatePin_Partial(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()
	desc := "Original description"
	p, err := repo.CreatePin(ctx, uuid.New(), "pins/key", CreatePinRequest{Title: "Old title", Description: &desc})
	require.NoError(t, err)

	title := "New title"
	require.NoError(t, repo.UpdatePin(ctx, UpdatePinRequest{Id: p.Id, Title: &title}))

	got, err := repo.GetPinByID(ctx, p.Id.String())
	require.NoError(t, err)
	assert.Equal(t, "New title", got.Title)
	assert.Equal(t, &desc, got.Description, "fields not in the request must be left untouched")
	assert.False(t, got.Updated_at.Before(p.Updated_at))
}

func TestPinRepository_GetStalePendingPins(t *testing.T) {
	repo := newTestRepository(t)
	owner := uuid.New()
	oldest := seedPin(t, repo, owner, "Oldest", "pending", 0, 3*time.Hour)
	old := seedPin(t, repo, owner, "Old", "pending", 0, 2*time.Hour)
	seedPin(t, repo, owner, "Fresh", "pending", 0, 0)
	seedPin(t, repo, owner, "Old but confirmed", "confirmed", 0, 3*time.Hour)

	cutoff := time.Now().Add(-time.Hour)

	tests := []struct {
		name    string
		limit   int
		wantIDs []uuid.UUID
	}{
		{name: "oldest first, fresh and confirmed excluded", limit: 10, wantIDs: []uuid.UUID{oldest.Id, old.Id}},
		{name: "respects limit", limit: 1, wantIDs: []uuid.UUID{oldest.Id}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stale, err := repo.GetStalePendingPins(context.Background(), cutoff, tt.limit)
			require.NoError(t, err)

			ids := make([]uuid.UUID, len(stale))
			for i, p := range stale {
				ids[i] = p.Id
			}
			assert.Equal(t, tt.wantIDs, ids)
		})
	}
}
