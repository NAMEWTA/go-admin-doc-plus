package account

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"go-admin/internal/platform/database"
)

var (
	ErrNotFound = errors.New("account not found")
	ErrConflict = errors.New("account conflicts with existing data")
)

type Profile struct {
	ID, Username, DisplayName, Email string
	AvatarRef                        *string
}

type Credential struct {
	Profile
	PasswordHash      string
	SessionGeneration int64
	Disabled          bool
}

func (c Credential) String() string   { return "account.Credential{PasswordHash:[redacted]}" }
func (c Credential) GoString() string { return c.String() }
func (c Credential) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Profile Profile `json:"profile"`
	}{Profile: c.Profile})
}

type Repository struct{ dialect database.Dialect }

func NewRepository(dialect database.Dialect) *Repository { return &Repository{dialect: dialect} }

func (r *Repository) Create(ctx context.Context, tx database.Tx, value Credential, now time.Time) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO iam_accounts
		(id, username, display_name, email, avatar_ref, password_hash, password_changed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, value.ID, normalizeUsername(value.Username), value.DisplayName,
		value.Email, value.AvatarRef, value.PasswordHash, now.UTC(), now.UTC(), now.UTC())
	if err != nil {
		return ErrConflict
	}
	return nil
}

func (r *Repository) FindCredential(ctx context.Context, tx database.Tx, username string, lock bool) (Credential, error) {
	query := `SELECT id, username, display_name, email, avatar_ref, password_hash, session_generation, disabled_at
		FROM iam_accounts WHERE username = ?`
	if lock && r.dialect == database.DialectPostgres {
		query += " FOR UPDATE"
	}
	return scanCredential(tx.QueryRowContext(ctx, query, normalizeUsername(username)))
}

func (r *Repository) FindByID(ctx context.Context, tx database.Tx, id string, lock bool) (Credential, error) {
	query := `SELECT id, username, display_name, email, avatar_ref, password_hash, session_generation, disabled_at
		FROM iam_accounts WHERE id = ?`
	if lock && r.dialect == database.DialectPostgres {
		query += " FOR UPDATE"
	}
	return scanCredential(tx.QueryRowContext(ctx, query, id))
}

func scanCredential(row *sql.Row) (Credential, error) {
	var value Credential
	var avatar sql.NullString
	var disabled sql.NullTime
	err := row.Scan(&value.ID, &value.Username, &value.DisplayName, &value.Email, &avatar, &value.PasswordHash, &value.SessionGeneration, &disabled)
	if errors.Is(err, sql.ErrNoRows) {
		return Credential{}, ErrNotFound
	}
	if err != nil {
		return Credential{}, errors.New("account lookup failed")
	}
	if avatar.Valid {
		value.AvatarRef = &avatar.String
	}
	value.Disabled = disabled.Valid
	return value, nil
}

func (r *Repository) UpdateProfile(ctx context.Context, tx database.Tx, id, displayName, email string, avatar *string, now time.Time) (Profile, error) {
	result, err := tx.ExecContext(ctx, `UPDATE iam_accounts SET display_name = ?, email = ?, avatar_ref = ?, updated_at = ? WHERE id = ? AND disabled_at IS NULL`,
		displayName, strings.ToLower(strings.TrimSpace(email)), avatar, now.UTC(), id)
	if err != nil {
		return Profile{}, errors.New("profile update failed")
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return Profile{}, ErrNotFound
	}
	value, err := r.FindByID(ctx, tx, id, false)
	return value.Profile, err
}

func (r *Repository) UpdatePasswordAndAdvanceGeneration(ctx context.Context, tx database.Tx, id, hash string, generation int64, now time.Time) error {
	result, err := tx.ExecContext(ctx, `UPDATE iam_accounts SET password_hash = ?, password_changed_at = ?, session_generation = session_generation + 1, updated_at = ? WHERE id = ? AND disabled_at IS NULL AND session_generation = ?`, hash, now.UTC(), now.UTC(), id, generation)
	if err != nil {
		return errors.New("password update failed")
	}
	count, err := result.RowsAffected()
	if err != nil {
		return errors.New("password update result failed")
	}
	if count != 1 {
		return ErrConflict
	}
	return nil
}

func (r *Repository) AdvanceSessionGeneration(ctx context.Context, tx database.Tx, id string, generation int64, now time.Time) error {
	result, err := tx.ExecContext(ctx, `UPDATE iam_accounts SET session_generation = session_generation + 1, updated_at = ? WHERE id = ? AND session_generation = ?`, now.UTC(), id, generation)
	if err != nil {
		return errors.New("session generation update failed")
	}
	count, err := result.RowsAffected()
	if err != nil {
		return errors.New("session generation result failed")
	}
	if count != 1 {
		return ErrConflict
	}
	return nil
}

func normalizeUsername(value string) string { return strings.ToLower(strings.TrimSpace(value)) }
