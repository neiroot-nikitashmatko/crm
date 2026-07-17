package service

import (
	"context"
	"errors"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

var allowedSalaryServices = map[string]struct{}{
	"covers_without_hem":     {},
	"covers_with_hem":        {},
	"seat_covers":            {},
	"steering_wheel":         {},
	"glass_repair":           {},
	"glass_headlight_polish": {},
}

type SalaryEntryService struct {
	repo     *repository.SalaryEntryRepository
	dealRepo *repository.DealRepository
	userRepo *repository.UserRepository
}

func NewSalaryEntryService(
	repo *repository.SalaryEntryRepository,
	dealRepo *repository.DealRepository,
	userRepo *repository.UserRepository,
) *SalaryEntryService {
	return &SalaryEntryService{repo: repo, dealRepo: dealRepo, userRepo: userRepo}
}

func (s *SalaryEntryService) List(ctx context.Context, actorID string, role string) ([]model.SalaryEntry, error) {
	if strings.TrimSpace(actorID) == "" {
		return nil, errors.New("authorization required")
	}
	if role == "admin" {
		return s.repo.ListAll(ctx)
	}
	return s.repo.ListByCreatedBy(ctx, actorID)
}

func (s *SalaryEntryService) Create(ctx context.Context, actorID string, role string, input model.UpsertSalaryEntryInput) (model.SalaryEntry, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.SalaryEntry{}, errors.New("authorization required")
	}

	ownerID := actorID
	if role == "admin" {
		employeeID := strings.TrimSpace(input.EmployeeID)
		if employeeID == "" {
			return model.SalaryEntry{}, errors.New("сотрудник обязателен")
		}
		user, err := s.userRepo.FindByID(ctx, employeeID)
		if err != nil {
			return model.SalaryEntry{}, errors.New("сотрудник не найден")
		}
		if !user.IsActive {
			return model.SalaryEntry{}, errors.New("сотрудник неактивен")
		}
		ownerID = user.ID
	} else if strings.TrimSpace(input.EmployeeID) != "" && strings.TrimSpace(input.EmployeeID) != actorID {
		return model.SalaryEntry{}, errors.New("недостаточно прав")
	}

	normalized, err := s.normalizeInput(ctx, input)
	if err != nil {
		return model.SalaryEntry{}, err
	}
	return s.repo.Create(ctx, ownerID, normalized)
}

func (s *SalaryEntryService) Update(ctx context.Context, id string, actorID string, role string, input model.UpsertSalaryEntryInput) (model.SalaryEntry, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.SalaryEntry{}, errors.New("authorization required")
	}
	if strings.TrimSpace(id) == "" {
		return model.SalaryEntry{}, errors.New("invalid entry id")
	}
	normalized, err := s.normalizeInput(ctx, input)
	if err != nil {
		return model.SalaryEntry{}, err
	}
	if role == "admin" {
		return s.repo.UpdateByID(ctx, id, normalized)
	}
	return s.repo.Update(ctx, id, actorID, normalized)
}

func (s *SalaryEntryService) Delete(ctx context.Context, id string, actorID string, role string) error {
	if strings.TrimSpace(actorID) == "" {
		return errors.New("authorization required")
	}
	if strings.TrimSpace(id) == "" {
		return errors.New("invalid entry id")
	}
	if role == "admin" {
		return s.repo.DeleteByID(ctx, id)
	}
	return s.repo.Delete(ctx, id, actorID)
}

func (s *SalaryEntryService) normalizeInput(ctx context.Context, input model.UpsertSalaryEntryInput) (model.UpsertSalaryEntryInput, error) {
	dealID := strings.TrimSpace(input.DealID)
	service := strings.TrimSpace(input.Service)
	comment := strings.TrimSpace(input.Comment)

	if input.Date <= 0 {
		return model.UpsertSalaryEntryInput{}, errors.New("дата обязательна")
	}
	if dealID == "" {
		return model.UpsertSalaryEntryInput{}, errors.New("сделка обязательна")
	}
	if _, ok := allowedSalaryServices[service]; !ok {
		return model.UpsertSalaryEntryInput{}, errors.New("некорректная услуга")
	}
	if input.Salary < 0 {
		return model.UpsertSalaryEntryInput{}, errors.New("зарплата не может быть отрицательной")
	}

	deal, err := s.dealRepo.GetByID(ctx, dealID)
	if err != nil {
		return model.UpsertSalaryEntryInput{}, errors.New("сделка не найдена")
	}
	if strings.ToLower(deal.Status) != "closed" {
		return model.UpsertSalaryEntryInput{}, errors.New("можно выбрать только закрытую сделку")
	}

	return model.UpsertSalaryEntryInput{
		Date:    input.Date,
		DealID:  dealID,
		Service: service,
		Salary:  input.Salary,
		Comment: comment,
	}, nil
}
