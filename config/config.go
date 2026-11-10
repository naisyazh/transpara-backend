package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config menyimpan semua konfigurasi aplikasi yang dibaca dari file .env.
type Config struct {
	AppPort string
	AppName string

	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	JWTSecret      string
	JWTIssuer      string
	JWTExpiredHour int // dalam satuan jam
}

// Load membaca file .env dan mengembalikan struct Config yang sudah terisi.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Gagal memuat file .env — pastikan file .env ada di root project")
	}

	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppName: getEnv("APP_NAME", "Transpara API"),

		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "3306"),
		DBUser: getEnv("DB_USER", "root"),
		DBPass: getEnv("DB_PASS", ""),
		DBName: getEnv("DB_NAME", "transpara"),

		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTIssuer:      getEnv("JWT_ISSUER", "transpara-api"),
		JWTExpiredHour: getEnvInt("JWT_EXPIRED_HOUR", 24),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt membaca environment variable sebagai integer.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
