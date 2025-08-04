package domain

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price"`
	Stock       int       `json:"stock" db:"stock"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type ProductRepository interface {
	Create(product *Product) error
	GetByID(id uuid.UUID) (*Product, error)
	GetAll(limit, offset int) ([]*Product, error)
	Update(product *Product) error
	Delete(id uuid.UUID) error
}

type ProductUsecase interface {
	Create(product *Product) error
	GetByID(id uuid.UUID) (*Product, error)
	GetAll(limit, offset int) ([]*Product, error)
	Update(product *Product) error
	Delete(id uuid.UUID) error
}
