package service

import (
	"context"
	"errors"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

type TaskService struct {
	repo        *repository.TaskRepository
	attachments *AttachmentService
	activities  *ActivityService
}

func NewTaskService(
	repo *repository.TaskRepository,
	attachments *AttachmentService,
	activities *ActivityService,
) *TaskService {
	return &TaskService{repo: repo, attachments: attachments, activities: activities}
}

func (s *TaskService) List(ctx context.Context) ([]model.Task, error) {
	tasks, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.attachments.AttachToTasks(ctx, tasks); err != nil {
		return nil, err
	}
	if err := s.activities.AttachToTasks(ctx, tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *TaskService) Create(ctx context.Context, input model.CreateTaskInput) (model.Task, error) {
	if strings.TrimSpace(input.Title) == "" {
		return model.Task{}, errors.New("title is required")
	}
	if strings.TrimSpace(input.CreatedBy) == "" {
		return model.Task{}, errors.New("createdBy is required")
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if input.LeadID == nil && input.DealID == nil {
		return model.Task{}, errors.New("leadId or dealId is required")
	}
	task, err := s.repo.Create(ctx, input)
	if err != nil {
		return model.Task{}, err
	}
	if _, err := s.activities.LogTaskCreated(ctx, task.ID, input.CreatedBy); err != nil {
		return model.Task{}, err
	}
	return s.enrichTask(ctx, task)
}

func (s *TaskService) enrichTask(ctx context.Context, task model.Task) (model.Task, error) {
	if err := s.attachments.AttachToTask(ctx, &task); err != nil {
		return model.Task{}, err
	}
	if err := s.activities.AttachToTask(ctx, &task); err != nil {
		return model.Task{}, err
	}
	return task, nil
}

func (s *TaskService) Update(ctx context.Context, taskID string, input model.UpdateTaskInput) (model.Task, error) {
	task, err := s.repo.Update(ctx, taskID, input)
	if err != nil {
		return model.Task{}, err
	}
	return s.enrichTask(ctx, task)
}

func (s *TaskService) Complete(ctx context.Context, taskID string) (model.Task, error) {
	task, err := s.repo.Complete(ctx, taskID)
	if err != nil {
		return model.Task{}, err
	}
	if _, err := s.activities.LogTaskCompleted(ctx, task.ID, task.CreatedBy); err != nil {
		return model.Task{}, err
	}
	return s.enrichTask(ctx, task)
}

func (s *TaskService) CompleteByLead(ctx context.Context, leadID string) error {
	if strings.TrimSpace(leadID) == "" {
		return errors.New("leadId is required")
	}
	return s.repo.CompleteByLead(ctx, leadID)
}

func (s *TaskService) AddComment(ctx context.Context, taskID string, authorID string, text string) (model.Activity, error) {
	return s.activities.CreateComment(ctx, model.ActivityEntityTask, taskID, authorID, text)
}

func (s *TaskService) ListActivities(ctx context.Context, taskID string) ([]model.Activity, error) {
	return s.activities.ListByTask(ctx, taskID)
}
