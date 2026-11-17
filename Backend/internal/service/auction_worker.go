package service

import (
	"context"
	"log"
	"time"

	"github.com/auction-studio/backend/internal/repository"
	"github.com/auction-studio/backend/internal/ws"
)

type AuctionWorker struct {
	itemRepo *repository.ItemRepository
	bidRepo  *repository.BidRepository
	wsHub    *ws.Hub
	interval time.Duration
}

func NewAuctionWorker(itemRepo *repository.ItemRepository, bidRepo *repository.BidRepository, wsHub *ws.Hub, interval time.Duration) *AuctionWorker {
	if interval <= 0 {
		interval = 10 * time.Second
	}
	return &AuctionWorker{
		itemRepo: itemRepo,
		bidRepo:  bidRepo,
		wsHub:    wsHub,
		interval: interval,
	}
}

func (w *AuctionWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processExpiredAuctions(ctx)
		}
	}
}

func (w *AuctionWorker) processExpiredAuctions(ctx context.Context) {
	items, err := w.itemRepo.GetExpiredActiveAuctions(ctx)
	if err != nil {
		log.Printf("AuctionWorker error fetching expired items: %v", err)
		return
	}

	for _, item := range items {
		// Mark as ended
		if err := w.itemRepo.UpdateStatus(ctx, item.ID, "ended"); err != nil {
			log.Printf("AuctionWorker failed to update status for item %s: %v", item.ID, err)
			continue
		}

		// Find winner
		highest, err := w.bidRepo.GetHighestActiveBid(ctx, nil, item.ID)
		winnerName := ""
		winnerID := ""
		var winningAmount int64
		if err == nil && highest != nil {
			winnerName = highest.Bidder
			winnerID = highest.UserID
			winningAmount = highest.Amount
		}

		log.Printf("Auction ended for item %s (%s). Winner: %s (NPR %d)", item.ID, item.Title, winnerName, winningAmount)

		// Broadcast event
		w.wsHub.Broadcast("AUCTION_ENDED", map[string]interface{}{
			"itemId":        item.ID,
			"title":         item.Title,
			"winnerName":    winnerName,
			"winnerId":      winnerID,
			"winningAmount": winningAmount,
		})
	}
}
