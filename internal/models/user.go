package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Plan string

const (
	PlanFree       Plan = "free"
	PlanPro        Plan = "pro"
	PlanEnterprise Plan = "enterprise"
)

type User struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email           string         `gorm:"uniqueIndex;not null;size:255"                  json:"email"`
	PasswordHash    string         `gorm:"not null"                                       json:"-"`
	FirstName       string         `gorm:"not null;size:100"                              json:"firstName"`
	LastName        string         `gorm:"not null;size:100"                              json:"lastName"`
	Currency        string         `gorm:"not null;default:'USD';size:3"                  json:"currency"`
	Timezone        string         `gorm:"not null;default:'UTC';size:50"                 json:"timezone"`
	AvatarURL       *string        `gorm:"size:500"                                       json:"avatarUrl,omitempty"`
	Plan            Plan           `gorm:"not null;default:'free';size:20"                json:"plan"`
	IsActive        bool           `gorm:"not null;default:true"                          json:"isActive"`
	EmailVerifiedAt *time.Time     `                                                      json:"emailVerifiedAt,omitempty"`
	CreatedAt       time.Time      `                                                      json:"createdAt"`
	UpdatedAt       time.Time      `                                                      json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index"                                          json:"-"`

	Categories   []Category   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Transactions []Transaction `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Budgets      []Budget     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type UserResponse struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	FirstName       string     `json:"firstName"`
	LastName        string     `json:"lastName"`
	Currency        string     `json:"currency"`
	Timezone        string     `json:"timezone"`
	AvatarURL       *string    `json:"avatarUrl,omitempty"`
	Plan            Plan       `json:"plan"`
	IsActive        bool       `json:"isActive"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:              u.ID,
		Email:           u.Email,
		FirstName:       u.FirstName,
		LastName:        u.LastName,
		Currency:        u.Currency,
		Timezone:        u.Timezone,
		AvatarURL:       u.AvatarURL,
		Plan:            u.Plan,
		IsActive:        u.IsActive,
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
	}
}
