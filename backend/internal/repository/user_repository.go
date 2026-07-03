package repository

import (
	"context"
	"errors"
	"time"

	"proclients/backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByCredentials(ctx context.Context, phone string, password string) (*model.AuthUser, error) {
	const query = `
SELECT
  id::text,
  phone,
  role::text,
  COALESCE(position, ''),
  COALESCE(first_name, ''),
  COALESCE(last_name, ''),
  COALESCE(patronymic, '')
FROM users
WHERE phone = $1
  AND password_hash = crypt($2, password_hash)
  AND is_active = true
LIMIT 1
`
	var user model.AuthUser
	err := r.db.QueryRow(ctx, query, phone, password).Scan(
		&user.ID,
		&user.Phone,
		&user.Role,
		&user.Position,
		&user.FirstName,
		&user.LastName,
		&user.Patronymic,
	)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

func (r *UserRepository) List(ctx context.Context) ([]model.User, error) {
	const query = `
SELECT
  id::text,
  COALESCE(first_name, ''),
  COALESCE(last_name, ''),
  COALESCE(patronymic, ''),
  phone,
  role::text,
  COALESCE(position, ''),
  birth_date,
  is_active,
  created_at,
  updated_at
FROM users
WHERE is_active = true
ORDER BY last_name ASC, first_name ASC, patronymic ASC
`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.User, 0)
	for rows.Next() {
		item, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (model.User, error) {
	const query = `
SELECT
  id::text,
  COALESCE(first_name, ''),
  COALESCE(last_name, ''),
  COALESCE(patronymic, ''),
  phone,
  role::text,
  COALESCE(position, ''),
  birth_date,
  is_active,
  created_at,
  updated_at
FROM users
WHERE id = $1::uuid
LIMIT 1
`
	row := r.db.QueryRow(ctx, query, id)
	item, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, err
	}
	return item, nil
}

func (r *UserRepository) Update(ctx context.Context, id string, input model.UpdateUserInput, updatePassword bool) (model.User, error) {
	const query = `
UPDATE users
SET
  first_name = $2,
  last_name = $3,
  patronymic = $4,
  phone = $5,
  role = $6::user_role,
  position = $7,
  birth_date = $8::date,
  password_hash = CASE WHEN $9 THEN crypt($10, gen_salt('bf')) ELSE password_hash END,
  updated_at = now()
WHERE id = $1::uuid
  AND is_active = true
RETURNING
  id::text,
  COALESCE(first_name, ''),
  COALESCE(last_name, ''),
  COALESCE(patronymic, ''),
  phone,
  role::text,
  COALESCE(position, ''),
  birth_date,
  is_active,
  created_at,
  updated_at
`
	row := r.db.QueryRow(
		ctx,
		query,
		id,
		input.FirstName,
		input.LastName,
		input.Patronymic,
		input.Phone,
		input.Role,
		input.Position,
		input.BirthDate,
		updatePassword,
		input.Password,
	)
	item, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		if isUniqueViolation(err) {
			return model.User{}, errors.New("сотрудник с таким телефоном уже существует")
		}
		return model.User{}, err
	}
	return item, nil
}

func (r *UserRepository) Deactivate(ctx context.Context, id string) error {
	const query = `
UPDATE users
SET is_active = false, updated_at = now()
WHERE id = $1::uuid
  AND is_active = true
`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) Create(ctx context.Context, input model.CreateUserInput) (model.User, error) {
	const query = `
INSERT INTO users (
  first_name,
  last_name,
  patronymic,
  phone,
  password_hash,
  role,
  position,
  birth_date
)
VALUES (
  $1,
  $2,
  $3,
  $4,
  crypt($5, gen_salt('bf')),
  $6::user_role,
  $7,
  $8::date
)
RETURNING
  id::text,
  COALESCE(first_name, ''),
  COALESCE(last_name, ''),
  COALESCE(patronymic, ''),
  phone,
  role::text,
  COALESCE(position, ''),
  birth_date,
  is_active,
  created_at,
  updated_at
`
	row := r.db.QueryRow(
		ctx,
		query,
		input.FirstName,
		input.LastName,
		input.Patronymic,
		input.Phone,
		input.Password,
		input.Role,
		input.Position,
		input.BirthDate,
	)
	item, err := scanUser(row)
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, errors.New("сотрудник с таким телефоном уже существует")
		}
		return model.User{}, err
	}
	return item, nil
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(row userScanner) (model.User, error) {
	var item model.User
	var birthDate *time.Time
	var createdAt time.Time
	var updatedAt time.Time

	err := row.Scan(
		&item.ID,
		&item.FirstName,
		&item.LastName,
		&item.Patronymic,
		&item.Phone,
		&item.Role,
		&item.Position,
		&birthDate,
		&item.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return model.User{}, err
	}

	if birthDate != nil {
		formatted := birthDate.Format("2006-01-02")
		item.BirthDate = &formatted
	}

	item.CreatedAt = createdAt.UnixMilli()
	item.UpdatedAt = updatedAt.UnixMilli()
	return item, nil
}
