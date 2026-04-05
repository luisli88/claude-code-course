package repository

import (
	"fmt"
	"myapp/internal/domain/entity"
	"time"
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
	return nil, ErrNotFound
}

func (m *MockUserRepository) Create(user entity.User) (*entity.User, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	user.ID = fmt.Sprintf("mock-id-%d", len(m.Users)+1)
	user.CreatedAt = time.Now()
	m.Users = append(m.Users, user)
	return &user, nil
}
