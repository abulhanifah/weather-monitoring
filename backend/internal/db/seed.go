package db

import (
	"log"

	"github.com/abulhanifah/weather-monitoring/internal/constants"
	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/internal/services"
	"gorm.io/gorm"
)

// Seed mengisi data awal device (3) dan user (2).
// Idempotent: baris yang sudah ada (berdasar ID / email) dilewati.
func Seed(database *gorm.DB) error {
	devices := []models.Device{
		{
			ID:          "DEV-001",
			Name:        "Device 1 - Stasiun Jakarta",
			Description: "Stasiun pemantau cuaca Jakarta",
			Status:      constants.StatusActive,
		},
		{
			ID:          "DEV-002",
			Name:        "Device 2 - Stasiun Depok",
			Description: "Stasiun pemantau cuaca Depok",
			Status:      constants.StatusMaintenance,
		},
		{
			ID:          "DEV-003",
			Name:        "Device 2 - Stasiun Bekasi",
			Description: "Stasiun pemantau cuaca Bekasi",
			Status:      constants.StatusOffline,
		},
	}

	for _, d := range devices {
		dev := d
		res := database.Where("id = ?", dev.ID).Attrs(dev).FirstOrCreate(&dev)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			log.Printf("Seed device %s sudah ada, dilewati", dev.ID)
		} else {
			log.Printf("Seed device %s berhasil", dev.ID)
		}
	}

	type seedUser struct {
		Name     string
		Email    string
		Password string
		Role     string
	}
	users := []seedUser{
		{Name: "Admin", Email: "admin@weather.local", Password: "admin123", Role: "admin"},
		{Name: "Operator", Email: "operator@weather.local", Password: "operator123", Role: "operator"},
	}

	for _, u := range users {
		var count int64
		if err := database.Model(&models.User{}).Where("email = ?", u.Email).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			log.Printf("Seed user %s sudah ada, dilewati", u.Email)
			continue
		}

		hash, err := services.HashPassword(u.Password)
		if err != nil {
			return err
		}

		if err := database.Create(&models.User{
			Name:     u.Name,
			Email:    u.Email,
			Password: hash,
			Role:     u.Role,
		}).Error; err != nil {
			return err
		}
		log.Printf("Seed user %s berhasil", u.Email)
	}

	log.Println("Seeding selesai")
	return nil
}
