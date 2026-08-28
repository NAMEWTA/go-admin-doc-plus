package generator

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
	"github.com/google/uuid"
)

type ConfigDatabase interface {
	WithinTx(context.Context, func(context.Context, database.Tx) error) error
	Dialect() database.Dialect
}

type ConfigStore interface {
	Save(context.Context, string, Draft, Preview) error
	Get(context.Context, string, string) (Draft, string, error)
}

type SQLConfigStore struct {
	db  ConfigDatabase
	now func() time.Time
}

func NewSQLConfigStore(db ConfigDatabase) (*SQLConfigStore, error) {
	if db == nil || (db.Dialect() != database.DialectPostgres && db.Dialect() != database.DialectSQLite) {
		return nil, ErrInvalid
	}
	return &SQLConfigStore{db: db, now: time.Now}, nil
}

func (store *SQLConfigStore) Save(ctx context.Context, actorID string, draft Draft, preview Preview) error {
	if actorID == "" || preview.Module != draft.Module || preview.Digest == "" {
		return ErrInvalid
	}
	encoded, err := json.Marshal(draft)
	if err != nil {
		return ErrInvalid
	}
	now := store.now().UTC()
	err = store.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		result, err := tx.ExecContext(ctx, configSaveStatement(store.db.Dialect()), uuid.NewString(), actorID, draft.Module, draft.Table.Schema, draft.Table.Name, string(encoded), preview.Digest, now, now)
		if err != nil {
			return err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if count != 1 {
			return ErrDenied
		}
		return nil
	})
	if err == nil || errors.Is(err, ErrDenied) || errors.Is(err, ErrInvalid) {
		return err
	}
	if context.Cause(ctx) != nil {
		return context.Cause(ctx)
	}
	return ErrInternal
}

func configSaveStatement(dialect database.Dialect) string {
	_ = dialect
	return `INSERT INTO generator_configs
			(id, actor_account_id, module_name, source_schema, source_table, normalized_config, preview_digest, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(module_name) DO UPDATE SET source_schema = excluded.source_schema, source_table = excluded.source_table,
			normalized_config = excluded.normalized_config, preview_digest = excluded.preview_digest, updated_at = excluded.updated_at
			WHERE generator_configs.actor_account_id = excluded.actor_account_id`
}

func (store *SQLConfigStore) Get(ctx context.Context, actorID, module string) (Draft, string, error) {
	if actorID == "" || !modulePattern.MatchString(module) {
		return Draft{}, "", ErrInvalid
	}
	var encoded, digest string
	err := store.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		query := `SELECT normalized_config, preview_digest FROM generator_configs WHERE module_name = ? AND actor_account_id = ?`
		return tx.QueryRowContext(ctx, query, module, actorID).Scan(&encoded, &digest)
	})
	if err != nil {
		if context.Cause(ctx) != nil {
			return Draft{}, "", context.Cause(ctx)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return Draft{}, "", ErrNotFound
		}
		return Draft{}, "", ErrInternal
	}
	var draft Draft
	if json.Unmarshal([]byte(encoded), &draft) != nil {
		return Draft{}, "", ErrInternal
	}
	return draft, digest, nil
}
