package service

import (
	"context"
	"errors"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

type LeadService struct {
	repo *repository.LeadRepository
}

func NewLeadService(repo *repository.LeadRepository) *LeadService {
	return &LeadService{repo: repo}
}

func (s *LeadService) List(ctx context.Context) ([]model.Lead, error) {
	return s.repo.List(ctx)
}

func (s *LeadService) Create(ctx context.Context, input model.CreateLeadInput) (model.Lead, error) {
	if strings.TrimSpace(input.FirstName) == "" || strings.TrimSpace(input.Phone) == "" {
		return model.Lead{}, errors.New("firstName and phone are required")
	}
	if strings.TrimSpace(input.ColumnID) == "" {
		input.ColumnID = "new"
	}
	return s.repo.Create(ctx, input)
}

func (s *LeadService) UpdateColumn(ctx context.Context, leadID string, columnID string, failureReason *string) (model.Lead, error) {
	if strings.TrimSpace(columnID) == "" {
		return model.Lead{}, errors.New("columnId is required")
	}
	if columnID == "failed" {
		if failureReason == nil || strings.TrimSpace(*failureReason) == "" {
			return model.Lead{}, errors.New("failureReason is required")
		}
	}
	return s.repo.UpdateColumn(ctx, leadID, columnID, failureReason)
}

func (s *LeadService) UpdateComment(ctx context.Context, leadID string, comment string) (model.Lead, error) {
	return s.repo.UpdateComment(ctx, leadID, comment)
}

func (s *LeadService) UpdatePickupDelivery(ctx context.Context, leadID string, input model.PickupDelivery) (model.Lead, error) {
	if strings.TrimSpace(leadID) == "" {
		return model.Lead{}, errors.New("leadId is required")
	}
	return s.repo.UpdatePickupDelivery(ctx, leadID, input)
}

func (s *LeadService) UpdateProducts(ctx context.Context, leadID string, products []model.DealProduct) (model.Lead, error) {
	if strings.TrimSpace(leadID) == "" {
		return model.Lead{}, errors.New("leadId is required")
	}
	return s.repo.ReplaceProducts(ctx, leadID, products)
}

func (s *LeadService) UpdateProduction(ctx context.Context, leadID string, input model.DealProduction) (model.Lead, error) {
	if strings.TrimSpace(leadID) == "" {
		return model.Lead{}, errors.New("leadId is required")
	}
	return s.repo.UpdateProduction(ctx, leadID, input)
}

func (s *LeadService) Delete(ctx context.Context, leadID string) error {
	return s.repo.SoftDelete(ctx, leadID)
}
