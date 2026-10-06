package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/abulhanifah/weather-monitoring/internal/middleware"
	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/internal/services"
	"github.com/abulhanifah/weather-monitoring/pkg"
)

// Telemetry handler untuk POST /api/v1/ingest/telemetry (auth API key device).
// Simpan readings telemetry ke sensor readings berdasar tipe terpasang aktif.
// Telemetry godoc
// @Summary Ingest telemetry device
// @Tags Ingest
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body models.TelemetryRequest true "Payload telemetry"
// @Success 200 {object} models.TelemetryResponse
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 403 {object} models.ErrorEnvelope
// @Failure 500 {object} models.ErrorEnvelope
// @Router /api/v1/ingest/telemetry [post]
func (h *SensorInstallationHandler) Telemetry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	deviceID, ok := middleware.DeviceIDFromContext(ctx)
	if !ok {
		pkg.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error": "Unauthorized: Invalid API Key",
		})
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid payload",
		})
		return
	}

	res, err := h.svc.IngestTelemetry(ctx, deviceID, payload)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrDeviceMismatch):
			pkg.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
				"error": err.Error(),
			})
		default:
			pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}
		return
	}

	pkg.WriteJSON(w, http.StatusOK, models.TelemetryResponse{
		Message: "OK",
		Saved:   res.Saved,
		Skipped: res.Skipped,
	})
}

// TelemetryBatch handler untuk POST /api/v1/ingest/telemetry/batch (auth API key).
// Simpan banyak paket telemetry sekaligus. device_id di root harus cocok api key;
// tiap entri batch wajib punya ts dan readings.
// TelemetryBatch godoc
// @Summary Ingest telemetry batch
// @Tags Ingest
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body models.TelemetryBatchRequest true "Payload telemetry batch"
// @Success 200 {object} models.TelemetryResponse
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 403 {object} models.ErrorEnvelope
// @Failure 500 {object} models.ErrorEnvelope
// @Router /api/v1/ingest/telemetry/batch [post]
func (h *SensorInstallationHandler) TelemetryBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	deviceID, ok := middleware.DeviceIDFromContext(ctx)
	if !ok {
		pkg.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error": "Unauthorized: Invalid API Key",
		})
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid payload",
		})
		return
	}

	res, err := h.svc.IngestTelemetryBatch(ctx, deviceID, payload)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrDeviceMismatch):
			pkg.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
				"error": err.Error(),
			})
		default:
			pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}
		return
	}

	pkg.WriteJSON(w, http.StatusOK, models.TelemetryResponse{
		Message: "OK",
		Saved:   res.Saved,
		Skipped: res.Skipped,
	})
}
