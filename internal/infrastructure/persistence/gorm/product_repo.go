package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/internal/domain/port"
)

// productModel is the GORM-specific persistence model for products.
// It carries gorm.DeletedAt for soft delete — the domain entity does not.
type productModel struct {
	ID          uint           `gorm:"primaryKey"`
	Name        string         `gorm:"not null;size:255"`
	Description string         `gorm:"not null"`
	Price       float64        `gorm:"not null;type:decimal(10,2)"`
	Stock       int            `gorm:"not null;default:0"`
	Status      string         `gorm:"not null;size:20;default:on_sale"`
	CreatedBy   uint           `gorm:"not null"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the default table name.
func (productModel) TableName() string {
	return "products"
}

// toEntity converts the GORM model to a domain entity.
func (m *productModel) toEntity() *entity.Product {
	return &entity.Product{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Price:       m.Price,
		Stock:       m.Stock,
		Status:      entity.ProductStatus(m.Status),
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// toProductModel converts a domain entity to the GORM model.
func toProductModel(p *entity.Product) *productModel {
	return &productModel{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Status:      string(p.Status),
		CreatedBy:   p.CreatedBy,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// ProductRepository implements port.ProductRepository using GORM.
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new GORM-backed ProductRepository.
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create persists a new product and populates the ID and timestamps.
func (r *ProductRepository) Create(ctx context.Context, product *entity.Product) error {
	model := toProductModel(product)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	// Copy generated fields back
	*product = *model.toEntity()
	return nil
}

// Update saves all fields of an existing product.
func (r *ProductRepository) Update(ctx context.Context, product *entity.Product) error {
	model := toProductModel(product)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	*product = *model.toEntity()
	return nil
}

// Delete performs a soft delete on a product by setting deleted_at.
func (r *ProductRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&productModel{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete product: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetByID retrieves a product by its primary key. Returns nil if not found.
func (r *ProductRepository) GetByID(ctx context.Context, id uint) (*entity.Product, error) {
	var model productModel
	err := r.db.WithContext(ctx).First(&model, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}
	return model.toEntity(), nil
}

// List returns a paginated, filtered, and sorted list of products.
func (r *ProductRepository) List(ctx context.Context, filter port.ProductFilter) (*port.ProductListResult, error) {
	// Defaults
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	// Build query with filter
	query := r.db.WithContext(ctx).Model(&productModel{})

	// Name fuzzy search (case-insensitive)
	if filter.Name != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+filter.Name+"%")
	}

	// Status filter
	if filter.Status != nil {
		query = query.Where("status = ?", string(*filter.Status))
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	// Sorting (whitelist-based)
	sortBy := "created_at"
	if filter.SortBy == "price" {
		sortBy = "price"
	}
	order := "DESC"
	if !filter.SortDesc {
		order = "ASC"
	}

	// Fetch page
	var models []productModel
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order(fmt.Sprintf("%s %s", sortBy, order)).
		Offset(offset).
		Limit(filter.PageSize).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Convert to entities
	products := make([]*entity.Product, len(models))
	for i := range models {
		products[i] = models[i].toEntity()
	}

	return &port.ProductListResult{
		Products: products,
		Total:    total,
	}, nil
}
