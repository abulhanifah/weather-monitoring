package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort            string
	AuthSecret         string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	MQTTBroker         string
	MQTTTopic          string
	MQTTClientID       string
	PrefixAPIKey       string
	SchedulerInterval  time.Duration
	HeartbeatThreshold time.Duration
	CORSAllowedOrigins string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("File .env tidak ditemukan, menggunakan environment variable sistem")
	}

	return &Config{
		AppPort:    getEnv("APP_PORT", "8080"),
		AuthSecret: getEnv("AUTH_SECRET", "top5ecret!"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "db$ecr3t"),
		DBName:     getEnv("DB_NAME", "weather_monitoring"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// TODO : MQTT
		// MQTTBroker:   getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		// MQTTTopic:    getEnv("MQTT_TOPIC", "telemetry/ingest"),
		// MQTTClientID: getEnv("MQTT_CLIENT_ID", "my-go-service"),

		PrefixAPIKey: getEnv("PREFIX_API_KEY", "apikey_"),

		SchedulerInterval:  getDuration("SCHEDULER_INTERVAL", 5*time.Minute),
		HeartbeatThreshold: getDuration("HEARTBEAT_THRESHOLD", 5*time.Minute),

		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		log.Printf("ENV %s tidak valid (%q), pakai default %s", key, value, fallback)
		return fallback
	}
	return d
}
