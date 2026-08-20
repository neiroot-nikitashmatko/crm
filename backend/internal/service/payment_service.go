package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

const paymentShortTitleMaxLength = 22

var (
	ErrPaymentForbidden = errors.New("недостаточно прав")
	ErrPaymentNotFound  = errors.New("оплата не найдена")
)

var allowedPaymentPayers = map[string]struct{}{
	"ip-panov-nikolay":  {},
	"ip-shmatko-nikita": {},
	"ip-panov-dmitry":   {},
}

var moscowLocation = mustLoadMoscowLocation()

type PaymentService struct {
	repo     *repository.PaymentRepository
	userRepo *repository.UserRepository
}

func NewPaymentService(repo *repository.PaymentRepository, userRepo *repository.UserRepository) *PaymentService {
	return &PaymentService{repo: repo, userRepo: userRepo}
}

func (s *PaymentService) List(ctx context.Context, actorID string, role string) ([]model.Payment, error) {
	if err := s.requireAccess(ctx, actorID, role); err != nil {
		return nil, err
	}
	return s.repo.List(ctx)
}

func (s *PaymentService) Create(ctx context.Context, actorID string, role string, input model.CreatePaymentInput) (model.Payment, error) {
	if err := s.requireAccess(ctx, actorID, role); err != nil {
		return model.Payment{}, err
	}
	normalized, err := normalizeCreatePaymentInput(input)
	if err != nil {
		return model.Payment{}, err
	}
	return s.repo.Create(ctx, actorID, normalized)
}

func (s *PaymentService) Update(ctx context.Context, id string, actorID string, role string, input model.CreatePaymentInput) (model.Payment, error) {
	if err := s.requireAccess(ctx, actorID, role); err != nil {
		return model.Payment{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return model.Payment{}, errors.New("некорректный идентификатор оплаты")
	}
	normalized, err := normalizeCreatePaymentInput(input)
	if err != nil {
		return model.Payment{}, err
	}
	item, err := s.repo.Update(ctx, id, normalized)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return model.Payment{}, ErrPaymentNotFound
		}
		return model.Payment{}, err
	}
	return item, nil
}

func (s *PaymentService) SetClosed(ctx context.Context, id string, actorID string, role string, isClosed bool) (model.Payment, error) {
	if err := s.requireAccess(ctx, actorID, role); err != nil {
		return model.Payment{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return model.Payment{}, errors.New("некорректный идентификатор оплаты")
	}
	item, err := s.repo.SetClosed(ctx, id, isClosed)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return model.Payment{}, ErrPaymentNotFound
		}
		return model.Payment{}, err
	}
	return item, nil
}

func (s *PaymentService) Delete(ctx context.Context, id string, actorID string, role string) error {
	if err := s.requireAccess(ctx, actorID, role); err != nil {
		return err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("некорректный идентификатор оплаты")
	}
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, repository.ErrPaymentNotFound) {
		return ErrPaymentNotFound
	}
	return err
}

func (s *PaymentService) requireAccess(ctx context.Context, actorID string, role string) error {
	if strings.TrimSpace(actorID) == "" {
		return errors.New("authorization required")
	}
	if role == "admin" {
		return nil
	}

	user, err := s.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return ErrPaymentForbidden
	}
	if !user.IsActive {
		return ErrPaymentForbidden
	}
	if isAccountantPosition(user.Position) {
		return nil
	}
	return ErrPaymentForbidden
}

func normalizeCreatePaymentInput(input model.CreatePaymentInput) (model.CreatePaymentInput, error) {
	shortTitle := strings.TrimSpace(input.ShortTitle)
	if shortTitle == "" {
		return model.CreatePaymentInput{}, errors.New("краткое описание обязательно")
	}
	if utf8.RuneCountInString(shortTitle) > paymentShortTitleMaxLength {
		return model.CreatePaymentInput{}, errors.New("краткое описание не длиннее 22 символов")
	}
	if input.Date <= 0 {
		return model.CreatePaymentInput{}, errors.New("дата платежа обязательна")
	}
	if input.Amount < 0 {
		return model.CreatePaymentInput{}, errors.New("сумма не может быть отрицательной")
	}

	var remindAt *int64
	if input.RemindAt != nil {
		if *input.RemindAt <= 0 {
			return model.CreatePaymentInput{}, errors.New("некорректная дата напоминания")
		}
		if moscowCalendarDay(*input.RemindAt).After(moscowCalendarDay(input.Date)) {
			return model.CreatePaymentInput{}, errors.New("дата напоминания не может быть позже даты платежа")
		}
		value := *input.RemindAt
		remindAt = &value
	}

	var payerID *string
	if input.PayerID != nil {
		payer := strings.TrimSpace(*input.PayerID)
		if payer != "" {
			if _, ok := allowedPaymentPayers[payer]; !ok {
				return model.CreatePaymentInput{}, errors.New("некорректный плательщик")
			}
			payerID = &payer
		}
	}

	return model.CreatePaymentInput{
		Date:         input.Date,
		RemindAt:     remindAt,
		PayerID:      payerID,
		Counterparty: strings.TrimSpace(input.Counterparty),
		Amount:       input.Amount,
		ShortTitle:   shortTitle,
		Comment:      strings.TrimSpace(input.Comment),
	}, nil
}

func isAccountantPosition(position string) bool {
	normalized := strings.ToLower(strings.TrimSpace(position))
	return strings.Contains(normalized, "бухгалтер")
}

func moscowCalendarDay(ms int64) time.Time {
	local := time.UnixMilli(ms).In(moscowLocation)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, moscowLocation)
}

func mustLoadMoscowLocation() *time.Location {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.FixedZone("MSK", 3*60*60)
	}
	return location
}
