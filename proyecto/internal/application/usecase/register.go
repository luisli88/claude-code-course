package usecase

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"myapp/internal/domain/entity"
	"myapp/internal/domain/repository"
)

var (
	ErrEmailAlreadyTaken = errors.New("email already taken")
	ErrInvalidInput      = errors.New("invalid input")
)

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type RegisterOutput struct {
	User entity.User
}

type Register struct {
	repo repository.UserRepository
}

func NewRegister(repo repository.UserRepository) *Register {
	return &Register{repo: repo}
}

func (uc *Register) Execute(input RegisterInput) (*RegisterOutput, error) {
	if err := validateRegisterInput(input); err != nil {
		return nil, err
	}

	_, err := uc.repo.FindByEmail(input.Email)
	if err == nil {
		return nil, ErrEmailAlreadyTaken
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	created, err := uc.repo.Create(entity.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return nil, err
	}

	return &RegisterOutput{User: *created}, nil
}

func validateRegisterInput(input RegisterInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if !strings.Contains(input.Email, "@") || strings.TrimSpace(input.Email) == "" {
		return fmt.Errorf("%w: email is invalid", ErrInvalidInput)
	}
	if len(input.Password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", ErrInvalidInput)
	}
	return nil
}
