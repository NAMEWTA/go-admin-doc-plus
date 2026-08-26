// Package authorization owns the database-backed Permission Code decision boundary.
package authorization

import (
	"context"
	"errors"
	"sort"

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
	query := `SELECT r.data_scope
		FROM iam_accounts a
		JOIN iam_account_roles ar ON ar.account_id = a.id
		JOIN iam_roles r ON r.id = ar.role_id AND r.enabled = ?
		JOIN iam_role_permissions rp ON rp.role_id = r.id
		WHERE a.id = ? AND a.disabled_at IS NULL AND rp.permission_code = ?`
	if s.dialect == database.DialectPostgres {
		query += ` FOR SHARE OF a, ar, r, rp`
	}
	rows, err := tx.QueryContext(ctx, query, true, accountID, permission)
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
		rows, err := tx.QueryContext(ctx, `SELECT DISTINCT rp.permission_code, r.data_scope
			FROM iam_accounts a
			JOIN iam_account_roles ar ON ar.account_id = a.id
			JOIN iam_roles r ON r.id = ar.role_id AND r.enabled = ?
			JOIN iam_role_permissions rp ON rp.role_id = r.id
			WHERE a.id = ? AND a.disabled_at IS NULL`, true, accountID)
		if err != nil {
			return err
		}
		permissions := map[string]struct{}{}
		for rows.Next() {
			var permission, scope string
			if err := rows.Scan(&permission, &scope); err != nil {
				_ = rows.Close()
				return err
			}
			permissions[permission] = struct{}{}
			if Scope(scope) == ScopeAll {
				manifest.Scope = ScopeAll
			}
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}
		for permission := range permissions {
			manifest.Permissions = append(manifest.Permissions, permission)
		}
		_, manifestAllowed = permissions[PermissionManifestRead]
		sort.Strings(manifest.Permissions)

		menuRows, err := tx.QueryContext(ctx, `SELECT DISTINCT m.menu_key, m.label, m.path, m.permission_code, m.sort_order
			FROM iam_accounts a
			JOIN iam_account_roles ar ON ar.account_id = a.id
			JOIN iam_roles r ON r.id = ar.role_id AND r.enabled = ?
			JOIN iam_role_menus rm ON rm.role_id = r.id
			JOIN iam_menus m ON m.id = rm.menu_id
			WHERE a.id = ? AND a.disabled_at IS NULL
			AND EXISTS (
				SELECT 1 FROM iam_account_roles permission_ar
				JOIN iam_roles permission_role ON permission_role.id = permission_ar.role_id AND permission_role.enabled = ?
				JOIN iam_role_permissions permission_rp ON permission_rp.role_id = permission_role.id
				WHERE permission_ar.account_id = a.id AND permission_rp.permission_code = m.permission_code
			)
			ORDER BY m.sort_order, m.menu_key`, true, accountID, true)
		if err != nil {
			return err
		}
		defer menuRows.Close()
		for menuRows.Next() {
			var menu Menu
			if err := menuRows.Scan(&menu.Key, &menu.Label, &menu.Path, &menu.PermissionCode, &menu.SortOrder); err != nil {
				return err
			}
			manifest.Menus = append(manifest.Menus, menu)
		}
		return menuRows.Err()
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
