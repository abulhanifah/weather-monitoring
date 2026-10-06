package services

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/internal/repositories"
	"github.com/abulhanifah/weather-monitoring/pkg"
)

var (
	// ErrSensorTypeAlreadyInstalled device sudah punya sensor aktif bertipe sama.
	ErrSensorTypeAlreadyInstalled = errors.New("device already has an active sensor of the same type")
)

type SensorInstallationService struct {
	repo       *repositories.SensorInstallationRepository
	deviceRepo *repositories.DeviceRepository
	sensorRepo *repositories.SensorRepository
	readingRepo *repositories.SensorReadingRepository
}

func NewSensorInstallationService(
	repo *repositories.SensorInstallationRepository,
	deviceRepo *repositories.DeviceRepository,
	sensorRepo *repositories.SensorRepository,
	readingRepo *repositories.SensorReadingRepository,
) *SensorInstallationService {
	return &SensorInstallationService{repo: repo, deviceRepo: deviceRepo, sensorRepo: sensorRepo, readingRepo: readingRepo}
}

// InstallSensor pasang sensor pada device.
// 1 device boleh punya banyak sensor, tapi tiap tipe hanya boleh 1 yang aktif.
func (s *SensorInstallationService) InstallSensor(ctx context.Context, deviceID string, sensorID uint) (*models.SensorInstallation, error) {
	if deviceID == "" {
		return nil, errors.New("device id is required")
	}
	if sensorID == 0 {
		return nil, errors.New("sensor id is required")
	}

	if _, err := s.deviceRepo.FindByID(ctx, deviceID); err != nil {
		return nil, err
	}

	sensor, err := s.sensorRepo.FindSensorByID(ctx, sensorID)
	if err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsActiveType(ctx, deviceID, sensor.SensorTypeID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSensorTypeAlreadyInstalled
	}

	inst := &models.SensorInstallation{
		DeviceID: deviceID,
		SensorID: sensorID,
		Status:   true,
	}
	if err := s.repo.Create(ctx, inst); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, inst.ID)
}

// GetInstallationByID ambil detail instalasi.
func (s *SensorInstallationService) GetInstallationByID(ctx context.Context, id uint) (*models.SensorInstallation, error) {
	if id == 0 {
		return nil, errors.New("installation id is required")
	}
	return s.repo.FindByID(ctx, id)
}

// UninstallSensor ubah status instalasi aktif device+sensor menjadi false.
func (s *SensorInstallationService) UninstallSensor(ctx context.Context, deviceID string, sensorID uint) error {
	if deviceID == "" {
		return errors.New("device id is required")
	}
	if sensorID == 0 {
		return errors.New("sensor id is required")
	}
	if _, err := s.deviceRepo.FindByID(ctx, deviceID); err != nil {
		return err
	}
	if _, err := s.sensorRepo.FindSensorByID(ctx, sensorID); err != nil {
		return err
	}
	return s.repo.DeactivateByDeviceSensor(ctx, deviceID, sensorID)
}

// TelemetryResult ringkasan ingest telemetry.
type TelemetryResult struct {
	Saved   int
	Skipped int
}

// IngestTelemetry simpan readings telemetry ke sensor readings.
// Payload map bebas, tapi device_id (harus sama dengan pemilik api key),
// ts (unix, jadi reading_time_origin), dan readings wajib ada.
// readings[i].s = nama sensor type -> dipetakan ke sensor_id yang
// terpasang aktif pada device (batch 1 query); readings[i].v = value.
// Baris yang tidak valid / tipenya tidak terpasang hanya di-log lalu dilewati.
func (s *SensorInstallationService) IngestTelemetry(ctx context.Context, authedDeviceID string, payload map[string]any) (*TelemetryResult, error) {
	if len(payload) == 0 {
		return nil, errors.New("payload is required")
	}

	deviceID, err := validatedDeviceID(authedDeviceID, payload)
	if err != nil {
		return nil, err
	}

	ts, ok := pkg.ToUintFilter(payload["ts"])
	if !ok || ts == 0 {
		return nil, errors.New("ts is required (unix timestamp)")
	}
	readingTime := time.Unix(int64(ts), 0).UTC()

	rows, typeSet, skipped := parseTelemetryReadings(ctx, payload["readings"])
	if len(rows) == 0 && skipped == 0 {
		return nil, errors.New("readings is required")
	}
	if len(rows) == 0 {
		return nil, errors.New("no valid readings")
	}

	sensorIDs, err := s.activeSensorIDs(ctx, deviceID, typeSet)
	if err != nil {
		return nil, err
	}

	readings, skipped := s.buildReadings(ctx, deviceID, sensorIDs, []telemetryBatch{{readingTime: readingTime, rows: rows}}, skipped)
	if len(readings) == 0 {
		return &TelemetryResult{Saved: 0, Skipped: skipped}, nil
	}

	if err := s.readingRepo.CreateBatch(ctx, readings); err != nil {
		return nil, err
	}
	return &TelemetryResult{Saved: len(readings), Skipped: skipped}, nil
}

// telemetryBatch satu entri batch: waktu + rows tervalidasi.
type telemetryBatch struct {
	readingTime time.Time
	rows        []telemetryRow
}

// IngestTelemetryBatch simpan banyak paket telemetry sekaligus.
// device_id di root (harus cocok api key); tiap entri batch wajib
// punya ts dan readings. Lookup sensor_id 1 query untuk semua tipe.
func (s *SensorInstallationService) IngestTelemetryBatch(ctx context.Context, authedDeviceID string, payload map[string]any) (*TelemetryResult, error) {
	if len(payload) == 0 {
		return nil, errors.New("payload is required")
	}

	deviceID, err := validatedDeviceID(authedDeviceID, payload)
	if err != nil {
		return nil, err
	}

	rawBatch, ok := payload["batch"].([]any)
	if !ok || len(rawBatch) == 0 {
		return nil, errors.New("batch is required")
	}

	var batches []telemetryBatch
	typeSet := map[string]bool{}
	skipped := 0
	for i, item := range rawBatch {
		entry, ok := item.(map[string]any)
		if !ok {
			slog.ErrorContext(ctx, "Skip telemetry batch: invalid entry", slog.Any("index", i))
			skipped++
			continue
		}
		ts, ok := pkg.ToUintFilter(entry["ts"])
		if !ok || ts == 0 {
			slog.ErrorContext(ctx, "Skip telemetry batch: ts is required", slog.Any("index", i))
			skipped++
			continue
		}
		rows, entryTypes, entrySkipped := parseTelemetryReadings(ctx, entry["readings"])
		skipped += entrySkipped
		if len(rows) == 0 {
			slog.ErrorContext(ctx, "Skip telemetry batch: no valid readings", slog.Any("index", i))
			skipped++
			continue
		}
		for name := range entryTypes {
			typeSet[name] = true
		}
		batches = append(batches, telemetryBatch{
			readingTime: time.Unix(int64(ts), 0).UTC(),
			rows:        rows,
		})
	}
	if len(batches) == 0 {
		return nil, errors.New("no valid batch entries")
	}

	sensorIDs, err := s.activeSensorIDs(ctx, deviceID, typeSet)
	if err != nil {
		return nil, err
	}

	readings, skipped := s.buildReadings(ctx, deviceID, sensorIDs, batches, skipped)
	if len(readings) == 0 {
		return &TelemetryResult{Saved: 0, Skipped: skipped}, nil
	}

	if err := s.readingRepo.CreateBatch(ctx, readings); err != nil {
		return nil, err
	}
	return &TelemetryResult{Saved: len(readings), Skipped: skipped}, nil
}

func (s *SensorInstallationService) activeSensorIDs(ctx context.Context, deviceID string, typeSet map[string]bool) (map[string]uint, error) {
	var typeNames []string
	for name := range typeSet {
		typeNames = append(typeNames, name)
	}
	return s.repo.FindActiveSensorIDsByTypeNames(ctx, deviceID, typeNames)
}

func (s *SensorInstallationService) buildReadings(ctx context.Context, deviceID string, sensorIDs map[string]uint, batches []telemetryBatch, skipped int) ([]models.SensorReading, int) {
	var readings []models.SensorReading
	for _, b := range batches {
		for _, row := range b.rows {
			sensorID, ok := sensorIDs[row.TypeName]
			if !ok {
				slog.ErrorContext(ctx, "Skip telemetry reading: no active sensor of type", slog.Any("s", row.TypeName), slog.Any("device_id", deviceID))
				skipped++
				continue
			}
			readings = append(readings, models.SensorReading{
				SensorID:          sensorID,
				ReadingTimeOrigin: b.readingTime,
				Value:             row.Value,
			})
		}
	}
	return readings, skipped
}

func toTelemetryFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(n), 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// telemetryRow satu baris readings tervalidasi.
type telemetryRow struct {
	TypeName string
	Value    float64
}

// parseTelemetryReadings validasi array readings menjadi rows.
// Baris invalid di-log + skip. Return rows valid dan jumlah skip.
func parseTelemetryReadings(ctx context.Context, raw any) ([]telemetryRow, map[string]bool, int) {
	rawReadings, ok := raw.([]any)
	if !ok || len(rawReadings) == 0 {
		return nil, nil, 0
	}
	var rows []telemetryRow
	typeSet := map[string]bool{}
	skipped := 0
	for i, item := range rawReadings {
		m, ok := item.(map[string]any)
		if !ok {
			slog.ErrorContext(ctx, "Skip telemetry reading: invalid row", slog.Any("index", i))
			skipped++
			continue
		}
		typeName, _ := m["s"].(string)
		if strings.TrimSpace(typeName) == "" {
			slog.ErrorContext(ctx, "Skip telemetry reading: s is required", slog.Any("index", i))
			skipped++
			continue
		}
		value, ok := toTelemetryFloat(m["v"])
		if !ok {
			slog.ErrorContext(ctx, "Skip telemetry reading: v must be a number", slog.Any("index", i), slog.Any("s", typeName))
			skipped++
			continue
		}
		rows = append(rows, telemetryRow{TypeName: typeName, Value: value})
		typeSet[typeName] = true
	}
	return rows, typeSet, skipped
}

// validatedDeviceID validasi device_id payload cocok dengan pemilik api key.
func validatedDeviceID(authedDeviceID string, payload map[string]any) (string, error) {
	if authedDeviceID == "" {
		return "", errors.New("device id is required")
	}
	deviceID, _ := payload["device_id"].(string)
	if strings.TrimSpace(deviceID) == "" {
		return "", errors.New("device_id is required")
	}
	if deviceID != authedDeviceID {
		return "", ErrDeviceMismatch
	}
	return deviceID, nil
}
