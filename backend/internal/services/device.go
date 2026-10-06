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

// UpdateDevice partial update device. Keys yang diizinkan:
// name, description, status, location_id,
// longitude, latitude, altitude (null untuk mengosongkan).
func (s *DeviceService) UpdateDevice(ctx context.Context, id string, data map[string]any) (*models.Device, error) {
	if id == "" {
		return nil, errors.New("device id is required")
	}
	if len(data) == 0 {
		return nil, errors.New("no fields to update")
	}

	allowedStatus := []string{
		constants.StatusInstalled,
		constants.StatusMaintenance,
		constants.StatusDecomissioned,
	}

	patch := map[string]any{}
	for key, val := range data {
		switch key {
		case "name", "description":
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("field %s must be a string", key)
			}
			patch[key] = s
		case "status":
			status, ok := val.(string)
			if !ok || !pkg.Contains(allowedStatus, status) {
				return nil, fmt.Errorf("invalid status, allowed: %v", allowedStatus)
			}
			patch[key] = status
		case "location_id":
			if val == nil {
				patch[key] = nil
				continue
			}
			locID, ok := pkg.ToUintFilter(val)
			if !ok {
				return nil, errors.New("invalid location_id")
			}
			patch[key] = locID
		case "longitude", "latitude", "altitude":
			if val == nil {
				patch[key] = nil
				continue
			}
			f, ok := val.(float64)
			if !ok {
				return nil, fmt.Errorf("field %s must be a number", key)
			}
			patch[key] = f
		default:
			return nil, fmt.Errorf("field %s cannot be updated", key)
		}
	}

	if len(patch) == 0 {
		return nil, errors.New("no fields to update")
	}

	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, id, patch); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *DeviceService) DeleteDevice(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("device id is required")
	}
	return s.repo.Delete(ctx, id)
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
	masking := fmt.Sprintf("%s%s", s.config.PrefixAPIKey, rawRandomHex[:6])

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

// RotateDeviceAPIKey revoke semua api key aktif lalu generate yang baru.
// Raw key hanya dikembalikan 1x ke caller.
func (s *DeviceService) RotateDeviceAPIKey(ctx context.Context, deviceID string) (*models.GenerateAPIKeyResponse, error) {
	if deviceID == "" {
		return nil, errors.New("device id is required")
	}

	if _, err := s.repo.FindByID(ctx, deviceID); err != nil {
		return nil, err
	}

	if err := s.repo.RevokeAPIKeysByDeviceID(ctx, deviceID); err != nil {
		return nil, err
	}

	raw, meta, err := s.CreateDeviceAPIKey(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	return &models.GenerateAPIKeyResponse{RawKey: raw, Meta: *meta}, nil
}
