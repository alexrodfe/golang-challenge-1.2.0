package models

import (
	"github.com/shopspring/decimal"
)

// Variant represents a product variant in the catalog.
// It includes a unique name, SKU, and an optional price.
// Variants can be used to represent different configurations or options for a product.
type Variant struct {
	ID        uint             `gorm:"primaryKey"`
	ProductID uint             `gorm:"not null"`
	Name      string           `gorm:"not null"`
	SKU       string           `gorm:"uniqueIndex;not null"`
	Price     *decimal.Decimal `gorm:"type:decimal(10,2);null"`
}

func (v *Variant) TableName() string {
	return "product_variants"
}

// EffectivePrice returns the variant's own price, or the product's price
// when the variant has none set.
func (v Variant) EffectivePrice(product Product) decimal.Decimal {
	if v.Price != nil {
		return *v.Price
	}
	return product.Price
}
