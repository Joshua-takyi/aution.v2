package service

import (
	"context"
	"log/slog"
	"time"
)

type WorkerService struct {
	auctionService *AuctionService
	logger         *slog.Logger
	stopChan       chan struct{}
}

func NewWorkerService(auctionService *AuctionService, logger *slog.Logger) *WorkerService {
	return &WorkerService{
		auctionService: auctionService,
		logger:         logger,
		stopChan:       make(chan struct{}),
	}
}

func (w *WorkerService) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	w.logger.Info("Auction status worker started", "interval", interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				w.RunUpdate()
			case <-w.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (w *WorkerService) RunUpdate() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := w.auctionService.UpdateAuctionStatuses(ctx)
	if err != nil {
		w.logger.Error("Failed to update auction statuses", "error", err)
		return
	}

	scheduledToLive := result["scheduled_to_live"].(float64)
	liveToEnded := result["live_to_ended"].(float64)

	if scheduledToLive > 0 || liveToEnded > 0 {
		w.logger.Info("Auction statuses updated",
			"scheduled_to_live", scheduledToLive,
			"live_to_ended", liveToEnded,
		)
	}
}

func (w *WorkerService) Stop() {
	close(w.stopChan)
}
