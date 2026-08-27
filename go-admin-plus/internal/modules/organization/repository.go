package organization

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"go-admin/internal/platform/database"
)

type repository struct{ dialect database.Dialect }

func (r repository) departments(ctx context.Context, tx database.Tx, lock bool) ([]Department, error) {
	query := `SELECT id, department_key, name, parent_id, sort_order, protected FROM organization_departments ORDER BY id`
	if lock && r.dialect == database.DialectPostgres {
		query += ` FOR UPDATE`
	}
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Department{}
	for rows.Next() {
		var value Department
		var parent sql.NullString
		if err := rows.Scan(&value.ID, &value.Key, &value.Name, &parent, &value.SortOrder, &value.Protected); err != nil {
			return nil, err
		}
		if parent.Valid {
			value.ParentID = &parent.String
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (repository) insertDepartment(ctx context.Context, tx database.Tx, value Department, now time.Time) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO organization_departments(id, department_key, name, parent_id, sort_order, protected, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, value.ID, value.Key, value.Name, value.ParentID, value.SortOrder, false, now, now)
	return err
}

func (repository) updateDepartment(ctx context.Context, tx database.Tx, value Department, now time.Time) error {
	result, err := tx.ExecContext(ctx, `UPDATE organization_departments SET department_key = ?, name = ?, parent_id = ?, sort_order = ?, updated_at = ? WHERE id = ?`, value.Key, value.Name, value.ParentID, value.SortOrder, now, value.ID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrNotFound
	}
	return nil
}

func (repository) departmentReferences(ctx context.Context, tx database.Tx, id string) (int, error) {
	var count int
	err := tx.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM organization_departments WHERE parent_id = ?) +
		(SELECT COUNT(*) FROM organization_positions WHERE department_id = ?)`, id, id).Scan(&count)
	return count, err
}

func (repository) deleteDepartment(ctx context.Context, tx database.Tx, id string) error {
	result, err := tx.ExecContext(ctx, `DELETE FROM organization_departments WHERE id = ?`, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrNotFound
	}
	return nil
}

func (r repository) positions(ctx context.Context, tx database.Tx, search string, page, pageSize int) (PositionPage, error) {
	where, arguments := ``, []any{}
	if search != "" {
		where = ` WHERE instr(lower(position_key), ?) > 0 OR instr(name_key, ?) > 0`
		if r.dialect == database.DialectPostgres {
			where = ` WHERE strpos(lower(position_key), ?) > 0 OR strpos(name_key, ?) > 0`
		}
		arguments = append(arguments, strings.ToLower(search), normalizedNameKey(search))
	}
	var result PositionPage
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM organization_positions`+where, arguments...).Scan(&result.Total); err != nil {
		return PositionPage{}, err
	}
	arguments = append(arguments, pageSize, (page-1)*pageSize)
	order := `position_key COLLATE BINARY`
	if r.dialect == database.DialectPostgres {
		order = `position_key COLLATE "C"`
	}
	rows, err := tx.QueryContext(ctx, `SELECT id, position_key, name, department_id, enabled, protected FROM organization_positions`+where+` ORDER BY `+order+` LIMIT ? OFFSET ?`, arguments...)
	if err != nil {
		return PositionPage{}, err
	}
	defer rows.Close()
	result.Rows = []Position{}
	for rows.Next() {
		var value Position
		if err := rows.Scan(&value.ID, &value.Key, &value.Name, &value.DepartmentID, &value.Enabled, &value.Protected); err != nil {
			return PositionPage{}, err
		}
		result.Rows = append(result.Rows, value)
	}
	return result, rows.Err()
}

func (r repository) position(ctx context.Context, tx database.Tx, id string, lock bool) (Position, error) {
	query := `SELECT id, position_key, name, department_id, enabled, protected FROM organization_positions WHERE id = ?`
	if lock && r.dialect == database.DialectPostgres {
		query += ` FOR UPDATE`
	}
	var value Position
	err := tx.QueryRowContext(ctx, query, id).Scan(&value.ID, &value.Key, &value.Name, &value.DepartmentID, &value.Enabled, &value.Protected)
	return value, err
}

func (repository) insertPosition(ctx context.Context, tx database.Tx, value Position, now time.Time) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO organization_positions(id, position_key, name, name_key, department_id, enabled, protected, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, value.ID, value.Key, value.Name, normalizedNameKey(value.Name), value.DepartmentID, value.Enabled, false, now, now)
	return err
}

func (repository) updatePosition(ctx context.Context, tx database.Tx, value Position, now time.Time) error {
	result, err := tx.ExecContext(ctx, `UPDATE organization_positions SET position_key = ?, name = ?, name_key = ?, department_id = ?, enabled = ?, updated_at = ? WHERE id = ?`, value.Key, value.Name, normalizedNameKey(value.Name), value.DepartmentID, value.Enabled, now, value.ID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrNotFound
	}
	return nil
}

func (repository) deletePosition(ctx context.Context, tx database.Tx, id string) error {
	result, err := tx.ExecContext(ctx, `DELETE FROM organization_positions WHERE id = ?`, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrNotFound
	}
	return nil
}
