// Package authorization owns the database-backed Permission Code decision boundary.
package authorization

import (
	"context"
	"database/sql"
	"errors"

	"go-admin/internal/platform/database"
)

const (
	PermissionUsersRead          = "iam.users.read"
	PermissionUsersWrite         = "iam.users.write"
	PermissionUsersDelete        = "iam.users.delete"
	PermissionUsersResetPassword = "iam.users.reset-password"
	PermissionRolesRead          = "iam.roles.read"
	PermissionRolesWrite         = "iam.roles.write"
	PermissionRolesDelete        = "iam.roles.delete"
	PermissionRolesAssign        = "iam.roles.assign"
	PermissionMenusRead          = "iam.menus.read"
	PermissionMenusWrite         = "iam.menus.write"
	PermissionMenusDelete        = "iam.menus.delete"
	PermissionPermissionsRead    = "iam.permissions.read"
	PermissionManifestRead       = "iam.manifest.read"
)

var (
	ErrDenied   = errors.New("authorization denied")
	ErrInternal = errors.New("authorization decision failed")
)

type Scope string

const (
	ScopeSelf Scope = "self"
	ScopeAll  Scope = "all"
)

type Decision struct{ Scope Scope }

type Menu struct {
	Key, Label, Path, PermissionCode string
	SortOrder                        int
}

type Manifest struct {
	Permissions []string
	Menus       []Menu
	Scope       Scope
}

type Database interface {
	WithinTx(context.Context, func(context.Context, database.Tx) error) error
	Dialect() database.Dialect
}

type Service struct {
	db      Database
	dialect database.Dialect
}

func NewService(db Database) *Service {
	service := &Service{db: db}
	if db != nil {
		service.dialect = db.Dialect()
	}
	return service
}

// Require deliberately reaches the database for every decision. Correctness never depends on a cache.
func (s *Service) Require(ctx context.Context, accountID, permission string) (Decision, error) {
	if s == nil || s.db == nil || accountID == "" || permission == "" {
		return Decision{}, ErrDenied
	}
	var decision Decision
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		var err error
		decision, err = s.RequireInTx(ctx, tx, accountID, permission)
		return err
	})
	if errors.Is(err, ErrDenied) {
		return Decision{}, ErrDenied
	}
	if err != nil {
		return Decision{}, sanitize(ctx, err)
	}
	return decision, nil
}

// RequireInTx is the final application-use-case boundary. PostgreSQL locks every contributing
// row so a concurrent revoke cannot commit between this decision and its protected mutation.
func (s *Service) RequireInTx(ctx context.Context, tx database.Tx, accountID, permission string) (Decision, error) {
	if s == nil || tx == nil || accountID == "" || permission == "" {
		return Decision{}, ErrDenied
	}
	query := `SELECT scope_role.data_scope
		FROM iam_accounts a
		JOIN iam_account_roles permission_ar ON permission_ar.account_id = a.id
		JOIN iam_roles permission_role ON permission_role.id = permission_ar.role_id AND permission_role.enabled = ?
		JOIN iam_role_permissions rp ON rp.role_id = permission_role.id AND rp.permission_code = ?
		JOIN iam_account_roles scope_ar ON scope_ar.account_id = a.id
		JOIN iam_roles scope_role ON scope_role.id = scope_ar.role_id AND scope_role.enabled = ?
		WHERE a.id = ? AND a.disabled_at IS NULL AND rp.permission_code = ?`
	if s.dialect == database.DialectPostgres {
		query += ` FOR SHARE OF a, permission_ar, permission_role, rp, scope_ar, scope_role`
	}
	rows, err := tx.QueryContext(ctx, query, true, permission, true, accountID, permission)
	if err != nil {
		return Decision{}, err
	}
	defer rows.Close()
	decision := Decision{Scope: ScopeSelf}
	found := false
	for rows.Next() {
		var scope string
		if err := rows.Scan(&scope); err != nil {
			return Decision{}, err
		}
		found = true
		if Scope(scope) == ScopeAll {
			decision.Scope = ScopeAll
		}
	}
	if err := rows.Err(); err != nil {
		return Decision{}, err
	}
	if !found {
		return Decision{}, ErrDenied
	}
	return decision, nil
}

func (s *Service) Manifest(ctx context.Context, accountID string) (Manifest, error) {
	if s == nil || s.db == nil || accountID == "" {
		return Manifest{}, ErrDenied
	}
	manifest := Manifest{Scope: ScopeSelf}
	manifestAllowed := false
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		rows, err := tx.QueryContext(ctx, `WITH enabled_roles AS (
				SELECT r.id, r.data_scope
				FROM iam_accounts a
				JOIN iam_account_roles ar ON ar.account_id = a.id
				JOIN iam_roles r ON r.id = ar.role_id AND r.enabled = ?
				WHERE a.id = ? AND a.disabled_at IS NULL
			), permissions AS (
				SELECT DISTINCT rp.permission_code
				FROM enabled_roles er
				JOIN iam_role_permissions rp ON rp.role_id = er.id
			), scope_projection AS (
				SELECT CASE WHEN SUM(CASE WHEN data_scope = 'all' THEN 1 ELSE 0 END) > 0 THEN 'all' ELSE 'self' END AS data_scope
				FROM enabled_roles
			), menus AS (
				SELECT DISTINCT m.menu_key, m.label, m.path, m.permission_code, m.sort_order
				FROM enabled_roles er
				JOIN iam_role_menus rm ON rm.role_id = er.id
				JOIN iam_menus m ON m.id = rm.menu_id
				WHERE EXISTS (SELECT 1 FROM permissions p WHERE p.permission_code = m.permission_code)
			)
			SELECT 0 AS row_kind, p.permission_code, s.data_scope,
				CAST(NULL AS TEXT) AS menu_key, CAST(NULL AS TEXT) AS label, CAST(NULL AS TEXT) AS path, CAST(NULL AS INTEGER) AS sort_order
			FROM permissions p CROSS JOIN scope_projection s
			UNION ALL
			SELECT 1 AS row_kind, m.permission_code, s.data_scope, m.menu_key, m.label, m.path, m.sort_order
			FROM menus m CROSS JOIN scope_projection s
			ORDER BY row_kind, sort_order, permission_code, menu_key`, true, accountID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var kind int
			var permission, scope string
			var key, label, path sql.NullString
			var sortOrder sql.NullInt64
			if err := rows.Scan(&kind, &permission, &scope, &key, &label, &path, &sortOrder); err != nil {
				return err
			}
			if Scope(scope) == ScopeAll {
				manifest.Scope = ScopeAll
			}
			if kind == 0 {
				manifest.Permissions = append(manifest.Permissions, permission)
				manifestAllowed = manifestAllowed || permission == PermissionManifestRead
				continue
			}
			manifest.Menus = append(manifest.Menus, Menu{Key: key.String, Label: label.String, Path: path.String, PermissionCode: permission, SortOrder: int(sortOrder.Int64)})
		}
		return rows.Err()
	})
	if err != nil {
		return Manifest{}, sanitize(ctx, err)
	}
	if !manifestAllowed {
		return Manifest{}, ErrDenied
	}
	return manifest, nil
}

func sanitize(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}
	return ErrInternal
}
