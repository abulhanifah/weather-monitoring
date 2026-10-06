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

type SensorTypeHandler struct {
	svc *services.SensorTypeService
}

func NewSensorTypeHandler(svc *services.SensorTypeService) *SensorTypeHandler {
	return &SensorTypeHandler{svc: svc}
}

// ListSensorTypes handler untuk GET /api/v1/sensor-types.
// Query params: page (default 1), limit (default 10),
// sort (contoh "name asc"), filter: name, id.
// ListSensorTypes godoc
// @Summary List sensor types (paginated)
// @Tags Sensors
// @Produce json
// @Security BearerAuth
// @Param page query int false "Halaman" default(1) minimum(1)
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param sort query string false "Sort" default(created_at desc)
// @Param name query string false "Filter nama (ILIKE)"
// @Param id query int false "Filter sensor type ID"
// @Success 200 {object} models.SensorTypeListResponse
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 500 {object} models.ErrorEnvelope
// @Router /api/v1/sensor-types [get]
func (h *SensorTypeHandler) ListSensorTypes(w http.ResponseWriter, r *http.Request) {
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
	if v := q.Get("name"); v != "" {
		filter["name"] = v
	}
	if v := q.Get("id"); v != "" {
		filter["id"] = v
	}

	params := map[string]any{
		"page":   page,
		"limit":  limit,
		"sort":   sort,
		"filter": filter,
	}

	types, total, err := h.svc.GetPaginated(ctx, params)
	if err != nil {
		pkg.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to fetch sensor types",
		})
		return
	}

	if types == nil {
		types = []models.SensorType{}
	}

	pkg.WriteJSON(w, http.StatusOK, models.SensorTypeListResponse{
		Data:      types,
		Page:      page,
		Limit:     limit,
		Total:     total,
		TotalPage: pkg.TotalPages(total, limit),
	})
}

// CreateSensorType handler untuk POST /api/v1/sensor-types.
// CreateSensorType godoc
// @Summary Buat sensor type baru
// @Tags Sensors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body models.SensorTypeInput true "Sensor type"
// @Success 201 {object} models.SensorTypeEnvelope
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Router /api/v1/sensor-types [post]
func (h *SensorTypeHandler) CreateSensorType(w http.ResponseWriter, r *http.Request) {
	var st models.SensorType

	if err := json.NewDecoder(r.Body).Decode(&st); err != nil {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid payload",
		})
		return
	}

	res, err := h.svc.CreateSensorType(r.Context(), &st)
	if err != nil {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	pkg.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Sensor type created successfully",
		"data":    res,
	})
}

type SensorHandler struct {
	svc *services.SensorService
}

func NewSensorHandler(svc *services.SensorService) *SensorHandler {
	return &SensorHandler{svc: svc}
}

// ListSensors handler untuk GET /api/v1/sensors.
// Query params: page (default 1), limit (default 10),
// sort, filter: name, id, sensor_type_id, status (true/false).
// ListSensors godoc
// @Summary List sensors (paginated)
// @Tags Sensors
// @Produce json
// @Security BearerAuth
// @Param page query int false "Halaman" default(1) minimum(1)
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param sort query string false "Sort" default(created_at desc)
// @Param name query string false "Filter nama (ILIKE)"
// @Param id query int false "Filter sensor ID"
// @Param sensor_type_id query int false "Filter sensor type ID"
// @Param status query boolean false "Filter status"
// @Success 200 {object} models.SensorListResponse
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 500 {object} models.ErrorEnvelope
// @Router /api/v1/sensors [get]
func (h *SensorHandler) ListSensors(w http.ResponseWriter, r *http.Request) {
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
	if v := q.Get("name"); v != "" {
		filter["name"] = v
	}
	if v := q.Get("id"); v != "" {
		filter["id"] = v
	}
	if v := q.Get("sensor_type_id"); v != "" {
		filter["sensor_type_id"] = v
	}
	if v := q.Get("status"); v != "" {
		if status, err := strconv.ParseBool(v); err == nil {
			filter["status"] = status
		}
	}

	params := map[string]any{
		"page":   page,
		"limit":  limit,
		"sort":   sort,
		"filter": filter,
	}

	sensors, total, err := h.svc.GetPaginatedSensors(ctx, params)
	if err != nil {
		pkg.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to fetch sensors",
		})
		return
	}

	if sensors == nil {
		sensors = []models.Sensor{}
	}

	pkg.WriteJSON(w, http.StatusOK, models.SensorListResponse{
		Data:      sensors,
		Page:      page,
		Limit:     limit,
		Total:     total,
		TotalPage: pkg.TotalPages(total, limit),
	})
}

// GetSensor handler untuk GET /api/v1/sensors/{id}.
// GetSensor godoc
// @Summary Detail sensor (termasuk tipe)
// @Tags Sensors
// @Produce json
// @Security BearerAuth
// @Param id path int true "Sensor ID" minimum(1)
// @Success 200 {object} models.Sensor
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Router /api/v1/sensors/{id} [get]
func (h *SensorHandler) GetSensor(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid sensor id",
		})
		return
	}

	sensor, err := h.svc.GetSensorByID(r.Context(), uint(id))
	if err != nil {
		pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"error": "Sensor not found",
		})
		return
	}

	pkg.WriteJSON(w, http.StatusOK, sensor)
}

// CreateSensor handler untuk POST /api/v1/sensors.
// CreateSensor godoc
// @Summary Buat sensor baru
// @Tags Sensors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body models.SensorInput true "Sensor"
// @Success 201 {object} models.SensorEnvelope
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Router /api/v1/sensors [post]
func (h *SensorHandler) CreateSensor(w http.ResponseWriter, r *http.Request) {
	var sensor models.Sensor

	if err := json.NewDecoder(r.Body).Decode(&sensor); err != nil {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid payload",
		})
		return
	}

	res, err := h.svc.CreateSensor(r.Context(), &sensor)
	if err != nil {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	pkg.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Sensor created successfully",
		"data":    res,
	})
}

// UpdateSensor handler untuk PATCH /api/v1/sensors/{id}.
// Body: subset dari {name, sensor_type_id, status}.
// UpdateSensor godoc
// @Summary Partial update sensor
// @Tags Sensors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Sensor ID" minimum(1)
// @Param body body models.SensorPatch true "Field yang diubah"
// @Success 200 {object} models.SensorEnvelope
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Router /api/v1/sensors/{id} [patch]
func (h *SensorHandler) UpdateSensor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid sensor id",
		})
		return
	}

	var data map[string]any
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid payload",
		})
		return
	}

	res, err := h.svc.UpdateSensor(ctx, uint(id), data)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"error": "Sensor not found",
			})
			return
		}
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	pkg.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Sensor updated successfully",
		"data":    res,
	})
}

// DeleteSensor handler untuk DELETE /api/v1/sensors/{id}.
// DeleteSensor godoc
// @Summary Hapus sensor
// @Tags Sensors
// @Produce json
// @Security BearerAuth
// @Param id path int true "Sensor ID" minimum(1)
// @Success 200 {object} models.MessageEnvelope
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Router /api/v1/sensors/{id} [delete]
func (h *SensorHandler) DeleteSensor(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid sensor id",
		})
		return
	}

	if err := h.svc.DeleteSensor(r.Context(), uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"error": "Sensor not found",
			})
			return
		}
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	pkg.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Sensor deleted successfully",
	})
}

// ListCalibrations handler untuk GET /api/v1/sensors/{id}/calibrations.
// ListCalibrations godoc
// @Summary Daftar kalibrasi sensor (terbaru dulu)
// @Tags Sensors
// @Produce json
// @Security BearerAuth
// @Param id path int true "Sensor ID" minimum(1)
// @Success 200 {object} models.SensorCalibrationListResponse
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Router /api/v1/sensors/{id}/calibrations [get]
func (h *SensorHandler) ListCalibrations(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid sensor id",
		})
		return
	}

	cals, err := h.svc.GetCalibrations(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"error": "Sensor not found",
			})
			return
		}
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if cals == nil {
		cals = []models.SensorCalibration{}
	}

	pkg.WriteJSON(w, http.StatusOK, models.SensorCalibrationListResponse{
		SensorID: uint(id),
		Total:    len(cals),
		Data:     cals,
	})
}

// CreateCalibration handler untuk POST /api/v1/sensors/{id}/calibrations.
// CreateCalibration godoc
// @Summary Tambah kalibrasi sensor
// @Tags Sensors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Sensor ID" minimum(1)
// @Param body body models.SensorCalibrationInput true "Kalibrasi (from_date default now, to_date default null)"
// @Success 201 {object} models.SensorCalibrationEnvelope
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Router /api/v1/sensors/{id}/calibrations [post]
func (h *SensorHandler) CreateCalibration(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid sensor id",
		})
		return
	}

	var cal models.SensorCalibration
	if err := json.NewDecoder(r.Body).Decode(&cal); err != nil {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid payload",
		})
		return
	}

	res, err := h.svc.CreateCalibration(r.Context(), uint(id), &cal)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"error": "Sensor not found",
			})
			return
		}
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	pkg.WriteJSON(w, http.StatusCreated, models.SensorCalibrationEnvelope{
		Message: "Calibration created successfully",
		Data:    *res,
	})
}
