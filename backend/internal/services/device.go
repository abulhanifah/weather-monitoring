package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/abulhanifah/weather-monitoring/internal/config"
	"github.com/abulhanifah/weather-monitoring/internal/constants"
	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/internal/repositories"
	"github.com/abulhanifah/weather-monitoring/pkg"
	"gorm.io/gorm"
)

type DeviceService struct {
	config *config.Config
	repo   *repositories.DeviceRepository
}

func NewService(cf *config.Config, repo *repositories.DeviceRepository) *DeviceService {
	return &DeviceService{config: cf, repo: repo}
}

func (s *DeviceService) GetByID(ctx context.Context, id string) (*models.Device, error) {
	return s.repo.FindByID(ctx, id)
}

// GetPaginated teruskan params filter/page/limit/sort ke repository.
// Return: list device, total data, error.
func (s *DeviceService) GetPaginated(ctx context.Context, params map[string]any) ([]models.Device, int, error) {
	return s.repo.GetPaginated(ctx, params)
}

func (s *DeviceService) SetStatus(ctx context.Context, id string, status string) error {
	// hanya boleh set status Active, Maintenance, atau Disabled dari admin
	if !pkg.Contains([]string{constants.StatusActive, constants.StatusMaintenance, constants.StatusDisabled}, status) {
		return errors.New("Invalid device status. You only can set status Active, Maintenance, atau Disabled")
	}
	return s.repo.Update(ctx, id, map[string]any{"status": status})
}

func (s *DeviceService) CreateDevice(ctx context.Context, dev *models.Device) (*models.CreateDeviceResponse, error) {
	// cek apakah device id sudah ada
	exist, err := s.repo.FindByID(ctx, dev.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		slog.ErrorContext(ctx, "Error on checking device id", slog.Any("id", dev.ID))
		return nil, fmt.Errorf("Error on Creating Device : %s", err.Error())
	}

	if exist != nil {
		return nil, fmt.Errorf("Error on Creating Device : Device ID was registered")
	}

	// simpan device
	err = s.repo.Create(ctx, dev)
	if err != nil {
		slog.ErrorContext(ctx, "Error create device", slog.Any("id", dev.ID), slog.Any("error", err.Error()))
		return nil, fmt.Errorf("Error on Creating Device : %s", err.Error())
	}

	// response
	res := &models.CreateDeviceResponse{Device: dev}

	// generate api key, jika berhasil langsung tampilkan ketika create device. Jika gagal bisa regenerate nanti di frontend
	raw, meta, err := s.CreateDeviceAPIKey(ctx, dev.ID)
	if err == nil {
		res.RawKey = raw
		res.Meta = *meta
	}

	return res, nil
}

func (s *DeviceService) CreateDeviceAPIKey(ctx context.Context, deviceID string) (string, *models.APIKeyMeta, error) {
	bytes, err := pkg.GetCryptoRandomBytes()
	if err != nil {
		slog.ErrorContext(ctx, "Error GetCryptoRandomBytes", slog.Any("id", deviceID), slog.Any("error", err.Error()))
		return "", nil, err
	}
	// generate raw key hanya akan ditampilkan pertama kali
	rawRandomHex := hex.EncodeToString(bytes)
	rawKey := fmt.Sprintf("%s%s", s.config.PrefixAPIKey, rawRandomHex)

	// simpan masking untuk keperluan frontend
	masking := fmt.Sprintf("%s...", rawRandomHex[:6])

	// hash api key untuk disimpan di db untuk validasi
	hashBytes := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hashBytes[:])

	meta := &models.APIKeyMeta{
		DeviceID:  deviceID,
		Masking:   masking,
		Hash:      keyHash,
		CreatedAt: time.Now(),
		IsRevoked: false,
	}

	if err = s.repo.SaveAPIKey(ctx, meta); err != nil {
		slog.ErrorContext(ctx, "Error SaveAPIKey", slog.Any("id", deviceID), slog.Any("error", err.Error()))
		return "", nil, err
	}
	return rawKey, meta, nil
}
