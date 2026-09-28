package models

import (
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

func (r *ProductsRepository) GetAllProducts(offset, limit int) ([]Product, int64, error) {
	var products []Product
	if err := r.db.Preload("Variants").Preload("Category").
		Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	var total int64
	if err := r.db.Model(&Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
