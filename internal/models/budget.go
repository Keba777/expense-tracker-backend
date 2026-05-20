package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BudgetPeriod string

const (
	BudgetWeekly  BudgetPeriod = "weekly"
	BudgetMonthly BudgetPeriod = "monthly"
	BudgetYearly  BudgetPeriod = "yearly"
)

type Budget struct {
	ID         uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"                                      json:"id"`
	UserID     uuid.UUID    `gorm:"type:uuid;not null;index"                                                            json:"userId"`
	CategoryID uuid.UUID    `gorm:"type:uuid;not null"                                                                  json:"categoryId"`
	Name       string       `gorm:"not null;size:100"                                                                   json:"name"`
	Amount     float64      `gorm:"not null;check:chk_budget_amount,amount > 0"                                         json:"amount"`
	Period     BudgetPeriod `gorm:"not null;size:10;check:chk_budget_period,period IN ('weekly','monthly','yearly')"    json:"period"`
	Year       int          `gorm:"not null"                                                                            json:"year"`
	Month      *int         `                                                                                           json:"month,omitempty"`
	CreatedAt  time.Time    `                                                                                           json:"createdAt"`
	UpdatedAt  time.Time    `                                                                                           json:"updatedAt"`

	User     User     `gorm:"foreignKey:UserID"     json:"-"`
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (b *Budget) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type BudgetWithSpent struct {
	Budget
	Spent      float64 `json:"spent"`
	Remaining  float64 `json:"remaining"`
	Percentage float64 `json:"percentage"`
}
