package models

import (
	"errors"

	"gorm.io/gorm"
)

// ErrCategoryAlreadyExists is returned when a category with the same code already exists.
var ErrCategoryAlreadyExists = errors.New("category already exists")

type CategoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) *CategoriesRepository {
	return &CategoriesRepository{
		db: db,
	}
}

func (r *CategoriesRepository) GetAllCategories() ([]ProductCategory, error) {
	var categories []ProductCategory
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoriesRepository) CreateCategory(category ProductCategory) (*ProductCategory, error) {
	if err := r.db.Create(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrCategoryAlreadyExists
		}
		return nil, err
	}
	return &category, nil
}
