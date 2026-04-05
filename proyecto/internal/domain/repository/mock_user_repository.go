package repository

import (
	"errors"
	"myapp/internal/domain/entity"
)

type MockUserRepository struct {
	Users []entity.User
	Err   error
}

func (m *MockUserRepository) FindAll(params PaginationParams) ([]entity.User, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	start := params.Offset
	if start > len(m.Users) {
		return []entity.User{}, nil
	}

	end := start + params.Limit
	if end > len(m.Users) {
		end = len(m.Users)
	}

	return m.Users[start:end], nil
}

func (m *MockUserRepository) Count() (int, error) {
	if m.Err != nil {
		return 0, m.Err
	}
	return len(m.Users), nil
}

func (m *MockUserRepository) FindByEmail(email string) (*entity.User, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	for _, u := range m.Users {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, errors.New("user not found")
}
