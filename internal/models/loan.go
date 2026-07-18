package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoanDirection string
type LoanStatus string

const (
	LoanLent     LoanDirection = "lent"
	LoanBorrowed LoanDirection = "borrowed"

	LoanOutstanding   LoanStatus = "outstanding"
	LoanPartiallyPaid LoanStatus = "partially_paid"
	LoanSettled       LoanStatus = "settled"
)

type Loan struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"                                                                            json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index"                                                                                                  json:"userId"`
	PersonID    uuid.UUID      `gorm:"type:uuid;not null;index"                                                                                                  json:"personId"`
	Direction   LoanDirection  `gorm:"not null;size:10;index;check:chk_loan_direction,direction IN ('lent','borrowed')"                                          json:"direction"`
	Amount      float64        `gorm:"not null;check:chk_loan_amount,amount > 0"                                                                                 json:"amount"`
	Description string         `gorm:"not null;size:255"                                                                                                         json:"description"`
	Notes       *string        `gorm:"size:1000"                                                                                                                 json:"notes,omitempty"`
	Date        time.Time      `gorm:"not null;index"                                                                                                            json:"date"`
	DueDate     *time.Time     `                                                                                                                                 json:"dueDate,omitempty"`
	Status      LoanStatus     `gorm:"not null;default:'outstanding';size:20;index;check:chk_loan_status,status IN ('outstanding','partially_paid','settled')" json:"status"`
	CreatedAt   time.Time      `                                                                                                                                 json:"createdAt"`
	UpdatedAt   time.Time      `                                                                                                                                 json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                                                                                                     json:"-"`

	User     User          `gorm:"foreignKey:UserID"   json:"-"`
	Person   Person        `gorm:"foreignKey:PersonID" json:"person,omitempty"`
	Payments []LoanPayment `gorm:"foreignKey:LoanID"   json:"payments,omitempty"`
}

func (l *Loan) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

// LoanWithBalance is the read model returned by every loan endpoint.
// PaidAmount/RemainingAmount are computed on read (Amount - SUM(payments)),
// never stored. Status IS stored on Loan and kept in sync by LoanService,
// since it's app-level state used for filtering/indexing, not a pure aggregate.
type LoanWithBalance struct {
	Loan
	PaidAmount      float64 `json:"paidAmount"`
	RemainingAmount float64 `json:"remainingAmount"`
}

type LoanFilter struct {
	PersonID  *uuid.UUID
	Direction LoanDirection
	Status    LoanStatus
	Page      int
	PerPage   int
}

func (f *LoanFilter) Offset() int {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PerPage <= 0 {
		f.PerPage = 20
	}
	return (f.Page - 1) * f.PerPage
}
