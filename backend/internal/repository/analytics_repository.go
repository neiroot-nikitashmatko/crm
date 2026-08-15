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
  AND created_at >= $1
  AND created_at <= $2
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
  AND created_at >= $1
  AND created_at <= $2
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
