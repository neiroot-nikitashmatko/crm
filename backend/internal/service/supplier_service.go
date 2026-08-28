package service

import (
	"context"
	"errors"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

type SupplierService struct {
	repo *repository.SupplierRepository
}

func NewSupplierService(repo *repository.SupplierRepository) *SupplierService {
	return &SupplierService{repo: repo}
}

func (s *SupplierService) List(ctx context.Context) ([]model.Supplier, error) {
	return s.repo.List(ctx)
}

func (s *SupplierService) Create(ctx context.Context, input model.UpsertSupplierInput) (model.Supplier, error) {
	normalized, err := normalizeSupplierInput(input)
	if err != nil {
		return model.Supplier{}, err
	}
	return s.repo.Create(ctx, normalized)
}

func (s *SupplierService) Update(ctx context.Context, id string, input model.UpsertSupplierInput) (model.Supplier, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.Supplier{}, errors.New("некорректный идентификатор поставщика")
	}
	normalized, err := normalizeSupplierInput(input)
	if err != nil {
		return model.Supplier{}, err
	}
	return s.repo.Update(ctx, id, normalized)
}

func (s *SupplierService) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("некорректный идентификатор поставщика")
	}
	return s.repo.Delete(ctx, id)
}

func normalizeSupplierInput(input model.UpsertSupplierInput) (model.UpsertSupplierInput, error) {
	name := strings.TrimSpace(input.Name)
	contactPerson := strings.TrimSpace(input.ContactPerson)
	phone := strings.TrimSpace(input.Phone)
	inn := strings.TrimSpace(input.INN)
	legalAddress := strings.TrimSpace(input.LegalAddress)
	actualAddress := strings.TrimSpace(input.ActualAddress)
	bik := strings.TrimSpace(input.BIK)
	settlementAccount := strings.TrimSpace(input.SettlementAccount)
	correspondentAccount := strings.TrimSpace(input.CorrespondentAccount)

	switch {
	case name == "":
		return model.UpsertSupplierInput{}, errors.New("название обязательно")
	case contactPerson == "":
		return model.UpsertSupplierInput{}, errors.New("контактное лицо обязательно")
	case phone == "" || !phonePattern.MatchString(phone):
		return model.UpsertSupplierInput{}, errors.New("некорректный телефон")
	case inn == "":
		return model.UpsertSupplierInput{}, errors.New("ИНН обязателен")
	case legalAddress == "":
		return model.UpsertSupplierInput{}, errors.New("юридический адрес обязателен")
	case actualAddress == "":
		return model.UpsertSupplierInput{}, errors.New("фактический адрес обязателен")
	case bik == "":
		return model.UpsertSupplierInput{}, errors.New("БИК обязателен")
	case settlementAccount == "":
		return model.UpsertSupplierInput{}, errors.New("расчётный счёт обязателен")
	case correspondentAccount == "":
		return model.UpsertSupplierInput{}, errors.New("корреспондентский счёт обязателен")
	}

	return model.UpsertSupplierInput{
		Name:                 name,
		ContactPerson:        contactPerson,
		Phone:                phone,
		INN:                  inn,
		KPP:                  strings.TrimSpace(input.KPP),
		OGRN:                 strings.TrimSpace(input.OGRN),
		LegalAddress:         legalAddress,
		ActualAddress:        actualAddress,
		BIK:                  bik,
		SettlementAccount:    settlementAccount,
		CorrespondentAccount: correspondentAccount,
	}, nil
}
