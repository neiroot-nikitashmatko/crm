package service

import (
	"context"
	"errors"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"

	"github.com/jackc/pgx/v5"
)

type QuickReplyService struct {
	repo *repository.QuickReplyRepository
}

func NewQuickReplyService(repo *repository.QuickReplyRepository) *QuickReplyService {
	return &QuickReplyService{repo: repo}
}

func (s *QuickReplyService) List(ctx context.Context) ([]model.QuickReplySection, error) {
	return s.repo.ListSectionsWithReplies(ctx)
}

func (s *QuickReplyService) CreateSection(ctx context.Context, input model.CreateQuickReplySectionInput) (model.QuickReplySection, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return model.QuickReplySection{}, errors.New("title is required")
	}
	return s.repo.CreateSection(ctx, title)
}

func (s *QuickReplyService) UpdateSection(ctx context.Context, id string, input model.UpdateQuickReplySectionInput) (model.QuickReplySection, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return model.QuickReplySection{}, errors.New("title is required")
	}
	section, err := s.repo.UpdateSection(ctx, id, title)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.QuickReplySection{}, errors.New("section not found")
	}
	return section, err
}

func (s *QuickReplyService) DeleteSection(ctx context.Context, id string) error {
	err := s.repo.DeleteSection(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.New("section not found")
	}
	return err
}

func (s *QuickReplyService) CreateReply(ctx context.Context, sectionID string, input model.CreateQuickReplyInput) (model.QuickReply, error) {
	title := strings.TrimSpace(input.Title)
	body := strings.TrimSpace(input.Body)
	if title == "" {
		return model.QuickReply{}, errors.New("title is required")
	}
	if body == "" {
		return model.QuickReply{}, errors.New("body is required")
	}
	exists, err := s.repo.SectionExists(ctx, sectionID)
	if err != nil {
		return model.QuickReply{}, err
	}
	if !exists {
		return model.QuickReply{}, errors.New("section not found")
	}
	return s.repo.CreateReply(ctx, sectionID, title, body)
}

func (s *QuickReplyService) UpdateReply(ctx context.Context, id string, input model.UpdateQuickReplyInput) (model.QuickReply, error) {
	title := strings.TrimSpace(input.Title)
	body := strings.TrimSpace(input.Body)
	if title == "" {
		return model.QuickReply{}, errors.New("title is required")
	}
	if body == "" {
		return model.QuickReply{}, errors.New("body is required")
	}
	reply, err := s.repo.UpdateReply(ctx, id, title, body)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.QuickReply{}, errors.New("reply not found")
	}
	return reply, err
}

func (s *QuickReplyService) DeleteReply(ctx context.Context, id string) error {
	err := s.repo.DeleteReply(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.New("reply not found")
	}
	return err
}
