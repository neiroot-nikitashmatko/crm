package service

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

type AttachmentService struct {
	repo       *repository.AttachmentRepository
	activities *ActivityService
}

func NewAttachmentService(repo *repository.AttachmentRepository, activities *ActivityService) *AttachmentService {
	return &AttachmentService{repo: repo, activities: activities}
}

func (s *AttachmentService) ListByDeal(ctx context.Context, dealID string) ([]model.Attachment, error) {
	dealID = strings.TrimSpace(dealID)
	if dealID == "" {
		return nil, errors.New("invalid deal id")
	}
	return s.repo.ListByEntity(ctx, model.AttachmentEntityDeal, dealID)
}

func (s *AttachmentService) ListByTask(ctx context.Context, taskID string) ([]model.Attachment, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, errors.New("invalid task id")
	}
	return s.repo.ListByEntity(ctx, model.AttachmentEntityTask, taskID)
}

func (s *AttachmentService) AttachToDeals(ctx context.Context, deals []model.Deal) error {
	if len(deals) == 0 {
		return nil
	}

	ids := make([]string, 0, len(deals))
	for index := range deals {
		deals[index].Attachments = nil
		if id := strings.TrimSpace(deals[index].ID); id != "" {
			ids = append(ids, id)
		}
	}

	attachmentsByEntity, err := s.repo.ListByEntityIDs(ctx, model.AttachmentEntityDeal, ids)
	if err != nil {
		return err
	}

	for index := range deals {
		if id := strings.TrimSpace(deals[index].ID); id != "" {
			deals[index].Attachments = attachmentsByEntity[id]
		}
	}
	return nil
}

func (s *AttachmentService) AttachToTasks(ctx context.Context, tasks []model.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	ids := make([]string, 0, len(tasks))
	for index := range tasks {
		tasks[index].Attachments = nil
		if id := strings.TrimSpace(tasks[index].ID); id != "" {
			ids = append(ids, id)
		}
	}

	attachmentsByEntity, err := s.repo.ListByEntityIDs(ctx, model.AttachmentEntityTask, ids)
	if err != nil {
		return err
	}

	for index := range tasks {
		if id := strings.TrimSpace(tasks[index].ID); id != "" {
			tasks[index].Attachments = attachmentsByEntity[id]
		}
	}
	return nil
}

func (s *AttachmentService) AttachToDeal(ctx context.Context, deal *model.Deal) error {
	if deal == nil {
		return nil
	}
	items, err := s.repo.ListByEntity(ctx, model.AttachmentEntityDeal, deal.ID)
	if err != nil {
		return err
	}
	deal.Attachments = items
	return nil
}

func (s *AttachmentService) AttachToTask(ctx context.Context, task *model.Task) error {
	if task == nil {
		return nil
	}
	items, err := s.repo.ListByEntity(ctx, model.AttachmentEntityTask, task.ID)
	if err != nil {
		return err
	}
	task.Attachments = items
	return nil
}

func (s *AttachmentService) UploadToDeal(
	ctx context.Context,
	dealID string,
	uploadedBy string,
	filename string,
	mimeType string,
	content []byte,
) (model.Attachment, error) {
	return s.upload(ctx, model.AttachmentEntityDeal, dealID, uploadedBy, filename, mimeType, content, s.repo.DealExists)
}

func (s *AttachmentService) UploadToTask(
	ctx context.Context,
	taskID string,
	uploadedBy string,
	filename string,
	mimeType string,
	content []byte,
) (model.Attachment, error) {
	return s.upload(ctx, model.AttachmentEntityTask, taskID, uploadedBy, filename, mimeType, content, s.repo.TaskExists)
}

func (s *AttachmentService) GetContent(ctx context.Context, attachmentID string) (model.AttachmentContent, error) {
	attachmentID = strings.TrimSpace(attachmentID)
	if attachmentID == "" {
		return model.AttachmentContent{}, errors.New("invalid attachment id")
	}
	return s.repo.GetContent(ctx, attachmentID)
}

func (s *AttachmentService) Delete(ctx context.Context, attachmentID string, authorID string) (*model.Activity, error) {
	attachmentID = strings.TrimSpace(attachmentID)
	if attachmentID == "" {
		return nil, errors.New("invalid attachment id")
	}

	meta, err := s.repo.GetMeta(ctx, attachmentID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SoftDelete(ctx, attachmentID); err != nil {
		return nil, err
	}

	if meta.EntityType == model.AttachmentEntityTask && strings.TrimSpace(authorID) != "" {
		activity, err := s.activities.LogTaskFileRemoved(ctx, meta.EntityID, authorID, meta.Name)
		if err != nil {
			return nil, err
		}
		return &activity, nil
	}

	return nil, nil
}

func (s *AttachmentService) LogDealUploadBatch(ctx context.Context, dealID string, authorID string, count int) (model.Activity, error) {
	return s.activities.LogDealFilesUploaded(ctx, dealID, authorID, count)
}

func (s *AttachmentService) LogTaskUploadBatch(ctx context.Context, taskID string, authorID string, count int) (model.Activity, error) {
	return s.activities.LogTaskFilesUploaded(ctx, taskID, authorID, count)
}

func (s *AttachmentService) upload(
	ctx context.Context,
	entityType string,
	entityID string,
	uploadedBy string,
	filename string,
	mimeType string,
	content []byte,
	exists func(context.Context, string) (bool, error),
) (model.Attachment, error) {
	entityID = strings.TrimSpace(entityID)
	uploadedBy = strings.TrimSpace(uploadedBy)
	filename = sanitizeAttachmentFilename(filename)
	mimeType = normalizeAttachmentMimeType(filename, mimeType)

	if entityID == "" {
		return model.Attachment{}, errors.New("invalid entity id")
	}
	if uploadedBy == "" {
		return model.Attachment{}, errors.New("uploadedBy is required")
	}
	if filename == "" {
		return model.Attachment{}, errors.New("file name is required")
	}
	if len(content) == 0 {
		return model.Attachment{}, errors.New("file is empty")
	}
	if len(content) > model.MaxAttachmentSizeBytes {
		return model.Attachment{}, errors.New("file is larger than 10 MB")
	}

	ok, err := exists(ctx, entityID)
	if err != nil {
		return model.Attachment{}, err
	}
	if !ok {
		return model.Attachment{}, errors.New("entity not found")
	}

	item, err := s.repo.Create(ctx, entityType, entityID, uploadedBy, filename, mimeType, content)
	if err != nil {
		return model.Attachment{}, err
	}

	items, err := s.repo.ListByEntity(ctx, entityType, entityID)
	if err != nil {
		return item, nil
	}
	for _, candidate := range items {
		if candidate.ID == item.ID {
			return candidate, nil
		}
	}
	return item, nil
}

func sanitizeAttachmentFilename(name string) string {
	name = strings.TrimSpace(filepath.Base(name))
	name = strings.ReplaceAll(name, "\x00", "")
	return name
}

func normalizeAttachmentMimeType(filename string, mimeType string) string {
	mimeType = strings.TrimSpace(mimeType)
	if mimeType == "" || mimeType == "application/octet-stream" {
		if detected := mime.TypeByExtension(filepath.Ext(filename)); detected != "" {
			return detected
		}
		return "application/octet-stream"
	}
	return mimeType
}

func ReadUploadFile(file io.Reader, maxSize int64) ([]byte, error) {
	limited := io.LimitReader(file, maxSize+1)
	content, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(content)) > maxSize {
		return nil, errors.New("file is larger than 10 MB")
	}
	if len(content) == 0 {
		return nil, errors.New("file is empty")
	}
	return content, nil
}

func ParseMultipartFiles(r *http.Request) ([]struct {
	Filename string
	MimeType string
	Content  []byte
}, error) {
	if err := r.ParseMultipartForm(model.MaxAttachmentSizeBytes + (1 << 20)); err != nil {
		return nil, errors.New("invalid multipart form")
	}

	if r.MultipartForm == nil || len(r.MultipartForm.File["file"]) == 0 {
		return nil, errors.New("file is required")
	}

	files := make([]struct {
		Filename string
		MimeType string
		Content  []byte
	}, 0, len(r.MultipartForm.File["file"]))

	for _, header := range r.MultipartForm.File["file"] {
		source, err := header.Open()
		if err != nil {
			return nil, err
		}

		content, err := ReadUploadFile(source, model.MaxAttachmentSizeBytes)
		_ = source.Close()
		if err != nil {
			return nil, err
		}

		mimeType := header.Header.Get("Content-Type")
		files = append(files, struct {
			Filename string
			MimeType string
			Content  []byte
		}{
			Filename: header.Filename,
			MimeType: mimeType,
			Content:  content,
		})
	}

	return files, nil
}
