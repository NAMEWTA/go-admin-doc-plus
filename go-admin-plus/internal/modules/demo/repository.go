package demo

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"go-admin/internal/platform/database"
)

// productRecord is private persistence state; transport/domain types never reach SQL directly.
type productRecord struct {
	ID, OwnerAccountID, SKU, Name, NameKey, Description, Status string
	PriceCents, Revision                                        int64
	CreatedAt, UpdatedAt                                        time.Time
}

type repository struct{ dialect database.Dialect }

func (r repository) create(ctx context.Context, tx database.Tx, record productRecord) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO demo_products
		(id, owner_account_id, sku, name, name_key, description, price_cents, status, revision, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, record.ID, record.OwnerAccountID, record.SKU, record.Name, record.NameKey,
		record.Description, record.PriceCents, record.Status, record.Revision, record.CreatedAt, record.UpdatedAt)
	return err
}

func (r repository) get(ctx context.Context, tx database.Tx, id, actorID string, scope Scope) (productRecord, error) {
	if !validScope(scope) {
		return productRecord{}, ErrDenied
	}
	query := `SELECT id, owner_account_id, sku, name, description, price_cents, status, revision, created_at, updated_at FROM demo_products WHERE id = ?`
	args := []any{id}
	if scope == ScopeSelf {
		query += ` AND owner_account_id = ?`
		args = append(args, actorID)
	}
	var value productRecord
	err := tx.QueryRowContext(ctx, query, args...).Scan(&value.ID, &value.OwnerAccountID, &value.SKU, &value.Name,
		&value.Description, &value.PriceCents, &value.Status, &value.Revision, &value.CreatedAt, &value.UpdatedAt)
	return value, err
}

func (r repository) list(ctx context.Context, tx database.Tx, actorID string, scope Scope, query ListQuery) (Page, error) {
	if !validScope(scope) {
		return Page{}, ErrDenied
	}
	clauses, args := []string{"1 = 1"}, []any{}
	if scope == ScopeSelf {
		clauses, args = append(clauses, "owner_account_id = ?"), append(args, actorID)
	}
	if query.Search != "" {
		skuValue := "%" + query.Search + "%"
		nameValue := "%" + normalizedNameKey(query.Search) + "%"
		clauses, args = append(clauses, "(lower(sku) LIKE ? OR name_key LIKE ?)"), append(args, skuValue, nameValue)
	}
	where := " WHERE " + strings.Join(clauses, " AND ")
	var result Page
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM demo_products`+where, args...).Scan(&result.Total); err != nil {
		return Page{}, err
	}
	columns := map[string]string{"sku": "sku", "name": "name_key", "priceCents": "price_cents", "updatedAt": "updated_at"}
	if query.Sort == "name" {
		if r.dialect == database.DialectPostgres {
			columns["name"] = `name_key COLLATE "C"`
		} else {
			columns["name"] = "name_key COLLATE BINARY"
		}
	}
	directions := map[string]string{"ascending": "ASC", "descending": "DESC"}
	statement := `SELECT id, owner_account_id, sku, name, description, price_cents, status, revision, created_at, updated_at FROM demo_products` + where +
		` ORDER BY ` + columns[query.Sort] + ` ` + directions[query.Direction] + `, id ASC LIMIT ? OFFSET ?`
	args = append(args, query.PageSize, (query.Page-1)*query.PageSize)
	rows, err := tx.QueryContext(ctx, statement, args...)
	if err != nil {
		return Page{}, err
	}
	defer rows.Close()
	result.Rows = []Product{}
	for rows.Next() {
		var record productRecord
		if err := rows.Scan(&record.ID, &record.OwnerAccountID, &record.SKU, &record.Name, &record.Description,
			&record.PriceCents, &record.Status, &record.Revision, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return Page{}, err
		}
		result.Rows = append(result.Rows, mapProduct(record))
	}
	return result, rows.Err()
}

func (r repository) update(ctx context.Context, tx database.Tx, record productRecord, expectedRevision int64, actorID string, scope Scope) error {
	if !validScope(scope) {
		return ErrDenied
	}
	query := `UPDATE demo_products SET sku = ?, name = ?, name_key = ?, description = ?, price_cents = ?, status = ?, revision = revision + 1, updated_at = ? WHERE id = ? AND revision = ?`
	args := []any{record.SKU, record.Name, record.NameKey, record.Description, record.PriceCents, record.Status, record.UpdatedAt, record.ID, expectedRevision}
	if scope == ScopeSelf {
		query, args = query+` AND owner_account_id = ?`, append(args, actorID)
	}
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 1 {
		return nil
	}
	var revision int64
	var owner string
	err = tx.QueryRowContext(ctx, `SELECT revision, owner_account_id FROM demo_products WHERE id = ?`, record.ID).Scan(&revision, &owner)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if scope == ScopeSelf && owner != actorID {
		return ErrDenied
	}
	return ErrConflict
}

func (r repository) remove(ctx context.Context, tx database.Tx, target DeleteTarget, actorID string, scope Scope) error {
	if !validScope(scope) {
		return ErrDenied
	}
	query := `DELETE FROM demo_products WHERE id = ? AND revision = ?`
	args := []any{target.ID, target.Revision}
	if scope == ScopeSelf {
		query, args = query+` AND owner_account_id = ?`, append(args, actorID)
	}
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 1 {
		return nil
	}
	var revision int64
	var owner string
	err = tx.QueryRowContext(ctx, `SELECT revision, owner_account_id FROM demo_products WHERE id = ?`, target.ID).Scan(&revision, &owner)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if scope == ScopeSelf && owner != actorID {
		return ErrDenied
	}
	return ErrConflict
}

func validScope(scope Scope) bool { return scope == ScopeSelf || scope == ScopeAll }

func mapProduct(value productRecord) Product {
	return Product{ID: value.ID, SKU: value.SKU, Name: value.Name, Description: value.Description, PriceCents: value.PriceCents,
		Status: value.Status, Revision: value.Revision, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}
