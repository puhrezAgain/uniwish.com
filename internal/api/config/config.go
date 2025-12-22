/*
uniwish.com/interal/api/config

centralizes our config specification
*/
package config

import "time"

type Config struct {
	// APP_ENV, dev
	Env string
	// APP_PORT, 8080
	Port int
	// DATABASE_URL,
	DBURL string
	// WORKER_POLL_INTERVAL int, 1
	WorkerPollInterval time.Duration
	// WORKER_FAILURE_TOLERANCE, 10
	WorkerFailureTolerance int
}
