package handler

import (
	"net/http"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/pkg"
)

// HealthCheck godoc
// @Summary Health check
// @Tags System
// @Produce json
// @Success 200 {object} models.HealthEnvelope
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	pkg.WriteJSON(w, http.StatusOK, models.HealthEnvelope{
		Status: "ok",
	})
}
