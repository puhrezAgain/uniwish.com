/*
uniwish.com/interal/config/load

module dedicated to logic around loading config
*/
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

func Load() (*Config, error) {
	cfg := &Config{}
	cfg.Env = getenv("APP_ENV", "dev")

	cfg.DBURL = getenv("DATABASE_URL", "")
	if cfg.DBURL == "" {
		return nil, errors.New("DATABASE_URL required")
	}

	port, err := getenvInt("APP_PORT", 8080)
	if err != nil {
		return nil, err
	}

	cfg.Port = port

	worker_interval, err := getenvInt("WORKER_POLL_INTERVAL", 1)
	if err != nil {
		return nil, err
	}

	cfg.WorkerPollInterval = time.Duration(worker_interval) * time.Second

	worker_tolerance, err := getenvInt("WORKER_FAILURE_TOLERANCE", 10)
	if err != nil {
		return nil, err
	}

	cfg.WorkerFailureTolerance = worker_tolerance
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
