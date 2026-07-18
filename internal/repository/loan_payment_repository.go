package repository

import (
	"context"
	"expense-tracker/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoanPaymentRepository interface {
	Create(ctx context.Context, p *models.LoanPayment) error
	ListByLoanID(ctx context.Context, loanID uuid.UUID) ([]models.LoanPayment, error)
	FindByID(ctx context.Context, id, loanID uuid.UUID) (*models.LoanPayment, error)
	Delete(ctx context.Context, id, loanID uuid.UUID) error
	SumByLoanID(ctx context.Context, loanID uuid.UUID) (float64, error)
}

type loanPaymentRepository struct {
	db *gorm.DB
}

func NewLoanPaymentRepository(db *gorm.DB) LoanPaymentRepository {
	return &loanPaymentRepository{db: db}
}

func (r *loanPaymentRepository) Create(ctx context.Context, p *models.LoanPayment) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *loanPaymentRepository) ListByLoanID(ctx context.Context, loanID uuid.UUID) ([]models.LoanPayment, error) {
	var payments []models.LoanPayment
	err := r.db.WithContext(ctx).
		Where("loan_id = ?", loanID).
		Order("date DESC, created_at DESC").
		Find(&payments).Error
	return payments, err
}

func (r *loanPaymentRepository) FindByID(ctx context.Context, id, loanID uuid.UUID) (*models.LoanPayment, error) {
	var p models.LoanPayment
	err := r.db.WithContext(ctx).
		Where("id = ? AND loan_id = ?", id, loanID).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *loanPaymentRepository) Delete(ctx context.Context, id, loanID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND loan_id = ?", id, loanID).
		Delete(&models.LoanPayment{}).Error
}

func (r *loanPaymentRepository) SumByLoanID(ctx context.Context, loanID uuid.UUID) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Model(&models.LoanPayment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("loan_id = ?", loanID).
		Scan(&total).Error
	return total, err
}
