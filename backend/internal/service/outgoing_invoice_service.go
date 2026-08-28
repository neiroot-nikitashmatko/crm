package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

type OutgoingInvoiceService struct {
	repo         *repository.OutgoingInvoiceRepository
	incomingRepo *repository.IncomingInvoiceRepository
	dealRepo     *repository.DealRepository
}

func NewOutgoingInvoiceService(
	repo *repository.OutgoingInvoiceRepository,
	incomingRepo *repository.IncomingInvoiceRepository,
	dealRepo *repository.DealRepository,
) *OutgoingInvoiceService {
	return &OutgoingInvoiceService{
		repo:         repo,
		incomingRepo: incomingRepo,
		dealRepo:     dealRepo,
	}
}

func (s *OutgoingInvoiceService) List(ctx context.Context) ([]model.OutgoingInvoice, error) {
	return s.repo.List(ctx)
}

func (s *OutgoingInvoiceService) Create(ctx context.Context, input model.UpsertOutgoingInvoiceInput) (model.OutgoingInvoice, error) {
	normalized, total, err := s.normalize(ctx, input, "")
	if err != nil {
		return model.OutgoingInvoice{}, err
	}
	return s.repo.Create(ctx, normalized, total)
}

func (s *OutgoingInvoiceService) Update(ctx context.Context, id string, input model.UpsertOutgoingInvoiceInput) (model.OutgoingInvoice, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.OutgoingInvoice{}, errors.New("некорректный идентификатор накладной")
	}
	normalized, total, err := s.normalize(ctx, input, id)
	if err != nil {
		return model.OutgoingInvoice{}, err
	}
	return s.repo.Update(ctx, id, normalized, total)
}

func (s *OutgoingInvoiceService) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("некорректный идентификатор накладной")
	}
	return s.repo.Delete(ctx, id)
}

func (s *OutgoingInvoiceService) normalize(
	ctx context.Context,
	input model.UpsertOutgoingInvoiceInput,
	excludeInvoiceID string,
) (model.UpsertOutgoingInvoiceInput, float64, error) {
	if input.Date <= 0 {
		return model.UpsertOutgoingInvoiceInput{}, 0, errors.New("дата обязательна")
	}
	dealID := strings.TrimSpace(input.DealID)
	if dealID == "" {
		return model.UpsertOutgoingInvoiceInput{}, 0, errors.New("сделка обязательна")
	}

	deal, err := s.dealRepo.GetByID(ctx, dealID)
	if err != nil {
		return model.UpsertOutgoingInvoiceInput{}, 0, errors.New("сделка не найдена")
	}
	if strings.EqualFold(deal.Status, "failed") {
		return model.UpsertOutgoingInvoiceInput{}, 0, errors.New("нельзя создать расходную накладную для проваленной сделки")
	}

	exists, err := s.repo.ExistsByDealID(ctx, dealID, excludeInvoiceID)
	if err != nil {
		return model.UpsertOutgoingInvoiceInput{}, 0, err
	}
	if exists {
		return model.UpsertOutgoingInvoiceInput{}, 0, repository.ErrOutgoingInvoiceDealTaken
	}

	items, total, err := normalizeInvoiceItems(input.Items)
	if err != nil {
		return model.UpsertOutgoingInvoiceInput{}, 0, err
	}

	if err := s.requireStock(ctx, items, excludeInvoiceID); err != nil {
		return model.UpsertOutgoingInvoiceInput{}, 0, err
	}

	return model.UpsertOutgoingInvoiceInput{
		Date:    input.Date,
		DealID:  dealID,
		Items:   items,
		Comment: strings.TrimSpace(input.Comment),
	}, total, nil
}

func (s *OutgoingInvoiceService) requireStock(ctx context.Context, items []model.InvoiceItem, excludeInvoiceID string) error {
	incoming, err := s.incomingRepo.List(ctx)
	if err != nil {
		return err
	}
	outgoing, err := s.repo.List(ctx)
	if err != nil {
		return err
	}

	quantities := buildStockQuantities(incoming, outgoing, excludeInvoiceID)
	shortages := outgoingStockShortageTitles(items, quantities)
	if len(shortages) == 0 {
		return nil
	}
	return fmt.Errorf("недостаточно товара на складе: %s", strings.Join(shortages, ", "))
}
