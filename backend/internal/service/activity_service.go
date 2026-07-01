package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

type ActivityService struct {
	repo *repository.ActivityRepository
}

func NewActivityService(repo *repository.ActivityRepository) *ActivityService {
	return &ActivityService{repo: repo}
}

func (s *ActivityService) ListByDeal(ctx context.Context, dealID string) ([]model.Activity, error) {
	dealID = strings.TrimSpace(dealID)
	if dealID == "" {
		return nil, errors.New("invalid deal id")
	}
	return s.repo.ListByEntity(ctx, model.ActivityEntityDeal, dealID)
}

func (s *ActivityService) ListByTask(ctx context.Context, taskID string) ([]model.Activity, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, errors.New("invalid task id")
	}
	return s.repo.ListByEntity(ctx, model.ActivityEntityTask, taskID)
}

func (s *ActivityService) AttachToDeals(ctx context.Context, deals []model.Deal) error {
	if len(deals) == 0 {
		return nil
	}

	ids := make([]string, 0, len(deals))
	for index := range deals {
		deals[index].Activities = nil
		if id := strings.TrimSpace(deals[index].ID); id != "" {
			ids = append(ids, id)
		}
	}

	activitiesByEntity, err := s.repo.ListByEntityIDs(ctx, model.ActivityEntityDeal, ids)
	if err != nil {
		return err
	}

	for index := range deals {
		if id := strings.TrimSpace(deals[index].ID); id != "" {
			deals[index].Activities = activitiesByEntity[id]
		}
	}
	return nil
}

func (s *ActivityService) AttachToTasks(ctx context.Context, tasks []model.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	ids := make([]string, 0, len(tasks))
	for index := range tasks {
		tasks[index].Activities = nil
		if id := strings.TrimSpace(tasks[index].ID); id != "" {
			ids = append(ids, id)
		}
	}

	activitiesByEntity, err := s.repo.ListByEntityIDs(ctx, model.ActivityEntityTask, ids)
	if err != nil {
		return err
	}

	for index := range tasks {
		if id := strings.TrimSpace(tasks[index].ID); id != "" {
			tasks[index].Activities = activitiesByEntity[id]
		}
	}
	return nil
}

func (s *ActivityService) AttachToDeal(ctx context.Context, deal *model.Deal) error {
	if deal == nil {
		return nil
	}
	items, err := s.repo.ListByEntity(ctx, model.ActivityEntityDeal, deal.ID)
	if err != nil {
		return err
	}
	deal.Activities = items
	return nil
}

func (s *ActivityService) AttachToTask(ctx context.Context, task *model.Task) error {
	if task == nil {
		return nil
	}
	items, err := s.repo.ListByEntity(ctx, model.ActivityEntityTask, task.ID)
	if err != nil {
		return err
	}
	task.Activities = items
	return nil
}

func (s *ActivityService) CreateComment(
	ctx context.Context,
	entityType string,
	entityID string,
	authorID string,
	text string,
) (model.Activity, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return model.Activity{}, errors.New("text is required")
	}
	return s.create(ctx, entityType, entityID, authorID, model.ActivityTypeComment, text)
}

func (s *ActivityService) LogDealCreated(ctx context.Context, dealID string, authorID string) (model.Activity, error) {
	return s.create(ctx, model.ActivityEntityDeal, dealID, authorID, model.ActivityTypeSystem, "Сделка создана")
}

func (s *ActivityService) LogDealFailureReason(ctx context.Context, dealID string, authorID string, reason string) (model.Activity, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return model.Activity{}, errors.New("reason is required")
	}
	return s.create(ctx, model.ActivityEntityDeal, dealID, authorID, model.ActivityTypeSystem, "Причина провала: "+reason)
}

func (s *ActivityService) LogDealProductionRescheduled(ctx context.Context, dealID string, authorID string) (model.Activity, error) {
	return s.create(ctx, model.ActivityEntityDeal, dealID, authorID, model.ActivityTypeSystem, "Перенесена дата и время производства")
}

func (s *ActivityService) LogDealFilesUploaded(ctx context.Context, dealID string, authorID string, count int) (model.Activity, error) {
	if count <= 0 {
		return model.Activity{}, errors.New("invalid files count")
	}
	return s.create(ctx, model.ActivityEntityDeal, dealID, authorID, model.ActivityTypeSystem, fmt.Sprintf("Прикреплен файл (%d)", count))
}

func (s *ActivityService) LogTaskCreated(ctx context.Context, taskID string, authorID string) (model.Activity, error) {
	return s.create(ctx, model.ActivityEntityTask, taskID, authorID, model.ActivityTypeSystem, "Задача создана")
}

func (s *ActivityService) LogTaskCompleted(ctx context.Context, taskID string, authorID string) (model.Activity, error) {
	return s.create(ctx, model.ActivityEntityTask, taskID, authorID, model.ActivityTypeSystem, "Задача завершена")
}

func (s *ActivityService) LogTaskFilesUploaded(ctx context.Context, taskID string, authorID string, count int) (model.Activity, error) {
	if count <= 0 {
		return model.Activity{}, errors.New("invalid files count")
	}
	return s.create(ctx, model.ActivityEntityTask, taskID, authorID, model.ActivityTypeSystem, fmt.Sprintf("Добавлено файлов: %d", count))
}

func (s *ActivityService) LogTaskFileRemoved(ctx context.Context, taskID string, authorID string, fileName string) (model.Activity, error) {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		fileName = "файл"
	}
	return s.create(ctx, model.ActivityEntityTask, taskID, authorID, model.ActivityTypeSystem, "Удалён файл: "+fileName)
}

func (s *ActivityService) create(
	ctx context.Context,
	entityType string,
	entityID string,
	authorID string,
	activityType string,
	text string,
) (model.Activity, error) {
	entityID = strings.TrimSpace(entityID)
	authorID = strings.TrimSpace(authorID)
	if entityID == "" {
		return model.Activity{}, errors.New("invalid entity id")
	}
	if authorID == "" {
		return model.Activity{}, errors.New("author is required")
	}

	var exists bool
	var err error
	switch entityType {
	case model.ActivityEntityDeal:
		exists, err = s.repo.DealExists(ctx, entityID)
	case model.ActivityEntityTask:
		exists, err = s.repo.TaskExists(ctx, entityID)
	default:
		return model.Activity{}, errors.New("invalid entity type")
	}
	if err != nil {
		return model.Activity{}, err
	}
	if !exists {
		return model.Activity{}, errors.New("entity not found")
	}

	return s.repo.Create(ctx, entityType, entityID, authorID, activityType, text)
}
