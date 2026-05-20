package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryType string

const (
	CategoryIncome  CategoryType = "income"
	CategoryExpense CategoryType = "expense"
)

type Category struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID    `gorm:"type:uuid;not null;index"                       json:"userId"`
	Name      string       `gorm:"not null;size:100"                              json:"name"`
	Icon      string       `gorm:"not null;size:10"                               json:"icon"`
	Color     string       `gorm:"not null;size:7"                                json:"color"`
	Type      CategoryType `gorm:"not null;size:10"                               json:"type"`
	IsDefault bool         `gorm:"not null;default:false"                         json:"isDefault"`
	CreatedAt time.Time    `                                                      json:"createdAt"`
	UpdatedAt time.Time    `                                                      json:"updatedAt"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

var DefaultCategories = []struct {
	Name      string
	Icon      string
	Color     string
	Type      CategoryType
	IsDefault bool
}{
	// Expense categories
	{Name: "Food & Dining", Icon: "🍔", Color: "#F59E0B", Type: CategoryExpense, IsDefault: true},
	{Name: "Transportation", Icon: "🚗", Color: "#3B82F6", Type: CategoryExpense, IsDefault: true},
	{Name: "Shopping", Icon: "🛍️", Color: "#8B5CF6", Type: CategoryExpense, IsDefault: true},
	{Name: "Housing & Rent", Icon: "🏠", Color: "#EF4444", Type: CategoryExpense, IsDefault: true},
	{Name: "Healthcare", Icon: "🏥", Color: "#10B981", Type: CategoryExpense, IsDefault: true},
	{Name: "Entertainment", Icon: "🎬", Color: "#F97316", Type: CategoryExpense, IsDefault: true},
	{Name: "Education", Icon: "📚", Color: "#6366F1", Type: CategoryExpense, IsDefault: true},
	{Name: "Utilities", Icon: "⚡", Color: "#14B8A6", Type: CategoryExpense, IsDefault: true},
	{Name: "Travel", Icon: "✈️", Color: "#0EA5E9", Type: CategoryExpense, IsDefault: true},
	{Name: "Personal Care", Icon: "💆", Color: "#EC4899", Type: CategoryExpense, IsDefault: true},
	{Name: "Subscriptions", Icon: "📱", Color: "#7C3AED", Type: CategoryExpense, IsDefault: true},
	{Name: "Other Expense", Icon: "📝", Color: "#6B7280", Type: CategoryExpense, IsDefault: true},
	// Income categories
	{Name: "Salary", Icon: "💼", Color: "#10B981", Type: CategoryIncome, IsDefault: true},
	{Name: "Freelance", Icon: "💻", Color: "#06B6D4", Type: CategoryIncome, IsDefault: true},
	{Name: "Investments", Icon: "📈", Color: "#8B5CF6", Type: CategoryIncome, IsDefault: true},
	{Name: "Gifts", Icon: "🎁", Color: "#F59E0B", Type: CategoryIncome, IsDefault: true},
	{Name: "Other Income", Icon: "💰", Color: "#6B7280", Type: CategoryIncome, IsDefault: true},
}

func SeedDefaultCategories(userID uuid.UUID) []Category {
	cats := make([]Category, len(DefaultCategories))
	for i, d := range DefaultCategories {
		cats[i] = Category{
			ID:        uuid.New(),
			UserID:    userID,
			Name:      d.Name,
			Icon:      d.Icon,
			Color:     d.Color,
			Type:      d.Type,
			IsDefault: d.IsDefault,
		}
	}
	return cats
}
