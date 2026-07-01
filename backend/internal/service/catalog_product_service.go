package service

import (
	"context"
	"errors"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

type CatalogProductService struct {
	repo *repository.CatalogProductRepository
}

func NewCatalogProductService(repo *repository.CatalogProductRepository) *CatalogProductService {
	return &CatalogProductService{repo: repo}
}

func (s *CatalogProductService) List(ctx context.Context) ([]model.CatalogProduct, error) {
	return s.repo.List(ctx)
}

func (s *CatalogProductService) Create(ctx context.Context, input model.UpsertCatalogProductInput) (model.CatalogProduct, error) {
	normalized, err := normalizeCatalogProductInput(input)
	if err != nil {
		return model.CatalogProduct{}, err
	}
	return s.repo.Create(ctx, normalized)
}

func (s *CatalogProductService) Update(ctx context.Context, id string, input model.UpsertCatalogProductInput) (model.CatalogProduct, error) {
	if strings.TrimSpace(id) == "" {
		return model.CatalogProduct{}, errors.New("invalid product id")
	}
	normalized, err := normalizeCatalogProductInput(input)
	if err != nil {
		return model.CatalogProduct{}, err
	}
	return s.repo.Update(ctx, id, normalized)
}

func (s *CatalogProductService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("invalid product id")
	}
	return s.repo.Delete(ctx, id)
}

func normalizeCatalogProductInput(input model.UpsertCatalogProductInput) (model.UpsertCatalogProductInput, error) {
	name := strings.TrimSpace(input.Name)
	sku := strings.TrimSpace(input.SKU)
	category := strings.TrimSpace(input.Category)
	cost := input.Cost

	if name == "" || sku == "" {
		return model.UpsertCatalogProductInput{}, errors.New("name and sku are required")
	}
	if cost < 0 {
		return model.UpsertCatalogProductInput{}, errors.New("cost must be non-negative")
	}

	return model.UpsertCatalogProductInput{
		Name:     name,
		SKU:      sku,
		Category: category,
		Cost:     cost,
	}, nil
}
