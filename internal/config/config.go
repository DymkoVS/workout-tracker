package config

import "os"

type Config struct {
	DatabaseURL    string
	ListenAddr     string
	SessionSecret  string
	UploadDir      string
	ImportAPIToken string
}

func Load() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://workout:workout_secret@localhost:5432/workout_tracker?sslmode=disable"),
		ListenAddr:     getEnv("LISTEN_ADDR", ":8080"),
		SessionSecret:  getEnv("SESSION_SECRET", "dev-secret-change-in-production!!"),
		UploadDir:      getEnv("UPLOAD_DIR", "./uploads"),
		ImportAPIToken: getEnv("IMPORT_API_TOKEN", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
