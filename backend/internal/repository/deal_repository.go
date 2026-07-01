package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"proclients/backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DealRepository struct {
	db *pgxpool.Pool
}

const dealSelectSQL = `
  id::text,
  lead_id::text,
  deal_number,
  first_name,
  COALESCE(patronymic, ''),
  phone,
  traffic_source,
  status::text,
  total_amount,
  COALESCE(deal_comments, ''),
  COALESCE(failure_reason, ''),
  created_by::text,
  created_at,
  updated_at,
  production_nomenclature,
  production_due_at,
  production_employee,
  COALESCE(pickup_address, ''),
  pickup_date,
  COALESCE(delivery_address, ''),
  delivery_date,
  COALESCE(courier, '')
`

func NewDealRepository(db *pgxpool.Pool) *DealRepository {
	return &DealRepository{db: db}
}

func (r *DealRepository) List(ctx context.Context) ([]model.Deal, error) {
	query := `
SELECT` + dealSelectSQL + `
FROM deals
WHERE deleted_at IS NULL
ORDER BY created_at DESC
`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.Deal, 0)
	for rows.Next() {
		deal, err := scanDeal(rows)
		if err != nil {
			return nil, err
		}
		products, err := r.listProducts(ctx, deal.ID)
		if err != nil {
			return nil, err
		}
		deal.Products = products
		result = append(result, deal)
	}
	return result, rows.Err()
}

func (r *DealRepository) CreateFromLead(ctx context.Context, input model.CreateDealFromLeadInput) (model.Deal, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Deal{}, err
	}
	defer tx.Rollback(ctx)

	const createQuery = `
INSERT INTO deals (
  lead_id,
  first_name,
  patronymic,
  phone,
  traffic_source,
  status,
  total_amount,
  deal_comments,
  production_nomenclature,
  production_due_at,
  production_employee,
  pickup_address,
  pickup_date,
  delivery_address,
  delivery_date,
  courier,
  created_by
)
SELECT
  l.id,
  l.first_name,
  l.patronymic,
  l.phone,
  l.traffic_source,
  'today',
  0,
  '',
  $2,
  to_timestamp($3::double precision / 1000),
  $4,
  $6,
  CASE WHEN $7::bigint IS NULL THEN NULL ELSE to_timestamp($7::double precision / 1000) END,
  $8,
  CASE WHEN $9::bigint IS NULL THEN NULL ELSE to_timestamp($9::double precision / 1000) END,
  $10,
  $5::uuid
FROM leads l
WHERE l.id = $1::uuid AND l.deleted_at IS NULL
RETURNING
  id::text
`
	var dueAtMillis *int64
	if input.Production.DueAt != nil {
		dueAtMillis = input.Production.DueAt
	}
	pickupAddress := ""
	var pickupDateMillis *int64
	deliveryAddress := ""
	var deliveryDateMillis *int64
	courier := ""
	if input.PickupDelivery != nil {
		pickupAddress = input.PickupDelivery.PickupAddress
		pickupDateMillis = input.PickupDelivery.PickupDate
		deliveryAddress = input.PickupDelivery.DeliveryAddress
		deliveryDateMillis = input.PickupDelivery.DeliveryDate
		courier = input.PickupDelivery.Courier
	}
	var createdDealID string
	err = tx.QueryRow(ctx, createQuery,
		input.LeadID,
		input.Production.Nomenclature,
		dueAtMillis,
		input.Production.Employee,
		input.CreatedBy,
		pickupAddress,
		pickupDateMillis,
		deliveryAddress,
		deliveryDateMillis,
		courier,
	).Scan(&createdDealID)
	if err != nil {
		return model.Deal{}, err
	}

	// Re-read row with consistent scanner and timestamps.
	deal, err := r.getByIDTx(ctx, tx, createdDealID)
	if err != nil {
		return model.Deal{}, err
	}

	var totalAmount float64
	position := 0
	for _, product := range input.Products {
		title := strings.TrimSpace(product.Title)
		if title == "" {
			continue
		}

		quantity := product.Quantity
		if quantity <= 0 {
			quantity = 1
		}

		unitPrice := product.UnitPrice
		if unitPrice < 0 {
			unitPrice = 0
		}

		amount := float64(quantity) * unitPrice
		totalAmount += amount
		_, err = tx.Exec(ctx, `
INSERT INTO deal_products (deal_id, position, title, quantity, unit_price, amount)
VALUES ($1::uuid, $2, $3, $4, $5, $6)
`,
			deal.ID,
			position,
			title,
			quantity,
			unitPrice,
			amount,
		)
		if err != nil {
			return model.Deal{}, err
		}
		position++
	}

	_, err = tx.Exec(ctx, `
UPDATE deals
SET total_amount = $2, updated_at = now()
WHERE id = $1::uuid
`, deal.ID, totalAmount)
	if err != nil {
		return model.Deal{}, err
	}

	_, err = tx.Exec(ctx, `
UPDATE tasks
SET status = 'completed', completed_at = now(), updated_at = now()
WHERE lead_id = $1::uuid AND status = 'active'
`, input.LeadID)
	if err != nil {
		return model.Deal{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Deal{}, err
	}

	deal.Products, err = r.listProducts(ctx, deal.ID)
	if err != nil {
		return model.Deal{}, err
	}
	return deal, nil
}

func (r *DealRepository) UpdateStatus(ctx context.Context, dealID string, status string, failureReason *string) (model.Deal, error) {
	_, err := r.db.Exec(ctx, `
UPDATE deals
SET status = $2::deal_status,
    failure_reason = CASE WHEN $2::text = 'failed' THEN COALESCE($3, '') ELSE failure_reason END,
    updated_at = now()
WHERE id = $1::uuid AND deleted_at IS NULL
`, dealID, status, failureReason)
	if err != nil {
		return model.Deal{}, err
	}
	deal, err := r.getByID(ctx, dealID)
	if err != nil {
		return model.Deal{}, err
	}
	deal.Products, err = r.listProducts(ctx, dealID)
	if err != nil {
		return model.Deal{}, err
	}
	return deal, nil
}

func (r *DealRepository) UpdateComment(ctx context.Context, dealID string, comment string) (model.Deal, error) {
	_, err := r.db.Exec(ctx, `
UPDATE deals SET deal_comments = $2, updated_at = now()
WHERE id = $1::uuid AND deleted_at IS NULL
`, dealID, comment)
	if err != nil {
		return model.Deal{}, err
	}
	return r.getDealWithProducts(ctx, dealID)
}

func (r *DealRepository) UpdateProductionDueAt(ctx context.Context, dealID string, dueAtMillis *int64) (model.Deal, error) {
	_, err := r.db.Exec(ctx, `
UPDATE deals
SET production_due_at = to_timestamp($2::double precision / 1000),
    updated_at = now()
WHERE id = $1::uuid AND deleted_at IS NULL
`, dealID, dueAtMillis)
	if err != nil {
		return model.Deal{}, err
	}
	return r.getDealWithProducts(ctx, dealID)
}

func (r *DealRepository) UpdateProduction(ctx context.Context, dealID string, input model.DealProduction) (model.Deal, error) {
	_, err := r.db.Exec(ctx, `
UPDATE deals
SET production_nomenclature = $2,
    production_due_at = CASE WHEN $3::bigint IS NULL THEN NULL ELSE to_timestamp($3::double precision / 1000) END,
    production_employee = $4,
    updated_at = now()
WHERE id = $1::uuid AND deleted_at IS NULL
`, dealID, strings.TrimSpace(input.Nomenclature), input.DueAt, strings.TrimSpace(input.Employee))
	if err != nil {
		return model.Deal{}, err
	}
	return r.getDealWithProducts(ctx, dealID)
}

func (r *DealRepository) UpdatePickupDelivery(ctx context.Context, dealID string, input model.PickupDelivery) (model.Deal, error) {
	_, err := r.db.Exec(ctx, `
UPDATE deals
SET pickup_address = $2,
    pickup_date = CASE WHEN $3::bigint IS NULL THEN NULL ELSE to_timestamp($3::double precision / 1000) END,
    delivery_address = $4,
    delivery_date = CASE WHEN $5::bigint IS NULL THEN NULL ELSE to_timestamp($5::double precision / 1000) END,
    courier = $6,
    updated_at = now()
WHERE id = $1::uuid AND deleted_at IS NULL
`, dealID, input.PickupAddress, input.PickupDate, input.DeliveryAddress, input.DeliveryDate, input.Courier)
	if err != nil {
		return model.Deal{}, err
	}
	return r.getDealWithProducts(ctx, dealID)
}

func (r *DealRepository) getDealWithProducts(ctx context.Context, dealID string) (model.Deal, error) {
	deal, err := r.getByID(ctx, dealID)
	if err != nil {
		return model.Deal{}, err
	}
	deal.Products, err = r.listProducts(ctx, dealID)
	if err != nil {
		return model.Deal{}, err
	}
	return deal, nil
}

func (r *DealRepository) UpdateProducts(ctx context.Context, dealID string, products []model.DealProduct) (model.Deal, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Deal{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM deal_products WHERE deal_id = $1::uuid`, dealID)
	if err != nil {
		return model.Deal{}, err
	}

	var totalAmount float64
	position := 0
	for _, product := range products {
		title := strings.TrimSpace(product.Title)
		if title == "" {
			continue
		}

		quantity := product.Quantity
		if quantity <= 0 {
			quantity = 1
		}

		unitPrice := product.UnitPrice
		if unitPrice < 0 {
			return model.Deal{}, fmt.Errorf("unitPrice must be non-negative")
		}

		amount := float64(quantity) * unitPrice
		totalAmount += amount

		_, err = tx.Exec(ctx, `
INSERT INTO deal_products (deal_id, position, title, quantity, unit_price, amount)
VALUES ($1::uuid, $2, $3, $4, $5, $6)
`, dealID, position, title, quantity, unitPrice, amount)
		if err != nil {
			return model.Deal{}, err
		}
		position++
	}

	_, err = tx.Exec(ctx, `
UPDATE deals
SET total_amount = $2, updated_at = now()
WHERE id = $1::uuid AND deleted_at IS NULL
`, dealID, totalAmount)
	if err != nil {
		return model.Deal{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Deal{}, err
	}

	return r.getDealWithProducts(ctx, dealID)
}

func (r *DealRepository) listProductsTx(ctx context.Context, tx pgx.Tx, dealID string) ([]model.DealProduct, error) {
	rows, err := tx.Query(ctx, `
SELECT title, quantity, unit_price
FROM deal_products
WHERE deal_id = $1::uuid
ORDER BY position ASC
`, dealID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.DealProduct, 0)
	for rows.Next() {
		var product model.DealProduct
		if err := rows.Scan(&product.Title, &product.Quantity, &product.UnitPrice); err != nil {
			return nil, err
		}
		result = append(result, product)
	}
	return result, rows.Err()
}

func (r *DealRepository) SoftDelete(ctx context.Context, dealID string) error {
	_, err := r.db.Exec(ctx, `
UPDATE deals SET deleted_at = now(), updated_at = now()
WHERE id = $1::uuid AND deleted_at IS NULL
`, dealID)
	return err
}

func (r *DealRepository) GetByID(ctx context.Context, dealID string) (model.Deal, error) {
	return r.getByID(ctx, dealID)
}

func (r *DealRepository) getByID(ctx context.Context, dealID string) (model.Deal, error) {
	query := `
SELECT` + dealSelectSQL + `
FROM deals
WHERE id = $1::uuid AND deleted_at IS NULL
`
	row := r.db.QueryRow(ctx, query, dealID)
	return scanDeal(row)
}

func (r *DealRepository) getByIDTx(ctx context.Context, tx pgx.Tx, dealID string) (model.Deal, error) {
	query := `
SELECT` + dealSelectSQL + `
FROM deals
WHERE id = $1::uuid AND deleted_at IS NULL
`
	row := tx.QueryRow(ctx, query, dealID)
	return scanDeal(row)
}

func (r *DealRepository) listProducts(ctx context.Context, dealID string) ([]model.DealProduct, error) {
	rows, err := r.db.Query(ctx, `
SELECT title, quantity, unit_price
FROM deal_products
WHERE deal_id = $1::uuid
ORDER BY position ASC
`, dealID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.DealProduct, 0)
	for rows.Next() {
		var product model.DealProduct
		if err := rows.Scan(&product.Title, &product.Quantity, &product.UnitPrice); err != nil {
			return nil, err
		}
		result = append(result, product)
	}
	return result, rows.Err()
}

type dealScanner interface {
	Scan(dest ...any) error
}

func scanDeal(scanner dealScanner) (model.Deal, error) {
	var deal model.Deal
	var createdAt time.Time
	var updatedAt time.Time
	var leadID *string
	var productionDueAt *time.Time
	var pickupDate *time.Time
	var deliveryDate *time.Time

	err := scanner.Scan(
		&deal.ID,
		&leadID,
		&deal.DealNumber,
		&deal.FirstName,
		&deal.Patronymic,
		&deal.Phone,
		&deal.TrafficSource,
		&deal.Status,
		&deal.TotalAmount,
		&deal.DealComments,
		&deal.FailureReason,
		&deal.CreatedBy,
		&createdAt,
		&updatedAt,
		&deal.Production.Nomenclature,
		&productionDueAt,
		&deal.Production.Employee,
		&deal.PickupDelivery.PickupAddress,
		&pickupDate,
		&deal.PickupDelivery.DeliveryAddress,
		&deliveryDate,
		&deal.PickupDelivery.Courier,
	)
	if err != nil {
		return model.Deal{}, err
	}

	deal.LeadID = leadID
	deal.CreatedAt = createdAt.UnixMilli()
	deal.UpdatedAt = updatedAt.UnixMilli()

	if productionDueAt != nil {
		value := productionDueAt.UnixMilli()
		deal.ProductionDueAt = &value
		deal.Production.DueAt = &value
	} else {
		deal.ProductionDueAt = nil
		deal.Production.DueAt = nil
	}

	if pickupDate != nil {
		value := pickupDate.UnixMilli()
		deal.PickupDelivery.PickupDate = &value
	} else {
		deal.PickupDelivery.PickupDate = nil
	}

	if deliveryDate != nil {
		value := deliveryDate.UnixMilli()
		deal.PickupDelivery.DeliveryDate = &value
	} else {
		deal.PickupDelivery.DeliveryDate = nil
	}

	return deal, nil
}

func statusToColumnID(status string) string {
	switch status {
	case "tomorrow":
		return "tomorrow"
	case "later":
		return "later"
	case "closed":
		return "closed"
	case "failed":
		return "failed"
	default:
		return "today"
	}
}

func (r *DealRepository) DebugString(ctx context.Context) (string, error) {
	items, err := r.List(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("deals: %d", len(items)), nil
}
