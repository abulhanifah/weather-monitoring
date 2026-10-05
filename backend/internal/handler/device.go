package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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

// ListDevices handler untuk GET /api/v1/devices.
// Query params: page (default 1), limit (default 10),
// sort (kolom devices, atau "location asc|desc" untuk sort nama lokasi),
// filter: status, name, id, location_id,
// q (search ILIKE di device name, location name, device id).
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
		Data:  devices,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}
