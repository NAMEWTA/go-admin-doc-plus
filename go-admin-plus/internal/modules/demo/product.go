// Package demo is the reference business CRUD vertical slice.
package demo

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

const (
	PermissionProductsRead   = "demo.products.read"
	PermissionProductsWrite  = "demo.products.write"
	PermissionProductsDelete = "demo.products.delete"
	MaximumPage              = 1_000_000
)

var (
	ErrDenied     = errors.New("demo authorization denied")
	ErrValidation = errors.New("demo request invalid")
	ErrNotFound   = errors.New("demo product not found")
	ErrConflict   = errors.New("demo product conflict")
	ErrInternal   = errors.New("demo operation failed")
	skuPattern    = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{2,31}$`)
)

type Scope string

const (
	ScopeSelf Scope = "self"
	ScopeAll  Scope = "all"
)

type Database interface {
	WithinTx(context.Context, func(context.Context, database.Tx) error) error
	Dialect() database.Dialect
}

// Authorizer is a consumer port. The final decision runs on the product transaction.
type Authorizer interface {
	Dialect() database.Dialect
	RequireInTx(context.Context, database.Tx, string, string) (Scope, error)
}

type Product struct {
	ID          string
	SKU         string
	Name        string
	Description string
	PriceCents  int64
	Status      string
	Revision    int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProductInput struct {
	SKU, Name, Description, Status string
	PriceCents                     int64
}

type ListQuery struct {
	Search, Sort, Direction string
	Page, PageSize          int
}

type Page struct {
	Rows  []Product
	Total int
}

type DeleteTarget struct {
	ID       string
	Revision int64
}

func normalizeInput(input ProductInput) (ProductInput, bool) {
	input.SKU = strings.ToUpper(strings.TrimSpace(input.SKU))
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	nameLength, descriptionLength := runeLength(input.Name), runeLength(input.Description)
	valid := skuPattern.MatchString(input.SKU) && nameLength >= 3 && nameLength <= 120 &&
		descriptionLength >= 0 && descriptionLength <= 500 && input.PriceCents >= 0 && input.PriceCents <= 100_000_000 &&
		(input.Status == "active" || input.Status == "inactive")
	return input, valid
}

func normalizeQuery(query ListQuery) (ListQuery, bool) {
	query.Search = strings.ToLower(strings.TrimSpace(query.Search))
	if query.Sort == "" {
		query.Sort = "updatedAt"
	}
	if query.Direction == "" {
		query.Direction = "descending"
	}
	validSort := query.Sort == "sku" || query.Sort == "name" || query.Sort == "priceCents" || query.Sort == "updatedAt"
	return query, runeLength(query.Search) >= 0 && runeLength(query.Search) <= 100 && query.Page >= 1 && query.Page <= MaximumPage &&
		query.PageSize >= 1 && query.PageSize <= 100 && validSort &&
		(query.Direction == "ascending" || query.Direction == "descending")
}

func runeLength(value string) int {
	if !utf8.ValidString(value) {
		return -1
	}
	return utf8.RuneCountInString(value)
}

// normalizedNameKey makes Unicode search and ordering independent of database case folding.
func normalizedNameKey(value string) string {
	return strings.ToLower(value)
}
