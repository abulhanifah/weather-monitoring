package services

import (
	"context"
	"errors"

	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/internal/repositories"
)

type LocationService struct {
	repo *repositories.LocationRepository
}

func NewLocationService(repo *repositories.LocationRepository) *LocationService {
	return &LocationService{repo: repo}
}

func (s *LocationService) GetByID(ctx context.Context, id uint) (*models.Location, error) {
	return s.repo.FindByID(ctx, id)
}

// GetPaginated teruskan params filter/page/limit/sort ke repository.
// Return: list location, total data, error.
func (s *LocationService) GetPaginated(ctx context.Context, params map[string]any) ([]models.Location, int, error) {
	return s.repo.GetPaginated(ctx, params)
}

func (s *LocationService) CreateLocation(ctx context.Context, loc *models.Location) (*models.Location, error) {
	if loc == nil {
		return nil, errors.New("location is required")
	}
	if loc.Name == "" {
		return nil, errors.New("name is required")
	}
	if err := s.repo.Create(ctx, loc); err != nil {
		return nil, err
	}
	return loc, nil
}
