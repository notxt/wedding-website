package config

import "os"

type Config struct {
	Port          string
	DatabaseURL   string
	AccessCode    string
	SessionSecret string
}

func Load() Config {
	return Config{
		Port:          getenv("PORT", "8080"),
		DatabaseURL:   getenv("DATABASE_URL", "postgres://wedding:wedding@localhost:5432/wedding?sslmode=disable"),
		AccessCode:    getenv("ACCESS_CODE", "letmein"),
		SessionSecret: getenv("SESSION_SECRET", "dev-secret-do-not-use-in-prod"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
