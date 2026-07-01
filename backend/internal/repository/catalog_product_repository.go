package repository

import (
	"context"
	"errors"
	"time"

	"proclients/backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CatalogProductRepository struct {
	db *pgxpool.Pool
}

func NewCatalogProductRepository(db *pgxpool.Pool) *CatalogProductRepository {
	return &CatalogProductRepository{db: db}
}

func (r *CatalogProductRepository) List(ctx context.Context) ([]model.CatalogProduct, error) {
	const query = `
SELECT
  id::text,
  name,
  sku,
  COALESCE(category, ''),
  cost,
  created_at,
  updated_at
FROM catalog_products
ORDER BY name ASC
`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.CatalogProduct, 0)
	for rows.Next() {
		item, err := scanCatalogProduct(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *CatalogProductRepository) Create(ctx context.Context, input model.UpsertCatalogProductInput) (model.CatalogProduct, error) {
	const query = `
INSERT INTO catalog_products (name, sku, category, cost)
VALUES ($1, $2, NULLIF($3, ''), $4)
RETURNING
  id::text,
  name,
  sku,
  COALESCE(category, ''),
  cost,
  created_at,
  updated_at
`
	row := r.db.QueryRow(ctx, query, input.Name, input.SKU, input.Category, input.Cost)
	item, err := scanCatalogProduct(row)
	if err != nil {
		if isUniqueViolation(err) {
			return model.CatalogProduct{}, errors.New("артикул уже используется")
		}
		return model.CatalogProduct{}, err
	}
	return item, nil
}

func (r *CatalogProductRepository) Update(ctx context.Context, id string, input model.UpsertCatalogProductInput) (model.CatalogProduct, error) {
	const query = `
UPDATE catalog_products
SET
  name = $2,
  sku = $3,
  category = NULLIF($4, ''),
  cost = $5,
  updated_at = now()
WHERE id = $1::uuid
RETURNING
  id::text,
  name,
  sku,
  COALESCE(category, ''),
  cost,
  created_at,
  updated_at
`
	row := r.db.QueryRow(ctx, query, id, input.Name, input.SKU, input.Category, input.Cost)
	item, err := scanCatalogProduct(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.CatalogProduct{}, errors.New("товар не найден")
		}
		if isUniqueViolation(err) {
			return model.CatalogProduct{}, errors.New("артикул уже используется")
		}
		return model.CatalogProduct{}, err
	}
	return item, nil
}

func (r *CatalogProductRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM catalog_products WHERE id = $1::uuid`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("товар не найден")
	}
	return nil
}

type catalogProductScanner interface {
	Scan(dest ...any) error
}

func scanCatalogProduct(row catalogProductScanner) (model.CatalogProduct, error) {
	var item model.CatalogProduct
	var createdAt time.Time
	var updatedAt time.Time
	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.SKU,
		&item.Category,
		&item.Cost,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return model.CatalogProduct{}, err
	}
	item.CreatedAt = createdAt.UnixMilli()
	item.UpdatedAt = updatedAt.UnixMilli()
	return item, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
