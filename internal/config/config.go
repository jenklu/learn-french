package config

import (
	"os"
)

type Config struct {
	DatabaseURL      string
	Port            string
	SessionSecret   string
	CSRFSecret      string
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	TTSAPIKey       string
	FrontendURL     string
	MistralAPIKey   string
}

func Load() *Config {
	return &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/learn_french?sslmode=disable"),
		Port:           getEnv("PORT", "8080"),
		SessionSecret:  getEnv("SESSION_SECRET", "super-secret-key-change-in-production"),
		CSRFSecret:     getEnv("CSRF_SECRET", "csrf-secret-key-change-in-production"),
		VAPIDPublicKey: getEnv("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey: getEnv("VAPID_PRIVATE_KEY", ""),
		TTSAPIKey:      getEnv("TTS_API_KEY", ""),
		FrontendURL:    getEnv("FRONTEND_URL", "http://localhost:5173"),
		MistralAPIKey:  getEnv("MISTRAL_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
