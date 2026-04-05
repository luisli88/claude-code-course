package usecase

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"myapp/internal/domain/repository"
	"myapp/internal/domain/service"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	Token string
}

type Login struct {
	repo         repository.UserRepository
	tokenService service.TokenService
}

func NewLogin(repo repository.UserRepository, tokenService service.TokenService) *Login {
	return &Login{repo: repo, tokenService: tokenService}
}

func (uc *Login) Execute(input LoginInput) (*LoginOutput, error) {
	user, err := uc.repo.FindByEmail(input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := uc.tokenService.Generate(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{Token: token}, nil
}
