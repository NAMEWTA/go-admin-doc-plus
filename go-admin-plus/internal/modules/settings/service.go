package settings

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
	"github.com/google/uuid"
)

type Observation struct{ Outcome string }
type Observer interface{ Observe(Observation) }
type discardObserver struct{}

func (discardObserver) Observe(Observation) {}

type Option func(*Service)

func WithObserver(observer Observer) Option {
	return func(service *Service) {
		if observer != nil {
			service.observer = observer
		}
	}
}

type Service struct {
	db       Database
	auth     Authorizer
	repo     repository
	observer Observer
}

// NewService constructs authorization from the same database owner used for business writes.
func NewService(db Database, options ...Option) (*Service, error) {
	auth, err := NewIAMAuthorizationAdapter(db)
	if err != nil {
		return nil, err
	}
	return newService(db, auth, options...)
}
func newService(db Database, auth Authorizer, options ...Option) (*Service, error) {
	if db == nil || auth == nil || db.Dialect() != auth.Dialect() {
		return nil, errors.New("settings database and authorizer must share a dialect")
	}
	service := &Service{db: db, auth: auth, repo: repository{dialect: db.Dialect()}, observer: discardObserver{}}
	for _, option := range options {
		option(service)
	}
	return service, nil
}

func (s *Service) authorize(ctx context.Context, tx database.Tx, actor, permission string) error {
	if actor == "" {
		return ErrValidation
	}
	scope, err := s.auth.RequireInTx(ctx, tx, actor, permission)
	if err != nil {
		return err
	}
	if scope != ScopeAll {
		return ErrDenied
	}
	return nil
}
func (s *Service) normalize(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		s.observe(ctxErr)
		return ctxErr
	}
	for _, stable := range []error{context.Canceled, context.DeadlineExceeded, ErrDenied, ErrValidation, ErrSensitive, ErrNotFound, ErrConflict} {
		if errors.Is(err, stable) {
			s.observe(stable)
			return stable
		}
	}
	if errors.Is(err, sql.ErrNoRows) {
		s.observe(ErrNotFound)
		return ErrNotFound
	}
	if isUnique(s.db.Dialect(), err) || isReference(s.db.Dialect(), err) {
		s.observe(ErrConflict)
		return ErrConflict
	}
	s.observe(ErrInternal)
	return ErrInternal
}

func (s *Service) reject(err error) error { s.observe(err); return err }
func (s *Service) observe(err error) {
	outcome := "internal"
	switch {
	case errors.Is(err, ErrSensitive):
		outcome = "sensitive_rejected"
	case errors.Is(err, ErrValidation):
		outcome = "validation_rejected"
	case errors.Is(err, ErrDenied):
		outcome = "authorization_denied"
	case errors.Is(err, ErrConflict):
		outcome = "conflict"
	case errors.Is(err, ErrNotFound):
		outcome = "not_found"
	case errors.Is(err, context.Canceled):
		outcome = "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		outcome = "deadline"
	}
	s.observer.Observe(Observation{Outcome: outcome})
}

func (s *Service) ListSettings(ctx context.Context, actor string, category Category, query ListQuery) (SettingPage, error) {
	query, ok := normalizeQuery(query)
	if !ok || (category != CategoryBusiness && category != CategoryUI) {
		return SettingPage{}, s.reject(ErrValidation)
	}
	var out SettingPage
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := s.authorize(ctx, tx, actor, PermissionValuesRead); err != nil {
			return err
		}
		var err error
		out, err = s.repo.listSettings(ctx, tx, category, query)
		return err
	})
	return out, s.normalize(ctx, err)
}
func (s *Service) CreateSetting(ctx context.Context, actor string, input SettingInput) (Setting, error) {
	value, err := normalizeSetting(input)
	if err != nil {
		return Setting{}, s.reject(err)
	}
	out := Setting{ID: uuid.NewString(), Category: value.Category, Key: value.Key, Label: value.Label, Value: value.Value, Description: value.Description, Enabled: value.Enabled, Revision: 1}
	err = s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := s.authorize(ctx, tx, actor, PermissionValuesWrite); err != nil {
			return err
		}
		return s.repo.createSetting(ctx, tx, out)
	})
	return out, s.normalize(ctx, err)
}
func (s *Service) UpdateSetting(ctx context.Context, actor, id string, revision int64, input SettingInput) (Setting, error) {
	value, err := normalizeSetting(input)
	if err != nil {
		return Setting{}, s.reject(err)
	}
	if uuid.Validate(id) != nil || revision < 1 {
		return Setting{}, s.reject(ErrValidation)
	}
	out := Setting{ID: id, Category: value.Category, Key: value.Key, Label: value.Label, Value: value.Value, Description: value.Description, Enabled: value.Enabled}
	err = s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := s.authorize(ctx, tx, actor, PermissionValuesWrite); err != nil {
			return err
		}
		if err := s.repo.updateSetting(ctx, tx, out, revision); err != nil {
			return err
		}
		out, err = s.repo.getSetting(ctx, tx, id)
		return err
	})
	return out, s.normalize(ctx, err)
}
func (s *Service) DeleteSetting(ctx context.Context, actor, id string, revision int64) error {
	if uuid.Validate(id) != nil || revision < 1 {
		return s.reject(ErrValidation)
	}
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := s.authorize(ctx, tx, actor, PermissionValuesDelete); err != nil {
			return err
		}
		return s.repo.deleteSetting(ctx, tx, id, revision)
	})
	return s.normalize(ctx, err)
}

func (s *Service) ListDictionaries(ctx context.Context, actor string, query ListQuery) (DictionaryPage, error) {
	query, ok := normalizeQuery(query)
	if !ok {
		return DictionaryPage{}, s.reject(ErrValidation)
	}
	var out DictionaryPage
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := s.authorize(ctx, tx, actor, PermissionDictionariesRead); err != nil {
			return err
		}
		var err error
		out, err = s.repo.listDictionaries(ctx, tx, query)
		return err
	})
	return out, s.normalize(ctx, err)
}
func (s *Service) CreateDictionary(ctx context.Context, actor string, input DictionaryInput) (Dictionary, error) {
	value, err := normalizeDictionary(input)
	if err != nil {
		return Dictionary{}, s.reject(err)
	}
	out := Dictionary{ID: uuid.NewString(), Key: value.Key, Name: value.Name, Description: value.Description, Enabled: value.Enabled, Revision: 1}
	err = s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := s.authorize(ctx, tx, actor, PermissionDictionariesWrite); err != nil {
			return err
		}
		return s.repo.createDictionary(ctx, tx, out)
	})
	return out, s.normalize(ctx, err)
}
func (s *Service) UpdateDictionary(ctx context.Context, actor, id string, revision int64, input DictionaryInput) (Dictionary, error) {
	value, err := normalizeDictionary(input)
	if err != nil {
		return Dictionary{}, s.reject(err)
	}
	if uuid.Validate(id) != nil || revision < 1 {
		return Dictionary{}, s.reject(ErrValidation)
	}
	out := Dictionary{ID: id, Key: value.Key, Name: value.Name, Description: value.Description, Enabled: value.Enabled}
	err = s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := s.authorize(ctx, tx, actor, PermissionDictionariesWrite); err != nil {
			return err
		}
		if err := s.repo.updateDictionary(ctx, tx, out, revision); err != nil {
			return err
		}
		out, err = s.repo.getDictionary(ctx, tx, id)
		return err
	})
	return out, s.normalize(ctx, err)
}
func (s *Service) DeleteDictionary(ctx context.Context, actor, id string, revision int64) error {
	if uuid.Validate(id) != nil || revision < 1 {
		return s.reject(ErrValidation)
	}
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := s.authorize(ctx, tx, actor, PermissionDictionariesDelete); err != nil {
			return err
		}
		return s.repo.deleteDictionary(ctx, tx, id, revision)
	})
	return s.normalize(ctx, err)
}

func (s *Service) ListItems(ctx context.Context, actor, dictionaryID string, query ListQuery) (DictionaryItemPage, error) {
	query, ok := normalizeQuery(query)
	if !ok || uuid.Validate(dictionaryID) != nil {
		return DictionaryItemPage{}, s.reject(ErrValidation)
	}
	var out DictionaryItemPage
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := s.authorize(ctx, tx, actor, PermissionDictionariesRead); err != nil {
			return err
		}
		if _, err := s.repo.getDictionary(ctx, tx, dictionaryID); err != nil {
			return err
		}
		value, err := s.repo.listItems(ctx, tx, dictionaryID, query)
		out = value
		return err
	})
	return out, s.normalize(ctx, err)
}
func (s *Service) CreateItem(ctx context.Context, actor, dictionaryID string, input DictionaryItemInput) (DictionaryItem, error) {
	value, err := normalizeItem(input)
	if err != nil {
		return DictionaryItem{}, s.reject(err)
	}
	if uuid.Validate(dictionaryID) != nil {
		return DictionaryItem{}, s.reject(ErrValidation)
	}
	out := DictionaryItem{ID: uuid.NewString(), DictionaryID: dictionaryID, Value: value.Value, Label: value.Label, SortOrder: value.SortOrder, Enabled: value.Enabled, Revision: 1}
	err = s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := s.authorize(ctx, tx, actor, PermissionDictionariesWrite); err != nil {
			return err
		}
		if _, err := s.repo.getDictionary(ctx, tx, dictionaryID); err != nil {
			return err
		}
		return s.repo.createItem(ctx, tx, out)
	})
	return out, s.normalize(ctx, err)
}
func (s *Service) UpdateItem(ctx context.Context, actor, id string, revision int64, input DictionaryItemInput) (DictionaryItem, error) {
	value, err := normalizeItem(input)
	if err != nil {
		return DictionaryItem{}, s.reject(err)
	}
	if uuid.Validate(id) != nil || revision < 1 {
		return DictionaryItem{}, s.reject(ErrValidation)
	}
	out := DictionaryItem{ID: id, Value: value.Value, Label: value.Label, SortOrder: value.SortOrder, Enabled: value.Enabled}
	err = s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := s.authorize(ctx, tx, actor, PermissionDictionariesWrite); err != nil {
			return err
		}
		if err := s.repo.updateItem(ctx, tx, out, revision); err != nil {
			return err
		}
		out, err = s.repo.getItem(ctx, tx, id)
		return err
	})
	return out, s.normalize(ctx, err)
}
func (s *Service) DeleteItem(ctx context.Context, actor, id string, revision int64) error {
	if uuid.Validate(id) != nil || revision < 1 {
		return s.reject(ErrValidation)
	}
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := s.authorize(ctx, tx, actor, PermissionDictionariesDelete); err != nil {
			return err
		}
		return s.repo.deleteItem(ctx, tx, id, revision)
	})
	return s.normalize(ctx, err)
}
func (s *Service) Options(ctx context.Context, actor, key string) ([]DictionaryOption, error) {
	key = strings.ToLower(strings.TrimSpace(key))
	if !validKey(key) {
		return nil, s.reject(ErrValidation)
	}
	var out []DictionaryOption
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		if err := s.authorize(ctx, tx, actor, PermissionOptionsRead); err != nil {
			return err
		}
		var err error
		out, err = s.repo.options(ctx, tx, key)
		return err
	})
	return out, s.normalize(ctx, err)
}
