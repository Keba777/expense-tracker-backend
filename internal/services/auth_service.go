package services

import (
	"context"
	"strings"

	"expense-tracker/internal/models"
	"expense-tracker/internal/repository"
	pkgerrors "expense-tracker/pkg/errors"
	"expense-tracker/pkg/jwt"
	"expense-tracker/pkg/password"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegisterInput struct {
	Email     string `json:"email"     validate:"omitempty,email"`
	Phone     string `json:"phone"     validate:"omitempty,e164"`
	Password  string `json:"password"  validate:"required,min=8"`
	FirstName string `json:"firstName" validate:"required,min=1,max=100"`
	LastName  string `json:"lastName"  validate:"required,min=1,max=100"`
	Currency  string `json:"currency"  validate:"omitempty,len=3"`
	Timezone  string `json:"timezone"  validate:"omitempty"`
}

type LoginInput struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password"   validate:"required"`
}

type AuthResponse struct {
	User   *models.UserResponse `json:"user"`
	Tokens *jwt.TokenPair       `json:"tokens"`
}

type UpdateProfileInput struct {
	FirstName string `json:"firstName" validate:"required,min=1,max=100"`
	LastName  string `json:"lastName"  validate:"required,min=1,max=100"`
	Currency  string `json:"currency"  validate:"required,len=3"`
	Timezone  string `json:"timezone"  validate:"required"`
}

type AuthService interface {
	Register(ctx context.Context, input *RegisterInput) (*AuthResponse, error)
	Login(ctx context.Context, input *LoginInput) (*AuthResponse, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*jwt.TokenPair, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, input *UpdateProfileInput) (*models.UserResponse, error)
}

type authService struct {
	userRepo     repository.UserRepository
	categoryRepo repository.CategoryRepository
	jwtManager   *jwt.Manager
}

func NewAuthService(
	userRepo repository.UserRepository,
	categoryRepo repository.CategoryRepository,
	jwtManager *jwt.Manager,
) AuthService {
	return &authService{
		userRepo:     userRepo,
		categoryRepo: categoryRepo,
		jwtManager:   jwtManager,
	}
}

func (s *authService) Register(ctx context.Context, input *RegisterInput) (*AuthResponse, error) {
	if input.Email == "" && input.Phone == "" {
		return nil, pkgerrors.ErrValidation
	}

	if input.Email != "" {
		existing, err := s.userRepo.FindByEmail(ctx, input.Email)
		if err != nil && !stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrInternalServer
		}
		if existing != nil {
			return nil, pkgerrors.ErrConflict
		}
	}

	if input.Phone != "" {
		existing, err := s.userRepo.FindByPhone(ctx, input.Phone)
		if err != nil && !stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrInternalServer
		}
		if existing != nil {
			return nil, pkgerrors.ErrConflict
		}
	}

	hash, err := password.Hash(input.Password)
	if err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}
	timezone := input.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	user := &models.User{
		PasswordHash: hash,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Currency:     currency,
		Timezone:     timezone,
		Plan:         models.PlanFree,
		IsActive:     true,
	}
	if input.Email != "" {
		user.Email = &input.Email
	}
	if input.Phone != "" {
		user.Phone = &input.Phone
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	defaultCats := models.SeedDefaultCategories(user.ID)
	if err := s.categoryRepo.BulkCreate(ctx, defaultCats); err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	tokens, err := s.jwtManager.GeneratePair(user.ID, user.Identifier(), string(user.Plan))
	if err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	return &AuthResponse{User: user.ToResponse(), Tokens: tokens}, nil
}

func (s *authService) Login(ctx context.Context, input *LoginInput) (*AuthResponse, error) {
	var user *models.User
	var err error

	if strings.Contains(input.Identifier, "@") {
		user, err = s.userRepo.FindByEmail(ctx, input.Identifier)
	} else {
		user, err = s.userRepo.FindByPhone(ctx, input.Identifier)
	}
	if err != nil {
		return nil, pkgerrors.ErrInvalidCredentials
	}

	if !password.Verify(input.Password, user.PasswordHash) {
		return nil, pkgerrors.ErrInvalidCredentials
	}

	tokens, err := s.jwtManager.GeneratePair(user.ID, user.Identifier(), string(user.Plan))
	if err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	return &AuthResponse{User: user.ToResponse(), Tokens: tokens}, nil
}

func (s *authService) RefreshTokens(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	claims, err := s.jwtManager.ValidateRefresh(refreshToken)
	if err != nil {
		return nil, pkgerrors.ErrTokenInvalid
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, pkgerrors.ErrUnauthorized
	}

	tokens, err := s.jwtManager.GeneratePair(user.ID, user.Identifier(), string(user.Plan))
	if err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	return tokens, nil
}

func (s *authService) GetUser(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, pkgerrors.ErrNotFound
	}
	return user.ToResponse(), nil
}

func (s *authService) UpdateProfile(ctx context.Context, userID uuid.UUID, input *UpdateProfileInput) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, pkgerrors.ErrNotFound
	}

	user.FirstName = input.FirstName
	user.LastName = input.LastName
	user.Currency = input.Currency
	user.Timezone = input.Timezone

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, pkgerrors.ErrInternalServer
	}
	return user.ToResponse(), nil
}

// stdErrors wraps stdlib errors.Is to avoid naming conflict with the pkgerrors alias.
var stdErrors = struct {
	Is func(error, error) bool
}{
	Is: pkgerrors.Is,
}
