package router

import (
	"net/http"

	"github.com/abulhanifah/weather-monitoring/internal/config"
	"github.com/abulhanifah/weather-monitoring/internal/handler"
	"github.com/abulhanifah/weather-monitoring/internal/middleware"
	"github.com/abulhanifah/weather-monitoring/internal/repositories"
	"github.com/abulhanifah/weather-monitoring/internal/services"
	"gorm.io/gorm"
)

// New mengembalikan HTTP router yang sudah dikonfigurasi
func New(cfg *config.Config, database *gorm.DB) *http.ServeMux {
	mux := http.NewServeMux()

	// Dependencies
	userRepo := repositories.NewUserRepository(database)
	userSvc := services.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(cfg, userSvc)

	deviceRepo := repositories.NewRepository(database)
	deviceSvc := services.NewService(cfg, deviceRepo)
	deviceHandler := handler.NewHandler(deviceSvc)

	// ----------------------------------------------------
	// 1. PUBLIC ROUTES (Tanpa Auth)
	// ----------------------------------------------------
	mux.HandleFunc("GET /health", handler.HealthCheck)
	mux.HandleFunc("POST /api/v1/login", userHandler.Login)

	// ----------------------------------------------------
	// 2. JWT ROUTES (Membutuhkan Auth JWT) - ADMIN
	// ----------------------------------------------------
	jwtAuth := middleware.JWTAuthMiddleware(cfg.AuthSecret)
	mux.Handle("GET /api/v1/devices", jwtAuth(http.HandlerFunc(deviceHandler.ListDevices)))

	// ----------------------------------------------------
	// 3. API KEY ROUTES (Membutuhkan Auth API Key) - Ingest Device
	// ----------------------------------------------------

	return mux
}
