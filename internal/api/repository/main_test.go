/*
uniwish.com/interal/api/repository/main_test

centralizes DB testing configuration and initiation
*/
package repository

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		log.Println("INTEGRATION_TESTS not set, skipping")
		os.Exit(0)
	}
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	var err error
	testDB, err = sql.Open("postgres", dsn)

	if err != nil {
		log.Fatal("database connection failed", err)
	}

	if err := testDB.Ping(); err != nil {
		log.Fatal("database ping failed", err)
	}

	code := m.Run()

	testDB.Close()
	os.Exit(code)
}

func requireIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("integration tests disabled")
	}
}
