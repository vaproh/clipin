package worker

import (
	"context"
	"log"
	"time"

	"clipin/apps/api/internal/service"
)

// StartPayoutProcessorWorker runs a background goroutine that periodically
// processes pending payout requests. The goroutine respects context
// cancellation for graceful shutdown.
func StartPayoutProcessorWorker(ctx context.Context, svc *service.PayoutService, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("payout processor worker started (interval: %s)", interval)

		for {
			select {
			case <-ctx.Done():
				log.Println("payout processor worker stopped")
				return
			case <-ticker.C:
				count, err := svc.ProcessPendingPayouts(ctx)
				if err != nil {
					log.Printf("payout processor worker error: %v", err)
				} else if count > 0 {
					log.Printf("payout processor worker: processed %d payouts", count)
				}
			}
		}
	}()
}
