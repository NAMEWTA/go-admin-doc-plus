package organization

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

const (
	PermissionDepartmentsRead   = "organization.departments.read"
	PermissionDepartmentsWrite  = "organization.departments.write"
	PermissionDepartmentsDelete = "organization.departments.delete"
	PermissionPositionsRead     = "organization.positions.read"
	PermissionPositionsWrite    = "organization.positions.write"
	PermissionPositionsDelete   = "organization.positions.delete"
	maximumPositionPage         = 1_000_000
)

var (
	ErrDenied     = authorization.ErrDenied
	ErrNotFound   = errors.New("organization resource not found")
	ErrValidation = errors.New("organization request invalid")
	ErrConflict   = errors.New("organization resource conflict")
	ErrInternal   = errors.New("organization operation failed")
	stableKey     = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
)

type Database interface {
	WithinTx(context.Context, func(context.Context, database.Tx) error) error
	Dialect() database.Dialect
}

type Authorizer interface {
	RequireInTx(context.Context, database.Tx, string, string) (authorization.Decision, error)
}

type Service struct {
	db         Database
	authorizer Authorizer
	repository repository
	now        func() time.Time
}

type Option func(*Service)

func WithClock(clock func() time.Time) Option { return func(service *Service) { service.now = clock } }

func NewService(db Database, options ...Option) (*Service, error) {
	if db == nil {
		return nil, errors.New("organization database is required")
	}
	return newServiceWithAuthorizer(db, authorization.NewService(db), options...)
}

func newServiceWithAuthorizer(db Database, authorizer Authorizer, options ...Option) (*Service, error) {
	if db == nil || authorizer == nil {
		return nil, errors.New("organization database and authorizer are required")
	}
	service := &Service{db: db, authorizer: authorizer, repository: repository{dialect: db.Dialect()}, now: time.Now}
	for _, option := range options {
		option(service)
	}
	if service.now == nil {
		return nil, errors.New("organization clock is required")
	}
	return service, nil
}

func (s *Service) ListDepartments(ctx context.Context, actorID string) ([]Department, error) {
	var result []Department
	err := s.read(ctx, actorID, PermissionDepartmentsRead, func(ctx context.Context, tx database.Tx) error {
		values, err := s.repository.departments(ctx, tx, false)
		if err != nil {
			return err
		}
		result, err = orderDepartmentTree(values)
		return err
	})
	return result, err
}

func (s *Service) CreateDepartment(ctx context.Context, actorID string, input DepartmentInput) (Department, error) {
	input = normalizeDepartmentInput(input)
	if !validDepartmentInput(input) {
		return Department{}, ErrValidation
	}
	parent := input.ParentID
	value := Department{ID: uuid.NewString(), Key: input.Key, Name: input.Name, ParentID: &parent, SortOrder: input.SortOrder}
	err := s.write(ctx, actorID, PermissionDepartmentsWrite, func(ctx context.Context, tx database.Tx) error {
		departments, err := s.repository.departments(ctx, tx, true)
		if err != nil {
			return err
		}
		if !departmentExists(departments, input.ParentID) {
			return ErrNotFound
		}
		return s.repository.insertDepartment(ctx, tx, value, s.now().UTC())
	})
	return value, err
}

func (s *Service) UpdateDepartment(ctx context.Context, actorID, id string, input DepartmentInput) (Department, error) {
	input = normalizeDepartmentInput(input)
	if id == "" || !validDepartmentInput(input) {
		return Department{}, ErrValidation
	}
	parent := input.ParentID
	value := Department{ID: id, Key: input.Key, Name: input.Name, ParentID: &parent, SortOrder: input.SortOrder}
	err := s.write(ctx, actorID, PermissionDepartmentsWrite, func(ctx context.Context, tx database.Tx) error {
		departments, err := s.repository.departments(ctx, tx, true)
		if err != nil {
			return err
		}
		current, ok := departmentByID(departments, id)
		if !ok {
			return ErrNotFound
		}
		if current.Protected {
			return ErrConflict
		}
		if !validParentChange(departments, id, input.ParentID) {
			return ErrConflict
		}
		return s.repository.updateDepartment(ctx, tx, value, s.now().UTC())
	})
	return value, err
}

func (s *Service) DeleteDepartment(ctx context.Context, actorID, id string) error {
	if id == "" {
		return ErrValidation
	}
	return s.write(ctx, actorID, PermissionDepartmentsDelete, func(ctx context.Context, tx database.Tx) error {
		departments, err := s.repository.departments(ctx, tx, true)
		if err != nil {
			return err
		}
		current, ok := departmentByID(departments, id)
		if !ok {
			return ErrNotFound
		}
		if current.Protected {
			return ErrConflict
		}
		references, err := s.repository.departmentReferences(ctx, tx, id)
		if err != nil {
			return err
		}
		if references > 0 {
			return ErrConflict
		}
		return s.repository.deleteDepartment(ctx, tx, id)
	})
}

func (s *Service) ListPositions(ctx context.Context, actorID, search string, page, pageSize int) (PositionPage, error) {
	search = strings.TrimSpace(search)
	if page < 1 || page > maximumPositionPage || pageSize < 1 || pageSize > 100 || runeLength(search) < 0 || runeLength(search) > 100 {
		return PositionPage{}, ErrValidation
	}
	var result PositionPage
	err := s.read(ctx, actorID, PermissionPositionsRead, func(ctx context.Context, tx database.Tx) error {
		var err error
		result, err = s.repository.positions(ctx, tx, search, page, pageSize)
		return err
	})
	return result, err
}

func (s *Service) CreatePosition(ctx context.Context, actorID string, input PositionInput) (Position, error) {
	input = normalizePositionInput(input)
	if !validPositionInput(input) {
		return Position{}, ErrValidation
	}
	value := Position{ID: uuid.NewString(), Key: input.Key, Name: input.Name, DepartmentID: input.DepartmentID, Enabled: input.Enabled}
	err := s.write(ctx, actorID, PermissionPositionsWrite, func(ctx context.Context, tx database.Tx) error {
		departments, err := s.repository.departments(ctx, tx, true)
		if err != nil {
			return err
		}
		if !departmentExists(departments, value.DepartmentID) {
			return ErrNotFound
		}
		return s.repository.insertPosition(ctx, tx, value, s.now().UTC())
	})
	return value, err
}

func (s *Service) UpdatePosition(ctx context.Context, actorID, id string, input PositionInput) (Position, error) {
	input = normalizePositionInput(input)
	if id == "" || !validPositionInput(input) {
		return Position{}, ErrValidation
	}
	value := Position{ID: id, Key: input.Key, Name: input.Name, DepartmentID: input.DepartmentID, Enabled: input.Enabled}
	err := s.write(ctx, actorID, PermissionPositionsWrite, func(ctx context.Context, tx database.Tx) error {
		departments, err := s.repository.departments(ctx, tx, true)
		if err != nil {
			return err
		}
		if !departmentExists(departments, value.DepartmentID) {
			return ErrNotFound
		}
		current, err := s.repository.position(ctx, tx, id, true)
		if err != nil {
			return err
		}
		if current.Protected {
			return ErrConflict
		}
		return s.repository.updatePosition(ctx, tx, value, s.now().UTC())
	})
	return value, err
}

func (s *Service) DeletePosition(ctx context.Context, actorID, id string) error {
	if id == "" {
		return ErrValidation
	}
	return s.write(ctx, actorID, PermissionPositionsDelete, func(ctx context.Context, tx database.Tx) error {
		current, err := s.repository.position(ctx, tx, id, true)
		if err != nil {
			return err
		}
		if current.Protected {
			return ErrConflict
		}
		return s.repository.deletePosition(ctx, tx, id)
	})
}

func (s *Service) read(ctx context.Context, actorID, permission string, operation func(context.Context, database.Tx) error) error {
	return s.transact(ctx, actorID, permission, operation)
}

func (s *Service) write(ctx context.Context, actorID, permission string, operation func(context.Context, database.Tx) error) error {
	return s.transact(ctx, actorID, permission, operation)
}

func (s *Service) transact(ctx context.Context, actorID, permission string, operation func(context.Context, database.Tx) error) error {
	if actorID == "" {
		return ErrDenied
	}
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		decision, err := s.authorizer.RequireInTx(ctx, tx, actorID, permission)
		if err != nil {
			return err
		}
		if decision.Scope != authorization.ScopeAll {
			return ErrDenied
		}
		return operation(ctx, tx)
	})
	return s.normalize(ctx, err)
}

func (s *Service) normalize(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if contextError := ctx.Err(); contextError != nil {
		return contextError
	}
	for _, stable := range []error{context.Canceled, context.DeadlineExceeded, ErrDenied, ErrNotFound, ErrValidation, ErrConflict} {
		if errors.Is(err, stable) {
			return stable
		}
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if constraintConflict(s.db.Dialect(), err) {
		return ErrConflict
	}
	return ErrInternal
}

func constraintConflict(dialect database.Dialect, err error) bool {
	if dialect == database.DialectPostgres {
		var state interface{ SQLState() string }
		return errors.As(err, &state) && (state.SQLState() == "23505" || state.SQLState() == "23503" || state.SQLState() == "23514")
	}
	if dialect == database.DialectSQLite {
		var coded interface{ Code() int }
		return errors.As(err, &coded) && (coded.Code() == 1555 || coded.Code() == 2067 || coded.Code() == 787 || coded.Code() == 275)
	}
	return false
}

func normalizeDepartmentInput(input DepartmentInput) DepartmentInput {
	input.Key = strings.ToLower(strings.TrimSpace(input.Key))
	input.Name = strings.TrimSpace(input.Name)
	input.ParentID = strings.TrimSpace(input.ParentID)
	return input
}

func normalizePositionInput(input PositionInput) PositionInput {
	input.Key = strings.ToLower(strings.TrimSpace(input.Key))
	input.Name = strings.TrimSpace(input.Name)
	input.DepartmentID = strings.TrimSpace(input.DepartmentID)
	return input
}

func validDepartmentInput(input DepartmentInput) bool {
	return validKey(input.Key) && runeLength(input.Name) >= 1 && runeLength(input.Name) <= 100 && input.ParentID != "" && input.SortOrder >= -1_000_000 && input.SortOrder <= 1_000_000
}

func validPositionInput(input PositionInput) bool {
	return validKey(input.Key) && runeLength(input.Name) >= 1 && runeLength(input.Name) <= 100 && input.DepartmentID != ""
}

func runeLength(value string) int {
	if !utf8.ValidString(value) {
		return -1
	}
	return utf8.RuneCountInString(value)
}

func normalizedNameKey(value string) string {
	const hexadecimal = "0123456789abcdef"
	bytes := []byte(strings.ToLower(value))
	var result strings.Builder
	result.Grow(len(bytes) * 3)
	for _, value := range bytes {
		result.WriteByte(hexadecimal[value>>4])
		result.WriteByte(hexadecimal[value&0x0f])
		result.WriteByte('.')
	}
	return result.String()
}

func validKey(value string) bool {
	return len(value) >= 3 && len(value) <= 64 && stableKey.MatchString(value)
}

func departmentByID(values []Department, id string) (Department, bool) {
	for _, value := range values {
		if value.ID == id {
			return value, true
		}
	}
	return Department{}, false
}

func departmentExists(values []Department, id string) bool {
	_, ok := departmentByID(values, id)
	return ok
}

func validParentChange(values []Department, id, parentID string) bool {
	if id == parentID || !departmentExists(values, parentID) {
		return false
	}
	byID := make(map[string]Department, len(values))
	for _, value := range values {
		byID[value.ID] = value
	}
	seen := map[string]bool{id: true}
	for current := parentID; current != ""; {
		if seen[current] {
			return false
		}
		seen[current] = true
		value, ok := byID[current]
		if !ok || value.ParentID == nil {
			return ok
		}
		current = *value.ParentID
	}
	return true
}

func orderDepartmentTree(values []Department) ([]Department, error) {
	children := make(map[string][]Department)
	roots := []Department{}
	for _, value := range values {
		if value.ParentID == nil {
			roots = append(roots, value)
			continue
		}
		children[*value.ParentID] = append(children[*value.ParentID], value)
	}
	order := func(values []Department) {
		sort.Slice(values, func(i, j int) bool {
			if values[i].SortOrder != values[j].SortOrder {
				return values[i].SortOrder < values[j].SortOrder
			}
			return values[i].Key < values[j].Key
		})
	}
	order(roots)
	for key := range children {
		order(children[key])
	}
	result := make([]Department, 0, len(values))
	seen := map[string]bool{}
	var visit func(Department) error
	visit = func(value Department) error {
		if seen[value.ID] {
			return ErrInternal
		}
		seen[value.ID] = true
		result = append(result, value)
		for _, child := range children[value.ID] {
			if err := visit(child); err != nil {
				return err
			}
		}
		return nil
	}
	for _, root := range roots {
		if err := visit(root); err != nil {
			return nil, err
		}
	}
	if len(result) != len(values) {
		return nil, ErrInternal
	}
	return result, nil
}
