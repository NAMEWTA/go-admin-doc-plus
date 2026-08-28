package files

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
	"github.com/google/uuid"
)

const recoveryBatchSize = 100

type Service struct {
	db         Database
	storage    Storage
	authorizer Authorizer
	repository repository
	now        func() time.Time
	observer   Observer
}

type Option func(*Service)

func WithClock(clock func() time.Time) Option { return func(service *Service) { service.now = clock } }
func WithObserver(observer Observer) Option {
	return func(service *Service) { service.observer = observer }
}

func NewService(db Database, storage Storage, options ...Option) (*Service, error) {
	adapter, err := NewIAMAuthorizationAdapter(db)
	if err != nil {
		return nil, err
	}
	return newServiceWithAuthorizer(db, storage, adapter, options...)
}

func newServiceWithAuthorizer(db Database, storage Storage, authorizer Authorizer, options ...Option) (*Service, error) {
	if db == nil || storage == nil || authorizer == nil || (db.Dialect() != database.DialectSQLite && db.Dialect() != database.DialectPostgres) {
		return nil, errors.New("files service dependencies are required")
	}
	service := &Service{db: db, storage: storage, authorizer: authorizer, repository: repository{dialect: db.Dialect()}, now: time.Now, observer: discardObserver{}}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	if service.now == nil || service.observer == nil {
		return nil, errors.New("files service options are invalid")
	}
	return service, nil
}

func (service *Service) Upload(ctx context.Context, actorID string, input UploadInput) (result Metadata, resultErr error) {
	defer func() { service.observe("upload", resultErr) }()
	name, valid := normalizeFilename(input.OriginalName)
	if actorID == "" || !valid || input.Content == nil {
		return Metadata{}, ErrValidation
	}
	// Keep the body outside a database transaction, but reject unauthorized callers before any
	// bytes are staged. The insert transaction repeats this check to fence concurrent revocation.
	if err := service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		scope, err := service.authorizer.RequireInTx(ctx, tx, actorID, PermissionFilesWrite)
		if err != nil {
			return err
		}
		if !validScope(scope) {
			return ErrDenied
		}
		return nil
	}); err != nil {
		return Metadata{}, service.normalize(ctx, err)
	}
	staged, err := service.storage.Stage(ctx, input.DeclaredMediaType, input.Content)
	if err != nil {
		return Metadata{}, service.normalize(ctx, err)
	}
	stageOwnedByRequest := true
	defer func() {
		if stageOwnedByRequest {
			cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = service.storage.Abort(cleanup, staged.TemporaryKey)
		}
	}()
	now := service.now().UTC()
	record := fileRecord{ID: uuid.NewString(), OwnerAccountID: actorID, OriginalName: name, NameKey: nameKey(name), MediaType: staged.MediaType,
		SizeBytes: staged.SizeBytes, SHA256: staged.SHA256, StorageKey: NewStorageKey(), TemporaryKey: &staged.TemporaryKey, State: statePending,
		Revision: 1, CreatedAt: now, UpdatedAt: now}
	err = service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		scope, err := service.authorizer.RequireInTx(ctx, tx, actorID, PermissionFilesWrite)
		if err != nil {
			return err
		}
		if !validScope(scope) {
			return ErrDenied
		}
		return service.repository.insert(ctx, tx, record)
	})
	if err != nil {
		return Metadata{}, service.normalize(ctx, err)
	}
	stageOwnedByRequest = false
	if err := service.storage.Publish(ctx, staged.TemporaryKey, record.StorageKey); err != nil {
		return Metadata{}, service.normalize(ctx, err)
	}
	record.TemporaryKey = nil
	readyAt := service.now().UTC()
	record.UpdatedAt = readyAt
	err = service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		return service.repository.markReady(ctx, tx, record.ID, readyAt)
	})
	if err != nil {
		return Metadata{}, service.normalize(ctx, err)
	}
	return metadata(record), nil
}

func (service *Service) List(ctx context.Context, actorID string, query ListQuery) (result Page, resultErr error) {
	defer func() { service.observe("list", resultErr) }()
	query, valid := normalizeListQuery(query)
	if actorID == "" || !valid {
		return Page{}, ErrValidation
	}
	var page Page
	err := service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		scope, err := service.authorizer.RequireInTx(ctx, tx, actorID, PermissionFilesRead)
		if err != nil {
			return err
		}
		if !validScope(scope) {
			return ErrDenied
		}
		page, err = service.repository.list(ctx, tx, actorID, scope, query)
		return err
	})
	return page, service.normalize(ctx, err)
}

func (service *Service) Download(ctx context.Context, actorID, id string) (resultValue Download, resultErr error) {
	defer func() { service.observe("download", resultErr) }()
	if actorID == "" || uuid.Validate(id) != nil {
		return Download{}, ErrValidation
	}
	var result Download
	err := service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		scope, err := service.authorizer.RequireInTx(ctx, tx, actorID, PermissionFilesRead)
		if err != nil {
			return err
		}
		if !validScope(scope) {
			return ErrDenied
		}
		record, err := service.repository.ready(ctx, tx, id, actorID, scope)
		if err != nil {
			return err
		}
		content, err := service.storage.Open(ctx, record.StorageKey)
		if err != nil {
			return err
		}
		result = Download{Metadata: metadata(record), Content: content}
		return nil
	})
	if err != nil {
		if result.Content != nil {
			_ = result.Content.Close()
		}
		return Download{}, service.normalize(ctx, err)
	}
	return result, nil
}

func (service *Service) GetMetadata(ctx context.Context, actorID, id string) (resultValue Metadata, resultErr error) {
	defer func() { service.observe("metadata", resultErr) }()
	if actorID == "" || uuid.Validate(id) != nil {
		return Metadata{}, ErrValidation
	}
	var result Metadata
	err := service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		scope, err := service.authorizer.RequireInTx(ctx, tx, actorID, PermissionFilesRead)
		if err != nil {
			return err
		}
		if !validScope(scope) {
			return ErrDenied
		}
		record, err := service.repository.ready(ctx, tx, id, actorID, scope)
		if err == nil {
			result = metadata(record)
		}
		return err
	})
	return result, service.normalize(ctx, err)
}

func (service *Service) Delete(ctx context.Context, actorID string, targets []DeleteTarget) (resultErr error) {
	defer func() { service.observe("delete", resultErr) }()
	if actorID == "" || len(targets) < 1 || len(targets) > 100 {
		return ErrValidation
	}
	seen := map[string]struct{}{}
	for _, target := range targets {
		if uuid.Validate(target.ID) != nil || target.Revision < 1 {
			return ErrValidation
		}
		if _, duplicate := seen[target.ID]; duplicate {
			return ErrValidation
		}
		seen[target.ID] = struct{}{}
	}
	var records []fileRecord
	err := service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		scope, err := service.authorizer.RequireInTx(ctx, tx, actorID, PermissionFilesDelete)
		if err != nil {
			return err
		}
		if !validScope(scope) {
			return ErrDenied
		}
		records, err = service.repository.markDeleting(ctx, tx, targets, actorID, scope, service.now().UTC())
		return err
	})
	if err != nil {
		return service.normalize(ctx, err)
	}
	for _, record := range records {
		if err := service.storage.Delete(ctx, record.StorageKey); err != nil {
			return service.normalize(ctx, err)
		}
		if err := service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
			return service.repository.removeDeleting(ctx, tx, record.ID)
		}); err != nil {
			return service.normalize(ctx, err)
		}
	}
	return nil
}

// Reconcile claims incomplete metadata before touching disk so concurrent PostgreSQL replicas do
// not repair the same object. SQLite's profile contract supplies the single process owner.
func (service *Service) Reconcile(ctx context.Context) (resultErr error) {
	defer func() { service.observe("reconcile", resultErr) }()
	for {
		now := service.now().UTC()
		token := uuid.NewString()
		var records []fileRecord
		err := service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
			var err error
			records, err = service.repository.claimRecovery(ctx, tx, now, now.Add(30*time.Second), token, recoveryBatchSize)
			return err
		})
		if err != nil {
			return service.normalize(ctx, err)
		}
		if len(records) == 0 {
			return nil
		}
		for _, record := range records {
			if err := service.reconcileRecord(ctx, record, token); err != nil {
				return err
			}
		}
	}
}

func (service *Service) reconcileRecord(ctx context.Context, record fileRecord, token string) error {
	switch record.State {
	case statePending:
		objectExists, err := service.storage.ObjectExists(ctx, record.StorageKey)
		if err != nil {
			return service.normalize(ctx, err)
		}
		temporaryExists := false
		if record.TemporaryKey != nil {
			temporaryExists, err = service.storage.TemporaryExists(ctx, *record.TemporaryKey)
			if err != nil {
				return service.normalize(ctx, err)
			}
		}
		if !objectExists && temporaryExists {
			if err := service.storage.Publish(ctx, *record.TemporaryKey, record.StorageKey); err != nil {
				return service.normalize(ctx, err)
			}
			objectExists = true
		}
		if objectExists && temporaryExists {
			if err := service.storage.Abort(ctx, *record.TemporaryKey); err != nil {
				return service.normalize(ctx, err)
			}
		}
		return service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
			if !objectExists {
				return service.repository.removeClaimed(ctx, tx, record.ID, statePending, token)
			}
			return service.repository.finishClaimedReady(ctx, tx, record.ID, token, service.now().UTC())
		})
	case stateDeleting:
		if err := service.storage.Delete(ctx, record.StorageKey); err != nil {
			return service.normalize(ctx, err)
		}
		return service.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
			return service.repository.removeClaimed(ctx, tx, record.ID, stateDeleting, token)
		})
	default:
		return ErrInternal
	}
}

func (service *Service) normalize(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	for _, stable := range []error{context.Canceled, context.DeadlineExceeded, ErrDenied, ErrValidation, ErrNotFound, ErrConflict,
		ErrAuthentication, ErrCSRF, ErrContentTooLarge, ErrMediaType} {
		if errors.Is(err, stable) {
			return stable
		}
	}
	if errors.Is(err, authorization.ErrDenied) {
		return ErrDenied
	}
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, ErrStorageNotFound) {
		return ErrNotFound
	}
	return ErrInternal
}

func (service *Service) observe(operation string, err error) {
	outcome := "succeeded"
	if err != nil {
		outcome = "failed"
		for _, rejected := range []error{ErrDenied, ErrValidation, ErrNotFound, ErrConflict, ErrAuthentication, ErrCSRF, ErrContentTooLarge, ErrMediaType} {
			if errors.Is(err, rejected) {
				outcome = "rejected"
				break
			}
		}
	}
	service.observer.Observe(Observation{Operation: operation, Outcome: outcome})
}

type discardObserver struct{}

func (discardObserver) Observe(Observation) {}
