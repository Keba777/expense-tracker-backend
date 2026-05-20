package services

import (
	"context"
	"expense-tracker/internal/models"
	"expense-tracker/internal/repository"
	pkgerrors "expense-tracker/pkg/errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateTransactionInput struct {
	CategoryID        string     `json:"categoryId"        validate:"required,uuid"`
	Type              string     `json:"type"              validate:"required,oneof=income expense"`
	Amount            float64    `json:"amount"            validate:"required,gt=0"`
	Description       string     `json:"description"       validate:"required,min=1,max=255"`
	Notes             *string    `json:"notes"             validate:"omitempty,max=1000"`
	Date              string     `json:"date"              validate:"required"`
	Recurrence        string     `json:"recurrence"        validate:"omitempty,oneof=once daily weekly monthly"`
	RecurrenceEndDate *string    `json:"recurrenceEndDate" validate:"omitempty"`
	Tags              []string   `json:"tags"`
}

type UpdateTransactionInput struct {
	CategoryID        *string    `json:"categoryId"        validate:"omitempty,uuid"`
	Type              *string    `json:"type"              validate:"omitempty,oneof=income expense"`
	Amount            *float64   `json:"amount"            validate:"omitempty,gt=0"`
	Description       *string    `json:"description"       validate:"omitempty,min=1,max=255"`
	Notes             *string    `json:"notes"             validate:"omitempty,max=1000"`
	Date              *string    `json:"date"              validate:"omitempty"`
	Recurrence        *string    `json:"recurrence"        validate:"omitempty,oneof=once daily weekly monthly"`
	RecurrenceEndDate *string    `json:"recurrenceEndDate" validate:"omitempty"`
	Tags              []string   `json:"tags"`
}

type TransactionService interface {
	Create(ctx context.Context, userID uuid.UUID, input *CreateTransactionInput) (*models.Transaction, error)
	List(ctx context.Context, userID uuid.UUID, filter *models.TransactionFilter) ([]models.Transaction, int64, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Transaction, error)
	Update(ctx context.Context, id, userID uuid.UUID, input *UpdateTransactionInput) (*models.Transaction, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	GetSummary(ctx context.Context, userID uuid.UUID, from, to time.Time) (*repository.SummaryResult, error)
	GetCategoryBreakdown(ctx context.Context, userID uuid.UUID, from, to time.Time, txType models.TransactionType) ([]repository.CategoryTotal, error)
	GetMonthlyTrends(ctx context.Context, userID uuid.UUID, months int) ([]repository.MonthlyTrend, error)
	GetDailyTotals(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]repository.DailyTotal, error)
}

type transactionService struct {
	txRepo repository.TransactionRepository
}

func NewTransactionService(txRepo repository.TransactionRepository) TransactionService {
	return &transactionService{txRepo: txRepo}
}

func (s *transactionService) Create(ctx context.Context, userID uuid.UUID, input *CreateTransactionInput) (*models.Transaction, error) {
	catID, err := uuid.Parse(input.CategoryID)
	if err != nil {
		return nil, pkgerrors.ErrInvalidInput
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, pkgerrors.ErrInvalidInput
	}

	recurrence := models.RecurrenceOnce
	if input.Recurrence != "" {
		recurrence = models.Recurrence(input.Recurrence)
	}

	t := &models.Transaction{
		UserID:      userID,
		CategoryID:  catID,
		Type:        models.TransactionType(input.Type),
		Amount:      input.Amount,
		Description: input.Description,
		Notes:       input.Notes,
		Date:        date,
		Recurrence:  recurrence,
		Tags:        input.Tags,
	}

	if input.RecurrenceEndDate != nil {
		endDate, err := time.Parse("2006-01-02", *input.RecurrenceEndDate)
		if err != nil {
			return nil, pkgerrors.ErrInvalidInput
		}
		t.RecurrenceEndDate = &endDate
	}

	if err := s.txRepo.Create(ctx, t); err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	return s.txRepo.FindByID(ctx, t.ID, userID)
}

func (s *transactionService) List(ctx context.Context, userID uuid.UUID, filter *models.TransactionFilter) ([]models.Transaction, int64, error) {
	return s.txRepo.List(ctx, userID, filter)
}

func (s *transactionService) GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Transaction, error) {
	t, err := s.txRepo.FindByID(ctx, id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, pkgerrors.ErrInternalServer
	}
	return t, nil
}

func (s *transactionService) Update(ctx context.Context, id, userID uuid.UUID, input *UpdateTransactionInput) (*models.Transaction, error) {
	t, err := s.txRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, pkgerrors.ErrNotFound
	}

	if input.CategoryID != nil {
		catID, err := uuid.Parse(*input.CategoryID)
		if err != nil {
			return nil, pkgerrors.ErrInvalidInput
		}
		t.CategoryID = catID
	}
	if input.Type != nil {
		t.Type = models.TransactionType(*input.Type)
	}
	if input.Amount != nil {
		t.Amount = *input.Amount
	}
	if input.Description != nil {
		t.Description = *input.Description
	}
	if input.Notes != nil {
		t.Notes = input.Notes
	}
	if input.Date != nil {
		date, err := time.Parse("2006-01-02", *input.Date)
		if err != nil {
			return nil, pkgerrors.ErrInvalidInput
		}
		t.Date = date
	}
	if input.Recurrence != nil {
		t.Recurrence = models.Recurrence(*input.Recurrence)
	}
	if input.Tags != nil {
		t.Tags = input.Tags
	}

	if err := s.txRepo.Update(ctx, t); err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	return s.txRepo.FindByID(ctx, t.ID, userID)
}

func (s *transactionService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.txRepo.FindByID(ctx, id, userID); err != nil {
		return pkgerrors.ErrNotFound
	}
	return s.txRepo.Delete(ctx, id, userID)
}

func (s *transactionService) GetSummary(ctx context.Context, userID uuid.UUID, from, to time.Time) (*repository.SummaryResult, error) {
	return s.txRepo.Summary(ctx, userID, from, to)
}

func (s *transactionService) GetCategoryBreakdown(ctx context.Context, userID uuid.UUID, from, to time.Time, txType models.TransactionType) ([]repository.CategoryTotal, error) {
	return s.txRepo.CategoryBreakdown(ctx, userID, from, to, txType)
}

func (s *transactionService) GetMonthlyTrends(ctx context.Context, userID uuid.UUID, months int) ([]repository.MonthlyTrend, error) {
	return s.txRepo.MonthlyTrends(ctx, userID, months)
}

func (s *transactionService) GetDailyTotals(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]repository.DailyTotal, error) {
	return s.txRepo.DailyTotals(ctx, userID, from, to)
}
