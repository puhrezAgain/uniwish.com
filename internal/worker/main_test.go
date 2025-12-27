/*
uniwish.com/interal/api/worker/main_test

centralizes DB testing configuration and initiation
*/
package worker

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"uniwish.com/internal/testutil"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	if os.Getenv("INTEGRATION_TESTS") != "" {
		testDB = testutil.OpenDB()
	}
	code := m.Run()

	if testDB != nil {
		testDB.Close()
	}
	os.Exit(code)
}
