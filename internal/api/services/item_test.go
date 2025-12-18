/*
uniwish.com/interal/api/services/health_test

tests for item service
*/
package services

import (
	"context"
	"testing"

	"uniwish.com/internal/api/errors"
)

func TestItemService(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedError  error
		expectedResult string
	}{
		{
			name:           "no_url",
			url:            "",
			expectedError:  errors.ErrInputInvalid,
			expectedResult: "",
		},
		{
			name:           "parse_error",
			url:            "http://[::1",
			expectedError:  errors.ErrInputInvalid,
			expectedResult: "",
		},
		{
			name:           "no_scheme",
			url:            "example.com",
			expectedError:  errors.ErrInputInvalid,
			expectedResult: "",
		},
		{
			name:           "no_host",
			url:            "http://",
			expectedError:  errors.ErrInputInvalid,
			expectedResult: "",
		},
		{
			name:           "unsupported_host",
			url:            "http://whatever.com",
			expectedError:  errors.ErrStoreUnsupported,
			expectedResult: "",
		},
		{
			name:           "healthy",
			url:            "http://store.com",
			expectedError:  nil,
			expectedResult: "fakeid",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srv := NewItemsService()
			r, e := srv.Create(context.Background(), tt.url)

			if e != tt.expectedError {
				t.Fatalf("expected error %v, recieved %v", tt.expectedError, e)
			}

			if r != tt.expectedResult {
				t.Fatalf("exepected result %s, received %s", tt.expectedResult, r)
			}
		})
	}
}
