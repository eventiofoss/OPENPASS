package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/service"
)

const paymentSweepInterval = time.Minute

// startPaymentSweeper launches a background loop that periodically runs the
// abandoned-reservation sweeper. Shutdown stops scheduling new sweeps, but any
// in-flight database transaction is allowed to finish cleanly.
func startPaymentSweeper(
	ctx context.Context,
	wg *sync.WaitGroup,
	paymentSvc *service.PaymentService,
	interval time.Duration,
) {
	if paymentSvc == nil {
		return
	}

	wg.Add(1)

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		slog.Info(
			"Payment sweeper started",
			slog.Duration("interval", interval),
		)

		for {
			select {
			case <-ctx.Done():
				slog.Info("Payment sweeper stopped")
				return
			case <-ticker.C:
				// Allow the active sweep to finish even if shutdown begins while
				// the transaction is in progress, but bound the shutdown wait so
				// a hung database call cannot block termination forever.
				runCtx, cancel := context.WithTimeout(
					context.WithoutCancel(ctx),
					30*time.Second,
				)

				if err := paymentSvc.RunSweeper(runCtx); err != nil {
					slog.Error(
						"Payment sweeper run failed",
						slog.String("error", err.Error()),
					)
				}
				cancel()
			}
		}
	}()
}
