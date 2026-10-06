package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/internal/services"
	"github.com/abulhanifah/weather-monitoring/pkg"
)

type LocationHandler struct {
	svc *services.LocationService
}

func NewLocationHandler(svc *services.LocationService) *LocationHandler {
	return &LocationHandler{svc: svc}
}

// ListLocations handler untuk GET /api/v1/locations.
// Query params: page (default 1), limit (default 10),
// sort (contoh "name asc"), filter: name, id.
// ListLocations godoc
// @Summary List locations (paginated)
// @Tags Locations
// @Produce json
// @Security BearerAuth
// @Param page query int false "Halaman" default(1) minimum(1)
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param sort query string false "Sort" default(created_at desc)
// @Param name query string false "Filter nama (ILIKE)"
// @Param id query int false "Filter location ID"
// @Success 200 {object} models.LocationListResponse
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 500 {object} models.ErrorEnvelope
// @Router /api/v1/locations [get]
func (h *LocationHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
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

	locations, total, err := h.svc.GetPaginated(ctx, params)
	if err != nil {
		pkg.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to fetch locations",
		})
		return
	}

	if locations == nil {
		locations = []models.Location{}
	}

	pkg.WriteJSON(w, http.StatusOK, models.LocationListResponse{
		Data:      locations,
		Page:      page,
		Limit:     limit,
		Total:     total,
		TotalPage: pkg.TotalPages(total, limit),
	})
}

// GetLocation godoc
// @Summary Detail location
// @Tags Locations
// @Produce json
// @Security BearerAuth
// @Param id path int true "Location ID" minimum(1)
// @Success 200 {object} models.Location
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Failure 404 {object} models.ErrorEnvelope
// @Router /api/v1/locations/{id} [get]
func (h *LocationHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid location id",
		})
		return
	}

	loc, err := h.svc.GetByID(r.Context(), uint(id))
	if err != nil {
		pkg.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"error": "Location not found",
		})
		return
	}

	pkg.WriteJSON(w, http.StatusOK, loc)
}

// CreateLocation godoc
// @Summary Buat location baru
// @Tags Locations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body models.LocationInput true "Location"
// @Success 201 {object} models.LocationEnvelope
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Router /api/v1/locations [post]
func (h *LocationHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	var loc models.Location

	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid payload",
		})
		return
	}

	res, err := h.svc.CreateLocation(r.Context(), &loc)
	if err != nil {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	pkg.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Location created successfully",
		"data":    res,
	})
}
