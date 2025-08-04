package usecase

import (
	"fmt"

	"go-clean-v2/internal/domain"

	"github.com/google/uuid"
)

type productUsecase struct {
	productRepo domain.ProductRepository
}

func NewProductUsecase(productRepo domain.ProductRepository) domain.ProductUsecase {
	return &productUsecase{
		productRepo: productRepo,
	}
}

func (u *productUsecase) Create(product *domain.Product) error {
	if product.Name == "" {
		return fmt.Errorf("product name is required")
	}

	if product.Price < 0 {
		return fmt.Errorf("product price cannot be negative")
	}

	if product.Stock < 0 {
		return fmt.Errorf("product stock cannot be negative")
	}

	return u.productRepo.Create(product)
}

func (u *productUsecase) GetByID(id uuid.UUID) (*domain.Product, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid product ID")
	}

	return u.productRepo.GetByID(id)
}

func (u *productUsecase) GetAll(limit, offset int) ([]*domain.Product, error) {
	if limit <= 0 {
		limit = 10 // default limit
	}

	if limit > 100 {
		limit = 100 // max limit
	}

	if offset < 0 {
		offset = 0
	}

	return u.productRepo.GetAll(limit, offset)
}

func (u *productUsecase) Update(product *domain.Product) error {
	if product.ID == uuid.Nil {
		return fmt.Errorf("invalid product ID")
	}

	if product.Name == "" {
		return fmt.Errorf("product name is required")
	}

	if product.Price < 0 {
		return fmt.Errorf("product price cannot be negative")
	}

	if product.Stock < 0 {
		return fmt.Errorf("product stock cannot be negative")
	}

	// Check if product exists
	_, err := u.productRepo.GetByID(product.ID)
	if err != nil {
		return err
	}

	return u.productRepo.Update(product)
}

func (u *productUsecase) Delete(id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid product ID")
	}

	// Check if product exists
	_, err := u.productRepo.GetByID(id)
	if err != nil {
		return err
	}

	return u.productRepo.Delete(id)
}
