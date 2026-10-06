package db

import (
	"fmt"
	"log"

	"github.com/abulhanifah/weather-monitoring/internal/config"
	"github.com/abulhanifah/weather-monitoring/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitPostgres(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal terhubung ke database PostgreSQL: %v", err)
	}

	// Migrasi semua model domain
	if err := Migrate(database); err != nil {
		log.Fatalf("Gagal melakukan auto migration: %v", err)
	}

	log.Println("Koneksi PostgreSQL & Auto Migration berhasil")
	return database
}

// Migrate menjalankan auto migration untuk semua model domain.
// Dipakai oleh service utama (via InitPostgres) dan migrator/seed.
func Migrate(database *gorm.DB) error {
	return database.AutoMigrate(
		&models.Device{},
		&models.APIKeyMeta{},
		&models.User{},
		&models.Location{},
		&models.SensorType{},
		&models.Sensor{},
		&models.SensorInstallation{},
		&models.SensorCalibration{},
		&models.DeviceStatusHistory{},
	)
}
