package db

import (
	"log"

	"github.com/abulhanifah/weather-monitoring/internal/constants"
	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/internal/services"
	"gorm.io/gorm"
)

// Seed mengisi data awal location (4), sensor type (7), device (3) dan user (2).
// Idempotent: baris yang sudah ada (berdasar ID / name / email) dilewati.
func Seed(database *gorm.DB) error {
	locationIDs := map[string]uint{}
	locations := []string{"Jakarta", "Depok", "Bekasi", "Bogor"}
	for _, name := range locations {
		var loc models.Location
		res := database.Where("name = ?", name).Attrs(models.Location{Name: name}).FirstOrCreate(&loc)
		if res.Error != nil {
			return res.Error
		}
		locationIDs[name] = loc.ID
		if res.RowsAffected == 0 {
			log.Printf("Seed location %s sudah ada, dilewati", name)
		} else {
			log.Printf("Seed location %s berhasil", name)
		}
	}

	sensorTypes := []string{
		"temperature",
		"humidity",
		"air_pressure",
		"rainfall",
		"wind_speed",
		"wind_direction",
		"solar_radiation",
	}
	for _, name := range sensorTypes {
		var st models.SensorType
		res := database.Where("name = ?", name).Attrs(models.SensorType{Name: name}).FirstOrCreate(&st)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			log.Printf("Seed sensor type %s sudah ada, dilewati", name)
		} else {
			log.Printf("Seed sensor type %s berhasil", name)
		}
	}

	type seedDevice struct {
		Device   models.Device
		Location string
	}
	devices := []seedDevice{
		{
			Device: models.Device{
				ID:          "DEV-001",
				Name:        "Device 1 - Stasiun Jakarta",
				Description: "Stasiun pemantau cuaca Jakarta",
				Status:      constants.StatusActive,
			},
			Location: "Jakarta",
		},
		{
			Device: models.Device{
				ID:          "DEV-002",
				Name:        "Device 2 - Stasiun Depok",
				Description: "Stasiun pemantau cuaca Depok",
				Status:      constants.StatusMaintenance,
			},
			Location: "Depok",
		},
		{
			Device: models.Device{
				ID:          "DEV-003",
				Name:        "Device 2 - Stasiun Bekasi",
				Description: "Stasiun pemantau cuaca Bekasi",
				Status:      constants.StatusOffline,
			},
			Location: "Bekasi",
		},
	}

	for _, d := range devices {
		dev := d.Device
		var locID uint
		hasLoc := false
		if id, ok := locationIDs[d.Location]; ok {
			locID, hasLoc = id, true
			dev.LocationID = &locID
		}
		res := database.Where("id = ?", dev.ID).Attrs(dev).FirstOrCreate(&dev)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			log.Printf("Seed device %s sudah ada, dilewati", dev.ID)
			// Backfill location untuk device lama yang belum punya lokasi
			if hasLoc {
				database.Model(&models.Device{}).Where("id = ? AND location_id IS NULL", dev.ID).Update("location_id", locID)
			}
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
