package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Person struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index"                       json:"userId"`
	Name      string         `gorm:"not null;size:100"                              json:"name"`
	Phone     *string        `gorm:"size:30"                                        json:"phone,omitempty"`
	Notes     *string        `gorm:"size:1000"                                      json:"notes,omitempty"`
	CreatedAt time.Time      `                                                       json:"createdAt"`
	UpdatedAt time.Time      `                                                       json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                          json:"-"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (p *Person) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// PersonWithBalance is the read model for person list/detail endpoints.
// Balances are always computed on read from Loan+LoanPayment rows, never
// stored on Person, so they can't drift out of sync with payment activity.
type PersonWithBalance struct {
	Person
	TotalLent     float64 `json:"totalLent"`
	TotalBorrowed float64 `json:"totalBorrowed"`
	NetBalance    float64 `json:"netBalance"`
	LoanCount     int64   `json:"loanCount"`
}
