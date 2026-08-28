package service

import (
	"context"
	"errors"
	"math"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

type IncomingInvoiceService struct {
	repo         *repository.IncomingInvoiceRepository
	supplierRepo *repository.SupplierRepository
}

func NewIncomingInvoiceService(
	repo *repository.IncomingInvoiceRepository,
	supplierRepo *repository.SupplierRepository,
) *IncomingInvoiceService {
	return &IncomingInvoiceService{repo: repo, supplierRepo: supplierRepo}
}

func (s *IncomingInvoiceService) List(ctx context.Context) ([]model.IncomingInvoice, error) {
	return s.repo.List(ctx)
}

func (s *IncomingInvoiceService) Create(ctx context.Context, input model.UpsertIncomingInvoiceInput) (model.IncomingInvoice, error) {
	normalized, total, err := s.normalize(ctx, input)
	if err != nil {
		return model.IncomingInvoice{}, err
	}
	return s.repo.Create(ctx, normalized, total)
}

func (s *IncomingInvoiceService) Update(ctx context.Context, id string, input model.UpsertIncomingInvoiceInput) (model.IncomingInvoice, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.IncomingInvoice{}, errors.New("некорректный идентификатор накладной")
	}
	normalized, total, err := s.normalize(ctx, input)
	if err != nil {
		return model.IncomingInvoice{}, err
	}
	return s.repo.Update(ctx, id, normalized, total)
}

func (s *IncomingInvoiceService) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("некорректный идентификатор накладной")
	}
	return s.repo.Delete(ctx, id)
}

func (s *IncomingInvoiceService) normalize(ctx context.Context, input model.UpsertIncomingInvoiceInput) (model.UpsertIncomingInvoiceInput, float64, error) {
	if input.Date <= 0 {
		return model.UpsertIncomingInvoiceInput{}, 0, errors.New("дата обязательна")
	}
	supplierID := strings.TrimSpace(input.SupplierID)
	if supplierID == "" {
		return model.UpsertIncomingInvoiceInput{}, 0, errors.New("поставщик обязателен")
	}
	if _, err := s.supplierRepo.GetByID(ctx, supplierID); err != nil {
		if errors.Is(err, repository.ErrSupplierNotFound) {
			return model.UpsertIncomingInvoiceInput{}, 0, errors.New("поставщик не найден")
		}
		return model.UpsertIncomingInvoiceInput{}, 0, err
	}

	items, total, err := normalizeInvoiceItems(input.Items)
	if err != nil {
		return model.UpsertIncomingInvoiceInput{}, 0, err
	}

	return model.UpsertIncomingInvoiceInput{
		Date:       input.Date,
		SupplierID: supplierID,
		Items:      items,
		Comment:    strings.TrimSpace(input.Comment),
	}, total, nil
}

func normalizeInvoiceItems(items []model.InvoiceItem) ([]model.InvoiceItem, float64, error) {
	normalized := make([]model.InvoiceItem, 0, len(items))
	var total float64

	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		if title == "" && strings.TrimSpace(item.CatalogProductID) == "" {
			continue
		}
		if title == "" {
			return nil, 0, errors.New("укажите наименование товара")
		}
		if item.Quantity <= 0 {
			return nil, 0, errors.New("количество должно быть больше нуля")
		}
		if item.UnitPrice < 0 {
			return nil, 0, errors.New("цена не может быть отрицательной")
		}
		normalized = append(normalized, model.InvoiceItem{
			CatalogProductID: strings.TrimSpace(item.CatalogProductID),
			Title:            title,
			Quantity:         item.Quantity,
			UnitPrice:        item.UnitPrice,
		})
		total += item.Quantity * item.UnitPrice
	}

	if len(normalized) == 0 {
		return nil, 0, errors.New("добавьте хотя бы одну позицию")
	}

	return normalized, math.Round(total*100) / 100, nil
}
