package administration

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"time"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

type AccountOrganization struct {
	PrimaryDepartmentID *string
	PositionIDs         []string
}

type RoleDataScope struct {
	Scope         authorization.Scope
	DepartmentIDs []string
}

type DataScopeProjection interface {
	DepartmentSetInTx(context.Context, database.Tx, string, bool) ([]string, error)
	PositionDepartmentInTx(context.Context, database.Tx, string) (string, error)
	ValidateDepartmentIDsInTx(context.Context, database.Tx, []string) error
}

type DataScopeService struct {
	db         Database
	authorizer *authorization.Service
	projection DataScopeProjection
	now        func() time.Time
}

type DataScopeOption func(*DataScopeService)

func WithDataScopeClock(clock func() time.Time) DataScopeOption {
	return func(service *DataScopeService) { service.now = clock }
}

func NewDataScopeService(db Database, projection DataScopeProjection, options ...DataScopeOption) (*DataScopeService, error) {
	if db == nil || projection == nil {
		return nil, errors.New("iam data scope dependencies are required")
	}
	service := &DataScopeService{db: db, authorizer: authorization.NewService(db), projection: projection, now: time.Now}
	for _, option := range options {
		option(service)
	}
	if service.now == nil {
		return nil, errors.New("iam data scope clock is required")
	}
	return service, nil
}

func (s *DataScopeService) SetAccountOrganization(ctx context.Context, actorID, accountID string, value AccountOrganization) error {
	if s == nil || accountID == "" || duplicate(value.PositionIDs) {
		return ErrValidation
	}
	positionIDs := append([]string(nil), value.PositionIDs...)
	sort.Strings(positionIDs)
	primary := ""
	if value.PrimaryDepartmentID != nil {
		primary = *value.PrimaryDepartmentID
		if primary == "" {
			return ErrValidation
		}
	}
	if primary == "" && len(positionIDs) > 0 {
		return ErrConflict
	}
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		decision, err := s.authorizer.RequireInTx(ctx, tx, actorID, authorization.PermissionRolesAssign)
		if err != nil {
			return err
		}
		if decision.Scope != authorization.ScopeAll {
			return ErrDenied
		}
		if err := requireDataScopeAccount(ctx, tx, s.db.Dialect(), accountID); err != nil {
			return err
		}
		if primary != "" {
			if _, err := s.projection.DepartmentSetInTx(ctx, tx, primary, false); err != nil {
				return ErrConflict
			}
		}
		for _, positionID := range positionIDs {
			departmentID, err := s.projection.PositionDepartmentInTx(ctx, tx, positionID)
			if err != nil || departmentID != primary {
				return ErrConflict
			}
		}
		now := s.now().UTC()
		var primaryValue any
		if primary != "" {
			primaryValue = primary
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO iam_account_organization(account_id, primary_department_id, updated_at)
			VALUES (?, ?, ?) ON CONFLICT(account_id) DO UPDATE SET primary_department_id = excluded.primary_department_id, updated_at = excluded.updated_at`, accountID, primaryValue, now); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM iam_account_positions WHERE account_id = ?`, accountID); err != nil {
			return err
		}
		for _, positionID := range positionIDs {
			if _, err := tx.ExecContext(ctx, `INSERT INTO iam_account_positions(account_id, position_id) VALUES (?, ?)`, accountID, positionID); err != nil {
				return err
			}
		}
		return nil
	})
	return normalizeDataScope(ctx, s.db.Dialect(), err)
}

func (s *DataScopeService) SetRoleDataScope(ctx context.Context, actorID, roleID string, value RoleDataScope) error {
	if s == nil || roleID == "" || !validDataScope(value.Scope) || duplicate(value.DepartmentIDs) {
		return ErrValidation
	}
	departmentIDs := append([]string(nil), value.DepartmentIDs...)
	sort.Strings(departmentIDs)
	if value.Scope != authorization.ScopeCustom && len(departmentIDs) > 0 {
		return ErrValidation
	}
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		decision, err := s.authorizer.RequireInTx(ctx, tx, actorID, authorization.PermissionRolesWrite)
		if err != nil {
			return err
		}
		if decision.Scope != authorization.ScopeAll {
			return ErrDenied
		}
		query := `SELECT protected FROM iam_roles WHERE id = ?`
		if s.db.Dialect() == database.DialectPostgres {
			query += ` FOR UPDATE`
		}
		var protected bool
		if err := tx.QueryRowContext(ctx, query, roleID).Scan(&protected); err != nil {
			return err
		}
		var current string
		err = tx.QueryRowContext(ctx, `SELECT scope FROM iam_role_data_scopes WHERE role_id = ?`, roleID).Scan(&current)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if protected && (errors.Is(err, sql.ErrNoRows) || authorization.Scope(current) != value.Scope) {
			return ErrConflict
		}
		if value.Scope == authorization.ScopeCustom && len(departmentIDs) > 0 {
			if err := s.projection.ValidateDepartmentIDsInTx(ctx, tx, departmentIDs); err != nil {
				return ErrConflict
			}
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO iam_role_data_scopes(role_id, scope, updated_at) VALUES (?, ?, ?)
			ON CONFLICT(role_id) DO UPDATE SET scope = excluded.scope, updated_at = excluded.updated_at`, roleID, value.Scope, s.now().UTC()); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM iam_role_data_scope_departments WHERE role_id = ?`, roleID); err != nil {
			return err
		}
		for _, departmentID := range departmentIDs {
			if _, err := tx.ExecContext(ctx, `INSERT INTO iam_role_data_scope_departments(role_id, department_id) VALUES (?, ?)`, roleID, departmentID); err != nil {
				return err
			}
		}
		return nil
	})
	return normalizeDataScope(ctx, s.db.Dialect(), err)
}

func requireDataScopeAccount(ctx context.Context, tx database.Tx, dialect database.Dialect, accountID string) error {
	query := `SELECT id FROM iam_accounts WHERE id = ?`
	if dialect == database.DialectPostgres {
		query += ` FOR UPDATE`
	}
	var id string
	return tx.QueryRowContext(ctx, query, accountID).Scan(&id)
}

func normalizeDataScope(ctx context.Context, dialect database.Dialect, err error) error {
	if err == nil {
		return nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	for _, stable := range []error{context.Canceled, context.DeadlineExceeded, ErrDenied, ErrValidation, ErrConflict} {
		if errors.Is(err, stable) {
			return stable
		}
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if constraintConflict(dialect, err) || sqliteRestrictConflict(dialect, err) {
		return ErrConflict
	}
	return ErrInternal
}

func sqliteRestrictConflict(dialect database.Dialect, err error) bool {
	if dialect != database.DialectSQLite {
		return false
	}
	var coded interface{ Code() int }
	return errors.As(err, &coded) && coded.Code() == 1811
}

func validDataScope(value authorization.Scope) bool {
	switch value {
	case authorization.ScopeAll, authorization.ScopeSelf, authorization.ScopeOrganization, authorization.ScopeOrganizationTree, authorization.ScopeCustom:
		return true
	default:
		return false
	}
}
