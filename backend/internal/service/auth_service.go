package service

import (
	"context"
	"errors"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

type AuthService struct {
	repo *repository.UserRepository
}

func NewAuthService(repo *repository.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Login(ctx context.Context, phone string, password string) (*model.AuthUser, error) {
	phone = strings.TrimSpace(phone)
	password = strings.TrimSpace(password)
	if phone == "" || password == "" {
		return nil, errors.New("phone and password are required")
	}
	return s.repo.FindByCredentials(ctx, phone, password)
}
