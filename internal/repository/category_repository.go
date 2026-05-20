package repository

import (
	"context"
	"expense-tracker/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	BulkCreate(ctx context.Context, categories []models.Category) error
	Create(ctx context.Context, category *models.Category) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.Category, error)
	FindByID(ctx context.Context, id, userID uuid.UUID) (*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) BulkCreate(ctx context.Context, categories []models.Category) error {
	return r.db.WithContext(ctx).CreateInBatches(categories, 50).Error
}

func (r *categoryRepository) Create(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("type, is_default DESC, name").
		Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*models.Category, error) {
	var category models.Category
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND is_default = false", id, userID).
		Delete(&models.Category{}).Error
}
