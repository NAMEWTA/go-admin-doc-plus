package files

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

const (
	statePending  = "pending"
	stateReady    = "ready"
	stateDeleting = "deleting"
)

type fileRecord struct {
	ID, OwnerAccountID, OriginalName, NameKey, MediaType, SHA256, StorageKey, State string
	TemporaryKey                                                                    *string
	SizeBytes, Revision                                                             int64
	ClaimToken                                                                      *string
	ClaimExpiresAt                                                                  *time.Time
	CreatedAt, UpdatedAt                                                            time.Time
}

type repository struct{ dialect database.Dialect }

func (repository) insert(ctx context.Context, tx database.Tx, record fileRecord) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO files_objects
		(id, owner_account_id, original_name, name_key, media_type, size_bytes, sha256, storage_key, temporary_key, state, revision, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, record.ID, record.OwnerAccountID, record.OriginalName, record.NameKey,
		record.MediaType, record.SizeBytes, record.SHA256, record.StorageKey, record.TemporaryKey, record.State, record.Revision, record.CreatedAt, record.UpdatedAt)
	return err
}

func (repository) markReady(ctx context.Context, tx database.Tx, id string, now time.Time) error {
	result, err := tx.ExecContext(ctx, `UPDATE files_objects SET state = ?, temporary_key = NULL, updated_at = ? WHERE id = ? AND state = ?`, stateReady, now, id, statePending)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrConflict
	}
	return nil
}

func (repository repository) list(ctx context.Context, tx database.Tx, actorID string, scope Scope, query ListQuery) (Page, error) {
	if !validScope(scope) {
		return Page{}, ErrDenied
	}
	clauses, args := []string{"state = ?"}, []any{stateReady}
	if scope == ScopeSelf {
		clauses, args = append(clauses, "owner_account_id = ?"), append(args, actorID)
	}
	if query.Search != "" {
		clause := "instr(name_key, ?) > 0"
		if repository.dialect == database.DialectPostgres {
			clause = "strpos(name_key, ?) > 0"
		}
		clauses, args = append(clauses, clause), append(args, nameKey(query.Search))
	}
	where := " WHERE " + strings.Join(clauses, " AND ")
	var page Page
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM files_objects`+where, args...).Scan(&page.Total); err != nil {
		return Page{}, err
	}
	columns := map[string]string{"name": "name_key", "sizeBytes": "size_bytes", "createdAt": "created_at"}
	if query.Sort == "name" {
		if repository.dialect == database.DialectPostgres {
			columns["name"] = `name_key COLLATE "C"`
		} else {
			columns["name"] = "name_key COLLATE BINARY"
		}
	}
	directions := map[string]string{"ascending": "ASC", "descending": "DESC"}
	statement := `SELECT id, original_name, media_type, size_bytes, sha256, revision, created_at, updated_at FROM files_objects` + where +
		` ORDER BY ` + columns[query.Sort] + ` ` + directions[query.Direction] + `, id ASC LIMIT ? OFFSET ?`
	args = append(args, query.PageSize, (query.Page-1)*query.PageSize)
	rows, err := tx.QueryContext(ctx, statement, args...)
	if err != nil {
		return Page{}, err
	}
	defer rows.Close()
	page.Rows = []Metadata{}
	for rows.Next() {
		var metadata Metadata
		if err := rows.Scan(&metadata.ID, &metadata.OriginalName, &metadata.MediaType, &metadata.SizeBytes, &metadata.SHA256,
			&metadata.Revision, &metadata.CreatedAt, &metadata.UpdatedAt); err != nil {
			return Page{}, err
		}
		page.Rows = append(page.Rows, metadata)
	}
	return page, rows.Err()
}

func (repository repository) ready(ctx context.Context, tx database.Tx, id, actorID string, scope Scope) (fileRecord, error) {
	if !validScope(scope) {
		return fileRecord{}, ErrDenied
	}
	query := `SELECT id, owner_account_id, original_name, name_key, media_type, size_bytes, sha256, storage_key, state, revision, created_at, updated_at
		FROM files_objects WHERE id = ? AND state = ?`
	args := []any{id, stateReady}
	if scope == ScopeSelf {
		query += ` AND owner_account_id = ?`
		args = append(args, actorID)
	}
	if repository.dialect == database.DialectPostgres {
		query += ` FOR SHARE`
	}
	var record fileRecord
	err := tx.QueryRowContext(ctx, query, args...).Scan(&record.ID, &record.OwnerAccountID, &record.OriginalName, &record.NameKey,
		&record.MediaType, &record.SizeBytes, &record.SHA256, &record.StorageKey, &record.State, &record.Revision, &record.CreatedAt, &record.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) && scope == ScopeSelf {
		var owner string
		ownerErr := tx.QueryRowContext(ctx, `SELECT owner_account_id FROM files_objects WHERE id = ? AND state = ?`, id, stateReady).Scan(&owner)
		if ownerErr == nil && owner != actorID {
			return fileRecord{}, ErrDenied
		}
	}
	return record, err
}

func (repository repository) markDeleting(ctx context.Context, tx database.Tx, targets []DeleteTarget, actorID string, scope Scope, now time.Time) ([]fileRecord, error) {
	if !validScope(scope) {
		return nil, ErrDenied
	}
	records := make([]fileRecord, 0, len(targets))
	for _, target := range targets {
		query := `SELECT id, owner_account_id, storage_key, state, revision FROM files_objects WHERE id = ?`
		if repository.dialect == database.DialectPostgres {
			query += ` FOR UPDATE`
		}
		var record fileRecord
		err := tx.QueryRowContext(ctx, query, target.ID).Scan(&record.ID, &record.OwnerAccountID, &record.StorageKey, &record.State, &record.Revision)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, err
		}
		if scope == ScopeSelf && record.OwnerAccountID != actorID {
			return nil, ErrDenied
		}
		if record.State != stateReady || record.Revision != target.Revision {
			return nil, ErrConflict
		}
		records = append(records, record)
	}
	for _, record := range records {
		result, err := tx.ExecContext(ctx, `UPDATE files_objects SET state = ?, revision = revision + 1, updated_at = ? WHERE id = ? AND state = ? AND revision = ?`,
			stateDeleting, now, record.ID, stateReady, record.Revision)
		if err != nil {
			return nil, err
		}
		count, err := result.RowsAffected()
		if err != nil || count != 1 {
			if err != nil {
				return nil, err
			}
			return nil, ErrConflict
		}
	}
	return records, nil
}

func (repository) removeDeleting(ctx context.Context, tx database.Tx, id string) error {
	result, err := tx.ExecContext(ctx, `DELETE FROM files_objects WHERE id = ? AND state = ?`, id, stateDeleting)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrConflict
	}
	return nil
}

func (repository repository) claimRecovery(ctx context.Context, tx database.Tx, now, expires time.Time, token string, limit int) ([]fileRecord, error) {
	query := `SELECT id, owner_account_id, original_name, name_key, media_type, size_bytes, sha256, storage_key, temporary_key, state, revision, created_at, updated_at
		FROM files_objects WHERE state <> ? AND (claim_expires_at IS NULL OR claim_expires_at < ?) ORDER BY updated_at ASC, id ASC LIMIT ?`
	if repository.dialect == database.DialectPostgres {
		query += ` FOR UPDATE SKIP LOCKED`
	}
	rows, err := tx.QueryContext(ctx, query, stateReady, now, limit)
	if err != nil {
		return nil, err
	}
	records := []fileRecord{}
	for rows.Next() {
		var record fileRecord
		if err := rows.Scan(&record.ID, &record.OwnerAccountID, &record.OriginalName, &record.NameKey, &record.MediaType, &record.SizeBytes,
			&record.SHA256, &record.StorageKey, &record.TemporaryKey, &record.State, &record.Revision, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for index := range records {
		result, err := tx.ExecContext(ctx, `UPDATE files_objects SET claim_token = ?, claim_expires_at = ? WHERE id = ? AND state <> ? AND (claim_expires_at IS NULL OR claim_expires_at < ?)`, token, expires, records[index].ID, stateReady, now)
		if err != nil {
			return nil, err
		}
		count, err := result.RowsAffected()
		if err != nil || count != 1 {
			if err != nil {
				return nil, err
			}
			return nil, ErrConflict
		}
		records[index].ClaimToken = &token
	}
	return records, nil
}

func (repository) finishClaimedReady(ctx context.Context, tx database.Tx, id, token string, now time.Time) error {
	result, err := tx.ExecContext(ctx, `UPDATE files_objects SET state = ?, temporary_key = NULL, claim_token = NULL, claim_expires_at = NULL, updated_at = ? WHERE id = ? AND state = ? AND claim_token = ?`, stateReady, now, id, statePending, token)
	return exactlyOne(result, err)
}

func (repository) removeClaimed(ctx context.Context, tx database.Tx, id, state, token string) error {
	result, err := tx.ExecContext(ctx, `DELETE FROM files_objects WHERE id = ? AND state = ? AND claim_token = ?`, id, state, token)
	return exactlyOne(result, err)
}

func exactlyOne(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrConflict
	}
	return nil
}

func metadata(record fileRecord) Metadata {
	return Metadata{ID: record.ID, OriginalName: record.OriginalName, MediaType: record.MediaType, SizeBytes: record.SizeBytes,
		SHA256: record.SHA256, Revision: record.Revision, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt}
}
