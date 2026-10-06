package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/pkg"
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
	err := r.db.WithContext(ctx).Preload("Location").First(&dev, "id = ?", id).Error
	if err != nil {
		slog.ErrorContext(ctx, "Error FindByID", slog.Any("id", id), slog.Any("error", err.Error()))
		return nil, err
	}
	return &dev, nil
}

func (r *DeviceRepository) Update(ctx context.Context, id string, data map[string]any) error {
	// Updates adalah finisher: dieksekusi saat dipanggil, jadi Where
	// harus dipasang SEBELUM Updates agar masuk ke klausa WHERE.
	if err := r.db.WithContext(ctx).Model(&models.Device{}).Where("id = ?", id).Updates(data).Error; err != nil {
		slog.ErrorContext(ctx, "Error Update", slog.Any("id", id), slog.Any("data", data), slog.Any("error", err.Error()))
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

func (r *DeviceRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&models.Device{}, "id = ?", id)
	if res.Error != nil {
		slog.ErrorContext(ctx, "Error Delete", slog.Any("id", id), slog.Any("error", res.Error.Error()))
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *DeviceRepository) RevokeAPIKeysByDeviceID(ctx context.Context, deviceID string) error {
	if err := r.db.WithContext(ctx).Model(&models.APIKeyMeta{}).
		Where("device_id = ? AND is_revoked = ?", deviceID, false).
		Update("is_revoked", true).Error; err != nil {
		slog.ErrorContext(ctx, "Error RevokeAPIKeys", slog.Any("device_id", deviceID), slog.Any("error", err.Error()))
		return err
	}
	return nil
}

// FindAPIKeyByHash cari api key aktif berdasar hash.
func (r *DeviceRepository) FindAPIKeyByHash(ctx context.Context, hash string) (*models.APIKeyMeta, error) {
	var meta models.APIKeyMeta
	if err := r.db.WithContext(ctx).First(&meta, "hash = ? AND is_revoked = ?", hash, false).Error; err != nil {
		return nil, err
	}
	return &meta, nil
}

// CreateStatusHistory catat status device. ID = device_id + timeseries (unix nano).
func (r *DeviceRepository) CreateStatusHistory(ctx context.Context, deviceID, status string) (*models.DeviceStatusHistory, error) {
	hist := &models.DeviceStatusHistory{
		ID:        fmt.Sprintf("%s-%d", deviceID, time.Now().UnixNano()),
		DeviceID:  deviceID,
		Status:    status,
		CreatedAt: time.Now(),
	}
	if err := r.db.WithContext(ctx).Create(hist).Error; err != nil {
		slog.ErrorContext(ctx, "Error CreateStatusHistory", slog.Any("device_id", deviceID), slog.Any("error", err.Error()))
		return nil, err
	}
	return hist, nil
}

// FindStatusHistory ambil histori status device, terbaru dulu.
func (r *DeviceRepository) FindStatusHistory(ctx context.Context, deviceID string, limit int) ([]models.DeviceStatusHistory, error) {
	var hists []models.DeviceStatusHistory
	q := r.db.WithContext(ctx).Where("device_id = ?", deviceID).Order("created_at desc")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&hists).Error; err != nil {
		slog.ErrorContext(ctx, "Error FindStatusHistory", slog.Any("device_id", deviceID), slog.Any("error", err.Error()))
		return nil, err
	}
	return hists, nil
}

// FindLatestStatusHistory ambil histori status terakhir device.
// Return nil, nil bila belum ada histori.
func (r *DeviceRepository) FindLatestStatusHistory(ctx context.Context, deviceID string) (*models.DeviceStatusHistory, error) {
	var hist models.DeviceStatusHistory
	err := r.db.WithContext(ctx).Where("device_id = ?", deviceID).Order("created_at desc").First(&hist).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		slog.ErrorContext(ctx, "Error FindLatestStatusHistory", slog.Any("device_id", deviceID), slog.Any("error", err.Error()))
		return nil, err
	}
	return &hist, nil
}

// FindDevicesByStatus ambil semua device dengan status tertentu.
func (r *DeviceRepository) FindDevicesByStatus(ctx context.Context, status string) ([]models.Device, error) {
	var devices []models.Device
	if err := r.db.WithContext(ctx).Where("status = ?", status).Find(&devices).Error; err != nil {
		slog.ErrorContext(ctx, "Error FindDevicesByStatus", slog.Any("status", status), slog.Any("error", err.Error()))
		return nil, err
	}
	return devices, nil
}

// GetPaginated ambil daftar device dengan filter, page, limit, sort dari params.
// Keys params: "filter" (map[string]any: status/id exact, name LIKE),
// "page" (int, default 1), "limit" (int, default 10, max 100),
// "sort" (string, contoh "created_at desc", default "created_at desc").
// Return: list device, total data (sebelum page/limit), error.
func (r *DeviceRepository) GetPaginated(ctx context.Context, params map[string]any) ([]models.Device, int, error) {
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
	// Sort: "location [asc|desc]" sort by nama lokasi (butuh JOIN),
	// selain itu sort kolom devices (otomatis di-prefix "devices." agar
	// tidak ambiguous saat JOIN aktif).
	sort, needLocationJoin := deviceSortParam(params["sort"])

	base := r.db.WithContext(ctx).Model(&models.Device{}).Preload("Location")
	if needLocationJoin {
		base = base.Joins("LEFT JOIN locations ON locations.id = devices.location_id")
	}

	// Terapkan filter
	if filter, ok := params["filter"].(map[string]any); ok && filter != nil {
		if v, ok := filter["status"].(string); ok && v != "" {
			base = base.Where("devices.status = ?", v)
		}
		if v, ok := filter["id"].(string); ok && v != "" {
			base = base.Where("devices.id = ?", v)
		}
		if v, ok := filter["name"].(string); ok && v != "" {
			base = base.Where("devices.name ILIKE ?", "%"+v+"%")
		}
		if locID, ok := pkg.ToUintFilter(filter["location_id"]); ok {
			base = base.Where("devices.location_id = ?", locID)
		}
		if v, ok := filter["q"].(string); ok && v != "" {
			like := "%" + v + "%"
			if !needLocationJoin {
				base = base.Joins("LEFT JOIN locations ON locations.id = devices.location_id")
				needLocationJoin = true
			}
			base = base.Where(
				"devices.name ILIKE ? OR locations.name ILIKE ? OR devices.id ILIKE ?",
				like, like, like,
			)
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

// deviceSortParam bangun ORDER BY yang aman.
// "location [asc|desc]" → sort by locations.name (return needJoin=true).
// Kolom lain → kolom devices (di-prefix "devices." agar tidak ambiguous).
// Return: order clause + apakah butuh JOIN ke locations.
func deviceSortParam(v any) (string, bool) {
	const defaultSort = "devices.created_at desc"

	s, _ := v.(string)
	parts := strings.Fields(strings.TrimSpace(s))
	if len(parts) == 0 {
		return defaultSort, false
	}

	dir := "desc"
	if len(parts) > 1 && strings.ToLower(parts[1]) == "asc" {
		dir = "asc"
	}

	col := strings.ToLower(parts[0])
	if col == "location" || col == "location.name" || col == "location_name" {
		return "locations.name " + dir, true
	}

	validated := pkg.ToSortParam(v, "created_at desc", []string{"created_at", "updated_at", "name", "id", "status", "location_id"})
	return "devices." + validated, false
}
