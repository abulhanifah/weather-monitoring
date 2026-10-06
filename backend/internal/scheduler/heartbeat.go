package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/abulhanifah/weather-monitoring/internal/config"
	"github.com/abulhanifah/weather-monitoring/internal/services"
	"github.com/go-co-op/gocron/v2"
)

// StartHeartbeatChecker jalankan scheduler yang menandai device Active
// yang tidak mengirim heartbeat melewati threshold menjadi Disconnected.
// Kembalikan scheduler yang sudah start (panggil Shutdown saat stop).
func StartHeartbeatChecker(cfg *config.Config, svc *services.DeviceService) (gocron.Scheduler, error) {
	sch, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	_, err = sch.NewJob(
		gocron.DurationJob(cfg.SchedulerInterval),
		gocron.NewTask(func(ctx context.Context) {
			jobCtx, cancel := context.WithTimeout(ctx, time.Minute)
			defer cancel()

			marked, err := svc.MarkStaleDevicesDisconnected(jobCtx, cfg.HeartbeatThreshold)
			if err != nil {
				slog.ErrorContext(jobCtx, "Heartbeat checker failed", slog.Any("error", err.Error()))
				return
			}
			if marked > 0 {
				slog.InfoContext(jobCtx, "Heartbeat checker done", slog.Int("marked_disconnected", marked))
			}
		}),
	)
	if err != nil {
		return nil, err
	}

	sch.Start()
	return sch, nil
}
