package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"go-admin/internal/modules/iam/account"
	"go-admin/internal/platform/database"
)

var (
	ErrAuthentication = errors.New("authentication required")
	ErrCredentials    = errors.New("credentials invalid")
	ErrCSRF           = errors.New("request authorization failed")
	ErrValidation     = errors.New("request validation failed")
	ErrConflict       = errors.New("request conflicts with current state")
	ErrInternal       = errors.New("iam operation failed")
	errExpired        = errors.New("session expired")
)

var avatarReference = regexp.MustCompile(`^files/[A-Za-z0-9][A-Za-z0-9._/-]{0,254}$`)

type Policy struct {
	IdleTimeout      time.Duration
	AbsoluteTimeout  time.Duration
	RotationInterval time.Duration
}

func (p Policy) validate() error {
	if p.IdleTimeout <= 0 || p.AbsoluteTimeout <= 0 || p.RotationInterval <= 0 || p.IdleTimeout > p.AbsoluteTimeout || p.RotationInterval > p.AbsoluteTimeout {
		return errors.New("session policy is invalid")
	}
	return nil
}

type Database interface {
	WithinTx(context.Context, func(context.Context, database.Tx) error) error
	Dialect() database.Dialect
}

type Service struct {
	db       Database
	accounts *account.Repository
	policy   Policy
	now      func() time.Time
}

type Option func(*Service)

func WithClock(clock func() time.Time) Option { return func(s *Service) { s.now = clock } }

func NewService(db Database, policy Policy, options ...Option) (*Service, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}
	if err := policy.validate(); err != nil {
		return nil, err
	}
	s := &Service{db: db, accounts: account.NewRepository(db.Dialect()), policy: policy, now: time.Now}
	for _, option := range options {
		option(s)
	}
	if s.now == nil {
		return nil, errors.New("clock is required")
	}
	return s, nil
}

type Issued struct {
	Profile account.Profile
	Token   string
	CSRF    string
	Rotated bool
}

func (i Issued) String() string   { return "session.Issued{Token:[redacted], CSRF:[redacted]}" }
func (i Issued) GoString() string { return i.String() }
func (i Issued) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Profile account.Profile `json:"profile"`
		Rotated bool            `json:"rotated"`
	}{Profile: i.Profile, Rotated: i.Rotated})
}

type record struct {
	ID, AccountID, TokenHash, CSRFHash, State                         string
	CreatedAt, LastSeenAt, IdleExpiresAt, AbsoluteExpiresAt, RotateAt time.Time
}

func (s *Service) Login(ctx context.Context, username, password string) (Issued, error) {
	if len(strings.TrimSpace(username)) < 3 || len(username) > 64 || len(password) < 12 || len(password) > 128 {
		return Issued{}, ErrCredentials
	}
	var observed account.Credential
	err := s.withinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		credential, err := s.accounts.FindCredential(ctx, tx, username, false)
		if err == nil {
			observed = credential
		}
		return err
	})
	if errors.Is(err, account.ErrNotFound) {
		// A fixed valid hash keeps absent and wrong-password paths in the same expensive class.
		_ = account.VerifyPassword(dummyPasswordHash, password)
		return Issued{}, ErrCredentials
	}
	if err != nil {
		return Issued{}, sanitize(err)
	}
	passwordValid := account.VerifyPassword(observed.PasswordHash, password)
	if observed.Disabled || !passwordValid {
		return Issued{}, ErrCredentials
	}

	var result Issued
	now := s.now().UTC()
	err = s.withinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		credential, err := s.accounts.FindCredential(ctx, tx, username, true)
		if errors.Is(err, account.ErrNotFound) || credential.Disabled {
			return ErrCredentials
		}
		if err != nil {
			return err
		}
		if subtle.ConstantTimeCompare([]byte(credential.PasswordHash), []byte(observed.PasswordHash)) != 1 {
			return ErrCredentials
		}
		issued, err := s.create(ctx, tx, credential.Profile, now, now.Add(s.policy.AbsoluteTimeout))
		if err == nil {
			result = issued
		}
		return err
	})
	return result, sanitize(err)
}

func (s *Service) Current(ctx context.Context, token string) (Issued, error) {
	var result Issued
	now := s.now().UTC()
	err := s.withinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		rec, err := s.active(ctx, tx, token, now)
		if err != nil {
			return err
		}
		credential, err := s.accounts.FindByID(ctx, tx, rec.AccountID, false)
		if errors.Is(err, account.ErrNotFound) || credential.Disabled {
			return ErrAuthentication
		}
		if err != nil {
			return err
		}
		if !now.Before(rec.RotateAt) {
			issued, err := s.create(ctx, tx, credential.Profile, now, rec.AbsoluteExpiresAt)
			if err != nil {
				return err
			}
			updated, err := tx.ExecContext(ctx, `UPDATE iam_sessions SET state = 'rotated', replaced_by = ? WHERE id = ? AND state = 'active'`, tokenDigest(issued.Token), rec.ID)
			if err != nil {
				return errors.New("session rotation failed")
			}
			if count, _ := updated.RowsAffected(); count != 1 {
				return ErrAuthentication
			}
			issued.Rotated = true
			result = issued
			return nil
		}
		csrf, err := randomSecret()
		if err != nil {
			return err
		}
		updated, err := tx.ExecContext(ctx, `UPDATE iam_sessions SET csrf_hash = ?, last_seen_at = ?, idle_expires_at = ? WHERE id = ? AND state = 'active'`, tokenDigest(csrf), now, minTime(now.Add(s.policy.IdleTimeout), rec.AbsoluteExpiresAt), rec.ID)
		if err != nil {
			return errors.New("session refresh failed")
		}
		if count, _ := updated.RowsAffected(); count != 1 {
			return ErrAuthentication
		}
		result = Issued{Profile: credential.Profile, CSRF: csrf}
		return nil
	})
	return result, sanitize(err)
}

func (s *Service) Profile(ctx context.Context, token string) (account.Profile, error) {
	var profile account.Profile
	now := s.now().UTC()
	err := s.withinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		rec, err := s.active(ctx, tx, token, now)
		if err != nil {
			return err
		}
		value, err := s.accounts.FindByID(ctx, tx, rec.AccountID, false)
		if errors.Is(err, account.ErrNotFound) {
			return ErrAuthentication
		}
		if err != nil || value.Disabled {
			if err == nil {
				return ErrAuthentication
			}
			return err
		}
		profile = value.Profile
		return s.touch(ctx, tx, rec, now)
	})
	return profile, sanitize(err)
}

func (s *Service) UpdateProfile(ctx context.Context, token, csrf, displayName, email string, avatar *string) (account.Profile, error) {
	displayName = strings.TrimSpace(displayName)
	email = strings.ToLower(strings.TrimSpace(email))
	parsedEmail, emailErr := mail.ParseAddress(email)
	if displayName == "" || len(displayName) > 80 || emailErr != nil || parsedEmail.Address != email || len(email) > 254 || (avatar != nil && !avatarReference.MatchString(*avatar)) {
		return account.Profile{}, ErrValidation
	}
	var profile account.Profile
	now := s.now().UTC()
	err := s.withinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		rec, err := s.authorizeMutation(ctx, tx, token, csrf, now)
		if err != nil {
			return err
		}
		profile, err = s.accounts.UpdateProfile(ctx, tx, rec.AccountID, displayName, email, avatar, now)
		if errors.Is(err, account.ErrNotFound) {
			return ErrAuthentication
		}
		if err != nil {
			return err
		}
		return s.touch(ctx, tx, rec, now)
	})
	return profile, sanitize(err)
}

func (s *Service) ChangePassword(ctx context.Context, token, csrf, current, replacement string) error {
	if len(current) < 12 || len(current) > 128 || len(replacement) < 12 || len(replacement) > 128 {
		return ErrValidation
	}
	var accountID, observedHash string
	// Fence cheaply first, then perform the memory-hard work without holding a database transaction.
	err := s.withinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		rec, err := s.authorizeMutation(ctx, tx, token, csrf, s.now().UTC())
		if err != nil {
			return err
		}
		credential, err := s.accounts.FindByID(ctx, tx, rec.AccountID, false)
		if err != nil || credential.Disabled {
			if err == nil {
				return ErrAuthentication
			}
			return err
		}
		accountID, observedHash = rec.AccountID, credential.PasswordHash
		return nil
	})
	if err != nil {
		return sanitize(err)
	}
	if !account.VerifyPassword(observedHash, current) {
		return ErrCredentials
	}
	newHash, err := account.HashPassword(replacement)
	if err != nil {
		return ErrValidation
	}
	now := s.now().UTC()
	err = s.withinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		rec, err := s.authorizeMutation(ctx, tx, token, csrf, now)
		if err != nil {
			return err
		}
		if rec.AccountID != accountID {
			return ErrAuthentication
		}
		credential, err := s.accounts.FindByID(ctx, tx, accountID, true)
		if err != nil {
			return err
		}
		if credential.Disabled {
			return ErrAuthentication
		}
		if subtle.ConstantTimeCompare([]byte(credential.PasswordHash), []byte(observedHash)) != 1 {
			return ErrConflict
		}
		if err := s.accounts.UpdatePassword(ctx, tx, accountID, newHash, now); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE iam_sessions SET state = 'revoked', revoked_at = ? WHERE account_id = ? AND state = 'active'`, now, accountID)
		return err
	})
	return sanitize(err)
}

func (s *Service) Logout(ctx context.Context, token, csrf string) error {
	return sanitize(s.withinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		rec, err := s.authorizeMutation(ctx, tx, token, csrf, s.now().UTC())
		if err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `UPDATE iam_sessions SET state = 'revoked', revoked_at = ? WHERE id = ? AND state = 'active'`, s.now().UTC(), rec.ID)
		if err != nil {
			return errors.New("session revoke failed")
		}
		if n, _ := result.RowsAffected(); n != 1 {
			return ErrAuthentication
		}
		return nil
	}))
}

func (s *Service) RevokeAccount(ctx context.Context, accountID string) error {
	return sanitize(s.withinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		_, err := tx.ExecContext(ctx, `UPDATE iam_sessions SET state = 'revoked', revoked_at = ? WHERE account_id = ? AND state = 'active'`, s.now().UTC(), accountID)
		return err
	}))
}

func (s *Service) create(ctx context.Context, tx database.Tx, profile account.Profile, now, absolute time.Time) (Issued, error) {
	token, err := randomSecret()
	if err != nil {
		return Issued{}, err
	}
	csrf, err := randomSecret()
	if err != nil {
		return Issued{}, err
	}
	id, err := randomSecret()
	if err != nil {
		return Issued{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO iam_sessions
		(id, account_id, token_hash, csrf_hash, state, created_at, last_seen_at, idle_expires_at, absolute_expires_at, rotate_at)
		VALUES (?, ?, ?, ?, 'active', ?, ?, ?, ?, ?)`, id, profile.ID, tokenDigest(token), tokenDigest(csrf), now, now,
		minTime(now.Add(s.policy.IdleTimeout), absolute), absolute, minTime(now.Add(s.policy.RotationInterval), absolute))
	if err != nil {
		return Issued{}, errors.New("session creation failed")
	}
	return Issued{Profile: profile, Token: token, CSRF: csrf}, nil
}

func (s *Service) active(ctx context.Context, tx database.Tx, token string, now time.Time) (record, error) {
	if len(token) != 43 {
		return record{}, ErrAuthentication
	}
	query := `SELECT id, account_id, token_hash, csrf_hash, state, created_at, last_seen_at, idle_expires_at, absolute_expires_at, rotate_at
		FROM iam_sessions WHERE token_hash = ?`
	if s.db.Dialect() == database.DialectPostgres {
		query += " FOR UPDATE"
	}
	var rec record
	err := tx.QueryRowContext(ctx, query, tokenDigest(token)).Scan(&rec.ID, &rec.AccountID, &rec.TokenHash, &rec.CSRFHash, &rec.State, &rec.CreatedAt, &rec.LastSeenAt, &rec.IdleExpiresAt, &rec.AbsoluteExpiresAt, &rec.RotateAt)
	if err != nil || rec.State != "active" {
		return record{}, ErrAuthentication
	}
	if !now.Before(rec.IdleExpiresAt) || !now.Before(rec.AbsoluteExpiresAt) {
		updated, updateErr := tx.ExecContext(ctx, `UPDATE iam_sessions SET state = 'expired' WHERE id = ? AND state = 'active'`, rec.ID)
		if updateErr != nil {
			return record{}, errors.New("session expiry failed")
		}
		if count, _ := updated.RowsAffected(); count != 1 {
			return record{}, ErrAuthentication
		}
		return record{}, errExpired
	}
	return rec, nil
}

func (s *Service) touch(ctx context.Context, tx database.Tx, rec record, now time.Time) error {
	updated, err := tx.ExecContext(ctx, `UPDATE iam_sessions SET last_seen_at = ?, idle_expires_at = ? WHERE id = ? AND state = 'active'`,
		now, minTime(now.Add(s.policy.IdleTimeout), rec.AbsoluteExpiresAt), rec.ID)
	if err != nil {
		return errors.New("session refresh failed")
	}
	if count, _ := updated.RowsAffected(); count != 1 {
		return ErrAuthentication
	}
	return nil
}

// withinTx commits the active->expired transition while preserving an authentication result.
func (s *Service) withinTx(ctx context.Context, fn func(context.Context, database.Tx) error) error {
	expired := false
	err := s.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		err := fn(ctx, tx)
		if errors.Is(err, errExpired) {
			expired = true
			return nil
		}
		return err
	})
	if err == nil && expired {
		return ErrAuthentication
	}
	return err
}

func (s *Service) authorizeMutation(ctx context.Context, tx database.Tx, token, csrf string, now time.Time) (record, error) {
	rec, err := s.active(ctx, tx, token, now)
	if err != nil {
		return record{}, err
	}
	got := tokenDigest(csrf)
	if subtle.ConstantTimeCompare([]byte(got), []byte(rec.CSRFHash)) != 1 {
		return record{}, ErrCSRF
	}
	return rec, nil
}

func randomSecret() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", errors.New("secure random unavailable")
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func tokenDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func sanitize(err error) error {
	if err == nil || errors.Is(err, ErrAuthentication) || errors.Is(err, ErrCredentials) || errors.Is(err, ErrCSRF) || errors.Is(err, ErrValidation) || errors.Is(err, ErrConflict) {
		return err
	}
	if errors.Is(err, account.ErrConflict) {
		return ErrConflict
	}
	return ErrInternal
}

const dummyPasswordHash = "$argon2id$v=19$m=65536,t=3,p=4$c29tZXNhbHRzb21lc2FsdA$Bf7uFvyZ/+DZ3JZXG5q83gVsEScJTiI4btfDrxbJKUA"
