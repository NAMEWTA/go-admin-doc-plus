package authorization

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

const (
	ScopeOrganization     Scope = "organization"
	ScopeOrganizationTree Scope = "organization-tree"
	ScopeCustom           Scope = "custom"
)

type OrganizationProjection interface {
	DepartmentSetInTx(context.Context, database.Tx, string, bool) ([]string, error)
	ValidateDepartmentIDsInTx(context.Context, database.Tx, []string) error
}

type ResolvedScope struct {
	All           bool
	Self          bool
	DepartmentIDs []string
}

func (scope ResolvedScope) Allows(actorID, ownerAccountID, departmentID string) bool {
	if scope.All || scope.Self && actorID != "" && actorID == ownerAccountID {
		return true
	}
	position := sort.SearchStrings(scope.DepartmentIDs, departmentID)
	return departmentID != "" && position < len(scope.DepartmentIDs) && scope.DepartmentIDs[position] == departmentID
}

type ScopeResolver struct {
	db         Database
	dialect    database.Dialect
	projection OrganizationProjection
}

func NewScopeResolver(db Database, projection OrganizationProjection) (*ScopeResolver, error) {
	if db == nil || projection == nil {
		return nil, errors.New("authorization scope dependencies are required")
	}
	return &ScopeResolver{db: db, dialect: db.Dialect(), projection: projection}, nil
}

func (resolver *ScopeResolver) Resolve(ctx context.Context, accountID, permission string) (ResolvedScope, error) {
	if resolver == nil || resolver.db == nil || accountID == "" || permission == "" {
		return ResolvedScope{}, ErrDenied
	}
	var result ResolvedScope
	err := resolver.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		var err error
		result, err = resolver.ResolveInTx(ctx, tx, accountID, permission)
		return err
	})
	if errors.Is(err, ErrDenied) {
		return ResolvedScope{}, ErrDenied
	}
	if err != nil {
		return ResolvedScope{}, sanitize(ctx, err)
	}
	return result, nil
}

func (resolver *ScopeResolver) ResolveInTx(ctx context.Context, tx database.Tx, accountID, permission string) (ResolvedScope, error) {
	if resolver == nil || tx == nil || accountID == "" || permission == "" {
		return ResolvedScope{}, ErrDenied
	}
	query := `SELECT r.id, ds.scope
		FROM iam_accounts a
		JOIN iam_account_roles ar ON ar.account_id = a.id
		JOIN iam_roles r ON r.id = ar.role_id AND r.enabled = ?
		JOIN iam_role_permissions rp ON rp.role_id = r.id AND rp.permission_code = ?
		JOIN iam_role_data_scopes ds ON ds.role_id = r.id
		WHERE a.id = ? AND a.disabled_at IS NULL`
	if resolver.dialect == database.DialectPostgres {
		query += ` FOR SHARE OF a, ar, r, rp, ds`
	}
	rows, err := tx.QueryContext(ctx, query, true, permission, accountID)
	if err != nil {
		return ResolvedScope{}, err
	}
	type grant struct {
		roleID string
		scope  Scope
	}
	grants := []grant{}
	for rows.Next() {
		var value grant
		if err := rows.Scan(&value.roleID, &value.scope); err != nil {
			_ = rows.Close()
			return ResolvedScope{}, err
		}
		grants = append(grants, value)
	}
	if err := rows.Close(); err != nil {
		return ResolvedScope{}, err
	}
	if err := rows.Err(); err != nil {
		return ResolvedScope{}, err
	}
	if len(grants) == 0 {
		return ResolvedScope{}, ErrDenied
	}

	result := ResolvedScope{}
	departments := map[string]struct{}{}
	primaryLoaded, primaryValid, primaryDepartment := false, false, ""
	loadPrimary := func() error {
		if primaryLoaded {
			return nil
		}
		primaryLoaded = true
		var value sql.NullString
		err := tx.QueryRowContext(ctx, `SELECT primary_department_id FROM iam_account_organization WHERE account_id = ?`, accountID).Scan(&value)
		if errors.Is(err, sql.ErrNoRows) || err == nil && !value.Valid {
			return nil
		}
		if err != nil {
			return err
		}
		primaryValid, primaryDepartment = true, value.String
		return nil
	}
	for _, value := range grants {
		switch value.scope {
		case ScopeAll:
			result.All = true
		case ScopeSelf:
			result.Self = true
		case ScopeOrganization, ScopeOrganizationTree:
			if err := loadPrimary(); err != nil {
				return ResolvedScope{}, err
			}
			if !primaryValid {
				continue
			}
			ids, err := resolver.projection.DepartmentSetInTx(ctx, tx, primaryDepartment, value.scope == ScopeOrganizationTree)
			if err != nil {
				return ResolvedScope{}, ErrDenied
			}
			for _, id := range ids {
				departments[id] = struct{}{}
			}
		case ScopeCustom:
			custom, err := customDepartments(ctx, tx, value.roleID)
			if err != nil {
				return ResolvedScope{}, err
			}
			if err := resolver.projection.ValidateDepartmentIDsInTx(ctx, tx, custom); err != nil {
				return ResolvedScope{}, ErrDenied
			}
			for _, id := range custom {
				departments[id] = struct{}{}
			}
		default:
			return ResolvedScope{}, ErrDenied
		}
	}
	result.DepartmentIDs = make([]string, 0, len(departments))
	for id := range departments {
		result.DepartmentIDs = append(result.DepartmentIDs, id)
	}
	sort.Strings(result.DepartmentIDs)
	return result, nil
}

func customDepartments(ctx context.Context, tx database.Tx, roleID string) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT department_id FROM iam_role_data_scope_departments WHERE role_id = ? ORDER BY department_id`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, rows.Err()
}
