package config

import (
	"fmt"
	"os"
	"strconv"
)

func Load() (*Config, error) {
	cfg := &Config{}
	cfg.Env = getenv("APP_ENV", "dev")
	port, err := getenvInt("APP_PORT", 8080)

	if err != nil {
		return nil, err
	}

	cfg.Port = port

	return cfg, nil
}

func getenvInt(key string, def int) (int, error) {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("invalid int for %s: %w", key, err)
		}
		return i, nil
	}
	return def, nil
}
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
