/*
uniwish.com/interal/api/handlers/health_test

tests for health endpoint
*/
package services

import (
	"context"
	"testing"
)

func TestHealthService_Check_Ok(t *testing.T) {
	srv := NewHealthService()

	err := srv.Check(context.Background())

	if err != nil {
		t.Fatalf("expected nil got, %v", err)
	}
}
