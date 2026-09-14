package pin

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// CleanupWorker periodically reclaims pins whose image upload was never
// confirmed within ttl - both the orphaned MinIO object (if any) and the
// pending pin row.
type CleanupWorker struct {
	repo      PinRepository
	storage   *ImageStorage
	interval  time.Duration
	ttl       time.Duration
	batchSize int
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func NewCleanupWorker(repo PinRepository, storage *ImageStorage, interval, ttl time.Duration, batchSize int) *CleanupWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &CleanupWorker{
		repo:      repo,
		storage:   storage,
		interval:  interval,
		ttl:       ttl,
		batchSize: batchSize,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (w *CleanupWorker) Start() {
	ticker := time.NewTicker(w.interval)
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-ticker.C:
				w.runCycle(w.ctx)
			case <-w.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop cancels the worker and blocks until any in-flight cleanup cycle
// finishes, so callers can safely close shared resources (DB pool, MinIO
// client) right after it returns.
func (w *CleanupWorker) Stop() {
	w.cancel()
	w.wg.Wait()
}

func (w *CleanupWorker) runCycle(ctx context.Context) {
	cutoff := time.Now().Add(-w.ttl)
	stale, err := w.repo.GetStalePendingPins(ctx, cutoff, w.batchSize)
	if err != nil {
		log.Printf("cleanup: list stale pending pins: %v", err)
		return
	}

	for _, p := range stale {
		if err := w.cleanupPin(ctx, p); err != nil {
			log.Printf("cleanup: pin %s: %v", p.Id, err)
		}
	}
}

func (w *CleanupWorker) cleanupPin(ctx context.Context, p StalePin) error {
	// Remove the object first: if this fails, leave the row pending so
	// it's retried next cycle instead of orphaning the object in MinIO.
	if err := w.storage.RemoveObject(ctx, p.Image_url); err != nil && !isMinioNotFoundErr(err) {
		return fmt.Errorf("remove image object: %w", err)
	}

	if err := w.repo.DeletePin(ctx, p.Id.String()); err != nil {
		return fmt.Errorf("delete pin row: %w", err)
	}

	return nil
}
