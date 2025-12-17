package config

import (
	"os"
	"strconv"
)

func Load() (*Config, error) {
	cfg := &Config{}
	cfg.Env = getenv("APP_ENV", "dev")
	cfg.Port = getenvInt("Port", 8080)

	return cfg, nil
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			panic("invalid int for " + key)
		}
		return i
	}
	return def
}
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
