package repositories

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"gorm.io/gorm"
)

type DeviceRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

func (r *DeviceRepository) FindByID(ctx context.Context, id string) (*models.Device, error) {
	var dev models.Device
	err := r.db.First(&dev, "id = ?", id).Error
	if err != nil {
		slog.ErrorContext(ctx, "Error FindByID", slog.Any("id", id), slog.Any("error", err.Error()))
		return nil, err
	}
	return &dev, nil
}

func (r *DeviceRepository) Update(ctx context.Context, id string, data map[string]any) error {
	if err := r.db.Model(models.Device{}).Updates(data).Where("id", id).Error; err != nil {
		slog.ErrorContext(ctx, "Error SetStatus", slog.Any("id", id), slog.Any("data", data), slog.Any("error", err.Error()))
		return err
	}
	return nil
}

func (r *DeviceRepository) Create(ctx context.Context, dev *models.Device) error {
	if dev == nil {
		return errors.New("Device NULL")
	}
	if err := r.db.Create(dev).Error; err != nil {
		slog.ErrorContext(ctx, "Error Create", slog.Any("id", dev.ID), slog.Any("error", err.Error()))
		return err
	}
	return nil
}

func (r *DeviceRepository) SaveAPIKey(ctx context.Context, data *models.APIKeyMeta) error {
	if data == nil {
		return errors.New("APIKey is NULL")
	}
	if err := r.db.Create(data).Error; err != nil {
		slog.ErrorContext(ctx, "Error Create", slog.Any("id", data.ID), slog.Any("error", err.Error()))
		return err
	}
	return nil
}

// GetPaginated ambil daftar device dengan filter, page, limit, sort dari params.
// Keys params: "filter" (map[string]any: status/id exact, name LIKE),
// "page" (int, default 1), "limit" (int, default 10, max 100),
// "sort" (string, contoh "created_at desc", default "created_at desc").
// Return: list device, total data (sebelum page/limit), error.
func (r *DeviceRepository) GetPaginated(ctx context.Context, params map[string]any) ([]models.Device, int, error) {
	page := toIntParam(params["page"], 1)
	limit := toIntParam(params["limit"], 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	sort := toSortParam(params["sort"])

	base := r.db.WithContext(ctx).Model(&models.Device{})

	// Terapkan filter
	if filter, ok := params["filter"].(map[string]any); ok && filter != nil {
		if v, ok := filter["status"].(string); ok && v != "" {
			base = base.Where("status = ?", v)
		}
		if v, ok := filter["id"].(string); ok && v != "" {
			base = base.Where("id = ?", v)
		}
		if v, ok := filter["name"].(string); ok && v != "" {
			base = base.Where("name ILIKE ?", "%"+v+"%")
		}
	}

	// Hitung total sebelum pagination
	var total int64
	if err := base.Count(&total).Error; err != nil {
		slog.ErrorContext(ctx, "Error GetPaginated Count", slog.Any("params", params), slog.Any("error", err.Error()))
		return nil, 0, err
	}

	var devices []models.Device
	offset := (page - 1) * limit
	if err := base.Order(sort).Offset(offset).Limit(limit).Find(&devices).Error; err != nil {
		slog.ErrorContext(ctx, "Error GetPaginated Find", slog.Any("params", params), slog.Any("error", err.Error()))
		return nil, 0, err
	}

	return devices, int(total), nil
}

func toIntParam(v any, fallback int) int {
	switch n := v.(type) {
	case int:
		return n
	case int8:
		return int(n)
	case int16:
		return int(n)
	case int32:
		return int(n)
	case int64:
		return int(n)
	case uint:
		return int(n)
	case float64:
		return int(n)
	default:
		return fallback
	}
}

// toSortParam validasi kolom dan arah sort agar aman dari SQL injection.
func toSortParam(v any) string {
	allowedCols := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"name":       true,
		"id":         true,
		"status":     true,
	}
	s, _ := v.(string)
	s = strings.TrimSpace(s)
	if s == "" {
		return "created_at desc"
	}
	parts := strings.Fields(s)
	col := strings.ToLower(parts[0])
	if !allowedCols[col] {
		return "created_at desc"
	}
	dir := "desc"
	if len(parts) > 1 && strings.ToLower(parts[1]) == "asc" {
		dir = "asc"
	}
	return col + " " + dir
}
