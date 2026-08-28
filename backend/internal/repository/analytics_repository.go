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

func (r *AnalyticsRepository) TradeProfit(
	ctx context.Context,
	from time.Time,
	to time.Time,
) (model.TradeProfit, error) {
	const query = `
WITH product_costs AS (
  SELECT
    CASE
      WHEN catalog_product_id IS NOT NULL THEN 'id:' || catalog_product_id::text
      ELSE 'title:' || lower(btrim(title))
    END AS product_key,
    SUM(quantity * unit_price) / NULLIF(SUM(quantity), 0) AS avg_cost
  FROM incoming_invoice_items
  WHERE btrim(title) <> '' OR catalog_product_id IS NOT NULL
  GROUP BY 1
),
sales AS (
  SELECT
    o.id AS invoice_id,
    i.quantity,
    i.unit_price,
    CASE
      WHEN i.catalog_product_id IS NOT NULL THEN 'id:' || i.catalog_product_id::text
      ELSE 'title:' || lower(btrim(i.title))
    END AS product_key
  FROM outgoing_invoices o
  JOIN outgoing_invoice_items i ON i.invoice_id = o.id
  WHERE o.invoice_date >= $1
    AND o.invoice_date <= $2
    AND (btrim(i.title) <> '' OR i.catalog_product_id IS NOT NULL)
)
SELECT
  COALESCE(SUM(s.quantity * s.unit_price), 0)::float8 AS revenue,
  COALESCE(SUM(s.quantity * COALESCE(c.avg_cost, 0)), 0)::float8 AS cost,
  COALESCE(SUM(s.quantity * (s.unit_price - COALESCE(c.avg_cost, 0))), 0)::float8 AS profit,
  COUNT(DISTINCT s.invoice_id)::int AS invoices_count
FROM sales s
LEFT JOIN product_costs c ON c.product_key = s.product_key
`
	var item model.TradeProfit
	if err := r.db.QueryRow(ctx, query, from, to).Scan(
		&item.Revenue,
		&item.Cost,
		&item.Profit,
		&item.InvoicesCount,
	); err != nil {
		return model.TradeProfit{}, err
	}
	return item, nil
}

func (r *AnalyticsRepository) TradeProfitItems(
	ctx context.Context,
	from time.Time,
	to time.Time,
) ([]model.TradeProfitItem, error) {
	const query = `
WITH product_costs AS (
  SELECT
    CASE
      WHEN catalog_product_id IS NOT NULL THEN 'id:' || catalog_product_id::text
      ELSE 'title:' || lower(btrim(title))
    END AS product_key,
    SUM(quantity * unit_price) / NULLIF(SUM(quantity), 0) AS avg_cost
  FROM incoming_invoice_items
  WHERE btrim(title) <> '' OR catalog_product_id IS NOT NULL
  GROUP BY 1
),
sales AS (
  SELECT
    i.title,
    i.quantity,
    i.unit_price,
    CASE
      WHEN i.catalog_product_id IS NOT NULL THEN 'id:' || i.catalog_product_id::text
      ELSE 'title:' || lower(btrim(i.title))
    END AS product_key
  FROM outgoing_invoices o
  JOIN outgoing_invoice_items i ON i.invoice_id = o.id
  WHERE o.invoice_date >= $1
    AND o.invoice_date <= $2
    AND (btrim(i.title) <> '' OR i.catalog_product_id IS NOT NULL)
)
SELECT
  s.product_key,
  COALESCE(
    (
      ARRAY_AGG(NULLIF(btrim(s.title), '') ORDER BY s.quantity DESC)
      FILTER (WHERE btrim(s.title) <> '')
    )[1],
    'Без названия'
  ) AS title,
  COALESCE(SUM(s.quantity), 0)::float8 AS quantity,
  COALESCE(c.avg_cost, 0)::float8 AS cost_price,
  COALESCE(
    SUM(s.quantity * s.unit_price) / NULLIF(SUM(s.quantity), 0),
    0
  )::float8 AS sale_price,
  COALESCE(SUM(s.quantity * (s.unit_price - COALESCE(c.avg_cost, 0))), 0)::float8 AS profit,
  (c.avg_cost IS NOT NULL) AS has_cost
FROM sales s
LEFT JOIN product_costs c ON c.product_key = s.product_key
GROUP BY s.product_key, c.avg_cost
ORDER BY profit DESC, title ASC
`
	rows, err := r.db.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.TradeProfitItem, 0)
	for rows.Next() {
		var item model.TradeProfitItem
		if err := rows.Scan(
			&item.ProductKey,
			&item.Title,
			&item.Quantity,
			&item.CostPrice,
			&item.SalePrice,
			&item.Profit,
			&item.HasCost,
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
