package repositories

import (
	"context"
	"log/slog"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"gorm.io/gorm"
)

type SensorInstallationRepository struct {
	db *gorm.DB
}

func NewSensorInstallationRepository(db *gorm.DB) *SensorInstallationRepository {
	return &SensorInstallationRepository{db: db}
}

func (r *SensorInstallationRepository) FindByID(ctx context.Context, id uint) (*models.SensorInstallation, error) {
	var inst models.SensorInstallation
	if err := r.db.WithContext(ctx).Preload("Device").Preload("Sensor.SensorType").First(&inst, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Error SensorInstallation FindByID", slog.Any("id", id), slog.Any("error", err.Error()))
		return nil, err
	}
	return &inst, nil
}

// ExistsActiveType cek apakah device sudah punya sensor aktif dengan tipe yang sama.
func (r *SensorInstallationRepository) ExistsActiveType(ctx context.Context, deviceID string, sensorTypeID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.SensorInstallation{}).
		Joins("JOIN sensors ON sensors.id = sensor_installations.sensor_id AND sensors.deleted_at IS NULL").
		Where("sensor_installations.device_id = ?", deviceID).
		Where("sensor_installations.status = ?", true).
		Where("sensors.sensor_type_id = ?", sensorTypeID).
		Count(&count).Error
	if err != nil {
		slog.ErrorContext(ctx, "Error ExistsActiveType", slog.Any("device_id", deviceID), slog.Any("error", err.Error()))
		return false, err
	}
	return count > 0, nil
}

func (r *SensorInstallationRepository) Create(ctx context.Context, inst *models.SensorInstallation) error {
	if err := r.db.WithContext(ctx).Create(inst).Error; err != nil {
		slog.ErrorContext(ctx, "Error SensorInstallation Create", slog.Any("device_id", inst.DeviceID), slog.Any("error", err.Error()))
		return err
	}
	return nil
}

// FindActiveSensorIDsByTypeNames ambil sensor aktif milik device
// untuk daftar nama tipe sekaligus. Return map[namaTipe]sensorID.
func (r *SensorInstallationRepository) FindActiveSensorIDsByTypeNames(ctx context.Context, deviceID string, typeNames []string) (map[string]uint, error) {
	out := map[string]uint{}
	if len(typeNames) == 0 {
		return out, nil
	}
	var rows []struct {
		SensorID uint
		TypeName string
	}
	err := r.db.WithContext(ctx).Model(&models.SensorInstallation{}).
		Select("sensor_installations.sensor_id AS sensor_id, sensor_types.name AS type_name").
		Joins("JOIN sensors ON sensors.id = sensor_installations.sensor_id AND sensors.deleted_at IS NULL").
		Joins("JOIN sensor_types ON sensor_types.id = sensors.sensor_type_id AND sensor_types.deleted_at IS NULL").
		Where("sensor_installations.device_id = ?", deviceID).
		Where("sensor_installations.status = ?", true).
		Where("sensor_types.name IN ?", typeNames).
		Scan(&rows).Error
	if err != nil {
		slog.ErrorContext(ctx, "Error FindActiveSensorIDs", slog.Any("device_id", deviceID), slog.Any("error", err.Error()))
		return nil, err
	}
	for _, row := range rows {
		out[row.TypeName] = row.SensorID
	}
	return out, nil
}

// DeactivateByDeviceSensor ubah status instalasi aktif device+sensor menjadi false.
// Return ErrRecordNotFound bila tidak ada instalasi aktif yang cocok.
func (r *SensorInstallationRepository) DeactivateByDeviceSensor(ctx context.Context, deviceID string, sensorID uint) error {
	res := r.db.WithContext(ctx).Model(&models.SensorInstallation{}).
		Where("device_id = ? AND sensor_id = ? AND status = ?", deviceID, sensorID, true).
		Update("status", false)
	if res.Error != nil {
		slog.ErrorContext(ctx, "Error DeactivateInstallation", slog.Any("device_id", deviceID), slog.Any("sensor_id", sensorID), slog.Any("error", res.Error.Error()))
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
