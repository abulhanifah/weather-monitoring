package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/abulhanifah/weather-monitoring/internal/middleware"
	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/internal/services"
	"github.com/abulhanifah/weather-monitoring/pkg"
	"gorm.io/gorm"
)

type DeviceHandler struct {
	svc *services.DeviceService
}

func NewHandler(svc *services.DeviceService) *DeviceHandler {
	return &DeviceHandler{svc: svc}
}

// GetDevice godoc
// @Summary Detail device (termasuk location)
// @Tags Devices
// @Produce json
// @Security BearerAuth
// @Param id path string true "Device ID"
// @Success 200 {object} models.Device
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Router /api/v1/devices/{id} [get]
func (h *DeviceHandler) GetDevice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	dev, err := h.svc.GetByID(ctx, id)
	if err != nil {
		pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"error": "Device not found",
		})
		return
	}

	pkg.WriteJSON(w, http.StatusOK, dev)
}

// CreateDevice godoc
// @Summary Buat device baru
// @Tags Devices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body models.Device true "Device (id, name wajib)"
// @Success 201 {object} models.DeviceCreateEnvelope
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 500 {object} models.ErrorEnvelope
// @Router /api/v1/devices [post]
func (h *DeviceHandler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var dev models.Device

	if err := json.NewDecoder(r.Body).Decode(&dev); err != nil {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid payload",
		})
		return
	}

	res, err := h.svc.CreateDevice(ctx, &dev)
	if err != nil {
		pkg.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to create device",
		})
		return
	}

	pkg.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Device created successfully",
		"data":    res,
	})
}

// UpdateDevice handler untuk PATCH /api/v1/devices/{id}.
// Body: subset dari {name, description, status, location_id}.
// UpdateDevice godoc
// @Summary Partial update device
// @Tags Devices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Device ID"
// @Param body body models.DevicePatch true "Field yang diubah"
// @Success 200 {object} models.DeviceEnvelope
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Router /api/v1/devices/{id} [patch]
func (h *DeviceHandler) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	var data map[string]any
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid payload",
		})
		return
	}

	res, err := h.svc.UpdateDevice(ctx, id, data)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"error": "Device not found",
			})
			return
		}
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	pkg.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Device updated successfully",
		"data":    res,
	})
}

// DeleteDevice handler untuk DELETE /api/v1/devices/{id}.
// DeleteDevice godoc
// @Summary Hapus device
// @Tags Devices
// @Produce json
// @Security BearerAuth
// @Param id path string true "Device ID"
// @Success 200 {object} models.MessageEnvelope
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Router /api/v1/devices/{id} [delete]
func (h *DeviceHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.svc.DeleteDevice(r.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"error": "Device not found",
			})
			return
		}
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	pkg.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Device deleted successfully",
	})
}

// RotateCredentials handler untuk POST /api/v1/devices/{id}/credentials/rotate.
// Me-revoke api key lama lalu mengembalikan raw key baru (hanya 1x).
// RotateCredentials godoc
// @Summary Revoke API key lama dan generate yang baru
// @Tags Devices
// @Produce json
// @Security BearerAuth
// @Param id path string true "Device ID"
// @Success 200 {object} models.RotateEnvelope "raw_key hanya ditampilkan 1x"
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Failure 500 {object} models.ErrorEnvelope
// @Router /api/v1/devices/{id}/credentials/rotate [post]
func (h *DeviceHandler) RotateCredentials(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	res, err := h.svc.RotateDeviceAPIKey(r.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"error": "Device not found",
			})
			return
		}
		if err.Error() == "device id is required" {
			pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		pkg.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to rotate device credentials",
		})
		return
	}

	pkg.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Device credentials rotated successfully",
		"data":    res,
	})
}

// ListDevices handler untuk GET /api/v1/devices.
// Query params: page (default 1), limit (default 10),
// sort (kolom devices, atau "location asc|desc" untuk sort nama lokasi),
// filter: status, name, id, location_id,
// q (search ILIKE di device name, location name, device id).
// ListDevices godoc
// @Summary List devices (paginated)
// @Tags Devices
// @Produce json
// @Security BearerAuth
// @Param page query int false "Halaman" default(1) minimum(1)
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param sort query string false "Kolom devices atau 'location asc|desc'" default(created_at desc)
// @Param status query string false "Filter status" Enums(active, offline, degraded, maintenance, faulty, disabled)
// @Param name query string false "Filter nama (ILIKE)"
// @Param id query string false "Filter device ID"
// @Param location_id query int false "Filter location ID"
// @Param q query string false "Search ILIKE di device name, location name, device id"
// @Success 200 {object} models.DeviceListResponse
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 500 {object} models.ErrorEnvelope
// @Router /api/v1/devices [get]
func (h *DeviceHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	page, err := strconv.Atoi(q.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}

	sort := q.Get("sort")
	if sort == "" {
		sort = "created_at desc"
	}

	filter := map[string]any{}
	if v := q.Get("status"); v != "" {
		filter["status"] = v
	}
	if v := q.Get("name"); v != "" {
		filter["name"] = v
	}
	if v := q.Get("id"); v != "" {
		filter["id"] = v
	}
	if v := q.Get("location_id"); v != "" {
		filter["location_id"] = v
	}
	if v := q.Get("q"); v != "" {
		filter["q"] = v
	}

	params := map[string]any{
		"page":   page,
		"limit":  limit,
		"sort":   sort,
		"filter": filter,
	}

	devices, total, err := h.svc.GetPaginated(ctx, params)
	if err != nil {
		pkg.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to fetch devices",
		})
		return
	}

	if devices == nil {
		devices = []models.Device{}
	}

	pkg.WriteJSON(w, http.StatusOK, models.DeviceListResponse{
		Data:      devices,
		Page:      page,
		Limit:     limit,
		Total:     total,
		TotalPage: pkg.TotalPages(total, limit),
	})
}

// Heartbeat handler untuk POST /api/v1/ingest/heartbeat (auth API key device).
// Ubah status device jadi Active, simpan payload ke LatestHealth,
// dan catat device status history.
// Heartbeat godoc
// @Summary Device heartbeat
// @Tags Ingest
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body models.HeartbeatRequest true "Payload telemetri (device_id dan ts wajib, field lain bebas)"
// @Success 200 {object} models.MessageEnvelope
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 403 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Failure 500 {object} models.ErrorEnvelope
// @Router /api/v1/ingest/heartbeat [post]
func (h *DeviceHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
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

	_, err := h.svc.Heartbeat(ctx, deviceID, payload)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"error": "Device not found",
			})
		case errors.Is(err, services.ErrDeviceMismatch):
			pkg.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
				"error": err.Error(),
			})
		case errors.Is(err, services.ErrHeartbeatValidation):
			pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		default:
			pkg.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"error": "Failed to process heartbeat",
			})
		}
		return
	}

	pkg.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "OK",
	})
}

// HealthHistory handler untuk GET /api/v1/devices/{id}/health.
// Mengembalikan histori status device, terbaru dulu.
// HealthHistory godoc
// @Summary Histori status device
// @Tags Devices
// @Produce json
// @Security BearerAuth
// @Param id path string true "Device ID"
// @Param limit query int false "Limit" default(50) minimum(1) maximum(200)
// @Success 200 {object} models.DeviceHealthResponse
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Failure 500 {object} models.ErrorEnvelope
// @Router /api/v1/devices/{id}/health [get]
func (h *DeviceHandler) HealthHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 {
		limit = 50
	}

	hists, err := h.svc.GetStatusHistory(ctx, id, limit)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"error": "Device not found",
			})
			return
		}
		if err.Error() == "device id is required" {
			pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		pkg.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to fetch device health history",
		})
		return
	}

	if hists == nil {
		hists = []models.DeviceStatusHistory{}
	}

	pkg.WriteJSON(w, http.StatusOK, models.DeviceHealthResponse{
		DeviceID: id,
		Total:    len(hists),
		Data:     hists,
	})
}
