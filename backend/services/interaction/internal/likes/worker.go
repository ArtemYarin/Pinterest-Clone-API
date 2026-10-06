package likes

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Worker struct {
	db       *pgxpool.Pool
	redis    *redis.Client
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewWorker(pool *pgxpool.Pool, redis *redis.Client, interval time.Duration) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		db:       pool,
		redis:    redis,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (w *Worker) Start() {
	ticker := time.NewTicker(w.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				w.SyncToPostgres(w.ctx)
			case <-w.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.cancel()
}

func (w *Worker) SyncToPostgres(ctx context.Context) {
	// Get all pins whos like count(delta) was changed
	pinIDs, err := w.redis.SMembers(ctx, "likes:dirty").Result()
	if err != nil {
		log.Println("can't get pin IDs from redis: ", err)
		return
	}

	// Loop for each pin
	for _, pinID := range pinIDs {
		key := fmt.Sprintf("likes:count:%s", pinID)

		// Clear the dirty mark before taking the delta.
		if err := w.redis.SRem(ctx, "likes:dirty", pinID).Err(); err != nil {
			log.Printf("can't clear dirty mark for pin %s: %v", pinID, err)
			continue
		}

		// Get and delete delta
		delta, err := w.redis.GetDel(ctx, key).Int64()
		if err != nil && !errors.Is(err, redis.Nil) {
			log.Printf("can't take like delta for pin %s: %v", pinID, err)
			w.restoreDelta(ctx, pinID, key, 0)
			continue
		}
		if delta == 0 {
			continue
		}

		// Insert delta into db.
		_, err = w.db.Exec(ctx,
			`INSERT INTO like_counts (target_id, count, updated_at)
			 VALUES ($2, $1, now())
			 ON CONFLICT (target_id)
			 DO UPDATE SET count = like_counts.count + EXCLUDED.count, updated_at = now()`,
			delta, pinID)
		if err != nil {
			log.Printf("can't sync like delta %d for pin %s: %v", delta, pinID, err)
			w.restoreDelta(ctx, pinID, key, delta)
		}
	}
}

// restoreDelta gives a delta back to Redis and re-marks the pin dirty so it is retried next cycle.
func (w *Worker) restoreDelta(ctx context.Context, pinID, key string, delta int64) {
	_, err := w.redis.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		if delta != 0 {
			pipe.IncrBy(ctx, key, delta)
		}
		pipe.SAdd(ctx, "likes:dirty", pinID)
		return nil
	})
	if err != nil {
		log.Printf("can't restore like delta %d for pin %s: %v", delta, pinID, err)
	}
}
