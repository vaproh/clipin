package worker

import (
	"context"
	"log"
	"time"

	"clipin/apps/api/internal/service"
)

// StartAutoApproveWorker runs a background goroutine that periodically
// auto-approves pending submissions past their campaign's auto_approve_hours.
// The goroutine respects context cancellation for graceful shutdown.
func StartAutoApproveWorker(ctx context.Context, svc *service.SubmissionService, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("auto-approve worker started (interval: %s)", interval)

		for {
			select {
			case <-ctx.Done():
				log.Println("auto-approve worker stopped")
				return
			case <-ticker.C:
				tickCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
				count, err := svc.AutoApprove(tickCtx)
				cancel()
				if err != nil {
					log.Printf("auto-approve worker error: %v", err)
				} else if count > 0 {
					log.Printf("auto-approve worker: approved %d submissions", count)
				}
			}
		}
	}()
}
