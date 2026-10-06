package models

// CategoryCode identifies a product category (e.g. for filtering by category).
type CategoryCode string

const (
	CategoryClothing    CategoryCode = "CLOTHING"
	CategoryShoes       CategoryCode = "SHOES"
	CategoryAccessories CategoryCode = "ACCESSORIES"
)

// ProductCategory represents a category products can be classified under.
type ProductCategory struct {
	ID   uint         `gorm:"primaryKey"`
	Code CategoryCode `gorm:"uniqueIndex;not null"`
	Name string       `gorm:"not null"`
}

func (c *ProductCategory) TableName() string {
	return "product_categories"
}
