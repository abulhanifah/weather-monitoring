package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/internal/services"
	"github.com/abulhanifah/weather-monitoring/pkg"
	"gorm.io/gorm"
)

type SensorInstallationHandler struct {
	svc *services.SensorInstallationService
}

func NewSensorInstallationHandler(svc *services.SensorInstallationService) *SensorInstallationHandler {
	return &SensorInstallationHandler{svc: svc}
}

// InstallSensor handler untuk POST /api/v1/devices/{device_id}/sensors/{sensor_id}.
// Memasang sensor pada device. Tiap tipe sensor hanya boleh 1 yang aktif per device.
// InstallSensor godoc
// @Summary Pasang sensor pada device
// @Tags Sensors
// @Produce json
// @Security BearerAuth
// @Param device_id path string true "Device ID"
// @Param sensor_id path int true "Sensor ID" minimum(1)
// @Success 201 {object} models.SensorInstallationEnvelope
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Failure 409 {object} models.ErrorEnvelope
// @Router /api/v1/devices/{device_id}/sensors/{sensor_id} [post]
func (h *SensorInstallationHandler) InstallSensor(w http.ResponseWriter, r *http.Request) {
	deviceID := r.PathValue("device_id")
	sensorID, err := strconv.Atoi(r.PathValue("sensor_id"))
	if err != nil || sensorID < 1 {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid sensor id",
		})
		return
	}

	res, err := h.svc.InstallSensor(r.Context(), deviceID, uint(sensorID))
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"error": "Device or sensor not found",
			})
		case errors.Is(err, services.ErrSensorTypeAlreadyInstalled):
			pkg.WriteJSON(w, http.StatusConflict, map[string]interface{}{
				"error": err.Error(),
			})
		default:
			pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}
		return
	}

	pkg.WriteJSON(w, http.StatusCreated, models.SensorInstallationEnvelope{
		Message: "Sensor installed successfully",
		Data:    *res,
	})
}

// UninstallSensor handler untuk DELETE /api/v1/devices/{device_id}/sensors/{sensor_id}.
// Mengubah status instalasi aktif menjadi false (uninstall, baris tetap ada).
// UninstallSensor godoc
// @Summary Lepas sensor dari device
// @Tags Sensors
// @Produce json
// @Security BearerAuth
// @Param device_id path string true "Device ID"
// @Param sensor_id path int true "Sensor ID" minimum(1)
// @Success 200 {object} models.MessageEnvelope
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Router /api/v1/devices/{device_id}/sensors/{sensor_id} [delete]
func (h *SensorInstallationHandler) UninstallSensor(w http.ResponseWriter, r *http.Request) {
	deviceID := r.PathValue("device_id")
	sensorID, err := strconv.Atoi(r.PathValue("sensor_id"))
	if err != nil || sensorID < 1 {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid sensor id",
		})
		return
	}

	if err := h.svc.UninstallSensor(r.Context(), deviceID, uint(sensorID)); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"error": "Device, sensor, or active installation not found",
			})
		default:
			pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}
		return
	}

	pkg.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Sensor uninstalled successfully",
	})
}

// GetInstallation handler untuk GET /api/v1/installations/{id}.
// GetInstallation godoc
// @Summary Detail instalasi sensor
// @Tags Sensors
// @Produce json
// @Security BearerAuth
// @Param id path int true "Installation ID" minimum(1)
// @Success 200 {object} models.SensorInstallation
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Router /api/v1/installations/{id} [get]
func (h *SensorInstallationHandler) GetInstallation(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid installation id",
		})
		return
	}

	res, err := h.svc.GetInstallationByID(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"error": "Installation not found",
			})
			return
		}
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	pkg.WriteJSON(w, http.StatusOK, res)
}
