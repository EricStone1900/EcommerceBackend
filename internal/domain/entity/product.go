package entity

import "time"

// ProductStatus represents the sale status of a product.
type ProductStatus string

const (
	ProductStatusOnSale  ProductStatus = "on_sale"
	ProductStatusOffSale ProductStatus = "off_sale"
)

// Product represents a product in the ecommerce system.
// Note: DeletedAt is intentionally omitted — it's infra-only (gorm.DeletedAt).
type Product struct {
	ID          uint          `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Price       float64       `json:"price"`
	Stock       int           `json:"stock"`
	Status      ProductStatus `json:"status"`
	CreatedBy   uint          `json:"created_by"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
