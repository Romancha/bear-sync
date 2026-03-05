package main

import (
	"context"
	"log/slog"
	"time"
)

// runDaemon runs the bridge sync loop continuously until the context is cancelled.
// Errors in individual sync cycles are logged but do not stop the daemon.
func runDaemon(ctx context.Context, bridge *Bridge, interval time.Duration, logger *slog.Logger) error {
	// Run the first sync immediately.
	logger.Info("daemon: running initial sync")
	if err := bridge.Run(ctx); err != nil {
		logger.Error("daemon: sync cycle failed", "error", err)
	} else {
		logger.Info("daemon: sync cycle completed")
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("daemon: shutting down")
			return nil
		case <-ticker.C:
			logger.Info("daemon: starting sync cycle")
			if err := bridge.Run(ctx); err != nil {
				logger.Error("daemon: sync cycle failed", "error", err)
			} else {
				logger.Info("daemon: sync cycle completed")
			}
		}
	}
}
