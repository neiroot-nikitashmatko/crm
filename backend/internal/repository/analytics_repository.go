package repository

import (
	"context"
	"time"

	"proclients/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepository struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) CountLeadsByTrafficSource(
	ctx context.Context,
	from time.Time,
	to time.Time,
) ([]model.TrafficSourceMetric, error) {
	const query = `
SELECT
  COALESCE(NULLIF(BTRIM(traffic_source), ''), 'Без источника') AS source,
  COUNT(*)::int AS items_count
FROM leads
WHERE deleted_at IS NULL
  AND column_id <> 'low_quality'
  AND created_at >= $1
  AND created_at <= $2
GROUP BY 1
ORDER BY items_count DESC, source ASC
`
	return r.scanTrafficSourceCounts(ctx, query, from, to)
}

func (r *AnalyticsRepository) CountDealsByTrafficSource(
	ctx context.Context,
	from time.Time,
	to time.Time,
) ([]model.TrafficSourceMetric, error) {
	const query = `
SELECT
  COALESCE(NULLIF(BTRIM(traffic_source), ''), 'Без источника') AS source,
  COUNT(*)::int AS items_count
FROM deals
WHERE deleted_at IS NULL
  AND created_at >= $1
  AND created_at <= $2
GROUP BY 1
ORDER BY items_count DESC, source ASC
`
	return r.scanTrafficSourceCounts(ctx, query, from, to)
}

func (r *AnalyticsRepository) CountLeadToDealConversion(
	ctx context.Context,
	from time.Time,
	to time.Time,
) (int, int, error) {
	const query = `
SELECT
  COUNT(*)::int AS leads_count,
  COUNT(*) FILTER (
    WHERE EXISTS (
      SELECT 1
      FROM deals d
      WHERE d.lead_id = leads.id
        AND d.deleted_at IS NULL
    )
  )::int AS converted_count
FROM leads
WHERE deleted_at IS NULL
  AND column_id <> 'low_quality'
  AND created_at >= $1
  AND created_at <= $2
`
	var leadsCount int
	var convertedCount int
	if err := r.db.QueryRow(ctx, query, from, to).Scan(&leadsCount, &convertedCount); err != nil {
		return 0, 0, err
	}
	return leadsCount, convertedCount, nil
}

func (r *AnalyticsRepository) CountFailedLeadShare(
	ctx context.Context,
	from time.Time,
	to time.Time,
) (int, int, error) {
	const query = `
SELECT
  COUNT(*)::int AS leads_count,
  COUNT(*) FILTER (WHERE column_id = 'failed')::int AS failed_count
FROM leads
WHERE deleted_at IS NULL
  AND column_id <> 'low_quality'
  AND created_at >= $1
  AND created_at <= $2
`
	var leadsCount int
	var failedCount int
	if err := r.db.QueryRow(ctx, query, from, to).Scan(&leadsCount, &failedCount); err != nil {
		return 0, 0, err
	}
	return leadsCount, failedCount, nil
}

func (r *AnalyticsRepository) CountFailedDealShare(
	ctx context.Context,
	from time.Time,
	to time.Time,
) (int, int, error) {
	const query = `
SELECT
  COUNT(*)::int AS deals_count,
  COUNT(*) FILTER (WHERE status = 'failed')::int AS failed_count
FROM deals
WHERE deleted_at IS NULL
  AND created_at >= $1
  AND created_at <= $2
`
	var dealsCount int
	var failedCount int
	if err := r.db.QueryRow(ctx, query, from, to).Scan(&dealsCount, &failedCount); err != nil {
		return 0, 0, err
	}
	return dealsCount, failedCount, nil
}

func (r *AnalyticsRepository) CountClosedDealsByNomenclature(
	ctx context.Context,
	from time.Time,
	to time.Time,
) ([]model.NomenclatureCount, error) {
	const query = `
SELECT
  BTRIM(production_nomenclature) AS nomenclature,
  COUNT(*)::int AS items_count
FROM deals
WHERE deleted_at IS NULL
  AND status = 'closed'
  AND BTRIM(production_nomenclature) <> ''
  AND BTRIM(production_employee) <> ''
  AND closed_at IS NOT NULL
  AND closed_at >= $1
  AND closed_at <= $2
GROUP BY 1
ORDER BY items_count DESC, nomenclature ASC
`
	rows, err := r.db.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.NomenclatureCount, 0)
	for rows.Next() {
		var item model.NomenclatureCount
		if err := rows.Scan(&item.Nomenclature, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *AnalyticsRepository) CountClosedDealsByEmployee(
	ctx context.Context,
	from time.Time,
	to time.Time,
) ([]model.EmployeeShareMetric, error) {
	const query = `
SELECT
  BTRIM(production_employee) AS employee,
  COUNT(*)::int AS items_count
FROM deals
WHERE deleted_at IS NULL
  AND status = 'closed'
  AND BTRIM(production_employee) <> ''
  AND closed_at IS NOT NULL
  AND closed_at >= $1
  AND closed_at <= $2
GROUP BY 1
ORDER BY items_count DESC, employee ASC
`
	rows, err := r.db.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.EmployeeShareMetric, 0)
	for rows.Next() {
		var item model.EmployeeShareMetric
		if err := rows.Scan(&item.Employee, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *AnalyticsRepository) ListClosedDeals(
	ctx context.Context,
	from time.Time,
	to time.Time,
	requireEmployee bool,
	requireProduction bool,
) ([]model.ClosedDealListItem, error) {
	const query = `
SELECT
  id::text,
  deal_number,
  first_name,
  COALESCE(patronymic, ''),
  phone,
  BTRIM(production_nomenclature),
  BTRIM(production_employee),
  created_at,
  closed_at
FROM deals
WHERE deleted_at IS NULL
  AND status = 'closed'
  AND closed_at IS NOT NULL
  AND closed_at >= $1
  AND closed_at <= $2
  AND ($3::bool = false OR BTRIM(production_employee) <> '')
  AND (
    $4::bool = false
    OR (
      BTRIM(production_nomenclature) <> ''
      AND BTRIM(production_employee) <> ''
    )
  )
ORDER BY closed_at DESC, deal_number DESC
`
	rows, err := r.db.Query(ctx, query, from, to, requireEmployee, requireProduction)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ClosedDealListItem, 0)
	for rows.Next() {
		var item model.ClosedDealListItem
		var createdAt time.Time
		var closedAt *time.Time
		if err := rows.Scan(
			&item.ID,
			&item.DealNumber,
			&item.FirstName,
			&item.Patronymic,
			&item.Phone,
			&item.Nomenclature,
			&item.Employee,
			&createdAt,
			&closedAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.UnixMilli()
		if closedAt != nil {
			item.ClosedAt = closedAt.UnixMilli()
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *AnalyticsRepository) ListFailedDeals(
	ctx context.Context,
	from time.Time,
	to time.Time,
) ([]model.ClosedDealListItem, error) {
	const query = `
SELECT
  id::text,
  deal_number,
  first_name,
  COALESCE(patronymic, ''),
  phone,
  BTRIM(production_nomenclature),
  BTRIM(production_employee),
  BTRIM(failure_reason),
  created_at
FROM deals
WHERE deleted_at IS NULL
  AND status = 'failed'
  AND created_at >= $1
  AND created_at <= $2
ORDER BY created_at DESC, deal_number DESC
`
	rows, err := r.db.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ClosedDealListItem, 0)
	for rows.Next() {
		var item model.ClosedDealListItem
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.DealNumber,
			&item.FirstName,
			&item.Patronymic,
			&item.Phone,
			&item.Nomenclature,
			&item.Employee,
			&item.FailureReason,
			&createdAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.UnixMilli()
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *AnalyticsRepository) ListFailedLeads(
	ctx context.Context,
	from time.Time,
	to time.Time,
) ([]model.FailedLeadListItem, error) {
	const query = `
SELECT
  id::text,
  lead_number,
  first_name,
  COALESCE(patronymic, ''),
  phone,
  BTRIM(COALESCE(failure_reason, '')),
  created_at
FROM leads
WHERE deleted_at IS NULL
  AND column_id = 'failed'
  AND created_at >= $1
  AND created_at <= $2
ORDER BY created_at DESC, lead_number DESC
`
	rows, err := r.db.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.FailedLeadListItem, 0)
	for rows.Next() {
		var item model.FailedLeadListItem
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.LeadNumber,
			&item.FirstName,
			&item.Patronymic,
			&item.Phone,
			&item.FailureReason,
			&createdAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.UnixMilli()
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *AnalyticsRepository) ListDealsForTrafficPeriod(
	ctx context.Context,
	from time.Time,
	to time.Time,
) ([]model.DealTrafficListItem, error) {
	const query = `
SELECT
  id::text,
  deal_number,
  first_name,
  COALESCE(patronymic, ''),
  phone,
  COALESCE(NULLIF(BTRIM(traffic_source), ''), 'Без источника') AS traffic_source,
  created_at
FROM deals
WHERE deleted_at IS NULL
  AND created_at >= $1
  AND created_at <= $2
ORDER BY created_at DESC, deal_number DESC
`
	rows, err := r.db.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.DealTrafficListItem, 0)
	for rows.Next() {
		var item model.DealTrafficListItem
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.DealNumber,
			&item.FirstName,
			&item.Patronymic,
			&item.Phone,
			&item.TrafficSource,
			&createdAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.UnixMilli()
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *AnalyticsRepository) ListIncomingStockLots(ctx context.Context) ([]model.IncomingStockLot, error) {
	const query = `
SELECT
  CASE
    WHEN i.catalog_product_id IS NOT NULL THEN 'id:' || i.catalog_product_id::text
    ELSE 'title:' || lower(btrim(i.title))
  END AS product_key,
  i.quantity::float8,
  i.unit_price::float8,
  inv.invoice_date,
  inv.created_at,
  inv.invoice_number,
  i.position
FROM incoming_invoice_items i
JOIN incoming_invoices inv ON inv.id = i.invoice_id
WHERE btrim(i.title) <> '' OR i.catalog_product_id IS NOT NULL
ORDER BY inv.invoice_date ASC, inv.created_at ASC, inv.invoice_number ASC, i.position ASC
`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.IncomingStockLot, 0)
	for rows.Next() {
		var item model.IncomingStockLot
		if err := rows.Scan(
			&item.ProductKey,
			&item.Quantity,
			&item.UnitCost,
			&item.InvoiceDate,
			&item.CreatedAt,
			&item.InvoiceNumber,
			&item.Position,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *AnalyticsRepository) ListOutgoingSaleLines(ctx context.Context) ([]model.OutgoingSaleLine, error) {
	const query = `
SELECT
  o.id::text,
  i.title,
  CASE
    WHEN i.catalog_product_id IS NOT NULL THEN 'id:' || i.catalog_product_id::text
    ELSE 'title:' || lower(btrim(i.title))
  END AS product_key,
  i.quantity::float8,
  i.unit_price::float8,
  o.invoice_date,
  o.created_at,
  o.invoice_number,
  i.position
FROM outgoing_invoices o
JOIN outgoing_invoice_items i ON i.invoice_id = o.id
WHERE btrim(i.title) <> '' OR i.catalog_product_id IS NOT NULL
ORDER BY o.invoice_date ASC, o.created_at ASC, o.invoice_number ASC, i.position ASC
`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.OutgoingSaleLine, 0)
	for rows.Next() {
		var item model.OutgoingSaleLine
		if err := rows.Scan(
			&item.InvoiceID,
			&item.Title,
			&item.ProductKey,
			&item.Quantity,
			&item.UnitPrice,
			&item.InvoiceDate,
			&item.CreatedAt,
			&item.InvoiceNumber,
			&item.Position,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *AnalyticsRepository) scanTrafficSourceCounts(
	ctx context.Context,
	query string,
	from time.Time,
	to time.Time,
) ([]model.TrafficSourceMetric, error) {
	rows, err := r.db.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.TrafficSourceMetric, 0)
	for rows.Next() {
		var item model.TrafficSourceMetric
		if err := rows.Scan(&item.Source, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
