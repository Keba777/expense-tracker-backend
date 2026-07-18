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

const balanceEpsilon = 0.005

type CreateLoanInput struct {
	PersonID    string  `json:"personId"    validate:"required,uuid"`
	Direction   string  `json:"direction"   validate:"required,oneof=lent borrowed"`
	Amount      float64 `json:"amount"      validate:"required,gt=0"`
	Description string  `json:"description" validate:"required,min=1,max=255"`
	Notes       *string `json:"notes"       validate:"omitempty,max=1000"`
	Date        string  `json:"date"        validate:"required"`
	DueDate     *string `json:"dueDate"     validate:"omitempty"`
}

type UpdateLoanInput struct {
	Description *string `json:"description" validate:"omitempty,min=1,max=255"`
	Notes       *string `json:"notes"       validate:"omitempty,max=1000"`
	Date        *string `json:"date"        validate:"omitempty"`
	DueDate     *string `json:"dueDate"     validate:"omitempty"`
}

type CreateLoanPaymentInput struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
	Date   string  `json:"date"   validate:"required"`
	Notes  *string `json:"notes"  validate:"omitempty,max=500"`
}

type LoanService interface {
	CreateLoan(ctx context.Context, userID uuid.UUID, input *CreateLoanInput) (*models.LoanWithBalance, error)
	ListLoans(ctx context.Context, userID uuid.UUID, filter *models.LoanFilter) ([]models.LoanWithBalance, int64, error)
	GetLoan(ctx context.Context, id, userID uuid.UUID) (*models.LoanWithBalance, error)
	UpdateLoan(ctx context.Context, id, userID uuid.UUID, input *UpdateLoanInput) (*models.LoanWithBalance, error)
	DeleteLoan(ctx context.Context, id, userID uuid.UUID) error

	AddPayment(ctx context.Context, loanID, userID uuid.UUID, input *CreateLoanPaymentInput) (*models.LoanWithBalance, error)
	ListPayments(ctx context.Context, loanID, userID uuid.UUID) ([]models.LoanPayment, error)
	DeletePayment(ctx context.Context, loanID, paymentID, userID uuid.UUID) (*models.LoanWithBalance, error)

	ListPeopleWithBalances(ctx context.Context, userID uuid.UUID) ([]models.PersonWithBalance, error)
	GetPersonWithBalance(ctx context.Context, personID, userID uuid.UUID) (*models.PersonWithBalance, error)

	HasOutstandingLoans(ctx context.Context, personID, userID uuid.UUID) (bool, error)
}

type loanService struct {
	loanRepo        repository.LoanRepository
	loanPaymentRepo repository.LoanPaymentRepository
	personRepo      repository.PersonRepository
}

func NewLoanService(loanRepo repository.LoanRepository, loanPaymentRepo repository.LoanPaymentRepository, personRepo repository.PersonRepository) LoanService {
	return &loanService{loanRepo: loanRepo, loanPaymentRepo: loanPaymentRepo, personRepo: personRepo}
}

func recomputeLoanStatus(amount, paid float64) models.LoanStatus {
	switch {
	case paid <= balanceEpsilon:
		return models.LoanOutstanding
	case paid >= amount-balanceEpsilon:
		return models.LoanSettled
	default:
		return models.LoanPartiallyPaid
	}
}

func (s *loanService) withBalance(ctx context.Context, l *models.Loan) (*models.LoanWithBalance, error) {
	paid, err := s.loanPaymentRepo.SumByLoanID(ctx, l.ID)
	if err != nil {
		return nil, pkgerrors.ErrInternalServer
	}
	return &models.LoanWithBalance{
		Loan:            *l,
		PaidAmount:      paid,
		RemainingAmount: l.Amount - paid,
	}, nil
}

func (s *loanService) CreateLoan(ctx context.Context, userID uuid.UUID, input *CreateLoanInput) (*models.LoanWithBalance, error) {
	personID, err := uuid.Parse(input.PersonID)
	if err != nil {
		return nil, pkgerrors.ErrInvalidInput
	}
	if _, err := s.personRepo.FindByID(ctx, personID, userID); err != nil {
		return nil, pkgerrors.ErrInvalidInput
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, pkgerrors.ErrInvalidInput
	}

	l := &models.Loan{
		UserID:      userID,
		PersonID:    personID,
		Direction:   models.LoanDirection(input.Direction),
		Amount:      input.Amount,
		Description: input.Description,
		Notes:       input.Notes,
		Date:        date,
		Status:      models.LoanOutstanding,
	}

	if input.DueDate != nil {
		dueDate, err := time.Parse("2006-01-02", *input.DueDate)
		if err != nil {
			return nil, pkgerrors.ErrInvalidInput
		}
		l.DueDate = &dueDate
	}

	if err := s.loanRepo.Create(ctx, l); err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	created, err := s.loanRepo.FindByID(ctx, l.ID, userID)
	if err != nil {
		return nil, pkgerrors.ErrInternalServer
	}
	return s.withBalance(ctx, created)
}

func (s *loanService) ListLoans(ctx context.Context, userID uuid.UUID, filter *models.LoanFilter) ([]models.LoanWithBalance, int64, error) {
	return s.loanRepo.List(ctx, userID, filter)
}

func (s *loanService) GetLoan(ctx context.Context, id, userID uuid.UUID) (*models.LoanWithBalance, error) {
	l, err := s.loanRepo.FindByID(ctx, id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, pkgerrors.ErrInternalServer
	}
	return s.withBalance(ctx, l)
}

func (s *loanService) UpdateLoan(ctx context.Context, id, userID uuid.UUID, input *UpdateLoanInput) (*models.LoanWithBalance, error) {
	l, err := s.loanRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, pkgerrors.ErrNotFound
	}

	if input.Description != nil {
		l.Description = *input.Description
	}
	if input.Notes != nil {
		l.Notes = input.Notes
	}
	if input.Date != nil {
		date, err := time.Parse("2006-01-02", *input.Date)
		if err != nil {
			return nil, pkgerrors.ErrInvalidInput
		}
		l.Date = date
	}
	if input.DueDate != nil {
		dueDate, err := time.Parse("2006-01-02", *input.DueDate)
		if err != nil {
			return nil, pkgerrors.ErrInvalidInput
		}
		l.DueDate = &dueDate
	}

	if err := s.loanRepo.Update(ctx, l); err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	updated, err := s.loanRepo.FindByID(ctx, l.ID, userID)
	if err != nil {
		return nil, pkgerrors.ErrInternalServer
	}
	return s.withBalance(ctx, updated)
}

func (s *loanService) DeleteLoan(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.loanRepo.FindByID(ctx, id, userID); err != nil {
		return pkgerrors.ErrNotFound
	}
	return s.loanRepo.Delete(ctx, id, userID)
}

func (s *loanService) AddPayment(ctx context.Context, loanID, userID uuid.UUID, input *CreateLoanPaymentInput) (*models.LoanWithBalance, error) {
	l, err := s.loanRepo.FindByID(ctx, loanID, userID)
	if err != nil {
		return nil, pkgerrors.ErrNotFound
	}

	paid, err := s.loanPaymentRepo.SumByLoanID(ctx, loanID)
	if err != nil {
		return nil, pkgerrors.ErrInternalServer
	}
	remaining := l.Amount - paid
	if input.Amount > remaining+balanceEpsilon {
		return nil, pkgerrors.ErrInvalidInput
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, pkgerrors.ErrInvalidInput
	}

	payment := &models.LoanPayment{
		LoanID: loanID,
		UserID: userID,
		Amount: input.Amount,
		Date:   date,
		Notes:  input.Notes,
	}
	if err := s.loanPaymentRepo.Create(ctx, payment); err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	newPaid := paid + input.Amount
	l.Status = recomputeLoanStatus(l.Amount, newPaid)
	if err := s.loanRepo.Update(ctx, l); err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	return &models.LoanWithBalance{
		Loan:            *l,
		PaidAmount:      newPaid,
		RemainingAmount: l.Amount - newPaid,
	}, nil
}

func (s *loanService) ListPayments(ctx context.Context, loanID, userID uuid.UUID) ([]models.LoanPayment, error) {
	if _, err := s.loanRepo.FindByID(ctx, loanID, userID); err != nil {
		return nil, pkgerrors.ErrNotFound
	}
	return s.loanPaymentRepo.ListByLoanID(ctx, loanID)
}

func (s *loanService) DeletePayment(ctx context.Context, loanID, paymentID, userID uuid.UUID) (*models.LoanWithBalance, error) {
	l, err := s.loanRepo.FindByID(ctx, loanID, userID)
	if err != nil {
		return nil, pkgerrors.ErrNotFound
	}

	if _, err := s.loanPaymentRepo.FindByID(ctx, paymentID, loanID); err != nil {
		return nil, pkgerrors.ErrNotFound
	}
	if err := s.loanPaymentRepo.Delete(ctx, paymentID, loanID); err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	newPaid, err := s.loanPaymentRepo.SumByLoanID(ctx, loanID)
	if err != nil {
		return nil, pkgerrors.ErrInternalServer
	}
	l.Status = recomputeLoanStatus(l.Amount, newPaid)
	if err := s.loanRepo.Update(ctx, l); err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	return &models.LoanWithBalance{
		Loan:            *l,
		PaidAmount:      newPaid,
		RemainingAmount: l.Amount - newPaid,
	}, nil
}

func (s *loanService) ListPeopleWithBalances(ctx context.Context, userID uuid.UUID) ([]models.PersonWithBalance, error) {
	people, err := s.personRepo.List(ctx, userID)
	if err != nil {
		return nil, pkgerrors.ErrInternalServer
	}
	balances, err := s.loanRepo.PersonBalances(ctx, userID)
	if err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	balanceMap := make(map[uuid.UUID]repository.PersonBalanceRow, len(balances))
	for _, b := range balances {
		balanceMap[b.PersonID] = b
	}

	result := make([]models.PersonWithBalance, len(people))
	for i, p := range people {
		pwb := models.PersonWithBalance{Person: p}
		if b, ok := balanceMap[p.ID]; ok {
			pwb.TotalLent = b.TotalLent
			pwb.TotalBorrowed = b.TotalBorrowed
			pwb.NetBalance = b.TotalLent - b.TotalBorrowed
			pwb.LoanCount = b.LoanCount
		}
		result[i] = pwb
	}
	return result, nil
}

func (s *loanService) GetPersonWithBalance(ctx context.Context, personID, userID uuid.UUID) (*models.PersonWithBalance, error) {
	p, err := s.personRepo.FindByID(ctx, personID, userID)
	if err != nil {
		return nil, pkgerrors.ErrNotFound
	}
	balances, err := s.loanRepo.PersonBalances(ctx, userID)
	if err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	pwb := &models.PersonWithBalance{Person: *p}
	for _, b := range balances {
		if b.PersonID == personID {
			pwb.TotalLent = b.TotalLent
			pwb.TotalBorrowed = b.TotalBorrowed
			pwb.NetBalance = b.TotalLent - b.TotalBorrowed
			pwb.LoanCount = b.LoanCount
			break
		}
	}
	return pwb, nil
}

func (s *loanService) HasOutstandingLoans(ctx context.Context, personID, userID uuid.UUID) (bool, error) {
	loans, _, err := s.loanRepo.List(ctx, userID, &models.LoanFilter{PersonID: &personID, PerPage: 1000})
	if err != nil {
		return false, pkgerrors.ErrInternalServer
	}
	for _, l := range loans {
		if l.Status != models.LoanSettled {
			return true, nil
		}
	}
	return false, nil
}
