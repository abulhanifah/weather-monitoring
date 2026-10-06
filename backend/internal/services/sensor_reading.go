package services

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/internal/repositories"
	"github.com/abulhanifah/weather-monitoring/pkg"
)

type SensorReadingService struct {
	installRepo *repositories.SensorInstallationRepository
	readingRepo *repositories.SensorReadingRepository
}

func NewSensorReadingService(
	installRepo *repositories.SensorInstallationRepository,
	readingRepo *repositories.SensorReadingRepository,
) *SensorReadingService {
	return &SensorReadingService{installRepo: installRepo, readingRepo: readingRepo}
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
func (s *SensorReadingService) IngestTelemetry(ctx context.Context, authedDeviceID string, payload map[string]any) (*TelemetryResult, error) {
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
func (s *SensorReadingService) IngestTelemetryBatch(ctx context.Context, authedDeviceID string, payload map[string]any) (*TelemetryResult, error) {
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

func (s *SensorReadingService) activeSensorIDs(ctx context.Context, deviceID string, typeSet map[string]bool) (map[string]uint, error) {
	var typeNames []string
	for name := range typeSet {
		typeNames = append(typeNames, name)
	}
	return s.installRepo.FindActiveSensorIDsByTypeNames(ctx, deviceID, typeNames)
}

func (s *SensorReadingService) buildReadings(ctx context.Context, deviceID string, sensorIDs map[string]uint, batches []telemetryBatch, skipped int) ([]models.SensorReading, int) {
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
				DeviceID:          deviceID,
				SensorID:          sensorID,
				ReadingTimeOrigin: b.readingTime,
				Value:             row.Value,
			})
		}
	}
	return readings, skipped
}

// Interval agregasi yang didukung untuk GET /api/v1/readings.
const (
	IntervalRaw = "raw"
	Interval1m  = "1m"
	Interval1h  = "1h"
	Interval1d  = "1d"
)

// maxAggregatedRows batas baris yang ditarik untuk agregasi di service.
const maxAggregatedRows = 20000

// GetReadingsRaw ambil readings paginated apa adanya.
func (s *SensorReadingService) GetReadingsRaw(ctx context.Context, params map[string]any) ([]models.SensorReading, int, error) {
	return s.readingRepo.FindReadings(ctx, params)
}

// GetReadingsAggregated ambil readings lalu bucket per interval di service.
// Repo hanya memfilter datetime; agregasi (avg/min/max/count per
// device+sensor+bucket) dikerjakan di sini lalu hasilnya di-paginate.
func (s *SensorReadingService) GetReadingsAggregated(ctx context.Context, params map[string]any, interval string, page, limit int) ([]models.AggregatedReading, int, error) {
	var bucketSize time.Duration
	switch interval {
	case Interval1m:
		bucketSize = time.Minute
	case Interval1h:
		bucketSize = time.Hour
	case Interval1d:
		bucketSize = 24 * time.Hour
	default:
		return nil, 0, errors.New("invalid interval, allowed: raw, 1m, 1h, 1d")
	}
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	rows, total, err := s.readingRepo.FindReadings(ctx, withNoPagination(params))
	if err != nil {
		return nil, 0, err
	}
	if total > maxAggregatedRows {
		return nil, 0, errors.New("too many rows, narrow the time range or use a bigger interval")
	}

	type bucketKey struct {
		deviceID string
		sensorID uint
		bucket   time.Time
	}
	agg := map[bucketKey]*models.AggregatedReading{}
	for _, row := range rows {
		ts := row.ReadingTimeOrigin.UTC()
		var bucket time.Time
		if bucketSize == 24*time.Hour {
			bucket = time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.UTC)
		} else {
			bucket = ts.Truncate(bucketSize)
		}
		key := bucketKey{deviceID: row.DeviceID, sensorID: row.SensorID, bucket: bucket}
		b, ok := agg[key]
		if !ok {
			b = &models.AggregatedReading{
				Time:     bucket,
				DeviceID: row.DeviceID,
				SensorID: row.SensorID,
				Min:      row.Value,
				Max:      row.Value,
			}
			agg[key] = b
		}
		b.Count++
		b.Avg += (row.Value - b.Avg) / float64(b.Count)
		if row.Value < b.Min {
			b.Min = row.Value
		}
		if row.Value > b.Max {
			b.Max = row.Value
		}
	}

	buckets := make([]models.AggregatedReading, 0, len(agg))
	for _, b := range agg {
		buckets = append(buckets, *b)
	}
	sort.Slice(buckets, func(i, j int) bool {
		if buckets[i].Time.Equal(buckets[j].Time) {
			if buckets[i].DeviceID == buckets[j].DeviceID {
				return buckets[i].SensorID < buckets[j].SensorID
			}
			return buckets[i].DeviceID < buckets[j].DeviceID
		}
		return buckets[i].Time.After(buckets[j].Time)
	})

	totalBuckets := len(buckets)
	start := (page - 1) * limit
	if start >= totalBuckets {
		return []models.AggregatedReading{}, totalBuckets, nil
	}
	end := start + limit
	if end > totalBuckets {
		end = totalBuckets
	}
	return buckets[start:end], totalBuckets, nil
}

func withNoPagination(params map[string]any) map[string]any {
	out := make(map[string]any, len(params)+1)
	for k, v := range params {
		out[k] = v
	}
	out["page"] = 1
	out["limit"] = 0
	return out
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
