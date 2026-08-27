package settings

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"go-admin/internal/platform/database"
)

const (
	PermissionValuesRead         = "settings.values.read"
	PermissionValuesWrite        = "settings.values.write"
	PermissionValuesDelete       = "settings.values.delete"
	PermissionDictionariesRead   = "settings.dictionaries.read"
	PermissionDictionariesWrite  = "settings.dictionaries.write"
	PermissionDictionariesDelete = "settings.dictionaries.delete"
	PermissionOptionsRead        = "settings.options.read"
	MaximumPage                  = 1_000_000
)

var (
	ErrDenied     = errors.New("settings authorization denied")
	ErrValidation = errors.New("settings request invalid")
	ErrSensitive  = errors.New("settings sensitive material rejected")
	ErrNotFound   = errors.New("settings resource not found")
	ErrConflict   = errors.New("settings resource conflict")
	ErrInternal   = errors.New("settings operation failed")
	stableKey     = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*$`)
)

type Scope string

const ScopeAll Scope = "all"

type Database interface {
	WithinTx(context.Context, func(context.Context, database.Tx) error) error
	Dialect() database.Dialect
}

type Authorizer interface {
	Dialect() database.Dialect
	RequireInTx(context.Context, database.Tx, string, string) (Scope, error)
}

type Category string

const (
	CategoryBusiness Category = "business"
	CategoryUI       Category = "ui"
)

type ListQuery struct {
	Search        string
	Page, PerPage int
}

type Setting struct {
	ID, Key, Label, Value, Description string
	Category                           Category
	Enabled                            bool
	Revision                           int64
}

type SettingInput struct {
	Key, Label, Value, Description string
	Category                       Category
	Enabled                        bool
}

type SettingPage struct {
	Rows  []Setting
	Total int
}

type Dictionary struct {
	ID, Key, Name, Description string
	Enabled                    bool
	Revision                   int64
}

type DictionaryInput struct {
	Key, Name, Description string
	Enabled                bool
}

type DictionaryPage struct {
	Rows  []Dictionary
	Total int
}

type DictionaryItem struct {
	ID, DictionaryID, Value, Label string
	SortOrder                      int
	Enabled                        bool
	Revision                       int64
}

type DictionaryItemInput struct {
	Value, Label string
	SortOrder    int
	Enabled      bool
}

type DictionaryItemPage struct {
	Rows  []DictionaryItem
	Total int
}
type DictionaryOption struct{ Value, Label string }

func normalizeQuery(value ListQuery) (ListQuery, bool) {
	value.Search = strings.ToLower(strings.TrimSpace(value.Search))
	return value, runeLength(value.Search) <= 100 && value.Page >= 1 && value.Page <= MaximumPage && value.PerPage >= 1 && value.PerPage <= 100
}

func normalizeSetting(value SettingInput) (SettingInput, error) {
	value.Key = strings.ToLower(strings.TrimSpace(value.Key))
	value.Label = strings.TrimSpace(value.Label)
	value.Value = strings.TrimSpace(value.Value)
	value.Description = strings.TrimSpace(value.Description)
	if !validKey(value.Key) || !validText(value.Label, 1, 120) || !validText(value.Value, 1, 500) || !validText(value.Description, 0, 500) ||
		(value.Category != CategoryBusiness && value.Category != CategoryUI) {
		return SettingInput{}, ErrValidation
	}
	if sensitive(value.Key) || sensitive(value.Label) || sensitive(value.Value) || sensitive(value.Description) {
		return SettingInput{}, ErrSensitive
	}
	return value, nil
}

func normalizeDictionary(value DictionaryInput) (DictionaryInput, error) {
	value.Key = strings.ToLower(strings.TrimSpace(value.Key))
	value.Name = strings.TrimSpace(value.Name)
	value.Description = strings.TrimSpace(value.Description)
	if !validKey(value.Key) || !validText(value.Name, 1, 120) || !validText(value.Description, 0, 500) {
		return DictionaryInput{}, ErrValidation
	}
	if sensitive(value.Key) || sensitive(value.Name) || sensitive(value.Description) {
		return DictionaryInput{}, ErrSensitive
	}
	return value, nil
}

func normalizeItem(value DictionaryItemInput) (DictionaryItemInput, error) {
	value.Value = strings.TrimSpace(value.Value)
	value.Label = strings.TrimSpace(value.Label)
	if !validText(value.Value, 1, 120) || !validText(value.Label, 1, 120) || value.SortOrder < 0 || value.SortOrder > 100_000 {
		return DictionaryItemInput{}, ErrValidation
	}
	if sensitive(value.Value) || sensitive(value.Label) {
		return DictionaryItemInput{}, ErrSensitive
	}
	return value, nil
}

func validKey(value string) bool {
	return runeLength(value) >= 3 && runeLength(value) <= 80 && stableKey.MatchString(value)
}
func validText(value string, minimum, maximum int) bool {
	length := runeLength(value)
	if length < minimum || length > maximum {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
func runeLength(value string) int {
	if !utf8.ValidString(value) {
		return -1
	}
	return utf8.RuneCountInString(value)
}

// normalizedTextKey is byte-boundary explicit and sorts identically under BINARY/C collations.
func normalizedTextKey(value string) string {
	var result strings.Builder
	for _, current := range []byte(strings.ToLower(value)) {
		result.WriteString(".")
		result.WriteString("0123456789abcdef"[current>>4 : current>>4+1])
		result.WriteString("0123456789abcdef"[current&15 : current&15+1])
	}
	return result.String()
}
