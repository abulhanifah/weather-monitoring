package router

import (
	"net/http"

	"github.com/abulhanifah/weather-monitoring/internal/config"
	"github.com/abulhanifah/weather-monitoring/internal/handler"
	"github.com/abulhanifah/weather-monitoring/internal/middleware"
	"github.com/abulhanifah/weather-monitoring/internal/repositories"
	"github.com/abulhanifah/weather-monitoring/internal/services"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"

	_ "github.com/abulhanifah/weather-monitoring/docs"
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

	locationRepo := repositories.NewLocationRepository(database)
	locationSvc := services.NewLocationService(locationRepo)
	locationHandler := handler.NewLocationHandler(locationSvc)

	// ----------------------------------------------------
	// 1. PUBLIC ROUTES (Tanpa Auth)
	// ----------------------------------------------------
	mux.HandleFunc("GET /health", handler.HealthCheck)
	mux.HandleFunc("POST /api/v1/login", userHandler.Login)
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	// ----------------------------------------------------
	// 2. JWT ROUTES (Membutuhkan Auth JWT) - ADMIN
	// ----------------------------------------------------
	jwtAuth := middleware.JWTAuthMiddleware(cfg.AuthSecret)
	mux.Handle("GET /api/v1/devices", jwtAuth(http.HandlerFunc(deviceHandler.ListDevices)))
	mux.Handle("POST /api/v1/devices", jwtAuth(http.HandlerFunc(deviceHandler.CreateDevice)))
	mux.Handle("GET /api/v1/devices/{id}", jwtAuth(http.HandlerFunc(deviceHandler.GetDevice)))
	mux.Handle("PATCH /api/v1/devices/{id}", jwtAuth(http.HandlerFunc(deviceHandler.UpdateDevice)))
	mux.Handle("DELETE /api/v1/devices/{id}", jwtAuth(http.HandlerFunc(deviceHandler.DeleteDevice)))
	mux.Handle("POST /api/v1/devices/{id}/credentials/rotate", jwtAuth(http.HandlerFunc(deviceHandler.RotateCredentials)))
	mux.Handle("GET /api/v1/locations", jwtAuth(http.HandlerFunc(locationHandler.ListLocations)))
	mux.Handle("GET /api/v1/locations/{id}", jwtAuth(http.HandlerFunc(locationHandler.GetLocation)))
	mux.Handle("POST /api/v1/locations", jwtAuth(http.HandlerFunc(locationHandler.CreateLocation)))

	// ----------------------------------------------------
	// 3. API KEY ROUTES (Membutuhkan Auth API Key) - Ingest Device
	// ----------------------------------------------------
	apiKeyAuth := middleware.APIKeyAuthMiddleware(deviceRepo, cfg.PrefixAPIKey)
	mux.Handle("POST /api/v1/ingest/heartbeat", apiKeyAuth(http.HandlerFunc(deviceHandler.Heartbeat)))

	return mux
}
