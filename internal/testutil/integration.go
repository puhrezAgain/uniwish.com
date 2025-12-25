package testutil

import (
	"database/sql"
	"log"
	"os"
	"testing"
)

func RequireIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("integration tests disabled")
	}
}

func OpenDB() *sql.DB {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		log.Println("INTEGRATION_TESTS not set, skipping")
		os.Exit(0)
	}
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	var err error
	testDB, err := sql.Open("postgres", dsn)

	if err != nil {
		log.Fatal("database connection failed", err)
	}

	if err := testDB.Ping(); err != nil {
		log.Fatal("database ping failed", err)
	}

	return testDB
}

func TruncateTables(t *testing.T, db *sql.DB) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		log.Println("INTEGRATION_TESTS not set, skipping")
		os.Exit(0)
	}
	t.Helper()

	_, err := db.Exec(`
		TRUNCATE TABLE
			prices,
			products,
			scrape_requests
		RESTART IDENTITY
		CASCADE
	`)
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}
