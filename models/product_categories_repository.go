package models

import (
	"errors"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// ErrCategoryAlreadyExists is returned when a category with the same code already exists.
var ErrCategoryAlreadyExists = errors.New("category already exists")

// uniqueViolationCode is the Postgres error code for a unique constraint violation.
const uniqueViolationCode = "23505"

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
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == uniqueViolationCode {
			return nil, ErrCategoryAlreadyExists
		}
		return nil, err
	}
	return &category, nil
}
