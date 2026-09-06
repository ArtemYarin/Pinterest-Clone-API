package likes

import (
	"context"
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
	pinIDs, err := w.redis.SMembers(ctx, "likes:dirty").Result()
	if err != nil {
		log.Println("can't get pin IDs from redis: ", err)
	}

	for _, pinID := range pinIDs {
		key := fmt.Sprintf("likes:count:%s", pinID)

		delta, err := w.redis.GetDel(ctx, key).Int64()

		if err != nil && err != redis.Nil {
			continue // leave in dirty set, retry next cycle
		}
		if delta == 0 {
			w.redis.SRem(ctx, "likes:dirty", pinID)
			continue
		}

		_, err = w.db.Exec(ctx,
			"UPDATE like_counts SET count = count + $1 WHERE id = $2",
			delta, pinID)
		if err != nil {
			// Postgres write failed — give the delta back to Redis
			// so it isn't lost, and leave pinID dirty for retry.
			w.redis.IncrBy(ctx, key, delta)
			continue
		}

		w.redis.SRem(ctx, "likes:dirty", pinID)
	}
}
