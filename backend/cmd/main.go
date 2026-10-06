package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abulhanifah/weather-monitoring/internal/config"
	"github.com/abulhanifah/weather-monitoring/internal/db"
	"github.com/abulhanifah/weather-monitoring/internal/middleware"
	"github.com/abulhanifah/weather-monitoring/internal/repositories"
	"github.com/abulhanifah/weather-monitoring/internal/router"
	"github.com/abulhanifah/weather-monitoring/internal/scheduler"
	"github.com/abulhanifah/weather-monitoring/internal/services"
)

// @title Weather Monitoring API
// @version 1.0.0
// @description API monitoring cuaca (devices, locations, auth JWT).
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	withSeed := flag.Bool("with-seed", false, "jalankan seeder sebelum server start")
	flag.Parse()

	// Load config from env
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database := db.InitPostgres(cfg)

	if *withSeed {
		if err := db.Seed(database); err != nil {
			log.Fatalf("Seeding gagal: %v", err)
		}
	}

	r := router.New(cfg, database)

	// Scheduler cek heartbeat basi (jalan tiap SCHEDULER_INTERVAL)
	deviceSvc := services.NewService(cfg, repositories.NewRepository(database))
	sch, err := scheduler.StartHeartbeatChecker(cfg, deviceSvc)
	if err != nil {
		log.Fatalf("Scheduler gagal start: %v", err)
	}
	defer func() {
		_ = sch.Shutdown()
	}()
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: middleware.CORSMiddleware(cfg.CORSAllowedOrigins)(r),
	}

	go func() {
		log.Printf("REST Server berjalan di http://localhost:%s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful Shutdown
	<-ctx.Done()
	log.Println("Mematikan service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Service selesai dihentikan.")
}
