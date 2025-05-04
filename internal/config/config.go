package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// HTTP
	Port int

	// Postgres
	DBHost string
	DBPort int
	DBUser string
	DBPass string
	DBName string

	// PGP
	PGPPublicKeyPath  string
	PGPPrivateKeyPath string
	PGPPrivatePass    string

	// HMAC
	HMACSecret string

	// JWT
	JWTSecret string

	// SMTP
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
}

func Load() *Config {
	// подгружаем .env, если есть
	_ = godotenv.Load()

	getStr := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}
	getInt := func(key string, def int) int {
		if v := os.Getenv(key); v != "" {
			i, err := strconv.Atoi(v)
			if err != nil {
				log.Fatalf("invalid %s: %v", key, err)
			}
			return i
		}
		return def
	}

	cfg := &Config{
		Port:              getInt("SERVER_PORT", 8080),
		DBHost:            getStr("DB_HOST", ""),
		DBPort:            getInt("DB_PORT", 5432),
		DBUser:            getStr("DB_USER", ""),
		DBPass:            getStr("DB_PASS", ""),
		DBName:            getStr("DB_NAME", ""),
		PGPPublicKeyPath:  getStr("PGP_PUBLIC_KEY_PATH", "keys/pub.asc"),
		PGPPrivateKeyPath: getStr("PGP_PRIVATE_KEY_PATH", "keys/private.asc"),
		PGPPrivatePass:    getStr("PGP_PASSPHRASE", ""),
		HMACSecret:        getStr("HMAC_SECRET", ""),
		JWTSecret:         getStr("JWT_SECRET", ""),
		SMTPHost:          getStr("SMTP_HOST", ""),
		SMTPPort:          getInt("SMTP_PORT", 587),
		SMTPUser:          getStr("SMTP_USER", ""),
		SMTPPass:          getStr("SMTP_PASS", ""),
	}

	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBPass == "" || cfg.DBName == "" {
		log.Fatal("database configuration is not complete")
	}
	if cfg.HMACSecret == "" || cfg.JWTSecret == "" {
		log.Fatal("HMAC_SECRET and JWT_SECRET must be set")
	}
	return cfg
}
