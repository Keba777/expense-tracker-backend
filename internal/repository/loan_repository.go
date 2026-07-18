package repository

import (
	"context"
	"expense-tracker/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PersonBalanceRow struct {
	PersonID      uuid.UUID `json:"personId"`
	TotalLent     float64   `json:"totalLent"`
	TotalBorrowed float64   `json:"totalBorrowed"`
	LoanCount     int64     `json:"loanCount"`
}

type LoanRepository interface {
	Create(ctx context.Context, l *models.Loan) error
	FindByID(ctx context.Context, id, userID uuid.UUID) (*models.Loan, error)
	List(ctx context.Context, userID uuid.UUID, f *models.LoanFilter) ([]models.LoanWithBalance, int64, error)
	Update(ctx context.Context, l *models.Loan) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	PersonBalances(ctx context.Context, userID uuid.UUID) ([]PersonBalanceRow, error)
}

type loanRepository struct {
	db *gorm.DB
}

func NewLoanRepository(db *gorm.DB) LoanRepository {
	return &loanRepository{db: db}
}

func (r *loanRepository) Create(ctx context.Context, l *models.Loan) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *loanRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*models.Loan, error) {
	var l models.Loan
	err := r.db.WithContext(ctx).
		Preload("Person").
		Where("id = ? AND user_id = ?", id, userID).
		First(&l).Error
	if err != nil {
		return nil, err
	}
	return &l, nil
}

const paidSubquery = `LEFT JOIN (
	SELECT loan_id, SUM(amount) as paid FROM loan_payments WHERE deleted_at IS NULL GROUP BY loan_id
) paid ON paid.loan_id = loans.id`

func (r *loanRepository) List(ctx context.Context, userID uuid.UUID, f *models.LoanFilter) ([]models.LoanWithBalance, int64, error) {
	base := r.db.WithContext(ctx).Table("loans").Where("loans.user_id = ? AND loans.deleted_at IS NULL", userID)

	if f.PersonID != nil {
		base = base.Where("loans.person_id = ?", *f.PersonID)
	}
	if f.Direction != "" {
		base = base.Where("loans.direction = ?", f.Direction)
	}
	if f.Status != "" {
		base = base.Where("loans.status = ?", f.Status)
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var loans []models.LoanWithBalance
	err := base.Session(&gorm.Session{}).
		Select("loans.*, COALESCE(paid.paid, 0) as paid_amount, loans.amount - COALESCE(paid.paid, 0) as remaining_amount").
		Joins(paidSubquery).
		Preload("Person").
		Order("loans.date DESC, loans.created_at DESC").
		Limit(f.PerPage).
		Offset(f.Offset()).
		Find(&loans).Error

	return loans, total, err
}

func (r *loanRepository) Update(ctx context.Context, l *models.Loan) error {
	return r.db.WithContext(ctx).Save(l).Error
}

func (r *loanRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Loan{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Where("loan_id = ?", id).Delete(&models.LoanPayment{}).Error
	})
}

func (r *loanRepository) PersonBalances(ctx context.Context, userID uuid.UUID) ([]PersonBalanceRow, error) {
	var rows []PersonBalanceRow
	err := r.db.WithContext(ctx).
		Table("loans").
		Select(`
			loans.person_id,
			COALESCE(SUM(CASE WHEN loans.direction = 'lent'     THEN loans.amount - COALESCE(paid.paid, 0) ELSE 0 END), 0) as total_lent,
			COALESCE(SUM(CASE WHEN loans.direction = 'borrowed' THEN loans.amount - COALESCE(paid.paid, 0) ELSE 0 END), 0) as total_borrowed,
			COUNT(CASE WHEN loans.status != 'settled' THEN 1 END) as loan_count
		`).
		Joins(paidSubquery).
		Where("loans.user_id = ? AND loans.deleted_at IS NULL", userID).
		Group("loans.person_id").
		Scan(&rows).Error
	return rows, err
}
