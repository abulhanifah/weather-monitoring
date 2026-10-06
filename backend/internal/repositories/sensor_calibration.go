package repositories

import (
	"context"
	"log/slog"
	"time"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"gorm.io/gorm"
)

type SensorCalibrationRepository struct {
	db *gorm.DB
}

func NewSensorCalibrationRepository(db *gorm.DB) *SensorCalibrationRepository {
	return &SensorCalibrationRepository{db: db}
}

func (r *SensorCalibrationRepository) Create(ctx context.Context, cal *models.SensorCalibration) error {
	if err := r.db.WithContext(ctx).Create(cal).Error; err != nil {
		slog.ErrorContext(ctx, "Error SensorCalibration Create", slog.Any("sensor_id", cal.SensorID), slog.Any("error", err.Error()))
		return err
	}
	return nil
}

// CreateClosingPrevious buat kalibrasi baru sekaligus menutup kalibrasi lama
// yang masih terbuka (to_date null) dengan now, dalam satu transaksi.
func (r *SensorCalibrationRepository) CreateClosingPrevious(ctx context.Context, cal *models.SensorCalibration) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&models.SensorCalibration{}).
			Where("sensor_id = ? AND to_date IS NULL", cal.SensorID).
			Update("to_date", now).Error; err != nil {
			slog.ErrorContext(ctx, "Error CloseCalibration", slog.Any("sensor_id", cal.SensorID), slog.Any("error", err.Error()))
			return err
		}
		if err := tx.Create(cal).Error; err != nil {
			slog.ErrorContext(ctx, "Error SensorCalibration Create", slog.Any("sensor_id", cal.SensorID), slog.Any("error", err.Error()))
			return err
		}
		return nil
	})
}

// FindBySensorID ambil semua kalibrasi sensor, terbaru (from_date) dulu.
func (r *SensorCalibrationRepository) FindBySensorID(ctx context.Context, sensorID uint) ([]models.SensorCalibration, error) {
	var cals []models.SensorCalibration
	if err := r.db.WithContext(ctx).Where("sensor_id = ?", sensorID).Order("from_date desc").Find(&cals).Error; err != nil {
		slog.ErrorContext(ctx, "Error FindCalibrations", slog.Any("sensor_id", sensorID), slog.Any("error", err.Error()))
		return nil, err
	}
	return cals, nil
}
