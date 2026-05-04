package config

import (
	"errors"
	"os"
)

type Config struct {
	Port          string
	DBDSN         string
	SessionSecret string
	SRSHookSecret string
	SRSAPIUrl     string
	MediaRoot     string
	PublicBaseURL string
	HLSBaseURL    string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:          getEnv("PORT", "8888"),
		DBDSN:         os.Getenv("DB_DSN"),
		SessionSecret: os.Getenv("SESSION_SECRET"),
		SRSHookSecret: os.Getenv("SRS_HOOK_SECRET"),
		SRSAPIUrl:     getEnv("SRS_API_URL", "http://localhost:1985/api/v1"),
		MediaRoot:     getEnv("MEDIA_ROOT", "./uploads"),
		PublicBaseURL: getEnv("PUBLIC_BASE_URL", "http://localhost:8888"),
		HLSBaseURL:    getEnv("HLS_BASE_URL", "http://localhost:8080/live"),
	}

	var errs []error
	if cfg.DBDSN == "" {
		errs = append(errs, errors.New("DB_DSN is required"))
	}
	if cfg.SessionSecret == "" {
		errs = append(errs, errors.New("SESSION_SECRET is required"))
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
