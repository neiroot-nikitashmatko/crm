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

type LeadRepository struct {
	db *pgxpool.Pool
}

const leadSelectSQL = `
  id::text,
  lead_number,
  first_name,
  COALESCE(patronymic, ''),
  phone,
  traffic_source,
  column_id,
  COALESCE(lead_comments, ''),
  COALESCE(failure_reason, ''),
  created_by::text,
  created_at,
  updated_at,
  COALESCE(pickup_address, ''),
  pickup_date,
  COALESCE(delivery_address, ''),
  delivery_date,
  COALESCE(courier, ''),
  COALESCE(production_nomenclature, ''),
  production_due_at,
  COALESCE(production_employee, '')
`

func NewLeadRepository(db *pgxpool.Pool) *LeadRepository {
	return &LeadRepository{db: db}
}

func (r *LeadRepository) List(ctx context.Context) ([]model.Lead, error) {
	query := `
SELECT` + leadSelectSQL + `
FROM leads
WHERE deleted_at IS NULL
ORDER BY created_at DESC
`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.Lead, 0)
	for rows.Next() {
		lead, err := scanLead(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, lead)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return r.enrichLeads(ctx, result)
}

func (r *LeadRepository) Create(ctx context.Context, input model.CreateLeadInput) (model.Lead, error) {
	const query = `
INSERT INTO leads (
  first_name, patronymic, phone, traffic_source, column_id, created_by
)
VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6::uuid)
RETURNING id::text
`
	var leadID string
	err := r.db.QueryRow(ctx, query,
		input.FirstName,
		input.Patronymic,
		input.Phone,
		input.TrafficSource,
		input.ColumnID,
		input.CreatedBy,
	).Scan(&leadID)
	if err != nil {
		return model.Lead{}, err
	}
	return r.GetByID(ctx, leadID)
}

func (r *LeadRepository) UpdateColumn(ctx context.Context, leadID string, columnID string, failureReason *string) (model.Lead, error) {
	_, err := r.db.Exec(ctx, `
UPDATE leads
SET column_id = $2,
    failure_reason = CASE WHEN $2 = 'failed' THEN COALESCE($3, '') ELSE failure_reason END,
    updated_at = now()
WHERE id = $1::uuid AND deleted_at IS NULL
`, leadID, columnID, failureReason)
	if err != nil {
		return model.Lead{}, err
	}
	return r.GetByID(ctx, leadID)
}

func (r *LeadRepository) UpdateComment(ctx context.Context, leadID string, comment string) (model.Lead, error) {
	_, err := r.db.Exec(ctx, `
UPDATE leads
SET lead_comments = $2, updated_at = now()
WHERE id = $1::uuid AND deleted_at IS NULL
`, leadID, comment)
	if err != nil {
		return model.Lead{}, err
	}
	return r.GetByID(ctx, leadID)
}

func (r *LeadRepository) UpdatePickupDelivery(ctx context.Context, leadID string, input model.PickupDelivery) (model.Lead, error) {
	_, err := r.db.Exec(ctx, `
UPDATE leads
SET pickup_address = $2,
    pickup_date = CASE WHEN $3::bigint IS NULL THEN NULL ELSE to_timestamp($3::double precision / 1000) END,
    delivery_address = $4,
    delivery_date = CASE WHEN $5::bigint IS NULL THEN NULL ELSE to_timestamp($5::double precision / 1000) END,
    courier = $6,
    updated_at = now()
WHERE id = $1::uuid AND deleted_at IS NULL
`, leadID, input.PickupAddress, input.PickupDate, input.DeliveryAddress, input.DeliveryDate, input.Courier)
	if err != nil {
		return model.Lead{}, err
	}
	return r.GetByID(ctx, leadID)
}

func (r *LeadRepository) UpdateProduction(ctx context.Context, leadID string, input model.DealProduction) (model.Lead, error) {
	_, err := r.db.Exec(ctx, `
UPDATE leads
SET production_nomenclature = $2,
    production_due_at = CASE WHEN $3::bigint IS NULL THEN NULL ELSE to_timestamp($3::double precision / 1000) END,
    production_employee = $4,
    updated_at = now()
WHERE id = $1::uuid AND deleted_at IS NULL
`, leadID, input.Nomenclature, input.DueAt, input.Employee)
	if err != nil {
		return model.Lead{}, err
	}
	return r.GetByID(ctx, leadID)
}

func (r *LeadRepository) ReplaceProducts(ctx context.Context, leadID string, products []model.DealProduct) (model.Lead, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Lead{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM lead_products WHERE lead_id = $1::uuid`, leadID)
	if err != nil {
		return model.Lead{}, err
	}

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
			return model.Lead{}, fmt.Errorf("unitPrice must be non-negative")
		}

		amount := float64(quantity) * unitPrice

		_, err = tx.Exec(ctx, `
INSERT INTO lead_products (lead_id, position, title, quantity, unit_price, amount)
VALUES ($1::uuid, $2, $3, $4, $5, $6)
`, leadID, position, title, quantity, unitPrice, amount)
		if err != nil {
			return model.Lead{}, err
		}
		position++
	}

	_, err = tx.Exec(ctx, `
UPDATE leads SET updated_at = now()
WHERE id = $1::uuid AND deleted_at IS NULL
`, leadID)
	if err != nil {
		return model.Lead{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Lead{}, err
	}

	return r.GetByID(ctx, leadID)
}

func (r *LeadRepository) SoftDelete(ctx context.Context, leadID string) error {
	const query = `UPDATE leads SET deleted_at = now(), updated_at = now() WHERE id = $1::uuid AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, leadID)
	return err
}

func (r *LeadRepository) GetByID(ctx context.Context, leadID string) (model.Lead, error) {
	query := `
SELECT` + leadSelectSQL + `
FROM leads
WHERE id = $1::uuid AND deleted_at IS NULL
`
	row := r.db.QueryRow(ctx, query, leadID)
	lead, err := scanLead(row)
	if err != nil {
		return model.Lead{}, err
	}
	return r.enrichLead(ctx, lead)
}

func (r *LeadRepository) enrichLeads(ctx context.Context, leads []model.Lead) ([]model.Lead, error) {
	if len(leads) == 0 {
		return leads, nil
	}

	ids := make([]string, len(leads))
	for i, lead := range leads {
		ids[i] = lead.ID
	}

	productsByLead, err := r.listProductsByLeadIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	for i := range leads {
		leads[i].Products = productsByLead[leads[i].ID]
		if leads[i].Products == nil {
			leads[i].Products = []model.DealProduct{}
		}
	}

	return leads, nil
}

func (r *LeadRepository) enrichLead(ctx context.Context, lead model.Lead) (model.Lead, error) {
	products, err := r.listProducts(ctx, lead.ID)
	if err != nil {
		return model.Lead{}, err
	}
	lead.Products = products
	if lead.Products == nil {
		lead.Products = []model.DealProduct{}
	}
	return lead, nil
}

func (r *LeadRepository) listProducts(ctx context.Context, leadID string) ([]model.DealProduct, error) {
	rows, err := r.db.Query(ctx, `
SELECT title, quantity, unit_price
FROM lead_products
WHERE lead_id = $1::uuid
ORDER BY position ASC
`, leadID)
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

func (r *LeadRepository) listProductsByLeadIDs(ctx context.Context, leadIDs []string) (map[string][]model.DealProduct, error) {
	result := make(map[string][]model.DealProduct, len(leadIDs))
	if len(leadIDs) == 0 {
		return result, nil
	}

	rows, err := r.db.Query(ctx, `
SELECT lead_id::text, title, quantity, unit_price
FROM lead_products
WHERE lead_id = ANY($1::uuid[])
ORDER BY lead_id, position ASC
`, leadIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var leadID string
		var product model.DealProduct
		if err := rows.Scan(&leadID, &product.Title, &product.Quantity, &product.UnitPrice); err != nil {
			return nil, err
		}
		result[leadID] = append(result[leadID], product)
	}
	return result, rows.Err()
}

type leadScanner interface {
	Scan(dest ...any) error
}

func scanLead(scanner leadScanner) (model.Lead, error) {
	var lead model.Lead
	var createdAt time.Time
	var updatedAt time.Time
	var pickupDate *time.Time
	var deliveryDate *time.Time
	var productionDueAt *time.Time

	err := scanner.Scan(
		&lead.ID,
		&lead.LeadNumber,
		&lead.FirstName,
		&lead.Patronymic,
		&lead.Phone,
		&lead.TrafficSource,
		&lead.ColumnID,
		&lead.LeadComments,
		&lead.FailureReason,
		&lead.CreatedBy,
		&createdAt,
		&updatedAt,
		&lead.PickupDelivery.PickupAddress,
		&pickupDate,
		&lead.PickupDelivery.DeliveryAddress,
		&deliveryDate,
		&lead.PickupDelivery.Courier,
		&lead.Production.Nomenclature,
		&productionDueAt,
		&lead.Production.Employee,
	)
	if err != nil {
		return model.Lead{}, err
	}

	lead.CreatedAt = createdAt.UnixMilli()
	lead.UpdatedAt = updatedAt.UnixMilli()

	if pickupDate != nil {
		value := pickupDate.UnixMilli()
		lead.PickupDelivery.PickupDate = &value
	} else {
		lead.PickupDelivery.PickupDate = nil
	}

	if deliveryDate != nil {
		value := deliveryDate.UnixMilli()
		lead.PickupDelivery.DeliveryDate = &value
	} else {
		lead.PickupDelivery.DeliveryDate = nil
	}

	if productionDueAt != nil {
		value := productionDueAt.UnixMilli()
		lead.Production.DueAt = &value
	} else {
		lead.Production.DueAt = nil
	}

	lead.Products = []model.DealProduct{}

	return lead, nil
}
