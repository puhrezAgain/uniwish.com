package config

import "time"

type Config struct {
	ENV                  string
	PORT                 int
	DBURL                string
	WORKER_POLL_INTERVAL time.Duration
}
