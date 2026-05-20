package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type TransactionType string
type Recurrence string

const (
	TransactionIncome  TransactionType = "income"
	TransactionExpense TransactionType = "expense"

	RecurrenceOnce    Recurrence = "once"
	RecurrenceDaily   Recurrence = "daily"
	RecurrenceWeekly  Recurrence = "weekly"
	RecurrenceMonthly Recurrence = "monthly"
)

type Transaction struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"                              json:"id"`
	UserID            uuid.UUID      `gorm:"type:uuid;not null;index;index:idx_tx_user_date"                             json:"userId"`
	CategoryID        uuid.UUID      `gorm:"type:uuid;not null;index"                                                    json:"categoryId"`
	Type              TransactionType `gorm:"not null;size:10;index;check:chk_tx_type,type IN ('income','expense')"       json:"type"`
	Amount            float64        `gorm:"not null;check:chk_tx_amount,amount > 0"                                     json:"amount"`
	Description       string         `gorm:"not null;size:255"                                                           json:"description"`
	Notes             *string        `gorm:"size:1000"                                                                   json:"notes,omitempty"`
	Date              time.Time      `gorm:"not null;index;index:idx_tx_user_date"                                       json:"date"`
	Recurrence        Recurrence     `gorm:"not null;default:'once';size:20"                                             json:"recurrence"`
	RecurrenceEndDate *time.Time     `                                                                                   json:"recurrenceEndDate,omitempty"`
	Tags              pq.StringArray `gorm:"type:text[]"                                                                 json:"tags"`
	CreatedAt         time.Time      `                                                                                   json:"createdAt"`
	UpdatedAt         time.Time      `                                                                                   json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index"                                                                       json:"-"`

	User     User     `gorm:"foreignKey:UserID"     json:"-"`
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type TransactionWithCategory struct {
	Transaction
	CategoryName  string `json:"categoryName"`
	CategoryIcon  string `json:"categoryIcon"`
	CategoryColor string `json:"categoryColor"`
}

type TransactionFilter struct {
	Type       TransactionType
	CategoryID *uuid.UUID
	FromDate   *time.Time
	ToDate     *time.Time
	Search     string
	Page       int
	PerPage    int
}

func (f *TransactionFilter) Offset() int {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PerPage <= 0 {
		f.PerPage = 20
	}
	return (f.Page - 1) * f.PerPage
}
