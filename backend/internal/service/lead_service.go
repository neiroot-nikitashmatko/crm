package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

var leadPhonePattern = regexp.MustCompile(`^\+[1-9][0-9]{10,14}$`)

type LeadService struct {
	repo        *repository.LeadRepository
	attachments *AttachmentService
	activities  *ActivityService
}

func NewLeadService(repo *repository.LeadRepository, attachments *AttachmentService, activities *ActivityService) *LeadService {
	return &LeadService{repo: repo, attachments: attachments, activities: activities}
}

func (s *LeadService) List(ctx context.Context) ([]model.Lead, error) {
	leads, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.attachments.AttachToLeads(ctx, leads); err != nil {
		return nil, err
	}
	if err := s.activities.AttachToLeads(ctx, leads); err != nil {
		return nil, err
	}
	return leads, nil
}

func (s *LeadService) Create(ctx context.Context, input model.CreateLeadInput) (model.Lead, error) {
	if strings.TrimSpace(input.FirstName) == "" {
		return model.Lead{}, errors.New("firstName is required")
	}
	input.Phone = strings.TrimSpace(input.Phone)
	if strings.TrimSpace(input.ColumnID) == "" {
		input.ColumnID = "new"
	}
	lead, err := s.repo.Create(ctx, input)
	if err != nil {
		return model.Lead{}, err
	}
	if strings.TrimSpace(input.CreatedBy) != "" {
		if _, err := s.activities.LogLeadCreated(ctx, lead.ID, input.CreatedBy); err != nil {
			return model.Lead{}, err
		}
	}
	return s.enrichLead(ctx, lead)
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
	lead, err := s.repo.UpdateColumn(ctx, leadID, columnID, failureReason)
	if err != nil {
		return model.Lead{}, err
	}
	return s.enrichLead(ctx, lead)
}

func (s *LeadService) UpdateComment(ctx context.Context, leadID string, comment string) (model.Lead, error) {
	lead, err := s.repo.UpdateComment(ctx, leadID, comment)
	if err != nil {
		return model.Lead{}, err
	}
	return s.enrichLead(ctx, lead)
}

func (s *LeadService) UpdateProfile(
	ctx context.Context,
	leadID string,
	firstName string,
	patronymic string,
	phone string,
) (model.Lead, error) {
	firstName = strings.TrimSpace(firstName)
	patronymic = strings.TrimSpace(patronymic)
	phone = strings.TrimSpace(phone)
	if firstName == "" {
		return model.Lead{}, errors.New("имя обязательно")
	}
	if phone != "" && !leadPhonePattern.MatchString(phone) {
		return model.Lead{}, errors.New("Некорректный формат телефона. Укажите номер полностью, например +79001234567.")
	}
	lead, err := s.repo.UpdateProfile(ctx, leadID, firstName, patronymic, phone)
	if err != nil {
		return model.Lead{}, err
	}
	return s.enrichLead(ctx, lead)
}

func (s *LeadService) UpdatePickupDelivery(ctx context.Context, leadID string, input model.PickupDelivery) (model.Lead, error) {
	if strings.TrimSpace(leadID) == "" {
		return model.Lead{}, errors.New("leadId is required")
	}
	lead, err := s.repo.UpdatePickupDelivery(ctx, leadID, input)
	if err != nil {
		return model.Lead{}, err
	}
	return s.enrichLead(ctx, lead)
}

func (s *LeadService) UpdateProducts(ctx context.Context, leadID string, products []model.DealProduct) (model.Lead, error) {
	if strings.TrimSpace(leadID) == "" {
		return model.Lead{}, errors.New("leadId is required")
	}
	lead, err := s.repo.ReplaceProducts(ctx, leadID, products)
	if err != nil {
		return model.Lead{}, err
	}
	return s.enrichLead(ctx, lead)
}

func (s *LeadService) UpdateProduction(ctx context.Context, leadID string, input model.DealProduction) (model.Lead, error) {
	if strings.TrimSpace(leadID) == "" {
		return model.Lead{}, errors.New("leadId is required")
	}
	lead, err := s.repo.UpdateProduction(ctx, leadID, input)
	if err != nil {
		return model.Lead{}, err
	}
	return s.enrichLead(ctx, lead)
}

func (s *LeadService) Delete(ctx context.Context, leadID string) error {
	return s.repo.SoftDelete(ctx, leadID)
}

func (s *LeadService) AddComment(ctx context.Context, leadID string, authorID string, text string) (model.Activity, error) {
	return s.activities.CreateComment(ctx, model.ActivityEntityLead, leadID, authorID, text)
}

func (s *LeadService) ListActivities(ctx context.Context, leadID string) ([]model.Activity, error) {
	return s.activities.ListByLead(ctx, leadID)
}

func (s *LeadService) enrichLead(ctx context.Context, lead model.Lead) (model.Lead, error) {
	if err := s.attachments.AttachToLead(ctx, &lead); err != nil {
		return model.Lead{}, err
	}
	if err := s.activities.AttachToLead(ctx, &lead); err != nil {
		return model.Lead{}, err
	}
	return lead, nil
}
