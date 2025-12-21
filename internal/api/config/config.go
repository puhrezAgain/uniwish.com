package config

import "time"

type Config struct {
	Env                string
	Port               int
	DBURL              string
	WorkerPollInterval time.Duration
}
