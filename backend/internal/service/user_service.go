package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

var (
	phonePattern = regexp.MustCompile(`^\+[1-9][0-9]{10,14}$`)
	validRoles   = map[string]struct{}{
		"admin":   {},
		"manager": {},
	}
	validPositions = map[string]struct{}{
		"Менеджер": {},
		"Мастер":   {},
	}
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) List(ctx context.Context) ([]model.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) GetByID(ctx context.Context, id string) (model.User, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.User{}, errors.New("invalid user id")
	}
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) Update(ctx context.Context, id string, input model.UpdateUserInput) (model.User, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.User{}, errors.New("invalid user id")
	}

	normalized, updatePassword, err := normalizeUpdateUserInput(input)
	if err != nil {
		return model.User{}, err
	}

	return s.repo.Update(ctx, id, normalized, updatePassword)
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("invalid user id")
	}
	return s.repo.Deactivate(ctx, id)
}

func (s *UserService) Create(ctx context.Context, input model.CreateUserInput) (model.User, error) {
	normalized, err := normalizeCreateUserInput(input)
	if err != nil {
		return model.User{}, err
	}
	return s.repo.Create(ctx, normalized)
}

func normalizeCreateUserInput(input model.CreateUserInput) (model.CreateUserInput, error) {
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Patronymic = strings.TrimSpace(input.Patronymic)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Password = strings.TrimSpace(input.Password)
	input.Role = strings.TrimSpace(input.Role)
	input.Position = strings.TrimSpace(input.Position)

	if input.FirstName == "" {
		return model.CreateUserInput{}, errors.New("имя обязательно")
	}
	if input.LastName == "" {
		return model.CreateUserInput{}, errors.New("фамилия обязательна")
	}
	if input.Patronymic == "" {
		return model.CreateUserInput{}, errors.New("отчество обязательно")
	}
	if input.Position == "" {
		return model.CreateUserInput{}, errors.New("должность обязательна")
	}
	if _, ok := validPositions[input.Position]; !ok {
		return model.CreateUserInput{}, errors.New("некорректная должность")
	}
	if input.Phone == "" || !phonePattern.MatchString(input.Phone) {
		return model.CreateUserInput{}, errors.New("некорректный телефон")
	}
	if input.Password == "" {
		return model.CreateUserInput{}, errors.New("пароль обязателен")
	}
	if _, ok := validRoles[input.Role]; !ok {
		return model.CreateUserInput{}, errors.New("некорректная роль")
	}

	if input.BirthDate != nil {
		trimmed := strings.TrimSpace(*input.BirthDate)
		if trimmed == "" {
			input.BirthDate = nil
		} else {
			if _, err := time.Parse("2006-01-02", trimmed); err != nil {
				return model.CreateUserInput{}, errors.New("некорректная дата рождения")
			}
			input.BirthDate = &trimmed
		}
	}

	if input.BirthDate == nil {
		return model.CreateUserInput{}, errors.New("дата рождения обязательна")
	}

	return input, nil
}

func normalizeUpdateUserInput(input model.UpdateUserInput) (model.UpdateUserInput, bool, error) {
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Patronymic = strings.TrimSpace(input.Patronymic)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Password = strings.TrimSpace(input.Password)
	input.Role = strings.TrimSpace(input.Role)
	input.Position = strings.TrimSpace(input.Position)

	updatePassword := input.Password != ""

	if input.FirstName == "" {
		return model.UpdateUserInput{}, false, errors.New("имя обязательно")
	}
	if input.LastName == "" {
		return model.UpdateUserInput{}, false, errors.New("фамилия обязательна")
	}
	if input.Patronymic == "" {
		return model.UpdateUserInput{}, false, errors.New("отчество обязательно")
	}
	if input.Position == "" {
		return model.UpdateUserInput{}, false, errors.New("должность обязательна")
	}
	if _, ok := validPositions[input.Position]; !ok {
		return model.UpdateUserInput{}, false, errors.New("некорректная должность")
	}
	if input.Phone == "" || !phonePattern.MatchString(input.Phone) {
		return model.UpdateUserInput{}, false, errors.New("некорректный телефон")
	}
	if _, ok := validRoles[input.Role]; !ok {
		return model.UpdateUserInput{}, false, errors.New("некорректная роль")
	}

	if input.BirthDate != nil {
		trimmed := strings.TrimSpace(*input.BirthDate)
		if trimmed == "" {
			input.BirthDate = nil
		} else {
			if _, err := time.Parse("2006-01-02", trimmed); err != nil {
				return model.UpdateUserInput{}, false, errors.New("некорректная дата рождения")
			}
			input.BirthDate = &trimmed
		}
	}

	if input.BirthDate == nil {
		return model.UpdateUserInput{}, false, errors.New("дата рождения обязательна")
	}

	return input, updatePassword, nil
}
