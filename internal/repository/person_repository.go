package repository

import (
	"context"
	"expense-tracker/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PersonRepository interface {
	Create(ctx context.Context, p *models.Person) error
	FindByID(ctx context.Context, id, userID uuid.UUID) (*models.Person, error)
	List(ctx context.Context, userID uuid.UUID) ([]models.Person, error)
	Update(ctx context.Context, p *models.Person) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type personRepository struct {
	db *gorm.DB
}

func NewPersonRepository(db *gorm.DB) PersonRepository {
	return &personRepository{db: db}
}

func (r *personRepository) Create(ctx context.Context, p *models.Person) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *personRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*models.Person, error) {
	var p models.Person
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *personRepository) List(ctx context.Context, userID uuid.UUID) ([]models.Person, error) {
	var people []models.Person
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("name").
		Find(&people).Error
	return people, err
}

func (r *personRepository) Update(ctx context.Context, p *models.Person) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *personRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.Person{}).Error
}
