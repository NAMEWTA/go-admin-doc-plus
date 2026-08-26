package demo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"go-admin/internal/platform/database"
)

type Service struct {
	db   Database
	auth Authorizer
	repo repository
	now  func() time.Time
}

type Option func(*Service)

func WithClock(clock func() time.Time) Option { return func(service *Service) { service.now = clock } }

func NewService(db Database, auth Authorizer, options ...Option) (*Service, error) {
	if db == nil || auth == nil || db.Dialect() != auth.Dialect() {
		return nil, errors.New("demo database and authorizer are required on the same dialect")
	}
	service := &Service{db: db, auth: auth, repo: repository{dialect: db.Dialect()}, now: time.Now}
	for _, option := range options {
		option(service)
	}
	if service.now == nil {
		return nil, errors.New("demo clock is required")
	}
	return service, nil
}

func (s *Service) List(ctx context.Context, actorID string, query ListQuery) (Page, error) {
	query, valid := normalizeQuery(query)
	if actorID == "" || !valid {
		return Page{}, ErrValidation
	}
	var result Page
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		scope, err := s.auth.RequireInTx(ctx, tx, actorID, PermissionProductsRead)
		if err != nil {
			return err
		}
		result, err = s.repo.list(ctx, tx, actorID, scope, query)
		return err
	})
	return result, s.normalize(ctx, err)
}

func (s *Service) Get(ctx context.Context, actorID, id string) (Product, error) {
	if actorID == "" || uuid.Validate(id) != nil {
		return Product{}, ErrValidation
	}
	var record productRecord
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		scope, err := s.auth.RequireInTx(ctx, tx, actorID, PermissionProductsRead)
		if err != nil {
			return err
		}
		record, err = s.repo.get(ctx, tx, id, actorID, scope)
		return err
	})
	return mapProduct(record), s.normalize(ctx, err)
}

func (s *Service) Create(ctx context.Context, actorID string, input ProductInput) (Product, error) {
	input, valid := normalizeInput(input)
	if actorID == "" || !valid {
		return Product{}, ErrValidation
	}
	now := s.now().UTC()
	record := productRecord{ID: uuid.NewString(), OwnerAccountID: actorID, SKU: input.SKU, Name: input.Name,
		Description: input.Description, PriceCents: input.PriceCents, Status: input.Status, Revision: 1, CreatedAt: now, UpdatedAt: now}
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if _, err := s.auth.RequireInTx(ctx, tx, actorID, PermissionProductsWrite); err != nil {
			return err
		}
		return s.repo.create(ctx, tx, record)
	})
	return mapProduct(record), s.normalize(ctx, err)
}

func (s *Service) Update(ctx context.Context, actorID, id string, revision int64, input ProductInput) (Product, error) {
	input, valid := normalizeInput(input)
	if actorID == "" || uuid.Validate(id) != nil || revision < 1 || !valid {
		return Product{}, ErrValidation
	}
	var result productRecord
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		scope, err := s.auth.RequireInTx(ctx, tx, actorID, PermissionProductsWrite)
		if err != nil {
			return err
		}
		result = productRecord{ID: id, SKU: input.SKU, Name: input.Name, Description: input.Description, PriceCents: input.PriceCents, Status: input.Status, UpdatedAt: s.now().UTC()}
		if err := s.repo.update(ctx, tx, result, revision, actorID, scope); err != nil {
			return err
		}
		result, err = s.repo.get(ctx, tx, id, actorID, scope)
		return err
	})
	return mapProduct(result), s.normalize(ctx, err)
}

func (s *Service) Delete(ctx context.Context, actorID string, targets []DeleteTarget) error {
	if actorID == "" || len(targets) < 1 || len(targets) > 100 {
		return ErrValidation
	}
	seen := map[string]struct{}{}
	for _, target := range targets {
		if uuid.Validate(target.ID) != nil || target.Revision < 1 {
			return ErrValidation
		}
		if _, ok := seen[target.ID]; ok {
			return ErrValidation
		}
		seen[target.ID] = struct{}{}
	}
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		scope, err := s.auth.RequireInTx(ctx, tx, actorID, PermissionProductsDelete)
		if err != nil {
			return err
		}
		for _, target := range targets {
			if err := s.repo.remove(ctx, tx, target, actorID, scope); err != nil {
				return err
			}
		}
		return nil
	})
	return s.normalize(ctx, err)
}

func (s *Service) normalize(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	for _, stable := range []error{context.Canceled, context.DeadlineExceeded, ErrDenied, ErrValidation, ErrNotFound, ErrConflict} {
		if errors.Is(err, stable) {
			return stable
		}
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if uniqueConflict(s.db.Dialect(), err) {
		return ErrConflict
	}
	return ErrInternal
}

func uniqueConflict(dialect database.Dialect, err error) bool {
	if dialect == database.DialectPostgres {
		var state interface{ SQLState() string }
		return errors.As(err, &state) && state.SQLState() == "23505"
	}
	if dialect == database.DialectSQLite {
		var coded interface{ Code() int }
		return errors.As(err, &coded) && (coded.Code() == 1555 || coded.Code() == 2067)
	}
	return false
}
