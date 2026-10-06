package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/internal/services"
	"github.com/abulhanifah/weather-monitoring/pkg"
)

type ReadingHandler struct {
	svc *services.SensorReadingService
}

func NewReadingHandler(svc *services.SensorReadingService) *ReadingHandler {
	return &ReadingHandler{svc: svc}
}

// ListReadings handler untuk GET /api/v1/readings.
// Query params: page, limit, sort (reading_time_origin asc|desc),
// filter: device_id, sensor_id, sensor_type, from, to (RFC3339/tanggal),
// interval: raw (default), 1m, 1h, 1d (agregasi avg/min/max/count).
// ListReadings godoc
// @Summary List sensor readings (raw atau agregasi)
// @Tags Readings
// @Produce json
// @Security BearerAuth
// @Param page query int false "Halaman" default(1) minimum(1)
// @Param limit query int false "Limit" default(100) minimum(1) maximum(1000)
// @Param sort query string false "Sort" default(reading_time_origin desc)
// @Param device_id query string false "Filter device ID"
// @Param sensor_id query int false "Filter sensor ID"
// @Param sensor_type query string false "Filter nama tipe sensor"
// @Param from query string false "Dari (RFC3339, acuan reading_time_origin)"
// @Param to query string false "Sampai (RFC3339, acuan reading_time_origin)"
// @Param interval query string false "Interval agregasi" Enums(raw, 1m, 1h, 1d) default(raw)
// @Success 200 {object} models.SensorReadingListResponse "interval=raw"
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 500 {object} models.ErrorEnvelope
// @Router /api/v1/readings [get]
func (h *ReadingHandler) ListReadings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	page, err := strconv.Atoi(q.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil || limit < 1 {
		limit = 100
	}

	sort := q.Get("sort")
	if sort == "" {
		sort = "reading_time_origin desc"
	}

	interval := q.Get("interval")
	if interval == "" {
		interval = services.IntervalRaw
	}
	if interval != services.IntervalRaw && interval != services.Interval1m &&
		interval != services.Interval1h && interval != services.Interval1d {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "invalid interval, allowed: raw, 1m, 1h, 1d",
		})
		return
	}

	filter := map[string]any{}
	if v := q.Get("device_id"); v != "" {
		filter["device_id"] = v
	}
	if v := q.Get("sensor_id"); v != "" {
		filter["sensor_id"] = v
	}
	if v := q.Get("sensor_type"); v != "" {
		filter["sensor_type"] = v
	}
	if v := q.Get("from"); v != "" {
		from, err := parseDateTime(v)
		if err != nil {
			pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": "invalid from, use RFC3339",
			})
			return
		}
		filter["from"] = from
	}
	if v := q.Get("to"); v != "" {
		to, err := parseDateTime(v)
		if err != nil {
			pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": "invalid to, use RFC3339",
			})
			return
		}
		filter["to"] = to
	}
	if from, ok := filter["from"].(time.Time); ok {
		if to, ok := filter["to"].(time.Time); ok && to.Before(from) {
			pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": "to must not be before from",
			})
			return
		}
	}

	params := map[string]any{
		"page":   page,
		"limit":  limit,
		"sort":   sort,
		"filter": filter,
	}

	if interval != services.IntervalRaw {
		buckets, total, err := h.svc.GetReadingsAggregated(ctx, params, interval, page, limit)
		if err != nil {
			pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		if buckets == nil {
			buckets = []models.AggregatedReading{}
		}
		pkg.WriteJSON(w, http.StatusOK, models.AggregatedReadingListResponse{
			Data:      buckets,
			Page:      page,
			Limit:     limit,
			Total:     total,
			TotalPage: pkg.TotalPages(total, limit),
			Interval:  interval,
		})
		return
	}

	readings, total, err := h.svc.GetReadingsRaw(ctx, params)
	if err != nil {
		pkg.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to fetch readings",
		})
		return
	}
	if readings == nil {
		readings = []models.SensorReading{}
	}

	pkg.WriteJSON(w, http.StatusOK, models.SensorReadingListResponse{
		Data:      readings,
		Page:      page,
		Limit:     limit,
		Total:     total,
		TotalPage: pkg.TotalPages(total, limit),
	})
}

func parseDateTime(v string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02"} {
		if t, err := time.Parse(layout, v); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("invalid datetime")
}
