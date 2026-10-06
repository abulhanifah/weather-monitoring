package services

import (
	"context"
	"errors"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/internal/repositories"
)

var (
	// ErrSensorTypeAlreadyInstalled device sudah punya sensor aktif bertipe sama.
	ErrSensorTypeAlreadyInstalled = errors.New("device already has an active sensor of the same type")
)

type SensorInstallationService struct {
	repo       *repositories.SensorInstallationRepository
	deviceRepo *repositories.DeviceRepository
	sensorRepo *repositories.SensorRepository
}

func NewSensorInstallationService(
	repo *repositories.SensorInstallationRepository,
	deviceRepo *repositories.DeviceRepository,
	sensorRepo *repositories.SensorRepository,
) *SensorInstallationService {
	return &SensorInstallationService{repo: repo, deviceRepo: deviceRepo, sensorRepo: sensorRepo}
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
