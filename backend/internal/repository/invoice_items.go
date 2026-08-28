package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"proclients/backend/internal/model"

	"github.com/jackc/pgx/v5"
)

func invoiceItemsTable(kind string) (string, error) {
	switch kind {
	case "incoming":
		return "incoming_invoice_items", nil
	case "outgoing":
		return "outgoing_invoice_items", nil
	default:
		return "", fmt.Errorf("unknown invoice items kind: %s", kind)
	}
}

func replaceInvoiceItems(ctx context.Context, tx pgx.Tx, kind string, invoiceID string, items []model.InvoiceItem) error {
	table, err := invoiceItemsTable(kind)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM `+table+` WHERE invoice_id = $1::uuid`, invoiceID); err != nil {
		return err
	}

	const insertSQL = `
INSERT INTO %s (invoice_id, position, catalog_product_id, title, quantity, unit_price)
VALUES ($1::uuid, $2, NULLIF($3, '')::uuid, $4, $5, $6)
`
	query := fmt.Sprintf(insertSQL, table)
	for index, item := range items {
		if _, err := tx.Exec(
			ctx,
			query,
			invoiceID,
			index,
			strings.TrimSpace(item.CatalogProductID),
			item.Title,
			item.Quantity,
			item.UnitPrice,
		); err != nil {
			if isForeignKeyViolation(err) {
				return errors.New("товар из каталога не найден")
			}
			return err
		}
	}
	return nil
}

func listInvoiceItemsByKind(ctx context.Context, q interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}, kind string) (map[string][]model.InvoiceItem, error) {
	table, err := invoiceItemsTable(kind)
	if err != nil {
		return nil, err
	}

	query := `
SELECT
  invoice_id::text,
  COALESCE(catalog_product_id::text, ''),
  title,
  quantity,
  unit_price
FROM ` + table + `
ORDER BY invoice_id, position ASC
`
	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]model.InvoiceItem)
	for rows.Next() {
		var invoiceID string
		var item model.InvoiceItem
		if err := rows.Scan(
			&invoiceID,
			&item.CatalogProductID,
			&item.Title,
			&item.Quantity,
			&item.UnitPrice,
		); err != nil {
			return nil, err
		}
		result[invoiceID] = append(result[invoiceID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
