package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

// ProductFilters holds optional filters for listing products.
type ProductFilters struct {
	CategoryCode  string
	PriceLessThan *decimal.Decimal
}

func (r *ProductsRepository) GetAllProducts(offset, limit int, filters ProductFilters) ([]Product, int64, error) {
	var total int64
	if err := filteredProductsQuery(r.db, filters).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var products []Product
	query := filteredProductsQuery(r.db, filters).Preload("Variants").Preload("Category")
	if err := query.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// filteredProductsQuery builds a fresh query every time it's called, so the same
// filters can be applied independently to the count query and to the page query.
func filteredProductsQuery(db *gorm.DB, filters ProductFilters) *gorm.DB {
	query := db.Model(&Product{})

	if filters.CategoryCode != "" {
		query = query.Where("category_id = (SELECT id FROM product_categories WHERE code = ?)", filters.CategoryCode)
	}
	if filters.PriceLessThan != nil {
		query = query.Where("price < ?", *filters.PriceLessThan)
	}

	return query
}
