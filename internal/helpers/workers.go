package helpers

import (
	"context"
	"log/slog"
	"time"
)

// StartCleanupWorker initializes a background worker that runs the provided cleanup task every 5 minutes.
func StartCleanupWorker(ctx context.Context, logger *slog.Logger, cleanupTask func(context.Context) error) {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logger.Info("Starting cleanup of expired tokens")
				if err := cleanupTask(ctx); err != nil {
					logger.Error("Cleanup worker failed", "error", err)
				} else {
					logger.Info("Cleanup of expired tokens completed")
				}
			}
		}
	}()
}
