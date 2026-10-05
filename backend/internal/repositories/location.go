package repositories

import (
	"context"
	"errors"
	"log/slog"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/pkg"
	"gorm.io/gorm"
)

type LocationRepository struct {
	db *gorm.DB
}

func NewLocationRepository(db *gorm.DB) *LocationRepository {
	return &LocationRepository{db: db}
}

func (r *LocationRepository) FindByID(ctx context.Context, id uint) (*models.Location, error) {
	var loc models.Location
	if err := r.db.WithContext(ctx).First(&loc, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Error Location FindByID", slog.Any("id", id), slog.Any("error", err.Error()))
		return nil, err
	}
	return &loc, nil
}

func (r *LocationRepository) Create(ctx context.Context, loc *models.Location) error {
	if loc == nil {
		return errors.New("Location NULL")
	}
	if err := r.db.WithContext(ctx).Create(loc).Error; err != nil {
		slog.ErrorContext(ctx, "Error Location Create", slog.Any("name", loc.Name), slog.Any("error", err.Error()))
		return err
	}
	return nil
}

// GetPaginated ambil daftar location dengan filter, page, limit, sort dari params.
// Keys params: "filter" (map[string]any: id exact, name LIKE),
// "page" (int, default 1), "limit" (int, default 10, max 100),
// "sort" (string, contoh "name asc", default "created_at desc").
// Return: list location, total data (sebelum page/limit), error.
func (r *LocationRepository) GetPaginated(ctx context.Context, params map[string]any) ([]models.Location, int, error) {
	page := pkg.ToIntParam(params["page"], 1)
	limit := pkg.ToIntParam(params["limit"], 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	sort := pkg.ToSortParam(params["sort"], "created_at desc", []string{"created_at", "updated_at", "name", "id"})

	base := r.db.WithContext(ctx).Model(&models.Location{})

	if filter, ok := params["filter"].(map[string]any); ok && filter != nil {
		if v, ok := filter["name"].(string); ok && v != "" {
			base = base.Where("name ILIKE ?", "%"+v+"%")
		}
		if id, ok := pkg.ToUintFilter(filter["id"]); ok {
			base = base.Where("id = ?", id)
		}
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		slog.ErrorContext(ctx, "Error Location GetPaginated Count", slog.Any("params", params), slog.Any("error", err.Error()))
		return nil, 0, err
	}

	var locations []models.Location
	offset := (page - 1) * limit
	if err := base.Order(sort).Offset(offset).Limit(limit).Find(&locations).Error; err != nil {
		slog.ErrorContext(ctx, "Error Location GetPaginated Find", slog.Any("params", params), slog.Any("error", err.Error()))
		return nil, 0, err
	}

	return locations, int(total), nil
}
