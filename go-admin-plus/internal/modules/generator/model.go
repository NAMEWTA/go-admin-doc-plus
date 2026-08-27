// Package generator imports authorized database metadata and emits isolated CRUD modules.
package generator

import (
	"errors"
	"time"
)

const (
	PermissionMetadataRead = "generator.metadata.read"
	PermissionPreview      = "generator.preview"
	PermissionWrite        = "generator.write"
)

var (
	ErrDenied         = errors.New("generator authorization denied")
	ErrAuthentication = errors.New("generator authentication required")
	ErrCSRF           = errors.New("generator csrf rejected")
	ErrInvalid        = errors.New("generator request invalid")
	ErrNotFound       = errors.New("generator metadata not found")
	ErrConflict       = errors.New("generator output conflict")
	ErrPreviewStale   = errors.New("generator preview is stale")
	ErrGateFailed     = errors.New("generator output gate failed")
	ErrInternal       = errors.New("generator operation failed")
)

type ColumnKind string

const (
	KindString  ColumnKind = "string"
	KindInt64   ColumnKind = "int64"
	KindBoolean ColumnKind = "boolean"
	KindDecimal ColumnKind = "decimal"
	KindTime    ColumnKind = "time"
	KindUUID    ColumnKind = "uuid"
	KindBytes   ColumnKind = "bytes"
)

type TableRef struct{ Schema, Name string }

type Column struct {
	Name         string
	DatabaseType string
	Kind         ColumnKind
	Nullable     bool
	PrimaryKey   bool
	Ordinal      int
}

type Table struct {
	Ref     TableRef
	Columns []Column
}

type ColumnDraft struct {
	Name       string
	Field      string
	Include    bool
	Searchable bool
	Sortable   bool
}

type Draft struct {
	Module, Entity, Plural string
	Table                  TableRef
	Columns                []ColumnDraft
}

type NormalizedColumn struct {
	Name, Field, DatabaseType, GoType, TypeScriptType, OpenAPIType, OpenAPIFormat string
	Kind                                                                          ColumnKind
	Nullable, PrimaryKey, Searchable, Sortable                                    bool
	Ordinal                                                                       int
}

type Model struct {
	Module, Entity, EntityVariable, Plural, TableName, Schema string
	Columns                                                   []NormalizedColumn
	PrimaryKey                                                NormalizedColumn
}

type PreviewFile struct {
	Path, Content, SHA256 string
}

type Preview struct {
	Token, Digest, Module string
	CreatedAt, ExpiresAt  time.Time
	Files                 []PreviewFile
}

type WriteResult struct {
	Token, Directory string
	Files            []string
}
