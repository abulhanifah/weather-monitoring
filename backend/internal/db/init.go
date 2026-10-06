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
	if err := database.AutoMigrate(
		&models.Device{},
		&models.APIKeyMeta{},
		&models.User{},
		&models.Location{},
		&models.SensorType{},
		&models.Sensor{},
		&models.SensorInstallation{},
		&models.SensorCalibration{},
		&models.DeviceStatusHistory{},
		&models.SensorReading{},
	); err != nil {
		return err
	}

	enableTimescaleHypertable(database)
	return nil
}

// enableTimescaleHypertable ubah sensor_readings menjadi hypertable.
// Non-fatal: bila ekstensi TimescaleDB tidak tersedia, tabel tetap
// dipakai sebagai tabel biasa agar service tetap jalan.
func enableTimescaleHypertable(database *gorm.DB) {
	if err := database.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb").Error; err != nil {
		log.Printf("TimescaleDB tidak tersedia, sensor_readings dipakai sebagai tabel biasa: %v", err)
		return
	}
	if err := database.Exec("SELECT create_hypertable('sensor_readings', 'reading_time_origin', if_not_exists => TRUE)").Error; err != nil {
		log.Printf("Gagal membuat hypertable sensor_readings, dipakai sebagai tabel biasa: %v", err)
		return
	}
	log.Println("Hypertable sensor_readings aktif")
}
