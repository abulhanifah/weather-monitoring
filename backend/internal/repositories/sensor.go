package repositories

import (
	"context"
	"errors"
	"log/slog"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/pkg"
	"gorm.io/gorm"
)

type SensorTypeRepository struct {
	db *gorm.DB
}

func NewSensorTypeRepository(db *gorm.DB) *SensorTypeRepository {
	return &SensorTypeRepository{db: db}
}

func (r *SensorTypeRepository) FindByID(ctx context.Context, id uint) (*models.SensorType, error) {
	var st models.SensorType
	if err := r.db.WithContext(ctx).First(&st, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Error SensorType FindByID", slog.Any("id", id), slog.Any("error", err.Error()))
		return nil, err
	}
	return &st, nil
}

func (r *SensorTypeRepository) Create(ctx context.Context, st *models.SensorType) error {
	if st == nil {
		return errors.New("SensorType NULL")
	}
	if err := r.db.WithContext(ctx).Create(st).Error; err != nil {
		slog.ErrorContext(ctx, "Error SensorType Create", slog.Any("name", st.Name), slog.Any("error", err.Error()))
		return err
	}
	return nil
}

// GetPaginated ambil daftar sensor type dengan filter, page, limit, sort dari params.
// Keys params: "filter" (map[string]any: id exact, name LIKE),
// "page" (int, default 1), "limit" (int, default 10, max 100),
// "sort" (string, contoh "name asc", default "created_at desc").
// Return: list sensor type, total data (sebelum page/limit), error.
func (r *SensorTypeRepository) GetPaginated(ctx context.Context, params map[string]any) ([]models.SensorType, int, error) {
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

	base := r.db.WithContext(ctx).Model(&models.SensorType{})

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
		slog.ErrorContext(ctx, "Error SensorType GetPaginated Count", slog.Any("params", params), slog.Any("error", err.Error()))
		return nil, 0, err
	}

	var types []models.SensorType
	offset := (page - 1) * limit
	if err := base.Order(sort).Offset(offset).Limit(limit).Find(&types).Error; err != nil {
		slog.ErrorContext(ctx, "Error SensorType GetPaginated Find", slog.Any("params", params), slog.Any("error", err.Error()))
		return nil, 0, err
	}

	return types, int(total), nil
}

type SensorRepository struct {
	db *gorm.DB
}

func NewSensorRepository(db *gorm.DB) *SensorRepository {
	return &SensorRepository{db: db}
}

func (r *SensorRepository) FindSensorByID(ctx context.Context, id uint) (*models.Sensor, error) {
	var sensor models.Sensor
	if err := r.db.WithContext(ctx).Preload("SensorType").First(&sensor, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Error Sensor FindByID", slog.Any("id", id), slog.Any("error", err.Error()))
		return nil, err
	}
	return &sensor, nil
}

func (r *SensorRepository) CreateSensor(ctx context.Context, sensor *models.Sensor) error {
	if sensor == nil {
		return errors.New("Sensor NULL")
	}
	if err := r.db.WithContext(ctx).Create(sensor).Error; err != nil {
		slog.ErrorContext(ctx, "Error Sensor Create", slog.Any("name", sensor.Name), slog.Any("error", err.Error()))
		return err
	}
	return nil
}

func (r *SensorRepository) UpdateSensor(ctx context.Context, id uint, data map[string]any) error {
	if err := r.db.WithContext(ctx).Model(&models.Sensor{}).Where("id = ?", id).Updates(data).Error; err != nil {
		slog.ErrorContext(ctx, "Error Sensor Update", slog.Any("id", id), slog.Any("data", data), slog.Any("error", err.Error()))
		return err
	}
	return nil
}

func (r *SensorRepository) DeleteSensor(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&models.Sensor{}, "id = ?", id)
	if res.Error != nil {
		slog.ErrorContext(ctx, "Error Sensor Delete", slog.Any("id", id), slog.Any("error", res.Error.Error()))
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetPaginatedSensors ambil daftar sensor dengan filter, page, limit, sort.
// Keys params: "filter" (map[string]any: id/sensor_type_id exact,
// status bool exact, name LIKE), "page", "limit", "sort".
func (r *SensorRepository) GetPaginatedSensors(ctx context.Context, params map[string]any) ([]models.Sensor, int, error) {
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
	sort := pkg.ToSortParam(params["sort"], "created_at desc", []string{"created_at", "updated_at", "name", "id", "sensor_type_id", "status"})

	base := r.db.WithContext(ctx).Model(&models.Sensor{}).Preload("SensorType")

	if filter, ok := params["filter"].(map[string]any); ok && filter != nil {
		if v, ok := filter["name"].(string); ok && v != "" {
			base = base.Where("sensors.name ILIKE ?", "%"+v+"%")
		}
		if id, ok := pkg.ToUintFilter(filter["id"]); ok {
			base = base.Where("sensors.id = ?", id)
		}
		if typeID, ok := pkg.ToUintFilter(filter["sensor_type_id"]); ok {
			base = base.Where("sensors.sensor_type_id = ?", typeID)
		}
		if status, ok := filter["status"].(bool); ok {
			base = base.Where("sensors.status = ?", status)
		}
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		slog.ErrorContext(ctx, "Error Sensor GetPaginated Count", slog.Any("params", params), slog.Any("error", err.Error()))
		return nil, 0, err
	}

	var sensors []models.Sensor
	offset := (page - 1) * limit
	if err := base.Order("sensors." + sort).Offset(offset).Limit(limit).Find(&sensors).Error; err != nil {
		slog.ErrorContext(ctx, "Error Sensor GetPaginated Find", slog.Any("params", params), slog.Any("error", err.Error()))
		return nil, 0, err
	}

	return sensors, int(total), nil
}
