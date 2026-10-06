package repositories

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"gorm.io/gorm"
)

type SensorReadingRepository struct {
	db *gorm.DB
}

func NewSensorReadingRepository(db *gorm.DB) *SensorReadingRepository {
	return &SensorReadingRepository{db: db}
}

// Create simpan satu data sensor reading.
func (r *SensorReadingRepository) Create(ctx context.Context, reading *models.SensorReading) error {
	if reading == nil {
		return errors.New("SensorReading NULL")
	}
	if reading.SensorID == 0 {
		return errors.New("sensor_id is required")
	}
	if strings.TrimSpace(reading.DeviceID) == "" {
		return errors.New("device_id is required")
	}
	if reading.ReadingTimeOrigin.IsZero() {
		return errors.New("reading_time_origin is required")
	}
	reading.CreatedAt = time.Now()
	if err := r.db.WithContext(ctx).Create(reading).Error; err != nil {
		slog.ErrorContext(ctx, "Error SensorReading Create", slog.Any("sensor_id", reading.SensorID), slog.Any("error", err.Error()))
		return err
	}
	return nil
}

// readingKey kunci dedup (sensor_id, reading_time_origin).
type readingKey struct {
	sensorID uint
	at       time.Time
}

// CreateBatch simpan banyak data sensor reading via batch (max 500).
// Hanya baris dengan sensor_id & reading_time_origin valid yang dikirim;
// baris invalid / duplikat (dalam payload maupun sudah ada di DB) di-log
// lalu dilewati. Error DB saat insert tetap di-return.
func (r *SensorReadingRepository) CreateBatch(ctx context.Context, readings []models.SensorReading) error {
	if len(readings) == 0 {
		return errors.New("readings is required")
	}

	// 1. Filter hanya baris valid + dedup dalam payload.
	now := time.Now()
	seen := map[readingKey]bool{}
	var valid []models.SensorReading
	for i := range readings {
		row := readings[i]
		if strings.TrimSpace(row.DeviceID) == "" {
			slog.ErrorContext(ctx, "Skip SensorReading: device_id is required", slog.Any("index", i))
			continue
		}
		if row.SensorID == 0 {
			slog.ErrorContext(ctx, "Skip SensorReading: sensor_id is required", slog.Any("index", i))
			continue
		}
		if row.ReadingTimeOrigin.IsZero() {
			slog.ErrorContext(ctx, "Skip SensorReading: reading_time_origin is required", slog.Any("index", i), slog.Any("sensor_id", row.SensorID))
			continue
		}
		key := readingKey{sensorID: row.SensorID, at: row.ReadingTimeOrigin.UTC()}
		if seen[key] {
			slog.ErrorContext(ctx, "Skip SensorReading: duplicate in payload", slog.Any("index", i), slog.Any("sensor_id", row.SensorID))
			continue
		}
		seen[key] = true
		row.CreatedAt = now
		valid = append(valid, row)
	}
	if len(valid) == 0 {
		slog.InfoContext(ctx, "SensorReading CreateBatch: no valid rows", slog.Int("received", len(readings)))
		return nil
	}

	// 2. Cek yang sudah ada di DB (per sensor_id, pakai composite index).
	timesBySensor := map[uint][]time.Time{}
	for _, row := range valid {
		timesBySensor[row.SensorID] = append(timesBySensor[row.SensorID], row.ReadingTimeOrigin)
	}
	exists := map[readingKey]bool{}
	for sensorID, times := range timesBySensor {
		var stored []time.Time
		if err := r.db.WithContext(ctx).Model(&models.SensorReading{}).
			Where("sensor_id = ?", sensorID).
			Where("reading_time_origin IN ?", times).
			Pluck("reading_time_origin", &stored).Error; err != nil {
			slog.ErrorContext(ctx, "Error SensorReading dedup check", slog.Any("sensor_id", sensorID), slog.Any("error", err.Error()))
			return err
		}
		for _, at := range stored {
			exists[readingKey{sensorID: sensorID, at: at.UTC()}] = true
		}
	}

	var fresh []models.SensorReading
	for _, row := range valid {
		key := readingKey{sensorID: row.SensorID, at: row.ReadingTimeOrigin.UTC()}
		if exists[key] {
			slog.ErrorContext(ctx, "Skip SensorReading: already exists", slog.Any("sensor_id", row.SensorID))
			continue
		}
		fresh = append(fresh, row)
	}
	if len(fresh) == 0 {
		slog.InfoContext(ctx, "SensorReading CreateBatch: all rows already exist", slog.Int("received", len(readings)))
		return nil
	}

	// 3. Insert batch max 500.
	if err := r.db.WithContext(ctx).CreateInBatches(fresh, 500).Error; err != nil {
		slog.ErrorContext(ctx, "Error SensorReading CreateBatch", slog.Any("count", len(fresh)), slog.Any("error", err.Error()))
		return err
	}
	slog.InfoContext(ctx, "SensorReading CreateBatch done", slog.Int("saved", len(fresh)), slog.Int("skipped", len(readings)-len(fresh)))
	return nil
}
