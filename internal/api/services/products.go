/*
uniwish.com/internal/api/services/products

contains logic of application's product service
*/
package services

import "uniwish.com/internal/api/repository"

type ProductReaderService struct {
	repo repository.ProductReader
}
