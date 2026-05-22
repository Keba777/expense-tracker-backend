package repository

import (
	"context"
	"expense-tracker/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasswordResetRepository interface {
	Create(ctx context.Context, token *models.PasswordResetToken) error
	FindByToken(ctx context.Context, token string) (*models.PasswordResetToken, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type passwordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

func (r *passwordResetRepository) Create(ctx context.Context, token *models.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *passwordResetRepository) FindByToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	var t models.PasswordResetToken
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *passwordResetRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.PasswordResetToken{}).Error
}

func (r *passwordResetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.PasswordResetToken{}, "id = ?", id).Error
}

func (r *passwordResetRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.PasswordResetToken{}).Error
}
