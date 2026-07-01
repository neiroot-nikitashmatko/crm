package repository

import (
	"context"
	"time"

	"proclients/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) List(ctx context.Context) ([]model.Task, error) {
	rows, err := r.db.Query(ctx, `
SELECT
  t.id::text,
  t.title,
  t.text,
  t.due_at,
  t.status::text,
  t.lead_id::text,
  t.deal_id::text,
  CASE
    WHEN COALESCE(u.first_name, '') = '' AND COALESCE(u.last_name, '') = '' THEN t.created_by::text
    ELSE trim(concat_ws(' ', COALESCE(u.first_name, ''), COALESCE(u.last_name, '')))
  END AS created_by_name,
  t.created_at,
  t.updated_at
FROM tasks t
LEFT JOIN users u ON u.id = t.created_by
ORDER BY t.created_at DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, task)
	}
	return result, rows.Err()
}

func (r *TaskRepository) Create(ctx context.Context, input model.CreateTaskInput) (model.Task, error) {
	var taskID string
	row := r.db.QueryRow(ctx, `
INSERT INTO tasks (title, text, due_at, status, lead_id, deal_id, created_by)
VALUES ($1, $2, to_timestamp($3::double precision / 1000), $4::task_status, $5::uuid, $6::uuid, $7::uuid)
RETURNING id::text
`, input.Title, input.Text, input.DueAt, input.Status, input.LeadID, input.DealID, input.CreatedBy)
	if err := row.Scan(&taskID); err != nil {
		return model.Task{}, err
	}
	return r.getByID(ctx, taskID)
}

func (r *TaskRepository) Update(ctx context.Context, taskID string, input model.UpdateTaskInput) (model.Task, error) {
	_, err := r.db.Exec(ctx, `
UPDATE tasks
SET
  title = COALESCE($2, title),
  text = COALESCE($3, text),
  due_at = CASE WHEN $4::boolean THEN to_timestamp($5::double precision / 1000) ELSE due_at END,
  updated_at = now()
WHERE id = $1::uuid
`, taskID, input.Title, input.Text, input.HasDueAt, input.DueAt)
	if err != nil {
		return model.Task{}, err
	}
	return r.getByID(ctx, taskID)
}

func (r *TaskRepository) Complete(ctx context.Context, taskID string) (model.Task, error) {
	_, err := r.db.Exec(ctx, `
UPDATE tasks
SET status = 'completed', completed_at = now(), updated_at = now()
WHERE id = $1::uuid
`, taskID)
	if err != nil {
		return model.Task{}, err
	}
	return r.getByID(ctx, taskID)
}

func (r *TaskRepository) CompleteByLead(ctx context.Context, leadID string) error {
	_, err := r.db.Exec(ctx, `
UPDATE tasks
SET status = 'completed', completed_at = now(), updated_at = now()
WHERE lead_id = $1::uuid AND status = 'active'
`, leadID)
	return err
}

func (r *TaskRepository) getByID(ctx context.Context, taskID string) (model.Task, error) {
	row := r.db.QueryRow(ctx, `
SELECT
  t.id::text,
  t.title,
  t.text,
  t.due_at,
  t.status::text,
  t.lead_id::text,
  t.deal_id::text,
  CASE
    WHEN COALESCE(u.first_name, '') = '' AND COALESCE(u.last_name, '') = '' THEN t.created_by::text
    ELSE trim(concat_ws(' ', COALESCE(u.first_name, ''), COALESCE(u.last_name, '')))
  END AS created_by_name,
  t.created_at,
  t.updated_at
FROM tasks t
LEFT JOIN users u ON u.id = t.created_by
WHERE t.id = $1::uuid
`, taskID)
	return scanTask(row)
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (model.Task, error) {
	var task model.Task
	var dueAt *time.Time
	var createdAt time.Time
	var updatedAt time.Time
	var leadID *string
	var dealID *string
	err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Text,
		&dueAt,
		&task.Status,
		&leadID,
		&dealID,
		&task.CreatedBy,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return model.Task{}, err
	}

	task.LeadID = leadID
	task.DealID = dealID
	task.CreatedAt = createdAt.UnixMilli()
	task.UpdatedAt = updatedAt.UnixMilli()
	if dueAt != nil {
		value := dueAt.UnixMilli()
		task.DueAt = &value
	}
	return task, nil
}
