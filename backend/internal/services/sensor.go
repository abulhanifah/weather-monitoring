package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/internal/repositories"
	"github.com/abulhanifah/weather-monitoring/pkg"
	"gorm.io/gorm"
)

type SensorTypeService struct {
	repo *repositories.SensorTypeRepository
}

func NewSensorTypeService(repo *repositories.SensorTypeRepository) *SensorTypeService {
	return &SensorTypeService{repo: repo}
}

func (s *SensorTypeService) GetByID(ctx context.Context, id uint) (*models.SensorType, error) {
	return s.repo.FindByID(ctx, id)
}

// GetPaginated teruskan params filter/page/limit/sort ke repository.
// Return: list sensor type, total data, error.
func (s *SensorTypeService) GetPaginated(ctx context.Context, params map[string]any) ([]models.SensorType, int, error) {
	return s.repo.GetPaginated(ctx, params)
}

func (s *SensorTypeService) CreateSensorType(ctx context.Context, st *models.SensorType) (*models.SensorType, error) {
	if st == nil {
		return nil, errors.New("sensor type is required")
	}
	if st.Name == "" {
		return nil, errors.New("name is required")
	}
	if st.MinValue != nil && st.MaxValue != nil && *st.MinValue > *st.MaxValue {
		return nil, errors.New("min_value must be less than or equal to max_value")
	}
	if err := s.repo.Create(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}

type SensorService struct {
	repo     *repositories.SensorRepository
	typeRepo *repositories.SensorTypeRepository
	calRepo  *repositories.SensorCalibrationRepository
}

func NewSensorService(repo *repositories.SensorRepository, typeRepo *repositories.SensorTypeRepository, calRepo *repositories.SensorCalibrationRepository) *SensorService {
	return &SensorService{repo: repo, typeRepo: typeRepo, calRepo: calRepo}
}

func (s *SensorService) GetSensorByID(ctx context.Context, id uint) (*models.Sensor, error) {
	return s.repo.FindSensorByID(ctx, id)
}

// GetPaginatedSensors teruskan params filter/page/limit/sort ke repository.
func (s *SensorService) GetPaginatedSensors(ctx context.Context, params map[string]any) ([]models.Sensor, int, error) {
	return s.repo.GetPaginatedSensors(ctx, params)
}

func (s *SensorService) CreateSensor(ctx context.Context, sensor *models.Sensor) (*models.Sensor, error) {
	if sensor == nil {
		return nil, errors.New("sensor is required")
	}
	if sensor.Name == "" {
		return nil, errors.New("name is required")
	}
	if sensor.SensorTypeID == 0 {
		return nil, errors.New("sensor_type_id is required")
	}
	if _, err := s.typeRepo.FindByID(ctx, sensor.SensorTypeID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("sensor_type_id not found")
		}
		return nil, err
	}
	if err := s.repo.CreateSensor(ctx, sensor); err != nil {
		return nil, err
	}
	return s.repo.FindSensorByID(ctx, sensor.ID)
}

// UpdateSensor partial update sensor. Keys yang diizinkan: name, sensor_type_id, status.
func (s *SensorService) UpdateSensor(ctx context.Context, id uint, data map[string]any) (*models.Sensor, error) {
	if id == 0 {
		return nil, errors.New("sensor id is required")
	}
	if len(data) == 0 {
		return nil, errors.New("no fields to update")
	}

	patch := map[string]any{}
	for key, val := range data {
		switch key {
		case "name":
			name, ok := val.(string)
			if !ok || name == "" {
				return nil, errors.New("name must be a non-empty string")
			}
			patch[key] = name
		case "sensor_type_id":
			typeID, ok := pkg.ToUintFilter(val)
			if !ok {
				return nil, errors.New("invalid sensor_type_id")
			}
			if _, err := s.typeRepo.FindByID(ctx, typeID); err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, errors.New("sensor_type_id not found")
				}
				return nil, err
			}
			patch[key] = typeID
		case "status":
			status, ok := val.(bool)
			if !ok {
				return nil, errors.New("status must be a boolean")
			}
			patch[key] = status
		default:
			return nil, fmt.Errorf("field %s cannot be updated", key)
		}
	}

	if len(patch) == 0 {
		return nil, errors.New("no fields to update")
	}

	if _, err := s.repo.FindSensorByID(ctx, id); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateSensor(ctx, id, patch); err != nil {
		return nil, err
	}

	return s.repo.FindSensorByID(ctx, id)
}

func (s *SensorService) DeleteSensor(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("sensor id is required")
	}
	return s.repo.DeleteSensor(ctx, id)
}

// CreateCalibration tambah kalibrasi untuk sensor.
// from_date default now bila kosong, to_date default null.
func (s *SensorService) CreateCalibration(ctx context.Context, sensorID uint, cal *models.SensorCalibration) (*models.SensorCalibration, error) {
	if sensorID == 0 {
		return nil, errors.New("sensor id is required")
	}
	if cal == nil {
		return nil, errors.New("calibration is required")
	}
	if _, err := s.repo.FindSensorByID(ctx, sensorID); err != nil {
		return nil, err
	}
	cal.SensorID = sensorID
	if err := s.calRepo.CreateClosingPrevious(ctx, cal); err != nil {
		return nil, err
	}
	return cal, nil
}

// GetCalibrations ambil daftar kalibrasi sensor, terbaru dulu.
func (s *SensorService) GetCalibrations(ctx context.Context, sensorID uint) ([]models.SensorCalibration, error) {
	if sensorID == 0 {
		return nil, errors.New("sensor id is required")
	}
	if _, err := s.repo.FindSensorByID(ctx, sensorID); err != nil {
		return nil, err
	}
	return s.calRepo.FindBySensorID(ctx, sensorID)
}
