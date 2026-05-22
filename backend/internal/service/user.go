package service

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/repository"
)

type UserService struct {
	users *repository.UserRepository
}

type UserInput struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Status      string `json:"status"`
}

func NewUserService(users *repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) List(ctx context.Context) ([]entity.User, error) {
	return s.users.List(ctx)
}

func (s *UserService) Get(ctx context.Context, id int64) (entity.User, error) {
	return s.users.Get(ctx, id)
}

func (s *UserService) Create(ctx context.Context, input UserInput) (entity.User, error) {
	item, err := buildUser(input, true, "")
	if err != nil {
		return entity.User{}, err
	}
	err = s.users.Create(ctx, &item)
	return item, err
}

func (s *UserService) Update(ctx context.Context, id int64, input UserInput) (entity.User, error) {
	item, err := s.users.Get(ctx, id)
	if err != nil {
		return entity.User{}, err
	}

	next, err := buildUser(input, false, item.PasswordHash)
	if err != nil {
		return entity.User{}, err
	}

	item.Email = next.Email
	item.PasswordHash = next.PasswordHash
	item.DisplayName = next.DisplayName
	item.Role = next.Role
	item.Status = next.Status

	err = s.users.Update(ctx, &item)
	return item, err
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	return s.users.Delete(ctx, id)
}

func buildUser(input UserInput, requirePassword bool, existingPasswordHash string) (entity.User, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Role = strings.TrimSpace(input.Role)
	input.Status = strings.TrimSpace(input.Status)

	if input.Email == "" {
		return entity.User{}, validationError("user email is required")
	}
	if input.Role == "" {
		input.Role = "user"
	}
	if !allowed(input.Role, "admin", "user") {
		return entity.User{}, validationError("invalid user role")
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if !allowed(input.Status, "active", "disabled") {
		return entity.User{}, validationError("invalid user status")
	}

	passwordHash := existingPasswordHash
	if strings.TrimSpace(input.Password) != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return entity.User{}, err
		}
		passwordHash = string(hashed)
	}
	if requirePassword && passwordHash == "" {
		return entity.User{}, validationError("user password is required")
	}

	return entity.User{
		Email:        input.Email,
		PasswordHash: passwordHash,
		DisplayName:  input.DisplayName,
		Role:         input.Role,
		Status:       input.Status,
	}, nil
}
