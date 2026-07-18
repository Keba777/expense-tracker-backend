package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoanPayment struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LoanID    uuid.UUID      `gorm:"type:uuid;not null;index"                       json:"loanId"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index"                       json:"userId"`
	Amount    float64        `gorm:"not null;check:chk_loanpay_amount,amount > 0"   json:"amount"`
	Date      time.Time      `gorm:"not null;index"                                 json:"date"`
	Notes     *string        `gorm:"size:500"                                       json:"notes,omitempty"`
	CreatedAt time.Time      `                                                       json:"createdAt"`
	UpdatedAt time.Time      `                                                       json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                          json:"-"`

	Loan Loan `gorm:"foreignKey:LoanID" json:"-"`
}

func (p *LoanPayment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
