/*
uniwish.com/internal/api/services/health

contains logic of application's self health service
*/
package services

import "context"

type HealthService struct{}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) Check(ctx context.Context) error {
	return nil
}
