/*
uniwish.com/interal/api/config

centralizes our config specification
*/
package config

import "time"

type Config struct {
	// APP_ENV
	Env string
	// APP_PORT
	Port int
	// DATABASE_URL
	DBURL string
	// WORKER_POLL_INTERVAL int
	WorkerPollInterval time.Duration
}
