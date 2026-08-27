package generator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/uptrace/bun"
	"go-admin/internal/platform/database"
)

var databaseIdentifierPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type MetadataDatabase interface {
	Dialect() database.Dialect
	Bun() *bun.DB
}

type MetadataAllowlist struct {
	CurrentSchema string
	Tables        []string
}

type MetadataSource interface {
	Tables(context.Context) ([]TableRef, error)
	Describe(context.Context, TableRef) (Table, error)
}

type SQLMetadataSource struct {
	db      MetadataDatabase
	allowed map[TableRef]struct{}
}

func NewSQLMetadataSource(ctx context.Context, db MetadataDatabase, allowlist MetadataAllowlist) (*SQLMetadataSource, error) {
	if db == nil || db.Bun() == nil || (db.Dialect() != database.DialectPostgres && db.Dialect() != database.DialectSQLite) {
		return nil, ErrInvalid
	}
	if !validMetadataSchema(db.Dialect(), allowlist.CurrentSchema) || len(allowlist.Tables) == 0 {
		return nil, ErrInvalid
	}
	var currentSchema string
	if db.Dialect() == database.DialectSQLite {
		currentSchema = "main"
	} else if err := db.Bun().QueryRowContext(ctx, `SELECT current_schema()`).Scan(&currentSchema); err != nil {
		return nil, sanitizeMetadataError(ctx, err)
	}
	if currentSchema != allowlist.CurrentSchema {
		return nil, ErrDenied
	}
	allowed := make(map[TableRef]struct{})
	for _, table := range allowlist.Tables {
		ref := TableRef{Schema: currentSchema, Name: table}
		if !databaseIdentifierPattern.MatchString(table) {
			return nil, ErrInvalid
		}
		if _, duplicate := allowed[ref]; duplicate {
			return nil, ErrInvalid
		}
		allowed[ref] = struct{}{}
	}
	if len(allowed) == 0 {
		return nil, ErrInvalid
	}
	return &SQLMetadataSource{db: db, allowed: allowed}, nil
}

func (source *SQLMetadataSource) Tables(ctx context.Context) ([]TableRef, error) {
	refs := make([]TableRef, 0, len(source.allowed))
	for ref := range source.allowed {
		exists, err := source.tableExists(ctx, ref)
		if err != nil {
			return nil, sanitizeMetadataError(ctx, err)
		}
		if exists {
			refs = append(refs, ref)
		}
	}
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Schema != refs[j].Schema {
			return refs[i].Schema < refs[j].Schema
		}
		return refs[i].Name < refs[j].Name
	})
	return refs, nil
}

func (source *SQLMetadataSource) Describe(ctx context.Context, ref TableRef) (Table, error) {
	if _, allowed := source.allowed[ref]; !allowed || !validMetadataSchema(source.db.Dialect(), ref.Schema) || !databaseIdentifierPattern.MatchString(ref.Name) {
		return Table{}, ErrDenied
	}
	var columns []Column
	var err error
	if source.db.Dialect() == database.DialectPostgres {
		columns, err = source.postgresColumns(ctx, ref)
	} else {
		columns, err = source.sqliteColumns(ctx, ref)
	}
	if err != nil {
		return Table{}, sanitizeMetadataError(ctx, err)
	}
	if len(columns) == 0 {
		return Table{}, ErrNotFound
	}
	if len(columns) > 100 {
		return Table{}, ErrInvalid
	}
	sort.Slice(columns, func(i, j int) bool { return columns[i].Ordinal < columns[j].Ordinal })
	return Table{Ref: ref, Columns: columns}, nil
}

func (source *SQLMetadataSource) tableExists(ctx context.Context, ref TableRef) (bool, error) {
	var count int
	if source.db.Dialect() == database.DialectPostgres {
		err := source.db.Bun().QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ? AND table_type = 'BASE TABLE'`, ref.Schema, ref.Name).Scan(&count)
		return count == 1, err
	}
	err := source.db.Bun().QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_schema WHERE type = 'table' AND name = ? AND name NOT LIKE 'sqlite_%'`, ref.Name).Scan(&count)
	return count == 1, err
}

func (source *SQLMetadataSource) postgresColumns(ctx context.Context, ref TableRef) ([]Column, error) {
	rows, err := source.db.Bun().QueryContext(ctx, `SELECT c.column_name, c.data_type, c.is_nullable = 'YES', c.ordinal_position,
		EXISTS (SELECT 1 FROM information_schema.table_constraints tc JOIN information_schema.key_column_usage kcu
		ON tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema AND tc.table_name = kcu.table_name
		WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_schema = c.table_schema AND tc.table_name = c.table_name AND kcu.column_name = c.column_name)
		FROM information_schema.columns c WHERE c.table_schema = ? AND c.table_name = ? ORDER BY c.ordinal_position`, ref.Schema, ref.Name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var columns []Column
	for rows.Next() {
		var column Column
		if err := rows.Scan(&column.Name, &column.DatabaseType, &column.Nullable, &column.Ordinal, &column.PrimaryKey); err != nil {
			return nil, err
		}
		column.Kind, err = postgresKind(column.DatabaseType)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	return columns, rows.Err()
}

func (source *SQLMetadataSource) sqliteColumns(ctx context.Context, ref TableRef) ([]Column, error) {
	rows, err := source.db.Bun().QueryContext(ctx, `PRAGMA table_xinfo(`+quoteIdentifier(ref.Name)+`)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var columns []Column
	for rows.Next() {
		var cid, notNull, primaryKey, hidden int
		var name, declaredType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &declaredType, &notNull, &defaultValue, &primaryKey, &hidden); err != nil {
			return nil, err
		}
		if hidden != 0 {
			continue
		}
		kind, kindErr := sqliteKind(declaredType)
		if kindErr != nil {
			return nil, kindErr
		}
		columns = append(columns, Column{Name: name, DatabaseType: strings.ToUpper(strings.TrimSpace(declaredType)), Kind: kind, Nullable: notNull == 0 && primaryKey == 0, PrimaryKey: primaryKey > 0, Ordinal: cid + 1})
	}
	return columns, rows.Err()
}

func validMetadataSchema(dialect database.Dialect, schema string) bool {
	if !databaseIdentifierPattern.MatchString(schema) {
		return false
	}
	if dialect == database.DialectSQLite {
		return schema == "main"
	}
	lower := strings.ToLower(schema)
	return lower != "information_schema" && lower != "pg_catalog" && !strings.HasPrefix(lower, "pg_")
}

func quoteIdentifier(value string) string { return `"` + strings.ReplaceAll(value, `"`, `""`) + `"` }

func postgresKind(value string) (ColumnKind, error) {
	switch strings.ToLower(value) {
	case "text", "character varying", "character", "citext", "json", "jsonb":
		return KindString, nil
	case "smallint", "integer", "bigint", "smallserial", "serial", "bigserial":
		return KindInt64, nil
	case "boolean":
		return KindBoolean, nil
	case "numeric", "decimal", "real", "double precision":
		return KindDecimal, nil
	case "timestamp without time zone", "timestamp with time zone", "date":
		return KindTime, nil
	case "uuid":
		return KindUUID, nil
	case "bytea":
		return KindBytes, nil
	default:
		return "", fmt.Errorf("%w: unsupported database type", ErrInvalid)
	}
}

func sqliteKind(value string) (ColumnKind, error) {
	upper := strings.ToUpper(strings.TrimSpace(value))
	switch {
	case strings.Contains(upper, "BOOL"):
		return KindBoolean, nil
	case strings.Contains(upper, "INT"):
		return KindInt64, nil
	case strings.Contains(upper, "REAL"), strings.Contains(upper, "FLOA"), strings.Contains(upper, "DOUB"), strings.Contains(upper, "NUM"), strings.Contains(upper, "DEC"):
		return KindDecimal, nil
	case strings.Contains(upper, "BLOB"):
		return KindBytes, nil
	case strings.Contains(upper, "DATE"), strings.Contains(upper, "TIME"):
		return KindTime, nil
	case strings.Contains(upper, "UUID"):
		return KindUUID, nil
	case upper == "" || strings.Contains(upper, "CHAR"), strings.Contains(upper, "CLOB"), strings.Contains(upper, "TEXT"), strings.Contains(upper, "JSON"):
		return KindString, nil
	default:
		return "", fmt.Errorf("%w: unsupported database type", ErrInvalid)
	}
}

func sanitizeMetadataError(ctx context.Context, err error) error {
	if ctxErr := context.Cause(ctx); ctxErr != nil {
		return ctxErr
	}
	if errors.Is(err, ErrDenied) || errors.Is(err, ErrInvalid) || errors.Is(err, ErrNotFound) {
		return err
	}
	return ErrInternal
}
