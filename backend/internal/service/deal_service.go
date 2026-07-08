package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

type DealService struct {
	repo        *repository.DealRepository
	attachments *AttachmentService
	activities  *ActivityService
}

func NewDealService(
	repo *repository.DealRepository,
	attachments *AttachmentService,
	activities *ActivityService,
) *DealService {
	return &DealService{repo: repo, attachments: attachments, activities: activities}
}

func (s *DealService) List(ctx context.Context) ([]model.Deal, error) {
	deals, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.attachments.AttachToDeals(ctx, deals); err != nil {
		return nil, err
	}
	if err := s.activities.AttachToDeals(ctx, deals); err != nil {
		return nil, err
	}
	return deals, nil
}

func (s *DealService) CreateFromLead(ctx context.Context, input model.CreateDealFromLeadInput) (model.Deal, error) {
	if strings.TrimSpace(input.LeadID) == "" {
		return model.Deal{}, errors.New("leadId is required")
	}
	if strings.TrimSpace(input.CreatedBy) == "" {
		return model.Deal{}, errors.New("createdBy is required")
	}
	deal, err := s.repo.CreateFromLead(ctx, input)
	if err != nil {
		return model.Deal{}, err
	}
	if _, err := s.activities.LogDealCreated(ctx, deal.ID, input.CreatedBy); err != nil {
		return model.Deal{}, err
	}
	return s.enrichDeal(ctx, deal)
}

func (s *DealService) enrichDeal(ctx context.Context, deal model.Deal) (model.Deal, error) {
	if err := s.attachments.AttachToDeal(ctx, &deal); err != nil {
		return model.Deal{}, err
	}
	if err := s.activities.AttachToDeal(ctx, &deal); err != nil {
		return model.Deal{}, err
	}
	return deal, nil
}

func (s *DealService) UpdateStatus(ctx context.Context, dealID string, status string, failureReason *string) (model.Deal, error) {
	switch status {
	case "today", "tomorrow", "later", "closed", "failed":
	default:
		return model.Deal{}, errors.New("invalid status")
	}
	if status == "failed" {
		if failureReason == nil || strings.TrimSpace(*failureReason) == "" {
			return model.Deal{}, errors.New("failureReason is required")
		}
	}
	deal, err := s.repo.UpdateStatus(ctx, dealID, status, failureReason)
	if err != nil {
		return model.Deal{}, err
	}
	if status == "failed" && failureReason != nil {
		if _, err := s.activities.LogDealFailureReason(ctx, deal.ID, deal.CreatedBy, *failureReason); err != nil {
			return model.Deal{}, err
		}
	}
	return s.enrichDeal(ctx, deal)
}

func (s *DealService) UpdateComment(ctx context.Context, dealID string, comment string) (model.Deal, error) {
	deal, err := s.repo.UpdateComment(ctx, dealID, comment)
	if err != nil {
		return model.Deal{}, err
	}
	return s.enrichDeal(ctx, deal)
}

func (s *DealService) UpdateProfile(ctx context.Context, dealID string, firstName string, patronymic string) (model.Deal, error) {
	firstName = strings.TrimSpace(firstName)
	patronymic = strings.TrimSpace(patronymic)
	if firstName == "" {
		return model.Deal{}, errors.New("имя обязательно")
	}
	deal, err := s.repo.UpdateProfile(ctx, dealID, firstName, patronymic)
	if err != nil {
		return model.Deal{}, err
	}
	return s.enrichDeal(ctx, deal)
}

func (s *DealService) UpdateProductionDueAt(ctx context.Context, dealID string, dueAt *int64) (model.Deal, error) {
	if strings.TrimSpace(dealID) == "" {
		return model.Deal{}, errors.New("dealId is required")
	}

	previous, err := s.repo.GetByID(ctx, dealID)
	if err != nil {
		return model.Deal{}, err
	}

	deal, err := s.repo.UpdateProductionDueAt(ctx, dealID, dueAt)
	if err != nil {
		return model.Deal{}, err
	}

	if !sameTimestamp(previous.ProductionDueAt, dueAt) {
		if _, err := s.activities.LogDealProductionRescheduled(ctx, deal.ID, deal.CreatedBy); err != nil {
			return model.Deal{}, err
		}
	}

	return s.enrichDeal(ctx, deal)
}

func (s *DealService) UpdateProduction(ctx context.Context, dealID string, input model.DealProduction) (model.Deal, error) {
	if strings.TrimSpace(dealID) == "" {
		return model.Deal{}, errors.New("dealId is required")
	}

	deal, err := s.repo.UpdateProduction(ctx, dealID, input)
	if err != nil {
		return model.Deal{}, err
	}
	return s.enrichDeal(ctx, deal)
}

func (s *DealService) UpdatePickupDelivery(ctx context.Context, dealID string, input model.PickupDelivery) (model.Deal, error) {
	if strings.TrimSpace(dealID) == "" {
		return model.Deal{}, errors.New("dealId is required")
	}
	deal, err := s.repo.UpdatePickupDelivery(ctx, dealID, input)
	if err != nil {
		return model.Deal{}, err
	}
	return s.enrichDeal(ctx, deal)
}

func (s *DealService) UpdateProducts(ctx context.Context, dealID string, products []model.DealProduct) (model.Deal, error) {
	if strings.TrimSpace(dealID) == "" {
		return model.Deal{}, errors.New("dealId is required")
	}
	deal, err := s.repo.UpdateProducts(ctx, dealID, products)
	if err != nil {
		return model.Deal{}, err
	}
	return s.enrichDeal(ctx, deal)
}

func (s *DealService) Delete(ctx context.Context, dealID string) error {
	return s.repo.SoftDelete(ctx, dealID)
}

func (s *DealService) AddComment(ctx context.Context, dealID string, authorID string, text string) (model.Activity, error) {
	return s.activities.CreateComment(ctx, model.ActivityEntityDeal, dealID, authorID, text)
}

func (s *DealService) ListActivities(ctx context.Context, dealID string) ([]model.Activity, error) {
	return s.activities.ListByDeal(ctx, dealID)
}

func sameTimestamp(left *int64, right *int64) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	leftTime := time.UnixMilli(*left).UTC()
	rightTime := time.UnixMilli(*right).UTC()
	return leftTime.Equal(rightTime)
}
